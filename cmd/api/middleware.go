package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/9thDuck/chat_go.git/internal/domain"
	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) ValidateTokenMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessTokenCookie, err := r.Cookie("access_token")
			if err != nil {
				app.unauthorizedError(w, r, err)
				return
			}

			refreshTokenCookie, err := r.Cookie("refresh_token")
			if err != nil {
				app.unauthorizedError(w, r, err)
				return
			}
			if accessTokenCookie.Value == "" || refreshTokenCookie.Value == "" {
				app.unauthorizedError(w, r, nil)
				return
			}
			ctx := r.Context()

			accessToken, err := app.authenticator.ValidateTokenAndParse(accessTokenCookie.Value)
			if err != nil {
				if !errors.Is(err, jwt.ErrTokenExpired) {
					app.badRequestError(w, r, err, "")
					return
				}
				// access token expired, let's validate refreshToken
				refreshToken, err := app.authenticator.ValidateTokenAndParse(refreshTokenCookie.Value)
				if err != nil {
					app.unauthorizedError(w, r, err)
					return
				}
				// refresh token is not expired, let's make new cookies
				refreshTokenUserID, err := getUserIDFromToken(refreshToken)
				if err != nil {
					app.unauthorizedError(w, r, err)
					return
				}

				user, err := app.getUser(ctx, refreshTokenUserID)
				if err != nil {
					if errors.Is(err, store.ErrNotFound) {
						app.deleteCookie(w, "access_token")
						app.deleteCookie(w, "refresh_token")
						app.unauthorizedError(w, r, store.ErrUnautorized)
						return
					}
					app.internalError(w, r, err)
					return
				}
				ctx := context.WithValue(ctx, userCtxKey, user)
				accessTokenCookie, refreshTokenCookie, err := app.makeAuthCookiesSet(refreshTokenUserID)
				if err != nil {
					app.internalError(w, r, err)
					return
				}
				http.SetCookie(w, accessTokenCookie)
				http.SetCookie(w, refreshTokenCookie)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			accessTokenUserID, err := getUserIDFromToken(accessToken)
			if err != nil {
				app.unauthorizedError(w, r, err)
				return
			}

			user, err := app.getUser(ctx, accessTokenUserID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					app.deleteCookie(w, "access_token")
					app.deleteCookie(w, "refresh_token")
					app.unauthorizedError(w, r, store.ErrUnautorized)
					return
				}
				app.internalError(w, r, err)
				return
			}

			ctx = context.WithValue(ctx, userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (app *application) userDetailsUpdateGuardMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDFromParam := getUserIDParamFromCtx(r)
		user := getUserFromCtx(r)

		if user.ID != userIDFromParam {
			app.unauthorizedError(w, r, store.ErrUnautorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// func (app *application) deleteUserAuthorityCheckMiddleware(requiredRoleName string, next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(
// 		func(w http.ResponseWriter, r *http.Request) {
// 			user := getUserFromCtx(r)

// 			idParam := chi.URLParam(r, string(userIDCtxKey))
// 			if idParam == "" {
// 				app.BadRequestError(w, r, errors.New("id of user to be deleted is not specified"))
// 				return
// 			}

// 			userIDFromParam, err := strconv.ParseInt(idParam, 10, 64)
// 			if err != nil {
// 				app.BadRequestError(w, r, err)
// 				return
// 			}
// 			if userIDFromParam == user.ID {
// 				ctx := context.WithValue(r.Context(), deleteTokenCookiesCtxKey, true)
// 				next.ServeHTTP(w, r.WithContext(ctx))
// 				return
// 			}
// 			allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRoleName)
// 			if err != nil {
// 				app.internalError(w, r, err)
// 				return
// 			}
// 			if !allowed {
// 				app.forbiddenRequestError(w, r, fmt.Errorf("forbidden action by userID: %d", user.ID))
// 				return
// 			}

// 			next.ServeHTTP(w, r)
// 		})
// }

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	user := &store.User{ID: userID}
	user.Role = &domain.Role{}
	if app.config.cacheCfg.initialised {
		user, err := app.cache.Users.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			app.logger.Infow("Cache:Users hit", "userID", user.ID)
			return user, nil
		}
		app.logger.Infow("Cache:Users miss", "userID", userID)
	}
	if err := app.store.Users.GetByID(ctx, user); err != nil {
		return nil, err
	}

	if app.config.cacheCfg.initialised {
		err := app.cache.Users.Set(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *application) encryptionIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		encryptionKeyID, err := r.Cookie(fmt.Sprintf("%s_%d", encryptionKeyIDCtxKey, user.ID))
		if err != nil || encryptionKeyID.Value == "" {
			app.badRequestError(w, r, errors.New("encryption key ID is required"), "")
			return
		}
		ctx := context.WithValue(r.Context(), encryptionKeyIDCtxKey, encryptionKeyID.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getUserWithEncryptionKey(ctx context.Context, userID int64, encryptionKeyID string) (*store.UserWithEncryptionKey, error) {
	user, err := app.getUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if app.config.cacheCfg.initialised {
		encryptionKey, err := app.cache.EncryptionKeys.Get(ctx, userID, encryptionKeyID)
		if err != nil {
			return nil, err
		}
		if encryptionKey != nil {
			return store.NewUserWithEncryptionKey(user, encryptionKey), nil
		}
	}

	encryptionKey, err := app.store.EncryptionKeys.Get(ctx, userID, encryptionKeyID)

	if err != nil {
		return nil, err
	}

	if app.config.cacheCfg.initialised {
		err = app.cache.EncryptionKeys.Set(ctx, userID, encryptionKey)
		if err != nil {
			return nil, err
		}
	}

	return store.NewUserWithEncryptionKey(user, encryptionKey), nil
}

// func (app *application) roleChangingAuthorityMiddleware(requiredRoleName string, next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		user := getUserFromCtx(r)

// 		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRoleName)
// 		if err != nil {
// 			app.internalError(w, r, err)
// 			return
// 		}
// 		if !allowed {
// 			app.forbiddenRequestError(w, r, fmt.Errorf("forbidden action by userID: %d", user.ID))
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, requiredRoleName string) (bool, error) {
// 	requiredRole, err := app.store.Roles.GetByName(ctx, requiredRoleName)
// 	if err != nil {
// 		return false, err
// 	}

// 	return requiredRole.Level <= user.Role.Level, nil
// }

func getUserIDFromToken(token *jwt.Token) (int64, error) {
	claims, _ := token.Claims.(jwt.MapClaims)
	userID, err :=
		strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

	if err != nil {
		return int64(0), err
	}
	return userID, nil
}

func (app *application) getUserIDParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, string(userIDCtxKey))

		id, err := strconv.ParseInt(idParam, 10, 64)

		if err != nil {
			app.badRequestError(w, r, err, "")
			return
		}
		ctxWithValue := context.WithValue(r.Context(), userIDCtxKey, id)

		next.ServeHTTP(w, r.WithContext(ctxWithValue))
	})
}

