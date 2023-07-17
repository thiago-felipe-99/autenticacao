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
	database data.User
	role     *Role
	validate *validator.Validate
	argon2id argon2id.Params
}

func (u *User) rolesExists(roles []string) (bool, error) {
	for _, role := range roles {
		_, err := u.role.GetByName(role)
		if err != nil {
			if errors.Is(err, errs.ErrRoleNotFound) {
				return false, nil
			}

			return false, fmt.Errorf("error validating role: %w", err)
		}
	}

	return true, nil
}

func (u *User) GetByID(id model.ID) (*model.User, error) {
	user, err := u.database.GetByID(id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return user, nil
}

func (u *User) GetByUsername(username string) (*model.User, error) {
	user, err := u.database.GetByUsername(username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return user, nil
}

func (u *User) GetByEmail(email string) (*model.User, error) {
	user, err := u.database.GetByEmail(email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return user, nil
}

func (u *User) GetAll(paginate int, qt int) ([]model.User, error) {
	users, err := u.database.GetAll(paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return []model.User{}, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return users, nil
}

func (u *User) GetByRole(roles []string, paginate int, qt int) ([]model.User, error) {
	users, err := u.database.GetByRoles(roles, paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFoud) {
			return []model.User{}, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return users, nil
}

func (u *User) Create(createdBy model.ID, partial model.UserPartial) error {
	err := validate(u.validate, partial)
	if err != nil {
		return err
	}

	valid, err := u.rolesExists(partial.Roles)
	if err != nil {
		return err
	}

	if !valid {
		return errs.ErrRoleNotFound
	}

	_, err = u.GetByUsername(partial.Username)
	if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
		return err
	}

	if err == nil {
		return errs.ErrUsernameAlreadyExist
	}

	_, err = u.GetByEmail(partial.Email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
		return err
	}

	if err == nil {
		return errs.ErrEmailAlreadyExist
	}

	hash, err := argon2id.CreateHash(partial.Password, &u.argon2id)
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

	err = u.database.Create(user)
	if err != nil {
		return fmt.Errorf("error creating user in the database: %w", err)
	}

	return nil
}

func (u *User) Update(userID model.ID, partial model.UserUpdate) error {
	err := validate(u.validate, partial)
	if err != nil {
		return err
	}

	valid, err := u.rolesExists(partial.Roles)
	if err != nil {
		return err
	}

	if !valid {
		return errs.ErrRoleNotFound
	}

	user, err := u.GetByID(userID)
	if err != nil {
		return err
	}

	if partial.Name != "" {
		user.Name = partial.Name
	}

	if partial.Username != "" {
		_, err = u.GetByUsername(partial.Username)
		if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
			return err
		}

		if err == nil {
			return errs.ErrUsernameAlreadyExist
		}

		user.Username = partial.Username
	}

	if partial.Email != "" {
		_, err = u.GetByEmail(partial.Email)
		if err != nil && !errors.Is(err, errs.ErrUserNotFoud) {
			return err
		}

		if err == nil {
			return errs.ErrEmailAlreadyExist
		}

		user.Email = partial.Email
	}

	if partial.Password != "" {
		hash, err := argon2id.CreateHash(partial.Password, &u.argon2id)
		if err != nil {
			return fmt.Errorf("error creating password hash: %w", err)
		}

		user.Password = hash
	}

	if partial.Roles != nil {
		user.Roles = partial.Roles
	}

	err = u.database.Update(*user)
	if err != nil {
		return fmt.Errorf("error creating user in the database: %w", err)
	}

	return nil
}

func (u *User) Delete(userID model.ID, deleteByID model.ID) error {
	user, err := u.GetByID(userID)
	if err != nil {
		return err
	}

	err = u.database.Delete(user.ID, time.Now(), deleteByID)
	if err != nil {
		return fmt.Errorf("error deleting user from database: %w", err)
	}

	return nil
}

func NewUser(database data.User, role *Role, validate *validator.Validate) *User {
	return &User{
		database: database,
		role:     role,
		validate: validate,
		argon2id: *argon2id.DefaultParams,
	}
}
