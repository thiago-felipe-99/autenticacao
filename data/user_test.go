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

func createUserWithRoles(roles []string) model.User {
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
		DeletedBy: model.EmptyID,
	}
}

func createUser() model.User {
	qtRoles := gofakeit.Number(10, 20)
	roles := make([]string, 0, qtRoles)

	for i := 0; i < qtRoles; i++ {
		roles = append(roles, gofakeit.Name())
	}

	return createUserWithRoles(roles)
}

func TestUserCreate(t *testing.T) {
	t.Parallel()

	qtUser := 100

	user := data.NewUserSQL(createTempDB(t, "data_user_create"))

	for i := 0; i < qtUser; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			err := user.Create(createUser())
			require.NoError(t, err)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		err := user.Create(createUser())
		require.ErrorContains(t, err, "no such host")
	})
}

func checkUser(t *testing.T, expected, found model.User) {
	t.Helper()

	require.Equal(t, expected.ID, found.ID)
	require.Equal(t, expected.Name, found.Name)
	require.Equal(t, expected.Username, found.Username)
	require.Equal(t, expected.Email, found.Email)
	require.Equal(t, expected.Password, found.Password)
	require.Equal(t, expected.Roles, found.Roles)
	require.Equal(t, expected.IsActive, found.IsActive)
	require.LessOrEqual(t, expected.CreatedAt.Sub(found.CreatedAt), time.Second)
	require.Equal(t, expected.CreatedBy, found.CreatedBy)
	require.LessOrEqual(t, expected.DeletedAt.Sub(found.DeletedAt), time.Second)
	require.Equal(t, expected.DeletedBy, found.DeletedBy)
}

func TestGetBy(t *testing.T) {
	t.Parallel()

	qtUsers := 100

	user := data.NewUserSQL(createTempDB(t, "data_role_get_by"))

	for i := 0; i < qtUsers; i++ {
		t.Run("InvalidInput", func(t *testing.T) {
			t.Parallel()

			tempUser := createUser()

			err := user.Create(tempUser)
			require.NoError(t, err)

			found, err := user.GetByID(tempUser.ID)
			require.NoError(t, err)
			checkUser(t, tempUser, found)

			found, err = user.GetByUsername(tempUser.Username)
			require.NoError(t, err)
			checkUser(t, tempUser, found)

			found, err = user.GetByEmail(tempUser.Email)
			require.NoError(t, err)
			checkUser(t, tempUser, found)
		})
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		found, err := user.GetByID(model.NewID())
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, found, model.EmptyUser)

		found, err = user.GetByUsername("invalid-username")
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, found, model.EmptyUser)

		found, err = user.GetByEmail("invalid-email")
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, found, model.EmptyUser)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		found, err := user.GetByID(model.NewID())
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.Equal(t, found, model.EmptyUser)

		found, err = user.GetByUsername(gofakeit.Username())
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.Equal(t, found, model.EmptyUser)

		found, err = user.GetByEmail(gofakeit.Email())
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.Equal(t, found, model.EmptyUser)
	})
}

func TestUserGetAll(t *testing.T) { //nolint:dupl
	t.Parallel()

	qtUsers := 100
	createdUsers := make([]model.User, 0, qtUsers)

	user := data.NewUserSQL(createTempDB(t, "data_user_get_all"))

	users, err := user.GetAll(0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, users, model.EmptyUsers)

	for i := 0; i < qtUsers; i++ {
		tempUser := createUser()
		createdUsers = append(createdUsers, tempUser)

		err := user.Create(tempUser)
		require.NoError(t, err)
	}

	users, err = user.GetAll(0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, len(users), qtUsers)

	id := model.NewID()

	for _, createdUser := range createdUsers {
		index := slices.IndexFunc(users, func(userID model.User) bool {
			return userID.ID == createdUser.ID
		})
		require.GreaterOrEqual(t, index, 0)
		checkUser(t, createdUser, users[index])

		err := user.Delete(createdUser.ID, time.Now(), id)
		require.NoError(t, err)
	}

	users, err = user.GetAll(0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, len(users), qtUsers)

	for _, createdUser := range createdUsers {
		index := slices.IndexFunc(users, func(userID model.User) bool {
			return userID.ID == createdUser.ID
		})
		require.GreaterOrEqual(t, index, 0)

		createdUser.DeletedBy = id
		createdUser.DeletedAt = time.Now()

		checkUser(t, createdUser, users[index])
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		roles, err := user.GetAll(0, qtUsers)
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, roles, model.EmptyUsers)
	})
}

