package core_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func TestUserSessionCreate(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "user_session_create")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "qt4BLAnrrNSZp2ssRMkLzZjnaZQkcL22",
		DB:       0,
	})

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSession := core.NewUserSession(
		data.NewUserSessionRedis(redisClient, db, 100),
		user,
		validator.New(),
		time.Second,
	)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	userID, _, userInput := createTempUser(t, user, db, rolesTemp)

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
				assert.NoError(t, err)

				assert.Equal(t, userID, userSessionTemp.UserID)
				assert.LessOrEqual(t, time.Since(userSessionTemp.CreateaAt), time.Second)
				assert.Equal(t, time.Time{}, userSessionTemp.DeletedAt)
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

				_, err := userSession.Create(input.input)
				assert.ErrorAs(t, err, &core.ModelInvalidError{})
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

				_, err := userSession.Create(input.input)
				assert.ErrorIs(t, err, errs.ErrPasswordDoesNotMatch)
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

				_, err := userSession.Create(input.input)
				assert.ErrorIs(t, err, errs.ErrUserNotFound)
			})
		}
	})
}

func TestUserSessionRefresh(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_session_create")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "qt4BLAnrrNSZp2ssRMkLzZjnaZQkcL22",
		DB:       0,
	})

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSession := core.NewUserSession(
		data.NewUserSessionRedis(redisClient, db, 100),
		user,
		validator.New(),
		time.Second,
	)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	userID, _, userInput := createTempUser(t, user, db, rolesTemp)

	t.Run("ValiInputs", func(t *testing.T) {
		t.Parallel()

		UserSessionPartial := model.UserSessionPartial{ //nolint:exhaustruct
			Username: userInput.Username,
			Password: userInput.Password,
		}

		userSessionTemp1, err := userSession.Create(UserSessionPartial)
		assert.NoError(t, err)

		userSessionTemp2, err := userSession.Refresh(userSessionTemp1.ID)
		assert.NoError(t, err)

		assert.NotEqual(t, userSessionTemp1.ID, userSessionTemp2.ID)
		assert.Equal(t, userID, userSessionTemp2.UserID)
		assert.LessOrEqual(t, time.Since(userSessionTemp2.CreateaAt), time.Second)
		assert.Equal(t, time.Time{}, userSessionTemp2.DeletedAt)
	})

	t.Run("UserSessionNotFound", func(t *testing.T) {
		t.Parallel()

		userSessionTemp, err := userSession.Refresh(model.NewID())
		assert.Nil(t, userSessionTemp)
		assert.ErrorIs(t, err, errs.ErrUserSessionNotFoud)
	})
}

func TestUserSessionDelete(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_session_create")
	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "qt4BLAnrrNSZp2ssRMkLzZjnaZQkcL22",
		DB:       0,
	})

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)
	userSession := core.NewUserSession(
		data.NewUserSessionRedis(redisClient, db, 100),
		user,
		validator.New(),
		time.Second,
	)

	qtRoles := 5
	rolesTemp := make([]string, qtRoles)

	for i := range rolesTemp {
		_, role := createTempRole(t, role, db)
		rolesTemp[i] = role.Name
	}

	userID, _, userInput := createTempUser(t, user, db, rolesTemp)

	t.Run("ValiInputs", func(t *testing.T) {
		t.Parallel()

		UserSessionPartial := model.UserSessionPartial{ //nolint:exhaustruct
			Username: userInput.Username,
			Password: userInput.Password,
		}

		userSessionTemp1, err := userSession.Create(UserSessionPartial)
		assert.NoError(t, err)

		userSessionTemp2, err := userSession.Delete(userSessionTemp1.ID)
		assert.NoError(t, err)

		assert.Equal(t, userSessionTemp1.ID, userSessionTemp2.ID)
		assert.Equal(t, userID, userSessionTemp2.UserID)
		assert.LessOrEqual(t, time.Since(userSessionTemp2.CreateaAt), time.Second)
		assert.LessOrEqual(t, time.Since(userSessionTemp2.DeletedAt), time.Second)
	})

	t.Run("UserSessionNotFound", func(t *testing.T) {
		t.Parallel()

		userSessionTemp, err := userSession.Delete(model.NewID())
		assert.Nil(t, userSessionTemp)
		assert.ErrorIs(t, err, errs.ErrUserSessionNotFoud)
	})
}
