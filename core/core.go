package core

import (
	"errors"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/thiago-felipe-99/autenticacao/errs"
)

type ModelInvalidError struct {
	invalid validator.ValidationErrors
}

func (m ModelInvalidError) Error() string {
	return m.invalid.Error()
}

func (m ModelInvalidError) Translate(language ut.Translator) string {
	messages := m.invalid.Translate(language)

	messageSend := ""
	for _, message := range messages {
		messageSend += ", " + message
	}

	return messageSend[2:]
}

func validate(validate *validator.Validate, data any) error {
	err := validate.Struct(data)
	if err != nil {
		validationErrs := validator.ValidationErrors{}

		okay := errors.As(err, &validationErrs)
		if !okay {
			return errs.ErrBodyValidate
		}

		return ModelInvalidError{validationErrs}
	}

	return nil
}
