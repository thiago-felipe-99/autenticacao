package core_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DELETE FROM role WHERE name=$1", input.Name)
		require.NoError(t, err)
	})

	return id, input
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_create")
	role := core.NewRole(data.NewRoleSQL(db), validator.New())

	t.Run("ValidInputs", func(t *testing.T) {
		t.Parallel()

		createTempRole(t, role, db)
	})

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
				require.ErrorAs(t, err, &core.InvalidError{})
			})
		}
	})

	t.Run("DuplicateRole", func(t *testing.T) {
		t.Parallel()

		id, input := createTempRole(t, role, db)

		err := role.Create(id, input)
		require.ErrorIs(t, err, errs.ErrRoleAlreadyExist)
	})

	t.Run("DuplicateCreatedBy", func(t *testing.T) {
		t.Parallel()

		id, _ := createTempRole(t, role, db)
		input := model.RolePartial{Name: gofakeit.Name()}

		err := role.Create(id, input)
		require.NoError(t, err)

		t.Cleanup(func() {
			_, err = db.Exec("DELETE FROM role WHERE name=$1", input.Name)
			require.NoError(t, err)
		})
	})
}

func TestRoleGet(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get")
	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	qtRoles := 100
	rolesTmp := make([]string, qtRoles)

	for i := range rolesTmp {
		_, role := createTempRole(t, role, db)
		rolesTmp[i] = role.Name
	}

	for _, roleTmp := range rolesTmp {
		roleTmp := roleTmp

		t.Run("Get", func(t *testing.T) {
			t.Parallel()

			roledb, err := role.GetByName(roleTmp)
			require.NoError(t, err)

			require.Equal(t, roleTmp, roledb.Name)
			require.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second)
			require.True(t, time.Time{}.Equal(roledb.DeletedAt))
			require.Equal(t, model.EmptyID, roledb.DeletedBy)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		rolesdb, err := role.GetAll(0, qtRoles)
		require.NoError(t, err)

		require.Equal(t, qtRoles, len(rolesdb))

		for _, roledb := range rolesdb {
			require.True(t, slices.Contains(rolesTmp, roledb.Name))
			require.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second)
			require.True(t, time.Time{}.Equal(roledb.DeletedAt))
			require.Equal(t, model.EmptyID, roledb.DeletedBy)
		}
	})

	t.Run("GetAll/NoResult", func(t *testing.T) {
		t.Parallel()

		rolesdb, err := role.GetAll(qtRoles, qtRoles)
		require.NoError(t, err)
		require.Equal(t, model.EmptyRoles, rolesdb)
	})
}

func TestRoleDelete(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_delete")
	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	qtRoles := 100
	rolesTmp := make([]string, qtRoles)
	rolesID := make([]model.ID, qtRoles)

	for i := range rolesTmp {
		_, roleTemp := createTempRole(t, role, db)

		id := model.NewID()
		rolesID[i] = id
		rolesTmp[i] = roleTemp.Name

		err := role.Delete(id, roleTemp.Name)
		require.NoError(t, err)
	}

	for _, roleTmp := range rolesTmp {
		roleTmp := roleTmp

		t.Run("GetByName", func(t *testing.T) {
			t.Parallel()

			found, err := role.GetByName(roleTmp)
			require.ErrorIs(t, err, errs.ErrRoleNotFound)
			require.Equal(t, found, model.EmptyRole)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		rolesdb, err := role.GetAll(0, qtRoles)
		require.NoError(t, err)

		require.Equal(t, qtRoles, len(rolesdb))

		for _, roledb := range rolesdb {
			require.True(t, slices.Contains(rolesTmp, roledb.Name))
			require.LessOrEqual(t, time.Since(roledb.CreatedAt), time.Second*2)
			require.LessOrEqual(t, time.Since(roledb.DeletedAt), time.Second*2)
			require.True(t, slices.Contains(rolesID, roledb.DeletedBy))
		}
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < qtRoles; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				err := role.Delete(model.NewID(), gofakeit.Name())
				require.ErrorIs(t, err, errs.ErrRoleNotFound)
			})
		}
	})
}

func TestRoleWrongDB(t *testing.T) {
	t.Parallel()

	db := createWrongDB(t)
	role := core.NewRole(data.NewRoleSQL(db), validator.New())

	err := role.Create(model.NewID(), model.RolePartial{Name: gofakeit.Name()})
	require.ErrorContains(t, err, "no such host")

	roles, err := role.GetAll(0, 100)
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyRoles, roles)

	roleTemp, err := role.GetByName(gofakeit.Name())
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyRole, roleTemp)

	exist, err := role.Exist([]string{gofakeit.Name()})
	require.ErrorContains(t, err, "no such host")
	require.False(t, exist)

	err = role.Delete(model.NewID(), gofakeit.Name())
	require.ErrorContains(t, err, "no such host")
}
