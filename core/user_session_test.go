package core_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"golang.org/x/exp/slices"
)

func TestUserSessionCreate(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "user_session_create")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		validator.New(),
		time.Second,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	qtUsers := overflowBuffer
	usersTemp := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdBy, userTemp := createTempUser(t, user, db, rolesTemp)
		usersTemp[i] = partialUser{
			userID:    userid,
			input:     userTemp,
			createdBy: createdBy,
			deletedBy: model.NewID(),
		}
	}

	for _, userTemp := range usersTemp {
		userID, userInput := userTemp.userID, userTemp.input

		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			//nolint:exhaustruct
			inputs := []struct {
				name  string
				input model.UserSessionPartial
			}{
				{
					"email",
					model.UserSessionPartial{Email: userInput.Email, Password: userInput.Password},
				},
				{
					"username",
					model.UserSessionPartial{
						Username: userInput.Username,
						Password: userInput.Password,
					},
				},
			}

			for _, input := range inputs {
				input := input

				t.Run(input.name, func(t *testing.T) {
					t.Parallel()

					userSessionTemp, err := userSession.Create(input.input)
					require.NoError(t, err)

					require.Equal(t, userID, userSessionTemp.UserID)
					require.LessOrEqual(t, time.Since(userSessionTemp.CreateaAt), time.Second)
					require.Equal(t, time.Time{}, userSessionTemp.DeletedAt)
				})
			}
		})

		t.Run("InvalidInputs", func(t *testing.T) {
			t.Parallel()

			//nolint:exhaustruct
			inputs := []struct {
				name  string
				input model.UserSessionPartial
			}{
				{
					"UsernameWithEmail",
					model.UserSessionPartial{
						Email:    userInput.Email,
						Username: userInput.Email,
						Password: userInput.Password,
					},
				},
				{
					"UsernameWhitoutPassword",
					model.UserSessionPartial{Username: userInput.Username},
				},
				{
					"EmailWhitoutPassword",
					model.UserSessionPartial{Email: userInput.Email},
				},
			}

			for _, input := range inputs {
				input := input

				t.Run(input.name, func(t *testing.T) {
					t.Parallel()

					userSessionCreated, err := userSession.Create(input.input)
					require.ErrorAs(t, err, &core.ModelInvalidError{})
					require.Equal(t, userSessionCreated, model.EmptyUserSession)
				})
			}
		})

		t.Run("WrongPassword", func(t *testing.T) {
			t.Parallel()
			//nolint:exhaustruct
			inputs := []struct {
				name  string
				input model.UserSessionPartial
			}{
				{
					"Username",
					model.UserSessionPartial{
						Username: userInput.Username,
						Password: gofakeit.Password(true, true, true, true, true, 20),
					},
				},
				{
					"Email",
					model.UserSessionPartial{
						Email:    userInput.Email,
						Password: gofakeit.Password(true, true, true, true, true, 20),
					},
				},
			}

			for _, input := range inputs {
				input := input

				t.Run(input.name, func(t *testing.T) {
					t.Parallel()

					userSessionCreated, err := userSession.Create(input.input)
					require.ErrorIs(t, err, errs.ErrPasswordDoesNotMatch)
					require.Equal(t, userSessionCreated, model.EmptyUserSession)
				})
			}
		})

		t.Run("UserNotFound", func(t *testing.T) {
			t.Parallel()
			//nolint:exhaustruct
			inputs := []struct {
				name  string
				input model.UserSessionPartial
			}{
				{
					"Username",
					model.UserSessionPartial{
						Username: gofakeit.Username(),
						Password: gofakeit.Password(true, true, true, true, true, 20),
					},
				},
				{
					"Email",
					model.UserSessionPartial{
						Email:    gofakeit.Email(),
						Password: gofakeit.Password(true, true, true, true, true, 20),
					},
				},
			}

			for _, input := range inputs {
				input := input

				t.Run(input.name, func(t *testing.T) {
					t.Parallel()

					userSessionCreated, err := userSession.Create(input.input)
					require.ErrorIs(t, err, errs.ErrUserNotFound)
					require.Equal(t, userSessionCreated, model.EmptyUserSession)
				})
			}
		})
	}
}

