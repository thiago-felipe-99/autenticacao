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

type User struct {
	database  data.User
	role      *Role
	validator *validator.Validate
	argon2id  argon2id.Params
}

func (core *User) validateRoles(roles []string) (bool, error) {
	for _, role := range roles {
		_, err := core.role.GetByName(role)
		if err != nil {
			if errors.Is(err, errs.ErrRoleNotFoud) {
				return false, nil
			}

			return false, fmt.Errorf("Error validating role: %w", err)
		}
	}

	return true, nil
}

func (core *User) GetByID(id model.ID) (*model.User, error) {
	user, err := core.database.GetByID(id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error on getting user from database: %w", err)
	}

	return user, nil
}

func (core *User) GetByUsername(username string) (*model.User, error) {
	user, err := core.database.GetByUsername(username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error on getting user from database: %w", err)
	}

	return user, nil
}

func (core *User) GetByEmail(email string) (*model.User, error) {
	user, err := core.database.GetByEmail(email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error on getting user from database: %w", err)
	}

	return user, nil
}

func (core *User) GetAll(paginate int, qt int) ([]model.User, error) {
	users, err := core.database.GetAll(paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return []model.User{}, err
		}

		return nil, fmt.Errorf("Error on getting user from database: %w", err)
	}

	return users, nil
}

func (core *User) GetByRole(role string, paginate int, qt int) ([]model.User, error) {
	users, err := core.database.GetByRole(role, paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return []model.User{}, err
		}

		return nil, fmt.Errorf("Error on getting user from database: %w", err)
	}

	return users, nil
}

func (core *User) Create(createdBy model.ID, partial model.UserPartial) error {
	err := validate(core.validator, partial)
	if err != nil {
		return err
	}

	valid, err := core.validateRoles(partial.Roles)
	if err != nil {
		return err
	}

	if !valid {
		return errs.ErrInvalidRoles
	}

	_, err = core.GetByUsername(partial.Username)
	if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
		return err
	}

	if err == nil {
		return errs.ErrUsernameAlreadyExist
	}

	_, err = core.GetByEmail(partial.Email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
		return err
	}

	if err == nil {
		return errs.ErrEmailAlreadyExist
	}

	hash, err := argon2id.CreateHash(partial.Password, &core.argon2id)
	if err != nil {
		return fmt.Errorf("error creating password hash: %w", err)
	}

	user := model.User{
		ID:        model.NewID(),
		Name:      partial.Name,
		Username:  partial.Username,
		Email:     partial.Email,
		Password:  hash,
		Roles:     partial.Roles,
		IsActive:  true,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
		DeletedAt: time.Time{},
		DeletedBy: model.ID{},
	}

	err = core.database.Create(user)
	if err != nil {
		return fmt.Errorf("Error creating user in the database: %w", err)
	}

	return nil
}

func (core *User) Update(userID model.ID, partial model.UserUpdate) error {
	err := validate(core.validator, partial)
	if err != nil {
		return err
	}

	valid, err := core.validateRoles(partial.Roles)
	if err != nil {
		return err
	}

	if !valid {
		return errs.ErrInvalidRoles
	}

	user, err := core.GetByID(userID)
	if err != nil {
		return err
	}

	if partial.Name != "" {
		user.Name = partial.Name
	}

	if partial.Username != "" {
		_, err = core.GetByUsername(partial.Username)
		if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
			return err
		}

		if err == nil {
			return errs.ErrUsernameAlreadyExist
		}

		user.Username = partial.Username
	}

	if partial.Email != "" {
		_, err = core.GetByEmail(partial.Email)
		if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
			return err
		}

		if err == nil {
			return errs.ErrEmailAlreadyExist
		}

		user.Email = partial.Email
	}

	if partial.Password != "" {
		hash, err := argon2id.CreateHash(partial.Password, &core.argon2id)
		if err != nil {
			return fmt.Errorf("error creating password hash: %w", err)
		}

		user.Password = hash
	}

	if partial.Roles != nil {
		user.Roles = partial.Roles
	}

	err = core.database.Update(*user)
	if err != nil {
		return fmt.Errorf("Error creating user in the database: %w", err)
	}

	return nil
}

func (core *User) Delete(userID model.ID, deleteByID model.ID) error {
	user, err := core.GetByID(userID)
	if err != nil {
		return err
	}

	err = core.database.Delete(user.ID, time.Now(), deleteByID)
	if err != nil {
		return fmt.Errorf("error deleting user from database: %w", err)
	}

	return nil
}
