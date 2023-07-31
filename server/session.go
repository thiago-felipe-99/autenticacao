package server

import (
	"errors"
	"log"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/gofiber/fiber/v2"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type UserSession struct {
	core       *core.UserSession
	translator *ut.UniversalTranslator
	languages  []string
}

func (u *UserSession) getTranslator(handler *fiber.Ctx) ut.Translator { //nolint:ireturn
	accept := handler.AcceptsLanguages(u.languages...)
	if accept == "" {
		accept = u.languages[0]
	}

	language, _ := u.translator.GetTranslator(accept)

	return language
}

func setUserSession(handler *fiber.Ctx, userSession model.UserSession) {
	handler.Set("session", userSession.ID.String())
	handler.Set("session_expires", userSession.Expires.Format(time.RFC3339))
}

// Create a user session
//
//	@Summary		Create session
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	sent						"session created successfully"
//	@Failure		400		{object}	sent						"an invalid user param was sent"
//	@Failure		404		{object}	sent						"user does not exist"
//	@Failure		500		{object}	sent						"internal server error"
//	@Param			user	body		model.UserSessionPartial	true	"user params"
//	@Router			/session [post]
//	@Description	Create a user session and set in the response header.
func (u *UserSession) Create(handler *fiber.Ctx) error {
	body := &model.UserSessionPartial{} //nolint:exhaustruct

	err := handler.BodyParser(body)
	if err != nil {
		return handler.Status(fiber.StatusBadRequest).JSON(sent{err.Error()})
	}

	session := model.UserSession{} //nolint:exhaustruct

	funcCore := func() error {
		sessionTemp, err := u.core.Create(*body)
		session = sessionTemp

		return err
	}

	expectErrors := []expectError{
		{errs.ErrUserNotFound, fiber.StatusNotFound},
		{errs.ErrPasswordDoesNotMatch, fiber.StatusBadRequest},
	}

	unexpectMessageError := "error creating user session"

	okay := okay{"user session created", fiber.StatusCreated}

	err = callingCore(
		funcCore,
		expectErrors,
		unexpectMessageError,
		okay,
		u.getTranslator(handler),
		handler,
	)

	deleteSession, ok := handler.Locals("deleteSession").(bool)
	if !(ok && deleteSession) {
		setUserSession(handler, session)
	}

	return err
}

// Refresh a user session
//
//	@Summary		Refresh session
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sent	"user session refreshed successfully"
//	@Failure		401	{object}	sent	"user session has expired"
//	@Failure		500	{object}	sent	"internal server error"
//	@Router			/session [put]
//	@Description	Refresh a user session and set in the response header.
func (u *UserSession) Refresh(handler *fiber.Ctx) error {
	sessionIDRaw := handler.Get("session", "invalid_session")

	sessionID, err := model.ParseID(sessionIDRaw)
	if err != nil {
		return handler.Status(fiber.StatusUnauthorized).
			JSON(sent{errs.ErrUserNotFound.Error()})
	}

	session, err := u.core.Refresh(sessionID)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFound) {
			return handler.Status(fiber.StatusUnauthorized).
				JSON(sent{errs.ErrUserSessionNotFound.Error()})
		}

		log.Printf("[ERROR] - error refreshing session: %s", err)

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{"error refreshing session"})
	}

	setUserSession(handler, session)

	handler.Locals("userID", session.UserID)

	errNext := handler.Next()

	return errNext
}
