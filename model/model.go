package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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
	Name string `config:"name" json:"name" validate:"required,max=256"`
}

type Role struct {
	Name      string    `json:"name"                db:"name"`
	CreatedAt time.Time `json:"createdAt"           db:"created_at"`
	CreatedBy ID        `json:"createdBy"           db:"created_by"`
	DeletedAt time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
	DeletedBy ID        `json:"deletedBy,omitempty" db:"deleted_by"`
}

type UserPartial struct {
	Name     string   `config:"name"     json:"name"     validate:"required,max=256"`
	Username string   `config:"username" json:"username" validate:"required,alphanumunicode,max=256"`
	Email    string   `config:"email"    json:"email"    validate:"required,email,max=256"`
	Password string   `config:"password" json:"password" validate:"required"`
	Roles    []string `                  json:"roles"    validate:"omitempty"`
}

type UserUpdate struct {
	Name     string   `json:"name"     validate:"omitempty,max=256"`
	Username string   `json:"username" validate:"omitempty,alphanumunicode,max=256"`
	Email    string   `json:"email"    validate:"omitempty,email,max=256"`
	Password string   `json:"password" validate:"omitempty"`
	Roles    []string `json:"roles"    validate:"omitempty"`
}

type User struct {
	ID        ID        `json:"id"                  db:"id"`
	Name      string    `json:"name"                db:"name"`
	Username  string    `json:"username"            db:"username"`
	Email     string    `json:"email"               db:"email"`
	Password  string    `json:"password,omitempty"  db:"password"`
	Roles     []string  `json:"roles,omitempty"     db:"roles"`
	IsActive  bool      `json:"isActive"            db:"is_active"`
	CreatedAt time.Time `json:"createdAt"           db:"created_at"`
	CreatedBy ID        `json:"createdBy"           db:"created_by"`
	DeletedAt time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
	DeletedBy ID        `json:"deletedBy,omitempty" db:"deleted_by"`
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
