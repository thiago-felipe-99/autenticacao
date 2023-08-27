package core

import (
	"errors"
	"fmt"
	"time"

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

func (u *UserSession) GetAllActive(paginate int, qt int) ([]model.UserSession, error) {
	userSessions, err := u.database.GetAllActive(paginate, qt)
	if err != nil {
		return model.EmptyUserSessions, fmt.Errorf(
			"error getting users sessions from database: %w",
			err,
		)
	}

	return userSessions, nil
}

func (u *UserSession) GetByUserIDActive(
	userID model.ID,
	paginate int,
	qt int,
) ([]model.UserSession, error) {
	userSessions, err := u.database.GetByUserIDActive(userID, paginate, qt)
	if err != nil {
		return model.EmptyUserSessions, fmt.Errorf(
			"error getting users sessions from database: %w",
			err,
		)
	}

	return userSessions, nil
}

func (u *UserSession) GetAllInactive(paginate int, qt int) ([]model.UserSession, error) {
	userSessions, err := u.database.GetAllInactive(paginate, qt)
	if err != nil {
		return model.EmptyUserSessions, fmt.Errorf(
			"error getting users sessions from database: %w",
			err,
		)
	}

	return userSessions, nil
}

func (u *UserSession) GetByUserIDInactive(
	userID model.ID,
	paginate int,
	qt int,
) ([]model.UserSession, error) {
	userSessions, err := u.database.GetByUserIDInactive(userID, paginate, qt)
	if err != nil {
		return model.EmptyUserSessions, fmt.Errorf(
			"error getting users sessions from database: %w",
			err,
		)
	}

	return userSessions, nil
}

func (u *UserSession) GetByID(id model.ID) (model.UserSession, error) {
	userSession, err := u.database.GetByID(id)
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFound) {
			return model.EmptyUserSession, errs.ErrUserSessionNotFound
		}

		return model.EmptyUserSession, fmt.Errorf(
			"error getting user session from database: %w",
			err,
		)
	}

	return userSession, nil
}

func (u *UserSession) Create(partial model.UserSessionPartial) (model.UserSession, error) {
	err := Validate(u.validator, partial)
	if err != nil {
		return model.EmptyUserSession, err
	}

	var user model.User

	if partial.Username != "" {
		user, err = u.user.GetByUsername(partial.Username)
	} else {
		user, err = u.user.GetByEmail(partial.Email)
	}

	if err != nil {
		return model.EmptyUserSession, err
	}

	equal, err := u.user.EqualPassword(partial.Password, user.Password)
	if err != nil {
		return model.EmptyUserSession, fmt.Errorf("erro checking password: %w", err)
	}

	if !equal {
		return model.EmptyUserSession, errs.ErrPasswordDoesNotMatch
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
		return model.EmptyUserSession, fmt.Errorf(
			"error creating user session on database: %w",
			err,
		)
	}

	return userSession, nil
}

func (u *UserSession) Delete(id model.ID) (model.UserSession, error) {
	userSession, err := u.database.Delete(id, time.Now())
	if err != nil {
		if errors.Is(err, errs.ErrUserSessionNotFound) {
			return model.EmptyUserSession, errs.ErrUserSessionNotFound
		}

		return model.EmptyUserSession, fmt.Errorf(
			"error deleting user session from database: %w",
			err,
		)
	}

	return userSession, nil
}

func (u *UserSession) Refresh(id model.ID) (model.UserSession, error) {
	userSession, err := u.Delete(id)
	if err != nil {
		return model.EmptyUserSession, err
	}

	userSession = model.UserSession{
		ID:        model.NewID(),
		UserID:    userSession.UserID,
		CreateaAt: time.Now(),
		Expires:   time.Now().Add(u.expires),
		DeletedAt: time.Time{},
	}

	err = u.database.Create(userSession)
	if err != nil {
		return model.EmptyUserSession, fmt.Errorf(
			"error creating user session on database: %w",
			err,
		)
	}

	return userSession, nil
}

func NewUserSession(
	db data.UserSession,
	user *User,
	validate *validator.Validate,
	expires time.Duration,
) *UserSession {
	return &UserSession{
		database:  db,
		user:      user,
		validator: validate,
		expires:   expires,
	}
}
