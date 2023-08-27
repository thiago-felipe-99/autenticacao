package core_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"golang.org/x/exp/slices"
)

func logErros(t *testing.T, errs <-chan error) {
	t.Helper()

	go func() {
		for err := range errs {
			if assert.ErrorContains(t, err, "database is closed") {
				continue
			}

			require.NoError(t, err)
		}
	}()
}

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

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		model.Validate(),
		time.Second,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)
	logErros(t, userSessionRedis.Errors())

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
					require.ErrorAs(t, err, &core.InvalidError{})
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

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		model.Validate(),
		time.Second,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)
	logErros(t, userSessionRedis.Errors())

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
			require.ErrorIs(t, err, errs.ErrUserSessionNotFound)
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

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		model.Validate(),
		time.Second,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)
	logErros(t, userSessionRedis.Errors())

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
			require.ErrorIs(t, err, errs.ErrUserSessionNotFound)
			require.Equal(t, userSessionTemp, model.EmptyUserSession)
		})
	}
}

func TestUserSessionGetAll(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "user_session_get_all")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		model.Validate(),
		time.Second*10,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)
	logErros(t, userSessionRedis.Errors())

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	qtUsersSessions := overflowBuffer
	usersID := make([]model.ID, 0, qtUsersSessions)
	usersSessionsID := make([]model.ID, 0, qtUsersSessions)

	usersSessionsActive, err := userSession.GetAllActive(0, qtUsersSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsActive)

	usersSessionsInactive, err := userSession.GetAllInactive(0, qtUsersSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsInactive)

	for i := 0; i < qtUsersSessions; i++ {
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

	usersSessionsActive, err = userSession.GetAllActive(0, qtUsersSessions)
	require.NoError(t, err)
	require.Equal(t, qtUsersSessions, len(usersSessionsActive))

	usersSessionsInactive, err = userSession.GetAllInactive(0, qtUsersSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsInactive)

	usersIDDB := make([]model.ID, 0, qtUsersSessions)
	idDB := make([]model.ID, 0, qtUsersSessions)

	for _, userSessionTemp := range usersSessionsActive {
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

	usersSessionsActive, err = userSession.GetAllActive(0, qtUsersSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsActive)

	usersSessionsInactive, err = userSession.GetAllInactive(0, qtUsersSessions)
	require.NoError(t, err)
	require.Equal(t, qtUsersSessions, len(usersSessionsInactive))

	usersIDDB = usersIDDB[:0]
	idDB = idDB[:0]

	for _, userSessionTemp := range usersSessionsInactive {
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

func TestUserSessionGetByID(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "user_session_get_all")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 30
	overflowBuffer := buffer * 10

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)
	userSessionRedis := data.NewUserSessionRedis(redisClient, db, buffer)
	userSession := core.NewUserSession(
		userSessionRedis,
		user,
		model.Validate(),
		time.Second*10,
	)

	err := userSessionRedis.ConsumeQueues(time.Second, buffer/2)
	require.NoError(t, err)
	logErros(t, userSessionRedis.Errors())

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

	usersSessionsActive, err := userSession.GetByUserIDActive(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsActive)

	usersSessionsInactive, err := userSession.GetByUserIDInactive(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsInactive)

	for i := 0; i < qtUserSessions; i++ {
		userSessionTemp, err := userSession.Create(UserSessionPartial)
		require.NoError(t, err)

		userSessionsID = append(userSessionsID, userSessionTemp.ID)
	}

	time.Sleep(time.Second)

	usersSessionsActive, err = userSession.GetByUserIDActive(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, qtUserSessions, len(usersSessionsActive))

	usersSessionsInactive, err = userSession.GetByUserIDInactive(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsInactive)

	usersIDDB := make([]model.ID, 0, qtUserSessions)
	idDB := make([]model.ID, 0, qtUserSessions)

	for _, userSessionTemp := range usersSessionsActive {
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

	usersSessionsActive, err = userSession.GetByUserIDActive(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, model.EmptyUserSessions, usersSessionsActive)

	usersSessionsInactive, err = userSession.GetByUserIDInactive(userID, 0, qtUserSessions)
	require.NoError(t, err)
	require.Equal(t, qtUserSessions, len(usersSessionsInactive))

	usersIDDB = usersIDDB[:0]
	idDB = idDB[:0]

	for _, userSessionTemp := range usersSessionsInactive {
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

func checkUserSessionWrongDB(t *testing.T, userSession *core.UserSession, qtRoles int) {
	t.Helper()

	roles, err := userSession.GetAllActive(0, qtRoles)
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSessions, roles)

	roles, err = userSession.GetAllInactive(0, qtRoles)
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSessions, roles)

	roles, err = userSession.GetByUserIDActive(model.NewID(), 0, qtRoles)
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSessions, roles)

	roles, err = userSession.GetByUserIDInactive(model.NewID(), 0, qtRoles)
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSessions, roles)

	role, err := userSession.Delete(model.NewID())
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSession, role)

	role, err = userSession.Refresh(model.NewID())
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSession, role)
}

