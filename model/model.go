package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ID uuid.UUID

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func NewID() ID {
	return ID(uuid.New())
}

func ParseID(id string) (ID, error) {
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return ID(uuid.UUID{}), fmt.Errorf("error parsing ID: %w", err)
	}

	return ID(idUUID), nil
}

type RolePartial struct {
	Name string `config:"name" json:"name" validate:"required,max=255"`
}

type Role struct {
	Name      string    `json:"name"                db:"name"`
	CreatedAt time.Time `json:"createdAt"           db:"created_at"`
	CreatedBy ID        `json:"createdBy"           db:"created_by"`
	DeletedAt time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
	DeletedBy ID        `json:"deletedBy,omitempty" db:"deleted_by"`
}

type UserPartial struct {
	Name     string   `config:"name"     json:"name"     validate:"required,max=255"`
	Username string   `config:"username" json:"username" validate:"required,alphanumunicode,max=255"`
	Email    string   `config:"email"    json:"email"    validate:"required,email,max=255"`
	Password string   `config:"password" json:"password" validate:"required"`
	Roles    []string `                  json:"roles"    validate:"omitempty"`
}

type UserUpdate struct {
	Name     string   `json:"name"     validate:"omitempty,max=255"`
	Username string   `json:"username" validate:"omitempty,alphanumunicode,max=255"`
	Email    string   `json:"email"    validate:"omitempty,email,max=255"`
	Password string   `json:"password" validate:"omitempty"`
	Roles    []string `json:"roles"    validate:"omitempty"`
	IsActive *bool    `json:"isActive" validate:"omitempty"`
}

type User struct {
	ID        ID        `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy ID        `json:"createdBy"`
	DeletedAt time.Time `json:"deletedAt,omitempty"`
	DeletedBy ID        `json:"deletedBy,omitempty"`
}

func (user *User) Postgres() *UserPostgres {
	return &UserPostgres{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		IsActive:  user.IsActive,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt,
		CreatedBy: user.CreatedBy,
		DeletedAt: user.DeletedAt,
		DeletedBy: user.DeletedBy,
	}
}

type UserPostgres struct {
	ID        ID             `db:"id"`
	Name      string         `db:"name"`
	Username  string         `db:"username"`
	Email     string         `db:"email"`
	Password  string         `db:"password"`
	Roles     pq.StringArray `db:"roles"`
	IsActive  bool           `db:"is_active"`
	CreatedAt time.Time      `db:"created_at"`
	CreatedBy ID             `db:"created_by"`
	DeletedAt time.Time      `db:"deleted_at"`
	DeletedBy ID             `db:"deleted_by"`
}

func (user *UserPostgres) User() *User {
	return &User{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		IsActive:  user.IsActive,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt,
		CreatedBy: user.CreatedBy,
		DeletedAt: user.DeletedAt,
		DeletedBy: user.DeletedBy,
	}
}

type UserSessionPartial struct {
	Username string `json:"username" validate:"required_without=Email,excluded_with=Email"`
	Email    string `json:"email"    validate:"required_without=Username,excluded_with=Username,omitempty,email"`
	Password string `json:"password" validate:"required"`
}

type UserSession struct {
	ID        ID        `json:"id"                  db:"id"`
	UserID    ID        `json:"userId"              db:"userid"`
	CreateaAt time.Time `json:"createdAt"           db:"created_at"`
	Expires   time.Time `json:"expires"             db:"expires"`
	DeletedAt time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}
