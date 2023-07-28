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

	for i := 0; i < overflowBuffer+buffer/2; i++ {
		err := userSession.Create(createUserSession(userTemp.ID))
		require.NoError(t, err)
	}

	time.Sleep(time.Second * 2)

	t.Run("InvalidConsumeQueues", func(t *testing.T) {
		t.Parallel()

		err := userSession.ConsumeQueues(time.Second, overflowBuffer*2)
		require.ErrorIs(t, err, data.ErrMaxBiggerThanBuffer)

		err = userSession.ConsumeQueues(time.Second/2, buffer)
		require.ErrorIs(t, err, data.ErrSmallClock)
	})
}

func TestUserSessionWrongDB(t *testing.T) {
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

	userTemp := createUser()
	err = user.Create(userTemp)
	require.NoError(t, err)

	t.Run("Redis", func(t *testing.T) {
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

		userSessionTemp, err := userSession.Delete(model.NewID(), time.Now())
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, model.EmptyUserSession, userSessionTemp)
	})

	t.Run("SQL", func(t *testing.T) {
		t.Parallel()

		db := createWrongDB(t)

		userSession := data.NewUserSessionRedis(redisClient, db, overflowBuffer)
		err := userSession.ConsumeQueues(time.Second, buffer)
		require.NoError(t, err)

		userSessionTemp1 := createUserSession(userTemp.ID)

		err = userSession.Create(userSessionTemp1)
		require.NoError(t, err)

		err = <-userSession.Errors()
		require.ErrorContains(t, err, "no such host")

		userSessions, err := userSession.GetAllActive(0, buffer)
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, model.EmptyUserSessions, userSessions)

		userSessions, err = userSession.GetAllInactive(0, buffer)
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, model.EmptyUserSessions, userSessions)

		userSessions, err = userSession.GetByUserIDActive(model.NewID(), 0, buffer)
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, model.EmptyUserSessions, userSessions)

		userSessions, err = userSession.GetByUserIDInactive(model.NewID(), 0, buffer)
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, model.EmptyUserSessions, userSessions)

		userSessionTemp2, err := userSession.Delete(userSessionTemp1.ID, time.Now())
		require.NoError(t, err)
		require.Equal(t, userSessionTemp1.ID, userSessionTemp2.ID)
		require.Equal(t, userSessionTemp1.UserID, userSessionTemp2.UserID)
		require.LessOrEqual(
			t,
			userSessionTemp1.CreateaAt.Sub(userSessionTemp2.CreateaAt),
			time.Second,
		)
		require.LessOrEqual(t, userSessionTemp1.Expires.Sub(userSessionTemp2.Expires), time.Second)
		require.LessOrEqual(
			t,
			userSessionTemp1.DeletedAt.Sub(userSessionTemp2.DeletedAt),
			time.Second,
		)

		err = <-userSession.Errors()
		require.ErrorContains(t, err, "no such host")
	})
}
