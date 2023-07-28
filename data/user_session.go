package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrMaxBiggerThanBuffer    = fmt.Errorf("max must be less than buffer size")
	ErrSmallClock             = fmt.Errorf("clock must be greater o equal to 1 second")
	ErrInsertingUserSessionDB = fmt.Errorf("error inserting user session in db")
)

type UserSessionRedis struct {
	redis      *redis.Client
	database   *sqlx.DB
	created    chan model.UserSession
	deleted    chan model.UserSession
	bufferSize int
	errs       chan error
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

func (u *UserSessionRedis) consumeChan(
	clock time.Duration,
	max int,
	users <-chan model.UserSession,
	table string,
) {
	ticker := time.NewTicker(clock)

	usersSessions := make([]model.UserSession, 0, max)

	query := fmt.Sprintf(
		"INSERT INTO %s (id, userid, created_at, deleted_at) VALUES (:id, :userid, :created_at, :deleted_at)",
		table,
	)

	saveDatabase := func() {
		_, err := u.database.NamedExec(query, usersSessions)
		if err != nil {
			u.errs <- errors.Join(ErrInsertingUserSessionDB, err)
		}

		usersSessions = usersSessions[:0]
	}

	for {
		select {
		case userSession := <-users:
			usersSessions = append(usersSessions, userSession)

			if len(usersSessions) >= max {
				saveDatabase()

				ticker.Reset(clock)
			}

		case <-ticker.C:
			if len(usersSessions) > 0 {
				saveDatabase()
			}
		}
	}
}

func (u *UserSessionRedis) ConsumeQueues(clock time.Duration, max int) error {
	if max >= u.bufferSize {
		return ErrMaxBiggerThanBuffer
	}

	if clock < time.Second {
		return ErrSmallClock
	}

	go u.consumeChan(clock, max, u.created, "users_sessions_created")
	go u.consumeChan(clock, max, u.deleted, "users_sessions_deleted")

	return nil
}

func (u *UserSessionRedis) Errors() <-chan error {
	return u.errs
}

func (u *UserSessionRedis) LogErrors() {
	go func() {
		for err := range u.Errors() {
			log.Printf("[ERROR] - %v", err)
		}
	}()
}

var _ UserSession = &UserSessionRedis{} //nolint:exhaustruct

func NewUserSessionRedis(
	redis *redis.Client,
	database *sqlx.DB,
	bufferSize int,
) *UserSessionRedis {
	return &UserSessionRedis{
		redis:      redis,
		database:   database,
		created:    make(chan model.UserSession, bufferSize),
		deleted:    make(chan model.UserSession, bufferSize),
		bufferSize: bufferSize,
		errs:       make(chan error, bufferSize),
	}
}
