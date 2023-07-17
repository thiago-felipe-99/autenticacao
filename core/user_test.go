package core_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func createTempUser(
	t *testing.T,
	user *core.User,
	db *sqlx.DB,
	validRoles []string,
) (model.ID, model.UserPartial) {
	t.Helper()

	qtRoles := gofakeit.Number(1, 5)
	roles := make([]string, 0, qtRoles)

	for i := 0; i < qtRoles && len(validRoles) > 0; i++ {
		roles = append(roles, gofakeit.RandomString(validRoles))
	}

	id := model.NewID()
	input := model.UserPartial{
		Name:     gofakeit.Name(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
		Roles:    roles,
	}

	err := user.Create(id, input)
	assert.Nil(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DELETE FROM users WHERE username=$1", input.Username)
		assert.Nil(t, err)
	})

	return id, input
}

func TestUserCreate(t *testing.T) {
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

				err := user.Create(model.NewID(), test.input)
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

		err := user.Create(model.NewID(), input)
		assert.ErrorIs(t, err, errs.ErrRoleNotFound)
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()

		_, input := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := input
			input.Email = gofakeit.Email()

			err := user.Create(model.NewID(), input)
			assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
		})

		t.Run("Email", func(t *testing.T) {
			input := input
			input.Username = gofakeit.Username()

			err := user.Create(model.NewID(), input)
			assert.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
		})
	})
}
