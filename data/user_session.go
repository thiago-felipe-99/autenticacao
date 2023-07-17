package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"github.com/vmihailenco/msgpack/v5"
)

type UserSessionRedis struct {
	redis    *redis.Client
	database *sqlx.DB
	created  chan model.UserSession
	deleted  chan model.UserSession
}

func (u *UserSessionRedis) GetAll(paginate int, qt int) ([]model.UserSession, error) {
	role := []model.UserSession{}

	err := u.database.Select(
		&role,
		`SELECT id, userid, created_at, deleted_at
		FROM users_sessions_created 
		LIMIT $1 
		OFFSET $2`,
		qt,
		qt*paginate,
	)
	if err != nil {
		return nil, fmt.Errorf("error get user sessions in database: %w", err)
	}

	return role, nil
}

func (u *UserSessionRedis) GetByUserID(
	id model.ID,
	paginate int,
	qt int,
) ([]model.UserSession, error) {
	role := []model.UserSession{}

	err := u.database.Select(
		&role,
		`SELECT id, userid, created_at, deleted_at
		FROM users_sessions_created 
		WHERE userid = $1
		LIMIT $2 
		OFFSET $3`,
		id,
		qt,
		qt*paginate,
	)
	if err != nil {
		return nil, fmt.Errorf("error get user sessions in database: %w", err)
	}

	return role, nil
}

func (u *UserSessionRedis) Create(user model.UserSession, expires time.Duration) error {
	serial, err := msgpack.Marshal(&user)
	if err != nil {
		return fmt.Errorf("error marshaling user session: %w", err)
	}

	_, err = u.redis.Set(context.Background(), user.ID.String(), serial, expires).Result()
	if err != nil {
		return fmt.Errorf("error setting user session in redis: %w", err)
	}

	u.created <- user

	return nil
}

func (u *UserSessionRedis) Refresh(
	oldID model.ID,
	deletedAt time.Time,
	newID model.ID,
	expires time.Duration,
) (*model.UserSession, error) {
	serial, err := u.redis.GetDel(context.Background(), oldID.String()).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrUserSessionNotFoud
		}

		return nil, fmt.Errorf("error getting user session from redis: %w", err)
	}

	var userSession model.UserSession

	err = msgpack.Unmarshal(serial, &userSession)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user session: %w", err)
	}

	userDeleted := userSession
	userDeleted.DeletedAt = deletedAt
	u.deleted <- userDeleted

	userSession.ID = newID

	serial, err = msgpack.Marshal(&userSession)
	if err != nil {
		return nil, fmt.Errorf("error marshaling user session: %w", err)
	}

	_, err = u.redis.Set(context.Background(), newID.String(), serial, expires).Result()
	if err != nil {
		return nil, fmt.Errorf("error setting user session in redis: %w", err)
	}

	u.created <- userSession

	return &userSession, nil
}

func (u *UserSessionRedis) Delete(id model.ID, deletetAd time.Time) (*model.UserSession, error) {
	serial, err := u.redis.GetDel(context.Background(), id.String()).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrUserSessionNotFoud
		}

		return nil, fmt.Errorf("error getting user session from redis: %w", err)
	}

	var userSession model.UserSession

	err = msgpack.Unmarshal(serial, &userSession)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user session: %w", err)
	}

	userSession.DeletedAt = deletetAd
	u.deleted <- userSession

	return &userSession, nil
}

func (u *UserSessionRedis) consumeCreated() {}

func (u *UserSessionRedis) consumeDeleted() {}

func (u *UserSessionRedis) ConsumeQueues() {
	go u.consumeCreated()
	go u.consumeDeleted()
}

var _ UserSession = &UserSessionRedis{} //nolint:exhaustruct

func NewUserSessionRedis(
	redis *redis.Client,
	database *sqlx.DB,
	createdSize int,
	deletedSize int,
) *UserSessionRedis {
	return &UserSessionRedis{
		redis:    redis,
		database: database,
		created:  make(chan model.UserSession, createdSize),
		deleted:  make(chan model.UserSession, deletedSize),
	}
}
