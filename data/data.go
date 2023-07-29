package data

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/thiago-felipe-99/autenticacao/model"
)

type Role interface {
	GetByName(name string) (model.Role, error)
	GetAll(paginate int, qt int) ([]model.Role, error)
	Exist(roles []string) (bool, error)
	Create(role model.Role) error
	Delete(name string, deletedAt time.Time, deletedBy model.ID) error
}

type User interface {
	GetByID(id model.ID) (model.User, error)
	GetByUsername(username string) (model.User, error)
	GetByEmail(email string) (model.User, error)
	GetAll(paginate int, qt int) ([]model.User, error)
	GetByRoles(role []string, paginate int, qt int) ([]model.User, error)
	Create(user model.User) error
	Update(user model.User) error
	Delete(id model.ID, deletedAt time.Time, deletedBy model.ID) error
}

type UserSession interface {
	GetAllActive(paginate int, qt int) ([]model.UserSession, error)
	GetByUserIDActive(id model.ID, paginate int, qt int) ([]model.UserSession, error)
	GetAllInactive(paginate int, qt int) ([]model.UserSession, error)
	GetByUserIDInactive(id model.ID, paginate int, qt int) ([]model.UserSession, error)
	Create(user model.UserSession) error
	Delete(id model.ID, deletetAd time.Time) (model.UserSession, error)
}

type Data struct {
	Role
	User
	UserSession
}

func NewDataSQLRedis(
	db *sqlx.DB,
	redis *redis.Client,
	expires time.Duration,
	bufferSize int,
	queueSize int,
) (*Data, error) {
	role := NewRoleSQL(db)
	user := NewUserSQL(db)
	userSession := NewUserSessionRedis(redis, db, bufferSize)

	err := userSession.ConsumeQueues(expires, queueSize)
	userSession.LogErrors()

	return &Data{
		Role:        role,
		User:        user,
		UserSession: userSession,
	}, err
}
