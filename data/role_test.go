package data_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
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
		DeletedBy: model.NewID(),
	}
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_create"))

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

	role := data.NewRoleSQL(createTempDB(t, "data_role_get_by_name"))

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

func TestRoleExist(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_exist"))

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidInput", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			assert.NoError(t, err)

			found, err := role.Exist([]string{tempRole.Name})
			assert.NoError(t, err)
			assert.True(t, found)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		role := data.NewRoleSQL(createWrongDB(t))

		found, err := role.Exist([]string{"invalid-role"})
		assert.ErrorContains(t, err, "no such host")
		assert.False(t, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		found, err := role.Exist([]string{"invalid-role"})
		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("Multiples", func(t *testing.T) {
		t.Parallel()

		qtRoles := gofakeit.Number(10, 100)
		roles := make([]string, 0, qtRoles)

		for i := 0; i < qtRoles; i++ {
			tempRole := createRole()

			err := role.Create(tempRole)
			assert.NoError(t, err)

			roles = append(roles, tempRole.Name)
		}

		found, err := role.Exist(roles)
		assert.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("MultiplesNotFound", func(t *testing.T) {
		t.Parallel()

		qtRoles := gofakeit.Number(10, 100)
		roles := make([]string, 0, qtRoles)

		for i := 0; i < qtRoles; i++ {
			tempRole := createRole()

			err := role.Create(tempRole)
			assert.NoError(t, err)

			roles = append(roles, tempRole.Name)
		}

		found, err := role.Exist(append(roles, gofakeit.Name()))
		assert.NoError(t, err)
		assert.False(t, found)

		for i := 0; i < gofakeit.Number(10, 100); i++ {
			roles = append(roles, gofakeit.Name())
		}

		found, err = role.Exist(append(roles, gofakeit.Name()))
		assert.NoError(t, err)
		assert.False(t, found)
	})
}

func TestRoleGetAll(t *testing.T) {
	t.Parallel()

	qtRoles := 100
	createRoles := make([]model.Role, 0, qtRoles)
	rolesName := make([]string, 0, qtRoles)

	role := data.NewRoleSQL(createTempDB(t, "data_role_get_all"))

	roles, err := role.GetAll(0, qtRoles)
	assert.NoError(t, err)
	assert.Equal(t, roles, []model.Role{})

	for i := 0; i < qtRoles; i++ {
		tempRole := createRole()
		createRoles = append(createRoles, tempRole)

		err := role.Create(tempRole)
		assert.NoError(t, err)
	}

	roles, err = role.GetAll(0, qtRoles)
	assert.NoError(t, err)
	assert.Equal(t, len(roles), qtRoles)

	id := model.NewID()

	for _, roleDB := range roles {
		rolesName = append(rolesName, roleDB.Name)

		err := role.Delete(roleDB.Name, time.Now(), id)
		assert.NoError(t, err)
	}

	for _, createRole := range createRoles {
		assert.True(t, slices.Contains(rolesName, createRole.Name))
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		role := data.NewRoleSQL(createWrongDB(t))

		roles, err := role.GetAll(0, qtRoles)
		assert.ErrorContains(t, err, "no such host")
		assert.Nil(t, roles)
	})
}

func TestRoleDelete(t *testing.T) {
	t.Parallel()

	qtRoles := 100

	role := data.NewRoleSQL(createTempDB(t, "data_role_delete"))

	for i := 0; i < qtRoles; i++ {
		t.Run("ValidNames", func(t *testing.T) {
			t.Parallel()

			tempRole := createRole()

			err := role.Create(tempRole)
			assert.NoError(t, err)

			err = role.Delete(tempRole.Name, time.Now(), model.NewID())
			assert.NoError(t, err)

			foundRole, err := role.GetByName(tempRole.Name)
			assert.ErrorIs(t, err, errs.ErrRoleNotFound)
			assert.Nil(t, foundRole)

			found, err := role.Exist([]string{tempRole.Name})
			assert.NoError(t, err)
			assert.False(t, found)
		})
	}

	t.Run("InvalidName", func(t *testing.T) {
		t.Parallel()

		err := role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		assert.NoError(t, err)
	})

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		role := data.NewRoleSQL(createWrongDB(t))

		err := role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		assert.ErrorContains(t, err, "no such host")
	})

	t.Run("TableRoleDoesNotExist", func(t *testing.T) {
		t.Parallel()

		db := createTempDB(t, "data_role_delete_no_role")
		role := data.NewRoleSQL(db)

		_, err := db.Exec("DROP TABLE role")
		assert.NoError(t, err)

		err = role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		assert.ErrorContains(t, err, "relation \"role\" does not exist")
	})

	t.Run("TableUsersDoesNotExist", func(t *testing.T) {
		t.Parallel()

		db := createTempDB(t, "data_role_delete_no_role")
		role := data.NewRoleSQL(db)

		_, err := db.Exec(
			"DROP TABLE users_sessions_created; DROP TABLE users_sessions_deleted; DROP TABLE users",
		)
		assert.NoError(t, err)

		err = role.Delete(gofakeit.Name(), time.Now(), model.NewID())
		assert.ErrorContains(t, err, "relation \"users\" does not exist")
	})
}
