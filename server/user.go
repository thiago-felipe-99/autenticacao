package server

import (
	"log"

	ut "github.com/go-playground/universal-translator"
	"github.com/gofiber/fiber/v2"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type User struct {
	core       *core.User
	translator *ut.UniversalTranslator
	languages  []string
}

func (u *User) getTranslator(handler *fiber.Ctx) ut.Translator { //nolint:ireturn
	accept := handler.AcceptsLanguages(u.languages...)
	if accept == "" {
		accept = u.languages[0]
	}

	language, _ := u.translator.GetTranslator(accept)

	return language
}

// Get user by id
//
//	@Summary		Get user by id
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.User	"user return"
//	@Failure		401	{object}	sent		"user session has expired"
//	@Failure		404	{object}	sent		"user does not exist"
//	@Failure		500	{object}	sent		"internal server error"
//	@Param			id	path		string		true	"user id"
//	@Router			/user/{id} [get]
//	@Description	Get user by id.
func (u *User) GetByID(handler *fiber.Ctx) error {
	id, err := model.ParseID(handler.Params("id", "invalid-id"))
	if err != nil {
		return handler.Status(fiber.StatusNotFound).
			JSON(sent{errs.ErrUserNotFound.Error()})
	}

	funcCore := func() (model.User, error) { return u.core.GetByID(id) }

	expectErrors := []expectError{{errs.ErrUserNotFound, fiber.StatusNotFound}}

	unexpectMessageError := "error getting user"

	return callingCoreWithReturn(
		funcCore,
		expectErrors,
		unexpectMessageError,
		fiber.StatusOK,
		u.getTranslator(handler),
		handler,
	)
}

// Get users by roles
//
//	@Summary		Get users by roles
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		model.User	"user return"
//	@Failure		400		{object}	sent		"an invalid role param was sent"
//	@Failure		401		{object}	sent		"user session has expired"
//	@Failure		404		{object}	sent		"user does not exist"
//	@Failure		500		{object}	sent		"internal server error"
//	@Param			roles	query		[]string	true	"roles"
//	@Param			page	query		string		false	"result page number"
//	@Param			qt		query		string		false	"quantity user per page"
//	@Router			/user/roles [get]
//	@Description	Get users by roles.
func (u *User) GetByRole(handler *fiber.Ctx) error {
	query := &struct {
		Roles []string
		Page  int
		Qt    int
	}{}

	err := handler.QueryParser(query)
	if err != nil {
		return handler.Status(fiber.StatusBadRequest).JSON(sent{err.Error()})
	}

	funcCore := func() ([]model.User, error) { return u.core.GetByRole(query.Roles, query.Page, query.Qt) }

	expectErrors := []expectError{{errs.ErrUserNotFound, fiber.StatusNotFound}}

	unexpectMessageError := "error getting users"

	return callingCoreWithReturn(
		funcCore,
		expectErrors,
		unexpectMessageError,
		fiber.StatusOK,
		u.getTranslator(handler),
		handler,
	)
}

// Get all users
//
//	@Summary		Get users
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		model.User	"all roles"
//	@Failure		401		{object}	sent		"user session has expired"
//	@Failure		500		{object}	sent		"internal server error"
//	@Param			page	query		string		false	"result page number"
//	@Param			qt		query		string		false	"quantity user per page"
//	@Router			/user [get]
//	@Description	Get all user
func (u *User) GetAll(handler *fiber.Ctx) error {
	page, qt := handler.QueryInt("page"), handler.QueryInt("qt", defaultQtResults)

	funcCore := func() ([]model.User, error) { return u.core.GetAll(page, qt) }

	expectErrors := []expectError{}

	unexpectMessageError := "error getting users"

	return callingCoreWithReturn(
		funcCore,
		expectErrors,
		unexpectMessageError,
		fiber.StatusOK,
		u.getTranslator(handler),
		handler,
	)
}

// Create a user
//
//	@Summary		Create user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	sent				"create user successfully"
//	@Failure		400		{object}	sent				"an invalid user param was sent"
//	@Failure		401		{object}	sent				"user session has expired"
//	@Failure		403		{object}	sent				"current user is not admin"
//	@Failure		409		{object}	sent				"username/email already exist"
//	@Failure		500		{object}	sent				"internal server error"
//	@Param			user	body		model.UserPartial	true	"user params"
//	@Router			/user [post]
//	@Description	Create a user.
func (u *User) Create(handler *fiber.Ctx) error {
	userID, ok := handler.Locals("userID").(model.ID)
	if !ok {
		log.Printf("[ERROR] - error getting user ID")

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{"error refreshing session"})
	}

	body := &model.UserPartial{} //nolint:exhaustruct

	err := handler.BodyParser(body)
	if err != nil {
		return handler.Status(fiber.StatusBadRequest).JSON(sent{err.Error()})
	}

	funcCore := func() (any, error) {
		id, err := u.core.Create(userID, *body)
		response := struct {
			ID      model.ID `json:"id"`
			Message string   `json:"message"`
		}{
			ID:      id,
			Message: "user created",
		}

		return response, err
	}

	expectErrors := []expectError{
		{errs.ErrRoleNotFound, fiber.StatusBadRequest},
		{errs.ErrUsernameAlreadyExist, fiber.StatusConflict},
		{errs.ErrEmailAlreadyExist, fiber.StatusConflict},
	}

	unexpectMessageError := "error creating user"

	return callingCoreWithReturn(
		funcCore,
		expectErrors,
		unexpectMessageError,
		fiber.StatusCreated,
		u.getTranslator(handler),
		handler,
	)
}

