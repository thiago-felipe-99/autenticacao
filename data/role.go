package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type RoleSQL struct {
	db *sqlx.DB
}

func (database *RoleSQL) GetByName(name string) (*model.Role, error) {
	role := &model.Role{}

	err := database.db.Get(
		role,
		`SELECT 
			name, created_at, created_by, deleted_at, deleted_by 
		FROM 
			role 
		WHERE 
		name=$1`,
		name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrRoleNotFoud
		}

		return nil, fmt.Errorf("Error get role by name in database: %w", err)
	}

	return role, nil
}

func (database *RoleSQL) GetAll(paginate int, qt int) ([]model.Role, error) {
	role := []model.Role{}

	err := database.db.Select(
		role,
		`SELECT 
			name, created_at, created_by, deleted_at, deleted_by 
		FROM 
			role 
		LIMIT $1 
		OFFSET $2`,
		qt,
		qt*paginate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrRoleNotFoud
		}

		return nil, fmt.Errorf("Error get role by name in database: %w", err)
	}

	return role, nil
}

func (database *RoleSQL) Create(role model.Role) error {
	_, err := database.db.NamedExec(
		`INSERT INTO role 
			(name, created_at, created_by, deleted_at, deleted_by)
		VALUES 
			(:name, :created_at, :created_by, :deleted_at, :deleted_by)`,
		role,
	)
	if err != nil {
		return fmt.Errorf("Error inserting role: %w", err)
	}

	return nil
}

func (database *RoleSQL) Delete(name string, deletedAt time.Time, deletedBy model.ID) (err error) {
	tx, err := database.db.Begin()
	if err != nil {
		return fmt.Errorf("Error beging transaction: %w", err)
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			newErr := tx.Rollback()
			if newErr != nil {
				err = fmt.Errorf("Error roolback transaction: %w", errors.Join(newErr, err))
			}
		}
	}(tx)

	_, err = tx.Exec("UPDATE users SET roles = array_remove(roles, $1)", name)
	if err != nil {
		return fmt.Errorf("Error deleting users roles: %w", err)
	}

	_, err = tx.Exec(
		"UPDATE role SET deleted_at=$1, deleted_by=$2 WHERE name=$3",
		deletedAt,
		deletedBy,
		name,
	)
	if err != nil {
		return fmt.Errorf("Error deleting role: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Error commiting transaction: %w", err)
	}

	return nil
}

var _ Role = &RoleSQL{}

func NewRoleSQL(db *sqlx.DB) *RoleSQL {
	return &RoleSQL{
		db: db,
	}
}
