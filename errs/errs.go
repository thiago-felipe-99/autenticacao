package errs

import "errors"

var (
	ErrInvalidID            = errors.New("ID is not a string")
	ErrBodyValidate         = errors.New("unable to parse body")
	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameAlreadyExist = errors.New("username already exist")
	ErrEmailAlreadyExist    = errors.New("emails already exist")
	ErrRoleNotFound         = errors.New("role not found")
	ErrRoleAlreadyExist     = errors.New("role already exist")
	ErrUserSessionNotFoud   = errors.New("user session not found")
	ErrPasswordDoesNotMatch = errors.New("password does not match")
)
