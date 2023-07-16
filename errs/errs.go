package errs

import "errors"

var (
	ErrBodyValidate         = errors.New("unable to parse body")
	ErrUserNotFoud          = errors.New("user not found")
	ErrUsernameAlreadyExist = errors.New("username already exist")
	ErrEmailAlreadyExist    = errors.New("emails already exist")
	ErrRoleNotFoud          = errors.New("role not found")
	ErrRoleAlreadyExist     = errors.New("role already exist")
	ErrInvalidRoles         = errors.New("invalid roles")
	ErrUserSessionNotFoud   = errors.New("user session not found")
	ErrPasswordDoesNotMatch = errors.New("password does not match")
)
