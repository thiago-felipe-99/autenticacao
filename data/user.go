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

type UserSQL struct {
	database *sqlx.DB
}

func (u *UserSQL) GetByID(id model.ID) (*model.User, error) {
	user := model.UserPostgres{} //nolint: exhaustruct

	err := u.database.Get(
		&user,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM users
		WHERE deleted_at = $1 AND id = $2`,
		time.Time{},
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error get user by id in database: %w", err)
	}

	return user.User(), nil
}

func (u *UserSQL) GetByUsername(username string) (*model.User, error) {
	user := model.UserPostgres{} //nolint: exhaustruct

	err := u.database.Get(
		&user,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM users
		WHERE deleted_at = $1 AND username = $2`,
		time.Time{},
		username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error get user by username in database: %w", err)
	}

	return user.User(), nil
}

func (u *UserSQL) GetByEmail(email string) (*model.User, error) {
	user := model.UserPostgres{} //nolint: exhaustruct

	err := u.database.Get(
		&user,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM users
		WHERE deleted_at = $1 AND email = $2`,
		time.Time{},
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("error get user by email in database: %w", err)
	}

	return user.User(), nil
}

func (u *UserSQL) GetAll(paginate int, qt int) ([]model.User, error) {
	partial := []model.UserPostgres{}

	err := u.database.Select(
		&partial,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM users
		LIMIT $1 
		OFFSET $2`,
		qt,
		qt*paginate,
	)
	if err != nil {
		return nil, fmt.Errorf("error get users in database: %w", err)
	}

	users := make([]model.User, 0, len(partial))
	for _, user := range partial {
		users = append(users, *user.User())
	}

	return users, nil
}

func (u *UserSQL) GetByRoles(roles []string, paginate int, qt int) ([]model.User, error) {
	partial := []model.UserPostgres{}

	err := u.database.Select(
		&partial,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM users
		WHERE roles @> $1
		LIMIT $2 
		OFFSET $3`,
		pq.StringArray(roles),
		qt,
		qt*paginate,
	)
	if err != nil {
		return nil, fmt.Errorf("error get users by role in database: %w", err)
	}

	users := make([]model.User, 0, len(partial))
	for _, user := range partial {
		users = append(users, *user.User())
	}

	return users, nil
}

func (u *UserSQL) Create(user model.User) error {
	_, err := u.database.NamedExec(
		`INSERT INTO users
			(id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by)
		VALUES 
			(:id, :name, :username, :email, :password, :roles, :is_active, :created_at, :created_by, :deleted_at, :deleted_by)`,
		user.Postgres(),
	)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	return nil
}

func (u *UserSQL) Update(user model.User) error {
	_, err := u.database.NamedExec(
		`UPDATE users SET
			name = :name, 
			username = :username, 
			email = :email, 
			password = :password, 
			roles = :roles, 
			is_active = :is_active
		WHERE id = :id`,
		user.Postgres(),
	)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func (u *UserSQL) Delete(id model.ID, deletedAt time.Time, deletedBy model.ID) error {
	_, err := u.database.Exec(
		"UPDATE users SET deleted_at=$1, deleted_by=$2 WHERE id=$3",
		deletedAt,
		deletedBy,
		id,
	)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

var _ User = &UserSQL{} //nolint: exhaustruct

func NewUserSQL(db *sqlx.DB) *UserSQL {
	return &UserSQL{
		database: db,
	}
}
