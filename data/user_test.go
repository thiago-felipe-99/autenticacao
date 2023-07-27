package data_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func createUser() model.User {
	qtRoles := gofakeit.Number(10, 20)
	roles := make([]string, 0, qtRoles)

	for i := 0; i < qtRoles; i++ {
		roles = append(roles, gofakeit.Name())
	}

	return model.User{
		ID:        model.NewID(),
		Name:      gofakeit.Name(),
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		Password:  gofakeit.Password(true, true, true, true, true, gofakeit.Number(10, 255)),
		Roles:     roles,
		IsActive:  true,
		CreatedAt: time.Now(),
		CreatedBy: model.NewID(),
		DeletedAt: time.Time{},
		DeletedBy: model.ID{},
	}
}

func TestUserCreate(t *testing.T) {
	t.Parallel()

	qtUser := 100

	user := data.NewUserSQL(createTempDB(t, "data_user_create"))

	for i := 0; i < qtUser; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			err := user.Create(createUser())
			assert.NoError(t, err)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		err := user.Create(createUser())
		assert.ErrorContains(t, err, "no such host")
	})
}
