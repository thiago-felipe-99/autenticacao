package data_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func createRole() model.Role {
	return model.Role{
		Name:      gofakeit.Name(),
		CreatedAt: time.Now(),
		CreatedBy: model.NewID(),
		DeletedAt: time.Time{},
		DeletedBy: model.NewID(),
	}
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_create_role"))

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			assert.NoError(t, err)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		role := data.NewRoleSQL(createWrongDB(t))

		err := role.Create(createRole())
		assert.ErrorContains(t, err, "no such host")
	})
}

func TestRoleGetByName(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_get_role"))

	for i := 0; i < qtRoles; i++ {
		t.Run("InvalidInput", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			assert.NoError(t, err)

			found, err := role.GetByName(tempRole.Name)
			assert.NoError(t, err)
			assert.Equal(t, found.Name, tempRole.Name)
			assert.LessOrEqual(t, found.CreatedAt.Sub(tempRole.CreatedAt), time.Second)
			assert.Equal(t, found.CreatedBy, tempRole.CreatedBy)
			assert.True(t, found.DeletedAt.Equal(tempRole.DeletedAt))
			assert.Equal(t, found.DeletedBy, tempRole.DeletedBy)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		role := data.NewRoleSQL(createWrongDB(t))

		found, err := role.GetByName("invalid-role")
		assert.ErrorContains(t, err, "no such host")
		assert.Nil(t, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		found, err := role.GetByName("invalid-role")
		assert.ErrorIs(t, err, errs.ErrRoleNotFound)
		assert.Nil(t, found)
	})
}
