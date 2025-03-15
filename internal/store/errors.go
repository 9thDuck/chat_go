package store

import (
	"errors"

	"github.com/lib/pq"
)

const (
	// generic
	DefaultNotFoundErrMsg                     = "record not found"
	DefaultConflictErrMsg                     = "resource already exists"
	DefaultSomethingWentWrongErrMsg           = "something went wrong, try again later"
	DefaultUnauthorizedErrorMsg               = "unauthorized"
	DefaultAuthorizationHeaderMissingErrMsg   = "authorization header missing"
	DefaultAuthorizationHeaderMalformedErrMsg = "malformed authorization error"
	DefaultInvalidCredentialsErrMsg           = "invalid credentials"
	DefaultBasicAuthInvalidCredentialsErrMsg  = "invalid basic auth credentials"

	// users
	DefaultDuplicateMailErrMsg     = "user with given email already exist"
	DefaultDuplicateUsernameErrMsg = "user with given username already exists"

	// contact requests
	DefaultContactRequestAlreadyExistsErrMsg = "either contact request or contact already exists or the most recent request was rejected. If the most recent request was rejected, you can ask the user you want to add to send you a contact request"
	DefaultContactRequestNotFoundErrMsg      = "contact request not found"

	// contacts
	DefaultContactAlreadyExistsErrMsg = "contact already exists"
	DefaultContactNotFoundErrMsg      = "contact not found"
)

var (
	ErrNotFound          = errors.New(DefaultNotFoundErrMsg)
	ErrConflict          = errors.New(DefaultConflictErrMsg)
	ErrSomethingWenWrong = errors.New(DefaultSomethingWentWrongErrMsg)

	// auth
	ErrUnautorized                  = errors.New(DefaultUnauthorizedErrorMsg)
	ErrAuthorizationHeaderMissing   = errors.New(DefaultAuthorizationHeaderMissingErrMsg)
	ErrAuthorizationHeaderMalformed = errors.New(DefaultAuthorizationHeaderMalformedErrMsg)
	ErrInvalidCredentials           = errors.New(DefaultInvalidCredentialsErrMsg)
	ErrBasicAuthInvalidCredentials  = errors.New(DefaultBasicAuthInvalidCredentialsErrMsg)

	// users
	ErrDuplicateMail     = errors.New(DefaultDuplicateMailErrMsg)
	ErrDuplicateUsername = errors.New(DefaultDuplicateUsernameErrMsg)

	// contact requests
	ErrContactRequestAlreadyExists = errors.New(DefaultContactRequestAlreadyExistsErrMsg)
	ErrContactRequestNotFound      = errors.New(DefaultContactRequestNotFoundErrMsg)

	// contacts
	ErrContactAlreadyExists = errors.New(DefaultContactAlreadyExistsErrMsg)
	ErrContactNotFound      = errors.New(DefaultContactNotFoundErrMsg)
)

const (
	PQ_CODE_UNIQUE_CONSTRAINT_VIOLATION      pq.ErrorCode = "23505"
	PQ_CODE_FOREIGN_KEY_CONSTRAINT_VIOLATION pq.ErrorCode = "23503"
)