// Update a user
//
//	@Summary		Update user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	sent				"update user successfully"
//	@Failure		400		{object}	sent				"an invalid user param was sent"
//	@Failure		401		{object}	sent				"user session has expired"
//	@Failure		403		{object}	sent				"current user is not admin"
//	@Failure		404		{object}	sent				"user does not exist"
//	@Failure		409		{object}	sent				"username/email already exist"
//	@Failure		500		{object}	sent				"internal server error"
//	@Param			user	body		model.UserUpdate	true	"user params"
//	@Param			id		path		string				true	"user id"
//	@Router			/user/{id} [put]
//	@Description	Update a user informations.
func (u *User) Update(handler *fiber.Ctx) error {
	id, err := model.ParseID(handler.Params("id", "invalid-id"))
	if err != nil {
		return handler.Status(fiber.StatusNotFound).
			JSON(sent{errs.ErrUserNotFound.Error()})
	}

	body := &model.UserUpdate{} //nolint:exhaustruct

	err = handler.BodyParser(body)
	if err != nil {
		return handler.Status(fiber.StatusBadRequest).JSON(sent{err.Error()})
	}

	funcCore := func() error { return u.core.Update(id, *body) }

	expectErrors := []expectError{
		{errs.ErrUserNotFound, fiber.StatusNotFound},
		{errs.ErrUsernameAlreadyExist, fiber.StatusConflict},
		{errs.ErrEmailAlreadyExist, fiber.StatusConflict},
		{errs.ErrRoleNotFound, fiber.StatusBadRequest},
	}

	unexpectMessageError := "error updating user"

	okay := okay{"user updated", fiber.StatusOK}

	return callingCore(
		funcCore,
		expectErrors,
		unexpectMessageError,
		okay,
		u.getTranslator(handler),
		handler,
	)
}

// Delete a user
//
//	@Summary		Delete user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sent	"user deleted"
//	@Failure		401	{object}	sent	"user session has expired"
//	@Failure		403	{object}	sent	"current user is not admin"
//	@Failure		404	{object}	sent	"user does not exist"
//	@Failure		500	{object}	sent	"internal server error"
//	@Param			id	path		string	true	"user id"
//	@Router			/user/{id} [delete]
//	@Description	Delete a user.
func (u *User) Delete(handler *fiber.Ctx) error {
	userID, ok := handler.Locals("userID").(model.ID)
	if !ok {
		log.Printf("[ERROR] - error getting user ID")

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{"error refreshing session"})
	}

	id, err := model.ParseID(handler.Params("id", "invalid-id"))
	if err != nil {
		return handler.Status(fiber.StatusNotFound).
			JSON(sent{errs.ErrUserNotFound.Error()})
	}

	funcCore := func() error { return u.core.Delete(userID, id) }

	expectErrors := []expectError{{errs.ErrUserNotFound, fiber.StatusNotFound}}

	unexpectMessageError := "error deleting user"

	okay := okay{"user deleted", fiber.StatusCreated}

	return callingCore(
		funcCore,
		expectErrors,
		unexpectMessageError,
		okay,
		u.getTranslator(handler),
		handler,
	)
}
