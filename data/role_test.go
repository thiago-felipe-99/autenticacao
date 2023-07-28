package data_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"golang.org/x/exp/slices"
)

func createRole() model.Role {
	return model.Role{
		Name:      gofakeit.Name(),
		CreatedAt: time.Now(),
		CreatedBy: model.NewID(),
		DeletedAt: time.Time{},
		DeletedBy: model.EmptyID,
	}
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_create"))

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			err := role.Create(createRole())
			require.NoError(t, err)
		})
	}
}

func checkRole(t *testing.T, expected, found model.Role) {
	t.Helper()

	require.Equal(t, expected.Name, found.Name)
	require.LessOrEqual(t, expected.CreatedAt.Sub(found.CreatedAt), time.Second)
	require.Equal(t, expected.CreatedBy, found.CreatedBy)
	require.LessOrEqual(t, expected.DeletedAt.Sub(found.DeletedAt), time.Second)
	require.Equal(t, expected.DeletedBy, found.DeletedBy)
}

func TestRoleGetByName(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_get_by_name"))

	for i := 0; i < qtRoles; i++ {
		t.Run("InvalidInput", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			require.NoError(t, err)

			found, err := role.GetByName(tempRole.Name)
			require.NoError(t, err)
			checkRole(t, tempRole, found)
		})
	}

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		found, err := role.GetByName("invalid-role")
		require.ErrorIs(t, err, errs.ErrRoleNotFound)
		require.Equal(t, found, model.EmptyRole)
	})
}

func TestRoleExist(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_exist"))

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidInput", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			require.NoError(t, err)

			found, err := role.Exist([]string{tempRole.Name})
			require.NoError(t, err)
			require.True(t, found)
		})
	}

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		found, err := role.Exist([]string{"invalid-role"})
		require.NoError(t, err)
		require.False(t, found)
	})

	t.Run("Multiples", func(t *testing.T) {
		t.Parallel()

		qtRoles := gofakeit.Number(10, 100)
		roles := make([]string, 0, qtRoles)

		for i := 0; i < qtRoles; i++ {
			tempRole := createRole()

			err := role.Create(tempRole)
			require.NoError(t, err)

			roles = append(roles, tempRole.Name)
		}

		found, err := role.Exist(roles)
		require.NoError(t, err)
		require.True(t, found)
	})

	t.Run("MultiplesNotFound", func(t *testing.T) {
		t.Parallel()

		qtRoles := gofakeit.Number(10, 100)
		roles := make([]string, 0, qtRoles)

		for i := 0; i < qtRoles; i++ {
			tempRole := createRole()

			err := role.Create(tempRole)
			require.NoError(t, err)

			roles = append(roles, tempRole.Name)
		}

		found, err := role.Exist(append(roles, gofakeit.Name()))
		require.NoError(t, err)
		require.False(t, found)

		for i := 0; i < gofakeit.Number(10, 100); i++ {
			roles = append(roles, gofakeit.Name())
		}

		found, err = role.Exist(append(roles, gofakeit.Name()))
		require.NoError(t, err)
		require.False(t, found)
	})
}

func TestRoleGetAll(t *testing.T) { //nolint:dupl
	t.Parallel()

	qtRoles := 100
	createdRoles := make([]model.Role, 0, qtRoles)

	role := data.NewRoleSQL(createTempDB(t, "data_role_get_all"))

	roles, err := role.GetAll(0, qtRoles)
	require.NoError(t, err)
	require.Equal(t, roles, model.EmptyRoles)

	for i := 0; i < qtRoles; i++ {
		tempRole := createRole()
		createdRoles = append(createdRoles, tempRole)

		err := role.Create(tempRole)
		require.NoError(t, err)
	}

	roles, err = role.GetAll(0, qtRoles)
	require.NoError(t, err)
	require.Equal(t, len(roles), qtRoles)

	id := model.NewID()

	for _, createdRole := range createdRoles {
		index := slices.IndexFunc(roles, func(role model.Role) bool {
			return role.Name == createdRole.Name
		})
		require.GreaterOrEqual(t, index, 0)
		checkRole(t, createdRole, roles[index])

		err := role.Delete(createdRole.Name, time.Now(), id)
		require.NoError(t, err)
	}

	roles, err = role.GetAll(0, qtRoles)
	require.NoError(t, err)
	require.Equal(t, len(roles), qtRoles)

	for _, createdRole := range createdRoles {
		index := slices.IndexFunc(roles, func(role model.Role) bool {
			return role.Name == createdRole.Name
		})
		require.GreaterOrEqual(t, index, 0)

		createdRole.DeletedBy = id
		createdRole.DeletedAt = time.Now()

		checkRole(t, createdRole, roles[index])
	}
}

func TestRoleDelete(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_delete"))

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			require.NoError(t, err)

			err = role.Delete(tempRole.Name, time.Now(), model.NewID())
			require.NoError(t, err)

			foundRole, err := role.GetByName(tempRole.Name)
			require.ErrorIs(t, err, errs.ErrRoleNotFound)
			require.Equal(t, foundRole, model.EmptyRole)

			found, err := role.Exist([]string{tempRole.Name})
			require.NoError(t, err)
			require.False(t, found)
		})
	}

	t.Run("InvalidName", func(t *testing.T) {
		t.Parallel()

		err := role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		require.NoError(t, err)
	})

	t.Run("TableRoleDoesNotExist", func(t *testing.T) {
		t.Parallel()

		db := createTempDB(t, "data_role_delete_no_role")
		role := data.NewRoleSQL(db)

		_, err := db.Exec("DROP TABLE role")
		require.NoError(t, err)

		err = role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		require.ErrorContains(t, err, "relation \"role\" does not exist")
	})

	t.Run("TableUsersDoesNotExist", func(t *testing.T) {
		t.Parallel()

		db := createTempDB(t, "data_role_delete_no_role")
		role := data.NewRoleSQL(db)

		_, err := db.Exec(
			"DROP TABLE users_sessions_created; DROP TABLE users_sessions_deleted; DROP TABLE users",
		)
		require.NoError(t, err)

		err = role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		require.ErrorContains(t, err, "relation \"users\" does not exist")
	})
}

func TestRoleWrongDB(t *testing.T) {
	t.Parallel()

	role := data.NewRoleSQL(createWrongDB(t))

	err := role.Create(createRole())
	require.ErrorContains(t, err, "no such host")

	roleTemp, err := role.GetByName("invalid-role")
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, roleTemp, model.EmptyRole)

	found, err := role.Exist([]string{"invalid-role"})
	require.ErrorContains(t, err, "no such host")
	require.False(t, found)

	roles, err := role.GetAll(0, 100)
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, roles, model.EmptyRoles)

	err = role.Delete(gofakeit.Name(), time.Now(), model.NewID())
	require.ErrorContains(t, err, "no such host")
}
