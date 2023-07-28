package data_test

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func logErrors(t *testing.T, errs <-chan error) {
	t.Helper()

	err := <-errs
	t.Log(err)
	require.NoError(t, err)
}

func createUserSession(userID model.ID) model.UserSession {
	return model.UserSession{
		ID:        model.NewID(),
		UserID:    userID,
		CreateaAt: time.Now(),
		Expires:   time.Now().Add(time.Second),
		DeletedAt: time.Time{},
	}
}

func TestUserSessionCreate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "data_user_session_create")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	user := data.NewUserSQL(db)
	userSession := data.NewUserSessionRedis(redisClient, db, overflowBuffer)
	err := userSession.ConsumeQueues(time.Second, buffer)
	require.NoError(t, err)
	userSession.LogErrors()

	go logErrors(t, userSession.Errors())

	userTemp := createUser()
	err = user.Create(userTemp)
	require.NoError(t, err)

	for i := 0; i < overflowBuffer; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			err := userSession.Create(createUserSession(userTemp.ID))
			require.NoError(t, err)
		})
	}

	for i := 0; i < buffer/2; i++ {
		err := userSession.Create(createUserSession(userTemp.ID))
		require.NoError(t, err)
	}

	time.Sleep(time.Second)

	t.Run("WrongRedis", func(t *testing.T) {
		t.Parallel()
		redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
			Addr:     "wrong:6379",
			Password: "redis",
			DB:       0,
		})

		userSession := data.NewUserSessionRedis(redisClient, db, overflowBuffer)
		err := userSession.ConsumeQueues(time.Second, buffer)
		require.NoError(t, err)

		err = userSession.Create(createUserSession(userTemp.ID))
		require.ErrorContains(t, err, "no such host")
	})

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		db := createWrongDB(t)
		userSession := data.NewUserSessionRedis(redisClient, db, overflowBuffer)
		err := userSession.ConsumeQueues(time.Second, buffer)
		require.NoError(t, err)

		err = userSession.Create(createUserSession(userTemp.ID))
		require.NoError(t, err)

		errCh := <-userSession.Errors()

		require.ErrorContains(t, errCh, "no such host")

		userSession.LogErrors()

		err = userSession.Create(createUserSession(userTemp.ID))
		require.NoError(t, err)

		time.Sleep(time.Second)
	})

	t.Run("InvalidConsumeQueues", func(t *testing.T) {
		t.Parallel()

		err := userSession.ConsumeQueues(time.Second, overflowBuffer*2)
		require.ErrorIs(t, err, data.ErrMaxBiggerThanBuffer)

		err = userSession.ConsumeQueues(time.Second/2, buffer)
		require.ErrorIs(t, err, data.ErrSmallClock)
	})
}
