package core_test

import (
	"fmt"
	"testing"
	"time"

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

func TestRoleGet(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get")
	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	qtRoles := 100
	roles := make([]string, qtRoles)

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
	}

	for _, roleName := range roles {
		roleName := roleName

		t.Run("Get/"+roleName, func(t *testing.T) {
			t.Parallel()

			roledb, err := role.GetByName(roleName)
			assert.Nil(t, err)

			assert.Equal(t, roleName, roledb.Name)
			assert.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second)
			assert.True(t, time.Time{}.Equal(roledb.DeletedAt))
			assert.Equal(t, model.ID{}, roledb.DeletedBy)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		rolesdb, err := role.GetAll(0, qtRoles)
		assert.Nil(t, err)

		assert.Equal(t, qtRoles, len(rolesdb))

		for _, roledb := range rolesdb {
			assert.True(t, slices.Contains(roles, roledb.Name))
			assert.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second)
			assert.True(t, time.Time{}.Equal(roledb.DeletedAt))
			assert.Equal(t, model.ID{}, roledb.DeletedBy)
		}
	})

	t.Run("GetAll/NoResult", func(t *testing.T) {
		t.Parallel()

		rolesdb, err := role.GetAll(qtRoles, qtRoles)
		assert.Nil(t, err)
		assert.Equal(t, []model.Role{}, rolesdb)
	})
}

func TestRoleDelete(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get")
	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	qtRoles := 100
	rolesName := make([]string, qtRoles)
	rolesID := make([]model.ID, qtRoles)

	for i := range rolesName {
		_, roleTemp := createTempRole(t, role, db)

		id := model.NewID()
		rolesID[i] = id
		rolesName[i] = roleTemp.Name

		err := role.Delete(id, roleTemp.Name)
		assert.Nil(t, err)
	}

	for i, roleName := range rolesName {
		roleName := roleName
		i := i

		t.Run("Get/"+roleName, func(t *testing.T) {
			t.Parallel()

			roledb, err := role.GetByName(roleName)
			assert.Nil(t, err)

			assert.Equal(t, roleName, roledb.Name)
			assert.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second)
			assert.LessOrEqual(t, time.Since(roledb.DeletedAt), time.Second)
			assert.Equal(t, rolesID[i], roledb.DeletedBy)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		rolesdb, err := role.GetAll(0, qtRoles)
		assert.Nil(t, err)

		assert.Equal(t, qtRoles, len(rolesdb))

		for _, roledb := range rolesdb {
			assert.True(t, slices.Contains(rolesName, roledb.Name))
			assert.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second)
			assert.LessOrEqual(t, time.Since(roledb.DeletedAt), time.Second)
			assert.True(t, slices.Contains(rolesID, roledb.DeletedBy))
		}
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < qtRoles; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				err := role.Delete(model.NewID(), gofakeit.Name())
				assert.ErrorIs(t, err, errs.ErrRoleNotFound)
			})
		}
	})
}
