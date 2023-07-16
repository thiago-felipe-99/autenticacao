package data

import (
	"time"

	"github.com/thiago-felipe-99/autenticacao/model"
)

type Role interface {
	GetByName(name string) (*model.Role, error)
	GetAll(paginate int, qt int) ([]model.Role, error)
	Create(role model.Role) error
	Delete(name string, deletedAt time.Time, deletedBy model.ID) error
}

type User interface {
	GetByID(id model.ID) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetAll(paginate int, qt int) ([]model.User, error)
	GetByRoles(role []string, paginate int, qt int) ([]model.User, error)
	Create(user model.User) error
	Update(user model.User) error
	Delete(id model.ID, deletedAt time.Time, deletedBy model.ID) error
}

type UserSession interface {
	GetByID(id model.ID) (*model.UserSession, error)
	GetAll(paginate int, qt int) ([]model.UserSession, error)
	GetByUserID(id model.ID, paginate int, qt int) ([]model.UserSession, error)
	Create(user model.UserSession) error
	Delete(id model.ID, deletetAd time.Time) error
}
