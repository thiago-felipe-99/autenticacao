package server

import (
	"log"

	ut "github.com/go-playground/universal-translator"
	"github.com/gofiber/fiber/v2"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type Role struct {
	core       *core.Role
	translator *ut.UniversalTranslator
	languages  []string
}

func (r *Role) getTranslator(handler *fiber.Ctx) ut.Translator { //nolint:ireturn
	accept := handler.AcceptsLanguages(r.languages...)
	if accept == "" {
		accept = r.languages[0]
	}

	language, _ := r.translator.GetTranslator(accept)

	return language
}

// Get role by name
//
//	@Summary		Get role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	model.Role	"role return"
//	@Failure		401		{object}	sent		"user session has expired"
//	@Failure		404		{object}	sent		"role does not exist"
//	@Failure		500		{object}	sent		"internal server error"
//	@Param			name	path		string		true	"role name"
//	@Router			/role/{name} [get]
//	@Description	Get role by name.
//	@Security		BasicAuth
func (r *Role) GetByName(handler *fiber.Ctx) error {
	funcCore := func() (model.Role, error) { return r.core.GetByName(handler.Params("name")) }

	expectErrors := []expectError{{errs.ErrRoleNotFound, fiber.StatusNotFound}}

	unexpectMessageError := "error getting role"

	return callingCoreWithReturn(
		funcCore,
		expectErrors,
		unexpectMessageError,
		fiber.StatusOK,
		r.getTranslator(handler),
		handler,
	)
}

// Get all roles
//
//	@Summary		Get roles
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		model.Role	"all roles"
//	@Failure		401		{object}	sent		"user session has expired"
//	@Failure		500		{object}	sent		"internal server error"
//	@Param			page	query		string		false	"result page number"
//	@Param			qt		query		string		false	"quantity roles per page"
//	@Router			/role [get]
//	@Description	Get all roles
//	@Security		BasicAuth
func (r *Role) GetAll(handler *fiber.Ctx) error {
	page, qt := handler.QueryInt("page"), handler.QueryInt("qt", defaultQtResults)

	funcCore := func() ([]model.Role, error) { return r.core.GetAll(page, qt) }

	expectErrors := []expectError{}

	unexpectMessageError := "error getting roles"

	return callingCoreWithReturn(
		funcCore,
		expectErrors,
		unexpectMessageError,
		fiber.StatusOK,
		r.getTranslator(handler),
		handler,
	)
}

// Create a role
//
//	@Summary		Create role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	sent				"create role successfully"
//	@Failure		400		{object}	sent				"an invalid role param was sent"
//	@Failure		401		{object}	sent				"user session has expired"
//	@Failure		403		{object}	sent				"current user is not admin"
//	@Failure		409		{object}	sent				"role already exist"
//	@Failure		500		{object}	sent				"internal server error"
//	@Param			role	body		model.RolePartial	true	"role params"
//	@Router			/role [post]
//	@Description	Create a role.
//	@Security		BasicAuth
func (r *Role) Create(handler *fiber.Ctx) error {
	userID, ok := handler.Locals("userID").(model.ID)
	if !ok {
		log.Printf("[ERROR] - error getting user ID")

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{"error refreshing session"})
	}

	body := &model.RolePartial{} //nolint:exhaustruct

	err := handler.BodyParser(body)
	if err != nil {
		return handler.Status(fiber.StatusBadRequest).JSON(sent{err.Error()})
	}

	funcCore := func() error { return r.core.Create(userID, *body) }

	expectErrors := []expectError{{errs.ErrRoleAlreadyExist, fiber.StatusConflict}}

	unexpectMessageError := "error creating role"

	okay := okay{"role created", fiber.StatusCreated}

	return callingCore(
		funcCore,
		expectErrors,
		unexpectMessageError,
		okay,
		r.getTranslator(handler),
		handler,
	)
}

// Delete a role
//
//	@Summary		Delete role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	sent	"role deleted"
//	@Failure		401		{object}	sent	"user session has expired"
//	@Failure		403		{object}	sent	"current user is not admin"
//	@Failure		404		{object}	sent	"role does not exist"
//	@Failure		500		{object}	sent	"internal server error"
//	@Param			name	path		string	true	"role name"
//	@Router			/role/{name} [delete]
//	@Description	Delete a role.
//	@Security		BasicAuth
func (r *Role) Delete(handler *fiber.Ctx) error {
	userID, ok := handler.Locals("userID").(model.ID)
	if !ok {
		log.Printf("[ERROR] - error getting user ID")

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{"error refreshing session"})
	}

	funcCore := func() error { return r.core.Delete(userID, handler.Params("name")) }

	expectErrors := []expectError{{errs.ErrRoleNotFound, fiber.StatusNotFound}}

	unexpectMessageError := "error deleting role"

	okay := okay{"role deleted", fiber.StatusCreated}

	return callingCore(
		funcCore,
		expectErrors,
		unexpectMessageError,
		okay,
		r.getTranslator(handler),
		handler,
	)
}
