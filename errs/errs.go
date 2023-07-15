package errs

import "errors"

var (
	ErrBodyValidate         = errors.New("unable to parse body")
	ErrUserNotFoud          = errors.New("User Not Found")
	ErrUsernameAlreadyExist = errors.New("Username already exist")
	ErrEmailAlreadyExist    = errors.New("Emails already exist")
)