func TestUserSessionWrongDB(t *testing.T) {
	t.Parallel()

	db := createWrongDB(t)
	dbValid := createTempDB(t, "user_wrong_db")

	role1 := core.NewRole(data.NewRoleSQL(dbValid), model.Validate())
	role2 := core.NewRole(data.NewRoleSQL(db), model.Validate())

	qtRoles := 10
	rolesValid := make([]string, qtRoles)

	for i := range rolesValid {
		_, role := createTempRole(t, role1, dbValid)
		rolesValid[i] = role.Name
	}

	user1 := core.NewUser(data.NewUserSQL(dbValid), role1, model.Validate(), true)
	user2 := core.NewUser(data.NewUserSQL(db), role2, model.Validate(), true)

	inputUser := model.UserPartial{
		Name:     gofakeit.Name(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
		Roles:    rolesValid,
	}

	redisClient1 := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})
	redisClient2 := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "wrong:6379",
		Password: "redis",
		DB:       0,
	})
	buffer := 1
	overflowBuffer := buffer * 10

	userSessionRedis1 := data.NewUserSessionRedis(redisClient1, db, overflowBuffer)
	userSession1 := core.NewUserSession(
		userSessionRedis1,
		user1,
		model.Validate(),
		time.Second,
	)
	err := userSessionRedis1.ConsumeQueues(time.Second, buffer)
	assert.NoError(t, err)

	userSession2 := core.NewUserSession(
		data.NewUserSessionRedis(redisClient2, db, overflowBuffer),
		user1,
		model.Validate(),
		time.Second,
	)

	userSession3 := core.NewUserSession(
		data.NewUserSessionRedis(redisClient2, db, overflowBuffer),
		user2,
		model.Validate(),
		time.Second,
	)

	input := model.UserSessionPartial{ //nolint:exhaustruct
		Username: inputUser.Username,
		Password: inputUser.Password,
	}

	userID, err := user1.Create(model.NewID(), inputUser)
	require.NoError(t, err)

	userSessionTemp, err := userSession1.Create(input)
	require.NoError(t, err)
	require.Equal(t, userID, userSessionTemp.UserID)
	require.LessOrEqual(t, time.Since(userSessionTemp.CreateaAt), time.Second)
	require.Equal(t, time.Time{}, userSessionTemp.DeletedAt)

	errCh := <-userSessionRedis1.Errors()
	assert.ErrorContains(t, errCh, "no such host")

	userSessionTemp, err = userSession2.Create(input)
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSession, userSessionTemp)

	userSessionTemp, err = userSession3.Create(input)
	require.ErrorContains(t, err, "no such host")
	assert.Equal(t, model.EmptyUserSession, userSessionTemp)

	checkUserSessionWrongDB(t, userSession2, qtRoles)
	checkUserSessionWrongDB(t, userSession3, qtRoles)
}