func (app *application) paginationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := 10
		page := 1
		sort := "created_at"
		sortDirection := "DESC"

		query := r.URL.Query()

		limitQParam := query.Get("limit")
		if limitQParam != "" {
			val, err := strconv.Atoi(limitQParam)
			if err != nil {
				app.badRequestError(w, r, nil, "limit query param must be number greater than 0 and less than equal to 10")
				return
			} else if limit < 1 || limit > 10 {
				app.badRequestError(w, r, nil, "limit query param must be number greater than 0 and less than equal to 10")
				return
			}
			limit = val
		}

		pageQParam := query.Get("page")
		if pageQParam != "" {
			val, err := strconv.Atoi(pageQParam)
			if err != nil {
				app.badRequestError(w, r, nil, "page query param must be a number greater than 0")
				return
			}
			page = val
		}

		sortQParam := query.Get("sort")
		if sortQParam != "" {
			if sortQParam != "first_name" && sortQParam != "last_name" && sortQParam != "username" && sortQParam != "created_at" {
				app.badRequestError(w, r, nil, "sort query param must be one of the following: first_name, last_name, username, created_at")
				return
			}
			sort = sortQParam
		}

		sortDirectionQParam := query.Get("sort_direction")
		if sortDirectionQParam != "" {
			if sortDirectionQParam != "ASC" && sortDirectionQParam != "DESC" {
				app.badRequestError(w, r, nil, "sort_direction query param must be one of the following: ASC, DESC")
				return
			}
			sortDirection = sortDirectionQParam
		}

		pagination := &store.Pagination{
			Limit:         limit,
			Page:          page,
			Sort:          sort,
			SortDirection: sortDirection,
		}

		ctx := context.WithValue(r.Context(), paginationCtxKey, pagination)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getContactIDParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contactID := chi.URLParam(r, "contactID")
		contactIDInt, err := strconv.ParseInt(contactID, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err, "")
			return
		}
		ctx := context.WithValue(r.Context(), contactIDCtxKey, contactIDInt)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) blockSelfContactRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		contactID := getContactIDFromCtx(r)

		if user.ID == contactID {
			app.badRequestError(w, r, nil, "invalid contact_id parameter, contact_id cannot be your own id")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) getReceiverIDParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receiverID := chi.URLParam(r, "receiverID")
		if receiverID == "" {
			app.badRequestError(w, r, nil, "receiver_id is required to send message")
			return
		}
		receiverIDInt, err := strconv.ParseInt(receiverID, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err, "")
			return
		}
		ctx := context.WithValue(r.Context(), receiverIDCtxKey, receiverIDInt)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) preMessageCreationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		const payloadValidationErrMsg = "content must be between 1 and 1000 characters, attachments must be an array of strings not more than 10 elements, each string must be less than 255 characters"

		var payload createMessagePayload
		if err := readJson(w, r, &payload); err != nil {
			app.badRequestError(w, r, err, payloadValidationErrMsg)
			return
		}
		if err := Validate.Struct(&payload); err != nil {
			app.badRequestError(w, r, err, payloadValidationErrMsg)
			return
		}
		user := getUserFromCtx(r)
		receiverID := getReceiverIDFromCtx(r)
		areContacts, err := app.checkContactRelationship(r.Context(), user.ID, receiverID)
		if err != nil {
			app.internalError(w, r, err)
			return
		}

		if !areContacts {
			app.badRequestError(w, r, nil, "You can only send messages to users in your contacts list")
			return
		}

		ctx := context.WithValue(r.Context(), messageCreationPayloadCtxKey, &payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkContactRelationship(ctx context.Context, userID, contactID int64) (bool, error) {
	if app.config.cacheCfg.initialised {
		areContacts, err := app.cache.Contacts.GetContactExists(ctx, userID, contactID)
		if err != nil {
			return false, err
		}
		if areContacts {
			app.logger.Infow("Cache:Contacts hit", "userID", userID, "contactID", contactID)
			return true, nil
		}
		app.logger.Infow("Cache:Contacts miss", "userID", userID, "contactID", contactID)
	}

	areContacts, err := app.store.Contacts.GetContactExists(ctx, userID, contactID)
	if err != nil {
		return false, err
	}
	if app.config.cacheCfg.initialised {
		err = app.cache.Contacts.SetContactExists(ctx, userID, contactID, areContacts)
		if err != nil {
			app.logger.Errorw("Failed to update contacts cache", "error", err)
		}
	}

	return areContacts, nil
}
