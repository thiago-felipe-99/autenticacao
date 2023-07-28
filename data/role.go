package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type RoleSQL struct {
	database *sqlx.DB
}

func (r *RoleSQL) GetByName(name string) (model.Role, error) {
	role := model.Role{} //nolint: exhaustruct

	err := r.database.Get(
		&role,
		`SELECT name, created_at, created_by, deleted_at, deleted_by 
		FROM role 
		WHERE deleted_at = $1 AND name=$2`,
		time.Time{},
		name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.EmptyRole, errs.ErrRoleNotFound
		}

		return model.EmptyRole, fmt.Errorf("error get role by name in database: %w", err)
	}

	return role, nil
}

func (r *RoleSQL) GetAll(paginate int, qt int) ([]model.Role, error) {
	roles := make([]model.Role, 0, qt)

	err := r.database.Select(
		&roles,
		`SELECT name, created_at, created_by, deleted_at, deleted_by 
		FROM role 
		LIMIT $1 
		OFFSET $2`,
		qt,
		qt*paginate,
	)
	if err != nil {
		return model.EmptyRoles, fmt.Errorf("error get roles in database: %w", err)
	}

	return roles, nil
}

func (r *RoleSQL) Exist(roles []string) (bool, error) {
	count := 0

	err := r.database.Get(
		&count,
		`SELECT COUNT(i) FROM unnest($1::text[]) i
		LEFT JOIN role r ON i = r.name
		WHERE r.deleted_at = $2`,
		pq.StringArray(roles),
		time.Time{},
	)
	if err != nil {
		return false, fmt.Errorf("error verifying if roles exist: %w", err)
	}

	return count == len(roles), nil
}

func (r *RoleSQL) Create(role model.Role) error {
	_, err := r.database.NamedExec(
		`INSERT INTO role (name, created_at, created_by, deleted_at, deleted_by)
		VALUES (:name, :created_at, :created_by, :deleted_at, :deleted_by)`,
		role,
	)
	if err != nil {
		return fmt.Errorf("error inserting role: %w", err)
	}

	return nil
}

func (r *RoleSQL) Delete(name string, deletedAt time.Time, deletedBy model.ID) (err error) {
	tx, err := r.database.Begin()
	if err != nil {
		return fmt.Errorf("error beging transaction: %w", err)
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			newErr := tx.Rollback()
			if newErr != nil {
				err = fmt.Errorf("error roolback transaction: %w", errors.Join(newErr, err))
			}
		}
	}(tx)

	_, err = tx.Exec("UPDATE users SET roles = array_remove(roles, $1)", name)
	if err != nil {
		return fmt.Errorf("error deleting users roles: %w", err)
	}

	_, err = tx.Exec(
		"UPDATE role SET deleted_at=$1, deleted_by=$2 WHERE name=$3",
		deletedAt,
		deletedBy,
		name,
	)
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

var _ Role = &RoleSQL{} //nolint: exhaustruct

func NewRoleSQL(db *sqlx.DB) *RoleSQL {
	return &RoleSQL{
		database: db,
	}
}
