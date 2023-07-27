package data_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func createRole() model.Role {
	return model.Role{
		Name:      gofakeit.Name(),
		CreatedAt: time.Now(),
		CreatedBy: model.NewID(),
		DeletedAt: gofakeit.FutureDate(),
		DeletedBy: model.NewID(),
	}
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()

	qtRoles := 1000

	role := data.NewRoleSQL(createTempDB(t, "data_create_role"))
	roles := make([]model.Role, 0, qtRoles)

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()
			tempRole := createRole()
			err := role.Create(tempRole)
			assert.NoError(t, err)

			roles = append(roles, tempRole)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		role := data.NewRoleSQL(createWrongDB(t))

		err := role.Create(createRole())
		assert.Error(t, err)
	})
}
