package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	user.Role = &store.Role{}
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
		fmt.Println(idParam, len(idParam))

		id, err := strconv.ParseInt(idParam, 10, 64)

		if err != nil {
			app.badRequestError(w, r, err, "")
			return
		}
		ctxWithValue := context.WithValue(r.Context(), userIDCtxKey, id)

		next.ServeHTTP(w, r.WithContext(ctxWithValue))
	})
}
