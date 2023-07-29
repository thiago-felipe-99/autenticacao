package server

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/gofiber/fiber/v2"
	"github.com/thiago-felipe-99/autenticacao/core"
)

type Role struct {
	core       *core.Role
	translator *ut.UniversalTranslator
	languages  []string
}

func (r *Role) hello(handler *fiber.Ctx) error {
	return handler.SendString("hello")
}
