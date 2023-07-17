package core_test

import (
	"log"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"golang.org/x/exp/slices"
)

func createTempUser(
	t *testing.T,
	user *core.User,
	db *sqlx.DB,
	validRoles []string,
) (model.ID, model.ID, model.UserPartial) {
	t.Helper()

	qtRoles := gofakeit.Number(1, 5)
	roles := make([]string, 0, qtRoles)

	for i := 0; i < qtRoles && len(validRoles) > 0; i++ {
		role := gofakeit.RandomString(validRoles)
		for slices.Contains(roles, role) {
			role = gofakeit.RandomString(validRoles)
		}

		roles = append(roles, role)
	}

	id := model.NewID()
	input := model.UserPartial{
		Name:     gofakeit.Name(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
		Roles:    roles,
	}

	userID, err := user.Create(id, input)
	assert.Nil(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DELETE FROM users WHERE username=$1", input.Username)
		assert.Nil(t, err)
	})

	return userID, id, input
}

func TestUserCreate(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New())

	qtRoles := 10
	roles := make([]string, qtRoles)

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
	}

	t.Run("ValidInputs", func(t *testing.T) {
		t.Parallel()

		createTempUser(t, user, db, roles)
	})

	t.Run("ValidInputs/EmptyRoles", func(t *testing.T) {
		t.Parallel()

		createTempUser(t, user, db, []string{})
	})

	t.Run("InvalidInputs", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			input model.UserPartial
		}{
			{"EmptyName", model.UserPartial{
				Name:     "",
				Username: gofakeit.Username(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"EmptyUsername", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: "",
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"EmptyEmail", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.Username(),
				Email:    "",
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"EmptyPassword", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.Username(),
				Email:    gofakeit.Email(),
				Password: "",
				Roles:    []string{},
			}},
			{"LongName", model.UserPartial{
				Name:     gofakeit.LetterN(256),
				Username: gofakeit.Username(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"LongUsername", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.LetterN(256),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"LongEmail", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.Username(),
				Email:    gofakeit.LetterN(256) + gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"LongPassword", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.Username(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 300),
				Roles:    []string{},
			}},
			{"InvalidUsername", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.Username() + "%$#¨",
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
			{"InvalidEmail", model.UserPartial{
				Name:     gofakeit.Name(),
				Username: gofakeit.Username(),
				Email:    gofakeit.LetterN(25),
				Password: gofakeit.Password(true, true, true, true, true, 20),
				Roles:    []string{},
			}},
		}

		for _, test := range tests {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				_, err := user.Create(model.NewID(), test.input)
				assert.ErrorAs(t, err, &core.ModelInvalidError{})
			})
		}
	})

	t.Run("InvalidRole", func(t *testing.T) {
		t.Parallel()

		input := model.UserPartial{
			Name:     gofakeit.Name(),
			Username: gofakeit.Username(),
			Email:    gofakeit.Email(),
			Password: gofakeit.Password(true, true, true, true, true, 20),
			Roles:    []string{gofakeit.Name()},
		}

		_, err := user.Create(model.NewID(), input)
		assert.ErrorIs(t, err, errs.ErrRoleNotFound)
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()

		_, _, input := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := input
			input.Email = gofakeit.Email()

			_, err := user.Create(model.NewID(), input)
			assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
		})

		t.Run("Email", func(t *testing.T) {
			input := input
			input.Username = gofakeit.Username()

			_, err := user.Create(model.NewID(), input)
			assert.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
		})
	})
}

type partialUser struct {
	userID    model.ID
	input     model.UserPartial
	createdBy model.ID
}

func assertUser(t *testing.T, partial partialUser, userdb model.User) {
	t.Helper()

	assert.Equal(t, partial.userID, userdb.ID)
	assert.Equal(t, partial.input.Name, userdb.Name)
	assert.Equal(t, partial.input.Username, userdb.Username)
	assert.Equal(t, partial.input.Email, userdb.Email)
	assert.Equal(t, partial.input.Roles, userdb.Roles)
	assert.True(t, userdb.IsActive)
	assert.LessOrEqual(t, time.Since(userdb.CreatedAt), time.Minute)
	assert.Equal(t, partial.createdBy, userdb.CreatedBy)
	assert.True(t, time.Time{}.Equal(userdb.DeletedAt))
	assert.Equal(t, model.ID{}, userdb.DeletedBy)

	match, _, err := argon2id.CheckHash(partial.input.Password, userdb.Password)
	assert.Nil(t, err)
	assert.True(t, match)
}

func TestUserGet(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New())

	qtRoles := 10
	roles := make([]string, qtRoles)
	rolesSum := make(map[string]int, qtRoles)

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
		rolesSum[role.Name] = 0
	}

	qtUsers := 100
	users := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdBy, user := createTempUser(t, user, db, roles)
		users[i] = partialUser{
			userID:    userid,
			input:     user,
			createdBy: createdBy,
		}

		for _, role := range user.Roles {
			rolesSum[role]++
		}
	}

	for _, userTemp := range users {
		userTemp := userTemp

		t.Run("GetByID", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByID(userTemp.userID)
			assert.Nil(t, err)

			assertUser(t, userTemp, *userdb)
		})
	}

	t.Run("GetByID/NotFound", func(t *testing.T) {
		t.Parallel()

		_, err := user.GetByID(model.NewID())
		assert.ErrorIs(t, err, errs.ErrUserNotFoud)
	})

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetAll(0, qtUsers)
		assert.Nil(t, err)

		assert.Equal(t, qtUsers, len(usersdb))

		for _, userdb := range usersdb {
			index := slices.IndexFunc(users, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			assertUser(t, users[index], userdb)
		}
	})

	for _, role := range roles {
		role := role

		t.Run("GetByRole", func(t *testing.T) {
			t.Parallel()

			usersdb, err := user.GetByRole([]string{role}, 0, qtUsers)
			assert.Nil(t, err)

			assert.Equal(t, rolesSum[role], len(usersdb))

			ids := make([]model.ID, 0, len(usersdb))
			for _, user := range usersdb {
				ids = append(ids, user.ID)
			}

			for _, user := range users {
				if slices.Contains(user.input.Roles, role) {
					if !slices.Contains(ids, user.userID) {
						log.Println(user)
					}
				}
			}

			for _, userdb := range usersdb {
				index := slices.IndexFunc(users, func(p partialUser) bool {
					return p.userID == userdb.ID
				})
				assertUser(t, users[index], userdb)
			}
		})
	}

	t.Run("GetByRole/All", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetAll(0, qtUsers)
		assert.Nil(t, err)

		assert.Equal(t, qtUsers, len(usersdb))

		for _, userdb := range usersdb {
			index := slices.IndexFunc(users, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			assertUser(t, users[index], userdb)
		}
	})
}
