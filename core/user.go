package core

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-playground/validator/v10"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"golang.org/x/exp/slices"
)

type User struct {
	database    data.User
	role        *Role
	validate    *validator.Validate
	argon2id    argon2id.Params
	argonEnable bool
}

func (u *User) GetByID(id model.ID) (*model.User, error) {
	user, err := u.database.GetByID(id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, errs.ErrUserNotFound
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return user, nil
}

func (u *User) GetByUsername(username string) (*model.User, error) {
	user, err := u.database.GetByUsername(username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, errs.ErrUserNotFound
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return user, nil
}

func (u *User) GetByEmail(email string) (*model.User, error) {
	user, err := u.database.GetByEmail(email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, errs.ErrUserNotFound
		}

		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return user, nil
}

func (u *User) GetAll(paginate int, qt int) ([]model.User, error) {
	users, err := u.database.GetAll(paginate, qt)
	if err != nil {
		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return users, nil
}

func (u *User) GetByRole(roles []string, paginate int, qt int) ([]model.User, error) {
	exist, err := u.role.Exist(roles)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errs.ErrRoleNotFound
	}

	users, err := u.database.GetByRoles(roles, paginate, qt)
	if err != nil {
		return nil, fmt.Errorf("error on getting user from database: %w", err)
	}

	return users, nil
}

func (u *User) Create(createdBy model.ID, partial model.UserPartial) (model.ID, error) {
	err := validate(u.validate, partial)
	if err != nil {
		return model.ID{}, err
	}

	exist, err := u.role.Exist(partial.Roles)
	if err != nil {
		return model.ID{}, err
	}

	if !exist {
		return model.ID{}, errs.ErrRoleNotFound
	}

	_, err = u.GetByUsername(partial.Username)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return model.ID{}, err
	}

	if err == nil {
		return model.ID{}, errs.ErrUsernameAlreadyExist
	}

	_, err = u.GetByEmail(partial.Email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return model.ID{}, err
	}

	if err == nil {
		return model.ID{}, errs.ErrEmailAlreadyExist
	}

	hash, err := u.createHash(partial.Password)
	if err != nil {
		return model.ID{}, fmt.Errorf("error creating password hash: %w", err)
	}

	roles := make([]string, 0, len(partial.Roles))
	for _, role := range partial.Roles {
		if !slices.Contains(roles, role) {
			roles = append(roles, role)
		}
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
		return model.ID{}, fmt.Errorf("error creating user in the database: %w", err)
	}

	return user.ID, nil
}

func (u *User) Update(userID model.ID, partial model.UserUpdate) error {
	err := validate(u.validate, partial)
	if err != nil {
		return err
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
		if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
			return err
		}

		if err == nil {
			return errs.ErrUsernameAlreadyExist
		}

		user.Username = partial.Username
	}

	if partial.Email != "" {
		_, err = u.GetByEmail(partial.Email)
		if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
			return err
		}

		if err == nil {
			return errs.ErrEmailAlreadyExist
		}

		user.Email = partial.Email
	}

	if partial.Password != "" {
		hash, err := u.createHash(partial.Password)
		if err != nil {
			return fmt.Errorf("error creating password hash: %w", err)
		}

		user.Password = hash
	}

	if partial.Roles != nil {
		exist, err := u.role.Exist(partial.Roles)
		if err != nil {
			return err
		}

		if !exist {
			return errs.ErrRoleNotFound
		}

		user.Roles = partial.Roles
	}

	if partial.IsActive != nil {
		user.IsActive = *partial.IsActive
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

func (u *User) createHash(password string) (string, error) {
	if u.argonEnable {
		hash, err := argon2id.CreateHash(password, &u.argon2id)
		if err != nil {
			return "", fmt.Errorf("error creating password hash: %w", err)
		}

		return hash, nil
	}

	hash := sha256.Sum256([]byte(password))

	return fmt.Sprintf("%x", hash), nil
}

func (u *User) EqualPassword(password string, hash string) (bool, error) {
	if u.argonEnable {
		match, err := argon2id.ComparePasswordAndHash(password, hash)
		if err != nil {
			return false, fmt.Errorf("error comparaing hash: %w", err)
		}

		return match, nil
	}

	hashp := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	return hash == hashp, nil
}

func NewUser(database data.User, role *Role, validate *validator.Validate, argonEnable bool) *User {
	return &User{
		database:    database,
		role:        role,
		validate:    validate,
		argon2id:    *argon2id.DefaultParams,
		argonEnable: argonEnable,
	}
}
