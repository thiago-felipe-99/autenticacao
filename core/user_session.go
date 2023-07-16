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

func (u *UserSession) GetByID(id model.ID) (*model.UserSession, error) {
	userSession, err := u.database.GetByID(id)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFoud) {
			return nil, errs.ErrUserSessionNotFoud
		}

		return nil, fmt.Errorf("error getting user session from database: %w", err)
	}

	return userSession, nil
}

func (u *UserSession) GetAll(paginate int, qt int) ([]model.UserSession, error) {
	userSessions, err := u.database.GetAll(paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFoud) {
			return nil, errs.ErrUserSessionNotFoud
		}

		return nil, fmt.Errorf("error getting users sessions from database: %w", err)
	}

	return userSessions, nil
}

func (u *UserSession) GetByUserID(
	userID model.ID,
	paginate int,
	qt int,
) ([]model.UserSession, error) {
	userSessions, err := u.database.GetByUserID(userID, paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFoud) {
			return nil, errs.ErrUserSessionNotFoud
		}

		return nil, fmt.Errorf("error getting users sessions from database: %w", err)
	}

	return userSessions, nil
}

func (u *UserSession) Create(partial model.UserSessionPartial) (*model.UserSession, error) {
	err := validate(u.validator, partial)
	if err != nil {
		return nil, err
	}

	var user *model.User

	if partial.Username != "" {
		user, err = u.user.GetByUsername(partial.Username)
	} else {
		user, err = u.user.GetByEmail(partial.Email)
	}

	if err != nil {
		return nil, err
	}

	equal, _, err := argon2id.CheckHash(partial.Password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("erro checking password: %w", err)
	}

	if !equal {
		return nil, errs.ErrPasswordDoesNotMatch
	}

	userSession := model.UserSession{
		ID:        model.NewID(),
		UserID:    user.ID,
		CreateaAt: time.Now(),
		Expires:   time.Now().Add(u.expires),
		DeletedAt: time.Time{},
	}

	err = u.database.Create(userSession)
	if err != nil {
		return nil, fmt.Errorf("error creating user session on database: %w", err)
	}

	return &userSession, nil
}

func (u *UserSession) Refresh(id model.ID) (*model.UserSession, error) {
	current, err := u.Delete(id)
	if err != nil {
		return nil, err
	}

	userSession := model.UserSession{
		ID:        model.NewID(),
		UserID:    current.UserID,
		CreateaAt: time.Now(),
		Expires:   time.Now().Add(u.expires),
		DeletedAt: time.Time{},
	}

	err = u.database.Create(userSession)
	if err != nil {
		return nil, fmt.Errorf("error creating user session on database: %w", err)
	}

	return &userSession, nil
}

func (u *UserSession) Delete(id model.ID) (*model.UserSession, error) {
	userSession, err := u.GetByID(id)
	if err != nil {
		return nil, err
	}

	err = u.database.Delete(id, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error deleting user session from database: %w", err)
	}

	return userSession, nil
}
