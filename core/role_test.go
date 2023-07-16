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

func createTempRole(t *testing.T, role *core.Role, db *sqlx.DB) (model.ID, model.RolePartial) {
	t.Helper()

	id := model.NewID()
	input := model.RolePartial{Name: gofakeit.Name()}

	err := role.Create(id, input)
	assert.Nil(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DELETE FROM role WHERE name=$1", input.Name)
		assert.Nil(t, err)
	})

	return id, input
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_create")
	role := core.NewRole(data.NewRoleSQL(db), validator.New())

	t.Run("InvalidInputs", func(t *testing.T) {
		t.Parallel()

		id := model.NewID()

		tests := []struct {
			name  string
			input model.RolePartial
		}{
			{"EmptyName", model.RolePartial{Name: ""}},
			{"LongName", model.RolePartial{Name: gofakeit.LetterN(256)}},
		}

		for _, test := range tests {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				err := role.Create(id, test.input)
				assert.ErrorAs(t, err, &core.ModelInvalidError{})
			})
		}
	})

	t.Run("DuplicateRole", func(t *testing.T) {
		t.Parallel()

		id, input := createTempRole(t, role, db)

		err := role.Create(id, input)
		assert.ErrorIs(t, err, errs.ErrRoleAlreadyExist)
	})

	t.Run("ValidInputs", func(t *testing.T) {
		t.Parallel()

		createTempRole(t, role, db)
	})

	t.Run("DuplicateCreatedBy", func(t *testing.T) {
		t.Parallel()

		id, _ := createTempRole(t, role, db)
		input := model.RolePartial{Name: gofakeit.Name()}

		err := role.Create(id, input)
		assert.Nil(t, err)

		t.Cleanup(func() {
			_, err = db.Exec("DELETE FROM role WHERE name=$1", input.Name)
			assert.Nil(t, err)
		})
	})
}
