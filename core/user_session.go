package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-playground/validator/v10"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type UserSession struct {
	database  data.UserSession
	user      *User
	validator *validator.Validate
	expires   time.Duration
}

func (core *UserSession) GetByID(id model.ID) (*model.UserSession, error) {
	userSession, err := core.database.GetByID(id)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error getting user session from database: %w", err)
	}

	return userSession, nil
}

func (core *UserSession) GetAll(paginate int, qt int) ([]model.UserSession, error) {
	userSessions, err := core.database.GetAll(paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error getting users sessions from database: %w", err)
	}

	return userSessions, nil
}

func (core *UserSession) GetByUserID(
	userID model.ID,
	paginate int,
	qt int,
) ([]model.UserSession, error) {
	userSessions, err := core.database.GetByUserID(userID, paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error getting users sessions from database: %w", err)
	}

	return userSessions, nil
}

func (core *UserSession) Create(partial model.UserSessionPartial) (*model.UserSession, error) {
	err := validate(core.validator, partial)
	if err != nil {
		return nil, err
	}

	var user *model.User

	if partial.Username != "" {
		user, err = core.user.GetByUsername(partial.Username)
	} else {
		user, err = core.user.GetByEmail(partial.Email)
	}

	if err != nil {
		return nil, err
	}

	equal, _, err := argon2id.CheckHash(partial.Password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("Erro checking password: %w", err)
	}

	if !equal {
		return nil, errs.ErrPasswordDoesNotMatch
	}

	userSession := model.UserSession{
		ID:        model.NewID(),
		UserID:    user.ID,
		CreateaAt: time.Now(),
		Expires:   time.Now().Add(core.expires),
		DeletedAt: time.Time{},
	}

	err = core.database.Create(userSession)
	if err != nil {
		return nil, fmt.Errorf("Error creating user session on database: %w", err)
	}

	return &userSession, nil
}

func (core *UserSession) Refresh(id model.ID) (*model.UserSession, error) {
	current, err := core.Delete(id)
	if err != nil {
		return nil, err
	}

	userSession := model.UserSession{
		ID:        model.NewID(),
		UserID:    current.UserID,
		CreateaAt: time.Now(),
		Expires:   time.Now().Add(core.expires),
		DeletedAt: time.Time{},
	}

	err = core.database.Create(userSession)
	if err != nil {
		return nil, fmt.Errorf("Error creating user session on database: %w", err)
	}

	return &userSession, nil
}

func (core *UserSession) Delete(id model.ID) (*model.UserSession, error) {
	userSession, err := core.GetByID(id)
	if err != nil {
		return nil, err
	}

	err = core.database.Delete(id, time.Now())
	if err != nil {
		return nil, fmt.Errorf("Error deleting user session from database")
	}

	return userSession, nil
}