func TestUserSessionRefresh(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_session_refresh")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		validator.New(),
		time.Second,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	qtUsers := overflowBuffer
	usersTemp := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdBy, userTemp := createTempUser(t, user, db, rolesTemp)
		usersTemp[i] = partialUser{
			userID:    userid,
			input:     userTemp,
			createdBy: createdBy,
			deletedBy: model.NewID(),
		}
	}

	for _, userTemp := range usersTemp {
		userID, userInput := userTemp.userID, userTemp.input

		t.Run("ValiInputs", func(t *testing.T) {
			t.Parallel()

			UserSessionPartial := model.UserSessionPartial{ //nolint:exhaustruct
				Username: userInput.Username,
				Password: userInput.Password,
			}

			userSessionTemp1, err := userSession.Create(UserSessionPartial)
			require.NoError(t, err)

			userSessionTemp2, err := userSession.Refresh(userSessionTemp1.ID)
			require.NoError(t, err)

			require.NotEqual(t, userSessionTemp1.ID, userSessionTemp2.ID)
			require.Equal(t, userID, userSessionTemp2.UserID)
			require.LessOrEqual(t, time.Since(userSessionTemp2.CreateaAt), time.Second)
			require.Equal(t, time.Time{}, userSessionTemp2.DeletedAt)
		})

		t.Run("UserSessionNotFound", func(t *testing.T) {
			t.Parallel()

			userSessionTemp, err := userSession.Refresh(model.NewID())
			require.ErrorIs(t, err, errs.ErrUserSessionNotFoud)
			require.Equal(t, userSessionTemp, model.EmptyUserSession)
		})
	}
}

func TestUserSessionDelete(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_session_delete")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		validator.New(),
		time.Second,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	qtUsers := overflowBuffer
	usersTemp := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdBy, userTemp := createTempUser(t, user, db, rolesTemp)
		usersTemp[i] = partialUser{
			userID:    userid,
			input:     userTemp,
			createdBy: createdBy,
			deletedBy: model.NewID(),
		}
	}

	for _, userTemp := range usersTemp {
		userID, userInput := userTemp.userID, userTemp.input

		t.Run("ValiInputs", func(t *testing.T) {
			t.Parallel()

			UserSessionPartial := model.UserSessionPartial{ //nolint:exhaustruct
				Username: userInput.Username,
				Password: userInput.Password,
			}

			userSessionTemp1, err := userSession.Create(UserSessionPartial)
			require.NoError(t, err)

			userSessionTemp2, err := userSession.Delete(userSessionTemp1.ID)
			require.NoError(t, err)

			require.Equal(t, userSessionTemp1.ID, userSessionTemp2.ID)
			require.Equal(t, userID, userSessionTemp2.UserID)
			require.LessOrEqual(t, time.Since(userSessionTemp2.CreateaAt), time.Second)
			require.LessOrEqual(t, time.Since(userSessionTemp2.DeletedAt), time.Second)
		})

		t.Run("UserSessionNotFound", func(t *testing.T) {
			t.Parallel()

			userSessionTemp, err := userSession.Delete(model.NewID())
			require.ErrorIs(t, err, errs.ErrUserSessionNotFoud)
			require.Equal(t, userSessionTemp, model.EmptyUserSession)
		})
	}
}

