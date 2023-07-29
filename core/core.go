package core

import (
	"errors"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
)

type InvalidError struct {
	invalid validator.ValidationErrors
}

func (m InvalidError) Error() string {
	return m.invalid.Error()
}

func (m InvalidError) Translate(language ut.Translator) string {
	messages := m.invalid.Translate(language)

	messageSend := ""
	for _, message := range messages {
		messageSend += ", " + message
	}

	return messageSend[2:]
}

func NewInvalidError(errs validator.ValidationErrors) InvalidError {
	return InvalidError{errs}
}

func Validate(validate *validator.Validate, data any) error {
	err := validate.Struct(data)
	if err != nil {
		validationErrs := validator.ValidationErrors{}

		okay := errors.As(err, &validationErrs)
		if !okay {
			return errs.ErrBodyValidate
		}

		return NewInvalidError(validationErrs)
	}

	return nil
}

type Cores struct {
	*Role
	*User
	*UserSession
}

func NewCore(
	data *data.Data,
	validate *validator.Validate,
	argonEnable bool,
	expires time.Duration,
) *Cores {
	role := NewRole(data.Role, validate)
	user := NewUser(data.User, role, validate, argonEnable)
	userSession := NewUserSession(data.UserSession, user, validate, expires)

	return &Cores{
		Role:        role,
		User:        user,
		UserSession: userSession,
	}
}