func TestUserGetByRoles(t *testing.T) {
	t.Parallel()

	qtUsers := 100
	createdUsers := make([]model.User, 0, qtUsers)

	user := data.NewUserSQL(createTempDB(t, "data_user_get_by_roles"))

	users, err := user.GetAll(0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, users, model.EmptyUsers)

	qtRoles := gofakeit.Number(10, 20)
	roles := make([]string, 0, qtRoles)

	for i := 0; i < qtRoles; i++ {
		roles = append(roles, gofakeit.Name())
	}

	for i := 0; i < qtUsers; i++ {
		tempUser := createUserWithRoles(roles)
		createdUsers = append(createdUsers, tempUser)

		err := user.Create(tempUser)
		require.NoError(t, err)
	}

	users, err = user.GetByRoles(roles, 0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, len(users), qtUsers)

	id := model.NewID()

	for _, createdUser := range createdUsers {
		index := slices.IndexFunc(users, func(userID model.User) bool {
			return userID.ID == createdUser.ID
		})
		require.GreaterOrEqual(t, index, 0)
		checkUser(t, createdUser, users[index])

		err := user.Delete(createdUser.ID, time.Now(), id)
		require.NoError(t, err)
	}

	users, err = user.GetByRoles(roles, 0, qtUsers)
	require.NoError(t, err)
	require.Equal(t, len(users), qtUsers)

	for _, createdUser := range createdUsers {
		index := slices.IndexFunc(users, func(userID model.User) bool {
			return userID.ID == createdUser.ID
		})
		require.GreaterOrEqual(t, index, 0)

		createdUser.DeletedBy = id
		createdUser.DeletedAt = time.Now()

		checkUser(t, createdUser, users[index])
	}

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		roles, err := user.GetByRoles(roles, 0, qtUsers)
		require.ErrorContains(t, err, "no such host")
		require.Equal(t, roles, model.EmptyUsers)
	})
}

func TestUserDelete(t *testing.T) {
	t.Parallel()

	qtUsers := 100

	user := data.NewUserSQL(createTempDB(t, "data_user_delete"))

	for i := 0; i < qtUsers; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			tempUser := createUser()

			err := user.Create(tempUser)
			require.NoError(t, err)

			err = user.Delete(tempUser.ID, time.Now(), model.NewID())
			require.NoError(t, err)

			found, err := user.GetByID(tempUser.ID)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, found, model.EmptyUser)

			found, err = user.GetByUsername(tempUser.Username)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, found, model.EmptyUser)

			found, err = user.GetByEmail(tempUser.Email)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, found, model.EmptyUser)
		})
	}

	t.Run("InvalidID", func(t *testing.T) {
		t.Parallel()

		err := user.Delete(model.NewID(), time.Now(), model.NewID())
		require.NoError(t, err)
	})

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		err := user.Delete(model.NewID(), time.Now(), model.NewID())
		require.ErrorContains(t, err, "no such host")
	})
}

func TestUserUpdate(t *testing.T) {
	t.Parallel()

	qtUsers := 100

	user := data.NewUserSQL(createTempDB(t, "data_user_update"))

	for i := 0; i < qtUsers; i++ {
		t.Run("ValidInputs", func(t *testing.T) {
			t.Parallel()

			tempUser := createUser()

			err := user.Create(tempUser)
			require.NoError(t, err)

			qtRoles := gofakeit.Number(10, 20)
			roles := make([]string, 0, qtRoles)

			for i := 0; i < qtRoles; i++ {
				roles = append(roles, gofakeit.Name())
			}

			tempUser.Name = gofakeit.Name()
			tempUser.Username = gofakeit.Username()
			tempUser.Email = gofakeit.Email()
			tempUser.Roles = roles
			tempUser.IsActive = false
			tempUser.Password = gofakeit.Password(
				true,
				true,
				true,
				true,
				true,
				gofakeit.Number(10, 255),
			)

			err = user.Update(tempUser)
			require.NoError(t, err)

			foundRole, err := user.GetByID(tempUser.ID)
			require.NoError(t, err)
			checkUser(t, tempUser, foundRole)

			foundRole, err = user.GetByUsername(tempUser.Username)
			require.NoError(t, err)
			checkUser(t, tempUser, foundRole)

			foundRole, err = user.GetByEmail(tempUser.Email)
			require.NoError(t, err)
			checkUser(t, tempUser, foundRole)
		})
	}

	t.Run("InvalidID", func(t *testing.T) {
		t.Parallel()

		err := user.Update(createUser())
		require.NoError(t, err)
	})

	t.Run("WrongDB", func(t *testing.T) {
		t.Parallel()

		user := data.NewUserSQL(createWrongDB(t))

		err := user.Update(createUser())
		require.ErrorContains(t, err, "no such host")
	})
}
