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

func (core *Role) GetByName(name string) (*model.Role, error) {
	role, err := core.database.GetByName(name)
	if err != nil {
		if errors.Is(err, errs.ErrRoleNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error getting role from database: %w", err)
	}

	return role, nil
}

func (core *Role) GetAll(paginate int, qt int) ([]model.Role, error) {
	roles, err := core.database.GetAll(paginate, qt)
	if err != nil {
		if errors.Is(err, errs.ErrRoleNotFoud) {
			return nil, err
		}

		return nil, fmt.Errorf("Error getting role from database: %w", err)
	}

	return roles, nil
}

func (core *Role) Create(createdBy model.ID, partial model.RolePartial) error {
	err := validate(core.validate, partial)
	if err != nil {
		return err
	}

	_, err = core.GetByName(partial.Name)
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

	err = core.database.Create(role)
	if err != nil {
		return fmt.Errorf("Error creating role in the database: %w", err)
	}

	return nil
}

func (core *Role) Delete(deleteBy model.ID, name string) error {
	_, err := core.GetByName(name)
	if err != nil {
		return err
	}

	err = core.database.Delete(name, time.Now(), deleteBy)
	if err != nil {
		return fmt.Errorf("Error deleting role from database: %w", err)
	}

	return nil
}

func NewRole(database data.Role, validate *validator.Validate) *Role {
	return &Role{
		database: database,
		validate: validate,
	}
}