func TestUserSessionGetAll(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_session_get_all")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		validator.New(),
		time.Second*10,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	qtUsers := overflowBuffer
	usersID := make([]model.ID, 0, qtUsers)
	usersSessionsID := make([]model.ID, 0, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userID, _, userTemp := createTempUser(t, user, db, rolesTemp)
		usersID = append(usersID, userID)
		UserSessionPartial := model.UserSessionPartial{ //nolint:exhaustruct
			Username: userTemp.Username,
			Password: userTemp.Password,
		}

		userSessionTemp, err := userSession.Create(UserSessionPartial)
		require.NoError(t, err)

		usersSessionsID = append(usersSessionsID, userSessionTemp.ID)
	}

	time.Sleep(time.Second)

	usersSessionsDB, err := userSession.GetAll(0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, qtUsers, len(usersSessionsDB))

	usersIDDB := make([]model.ID, 0, qtUsers)
	idDB := make([]model.ID, 0, qtUsers)

	for _, userSessionTemp := range usersSessionsDB {
		usersIDDB = append(usersIDDB, userSessionTemp.UserID)
		idDB = append(idDB, userSessionTemp.ID)
	}

	for _, userID := range usersID {
		require.True(t, slices.Contains(usersIDDB, userID))
	}

	for _, id := range usersSessionsID {
		require.True(t, slices.Contains(idDB, id))

		_, err := userSession.Delete(id)
		require.NoError(t, err)
	}

	time.Sleep(time.Second)

	usersSessionsDB, err = userSession.GetAll(0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, qtUsers, len(usersSessionsDB))

	usersIDDB = make([]model.ID, 0, qtUsers)
	idDB = make([]model.ID, 0, qtUsers)

	for _, userSessionTemp := range usersSessionsDB {
		usersIDDB = append(usersIDDB, userSessionTemp.UserID)
		idDB = append(idDB, userSessionTemp.ID)
	}

	for _, userID := range usersID {
		require.True(t, slices.Contains(usersIDDB, userID))
	}

	for _, id := range usersSessionsID {
		require.True(t, slices.Contains(idDB, id))
	}
}

func TestUserSessionGetByID(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_session_get_all")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		validator.New(),
		time.Second*10,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	qtUserSessions := overflowBuffer
	userSessionsID := make([]model.ID, 0, qtUserSessions)

	userID, _, userTemp := createTempUser(t, user, db, rolesTemp)
	UserSessionPartial := model.UserSessionPartial{ //nolint:exhaustruct
		Username: userTemp.Username,
		Password: userTemp.Password,
	}

	for i := 0; i < qtUserSessions; i++ {
		userSessionTemp, err := userSession.Create(UserSessionPartial)
		require.NoError(t, err)

		userSessionsID = append(userSessionsID, userSessionTemp.ID)
	}

	time.Sleep(time.Second)

	usersSessionsDB, err := userSession.GetByUserID(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, qtUserSessions, len(usersSessionsDB))

	usersIDDB := make([]model.ID, 0, qtUserSessions)
	idDB := make([]model.ID, 0, qtUserSessions)

	for _, userSessionTemp := range usersSessionsDB {
		usersIDDB = append(usersIDDB, userSessionTemp.UserID)
		idDB = append(idDB, userSessionTemp.ID)
	}

	for _, userIDDB := range usersIDDB {
		require.Equal(t, userIDDB, userID)
	}

	for _, id := range userSessionsID {
		require.True(t, slices.Contains(idDB, id))

		_, err := userSession.Delete(id)
		require.NoError(t, err)
	}

	time.Sleep(time.Second)

	usersSessionsDB, err = userSession.GetByUserID(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, qtUserSessions, len(usersSessionsDB))

	usersIDDB = make([]model.ID, 0, qtUserSessions)
	idDB = make([]model.ID, 0, qtUserSessions)

	for _, userSessionTemp := range usersSessionsDB {
		usersIDDB = append(usersIDDB, userSessionTemp.UserID)
		idDB = append(idDB, userSessionTemp.ID)
	}

	for _, userIDDB := range usersIDDB {
		require.Equal(t, userIDDB, userID)
	}

	for _, id := range userSessionsID {
		require.True(t, slices.Contains(idDB, id))
	}
}
