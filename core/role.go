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

type Role struct {
	database data.Role
	validate *validator.Validate
}

func (r *Role) GetByName(name string) (*model.Role, error) {
	role, err := r.database.GetByName(name)
	if err != nil {
		if errors.Is(err, errs.ErrRoleNotFoud) {
			return nil, errs.ErrRoleNotFoud
		}

		return nil, fmt.Errorf("error getting role from database: %w", err)
	}

	return role, nil
}

func (r *Role) GetAll(paginate int, qt int) ([]model.Role, error) {
	roles, err := r.database.GetAll(paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrRoleNotFoud) {
			return nil, errs.ErrRoleNotFoud
		}

		return nil, fmt.Errorf("error getting role from database: %w", err)
	}

	return roles, nil
}

func (r *Role) Create(createdBy model.ID, partial model.RolePartial) error {
	err := validate(r.validate, partial)
	if err != nil {
		return err
	}

	_, err = r.GetByName(partial.Name)
	if err != nil && !errors.Is(err, errs.ErrRoleNotFoud) {
		return err
	}

	if err == nil {
		return errs.ErrRoleAlreadyExist
	}

	role := model.Role{
		Name:      partial.Name,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
		DeletedAt: time.Time{},
		DeletedBy: model.ID{},
	}

	err = r.database.Create(role)
	if err != nil {
		return fmt.Errorf("error creating role in the database: %w", err)
	}

	return nil
}

func (r *Role) Delete(deleteBy model.ID, name string) error {
	_, err := r.GetByName(name)
	if err != nil {
		return err
	}

	err = r.database.Delete(name, time.Now(), deleteBy)
	if err != nil {
		return fmt.Errorf("error deleting role from database: %w", err)
	}

	return nil
}

func NewRole(database data.Role, validate *validator.Validate) *Role {
	return &Role{
		database: database,
		validate: validate,
	}
}
