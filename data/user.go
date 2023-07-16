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
	db *sqlx.DB
}

func (database *UserSQL) GetByID(id model.ID) (*model.User, error) {
	user := &model.UserPostgres{}

	err := database.db.Get(
		user,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM 
			users
		WHERE 
			deleted_at = $1 AND
			id = $2`,
		time.Time{},
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("Error get user by id in database: %w", err)
	}

	return user.User(), nil
}

func (database *UserSQL) GetByUsername(username string) (*model.User, error) {
	user := &model.UserPostgres{}

	err := database.db.Get(
		user,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM 
			users
		WHERE 
			deleted_at = $1 AND
			username = $2`,
		time.Time{},
		username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("Error get user by username in database: %w", err)
	}

	return user.User(), nil
}

func (database *UserSQL) GetByEmail(email string) (*model.User, error) {
	user := &model.UserPostgres{}

	err := database.db.Get(
		user,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM 
			users
		WHERE 
			deleted_at = $1 AND
			email = $2`,
		time.Time{},
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("Error get user by email in database: %w", err)
	}

	return user.User(), nil
}

func (database *UserSQL) GetAll(paginate int, qt int) ([]model.User, error) {
	partial := []model.UserPostgres{}

	err := database.db.Select(
		partial,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM 
			users
		LIMIT $1 
		OFFSET $2`,
		qt,
		qt*paginate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("Error get users in database: %w", err)
	}

	users := make([]model.User, 0, len(partial))
	for _, user := range partial {
		users = append(users, *user.User())
	}

	return users, nil
}

func (database *UserSQL) GetByRoles(roles []string, paginate int, qt int) ([]model.User, error) {
	partial := []model.UserPostgres{}

	err := database.db.Select(
		partial,
		`SELECT 
			id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by
		FROM 
			users
		WHERE 
			roles @> $1
		LIMIT $2 
		OFFSET $3`,
		pq.StringArray(roles),
		qt,
		qt*paginate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFoud
		}

		return nil, fmt.Errorf("Error get users by role in database: %w", err)
	}

	users := make([]model.User, 0, len(partial))
	for _, user := range partial {
		users = append(users, *user.User())
	}

	return users, nil
}

func (database *UserSQL) Create(user model.User) error {
	_, err := database.db.NamedExec(
		`INSERT INTO users
			(id, name, username, email, password, roles, is_active, created_at, created_by, deleted_at, deleted_by)
		VALUES 
			(:id, :name, :username, :email, :password, :roles, :is_active, :created_at, :created_by, :deleted_at, :deleted_by)`,
		user.Postgres(),
	)
	if err != nil {
		return fmt.Errorf("Error inserting user: %w", err)
	}

	return nil
}

func (database *UserSQL) Update(user model.User) error {
	_, err := database.db.NamedExec(
		`UPDATE users SET
			name = :name, 
			username = :username, 
			email = :email, 
			password = :password, 
			roles = :roles, 
			is_active = :is_active
		WHERE 
			id = :id`,
		user.Postgres(),
	)
	if err != nil {
		return fmt.Errorf("Error updating user: %w", err)
	}

	return nil
}

func (database *UserSQL) Delete(id model.ID, deletedAt time.Time, deletedBy model.ID) error {
	_, err := database.db.Exec(
		"UPDATE users SET deleted_at=$1, deleted_by=$2 WHERE id=$3",
		deletedAt,
		deletedBy,
		id,
	)
	if err != nil {
		return fmt.Errorf("Error deleting user: %w", err)
	}

	return nil
}

var _ User = &UserSQL{}

func NewUserSQL(db *sqlx.DB) *UserSQL {
	return &UserSQL{
		db: db,
	}
}
