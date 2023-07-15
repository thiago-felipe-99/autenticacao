package main

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
	Name string `config:"name" json:"name" validate:"required"`
}

type Role struct {
	ID        ID        `json:"id"                  bson:"_id"`
	Name      string    `json:"name"                bson:"name"`
	CreatedAt time.Time `json:"createdAt"           bson:"created_at"`
	CreatedBy ID        `json:"createdBy"           bson:"created_by"`
	DeletedAt time.Time `json:"deletedAt,omitempty" bson:"deleted_at"`
	DeletedBy ID        `json:"deletedBy,omitempty" bson:"deleted_by"`
}

type UserPartial struct {
	Name     string   `config:"name"     json:"name"     validate:"required"`
	Username string   `config:"username" json:"username" validate:"required,alphanumunicode"`
	Email    string   `config:"email"    json:"email"    validate:"required,email"`
	Password string   `config:"password" json:"password" validate:"required"`
	Roles    []string `                  json:"roles"    validate:""`
}

type User struct {
	ID        ID        `json:"id"                  bson:"_id"`
	Name      string    `json:"name"                bson:"name"`
	Username  string    `json:"username"            bson:"username"`
	Email     string    `json:"email"               bson:"email"`
	Password  string    `json:"password,omitempty"  bson:"password"`
	Roles     []string  `json:"roles,omitempty"     bson:"roles"`
	IsActive  bool      `json:"isActive"            bson:"is_active"`
	CreatedAt time.Time `json:"createdAt"           bson:"created_at"`
	CreatedBy ID        `json:"createdBy"           bson:"created_by"`
	DeletedAt time.Time `json:"deletedAt,omitempty" bson:"deleted_at"`
	DeletedBy ID        `json:"deletedBy,omitempty" bson:"deleted_by"`
}

type UserSessionPartial struct {
	Name     string `json:"name"     validate:"required_without=Email,excluded_with=Email"`
	Email    string `json:"email"    validate:"required_without=Name,excluded_with=Name,omitempty,email"`
	Password string `json:"password" validate:"required"`
}

type UserSession struct {
	ID        ID        `json:"id"                  bson:"_id"`
	UserID    ID        `json:"userId"              bson:"user_id"`
	CreateaAt time.Time `json:"createdAt"           bson:"created_at"`
	Expires   time.Time `json:"expires"             bson:"expires"`
	DeletedAt time.Time `json:"deletedAt,omitempty" bson:"deleted_at"`
}
