package core_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"golang.org/x/exp/slices"
)

var invalidUserPartialInputs = []struct { //nolint:gochecknoglobals
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
		Password: gofakeit.Password(true, true, true, true, true, 256),
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

//nolint:gochecknoglobals,exhaustruct
var invalidUserUpdateInputs = []struct {
	name  string
	input model.UserUpdate
}{
	{"LongName", model.UserUpdate{Name: gofakeit.LetterN(256)}},
	{"LongUsername", model.UserUpdate{Username: gofakeit.LetterN(256)}},
	{"LongEmail", model.UserUpdate{Email: gofakeit.LetterN(256) + gofakeit.Email()}},
	{
		"LongPassword",
		model.UserUpdate{Password: gofakeit.Password(true, true, true, true, true, 300)},
	},
	{"InvalidUsername", model.UserUpdate{Username: gofakeit.Username() + "%$#¨"}},
	{"InvalidEmail", model.UserUpdate{Email: gofakeit.LetterN(25)}},
}

//nolint:gochecknoglobals,exhaustruct
var validUserUpdateInputs = []struct {
	name  string
	input model.UserUpdate
}{
	{"OnlyName", model.UserUpdate{Name: gofakeit.Name()}},
	{"OnlyUsername", model.UserUpdate{Username: gofakeit.Username()}},
	{"OnlyEmail", model.UserUpdate{Email: gofakeit.Email()}},
	{
		"OnlyPassword",
		model.UserUpdate{Password: gofakeit.Password(true, true, true, true, true, 20)},
	},
	{"OnlyIsActive", model.UserUpdate{IsActive: boolPointer(false)}},
}

func randomSliceString(valids []string) []string {
	qt := gofakeit.Number(1, 5)
	slice := make([]string, 0, qt)

	for i := 0; i < qt && len(valids) > 0; i++ {
		element := gofakeit.RandomString(valids)
		for slices.Contains(slice, element) {
			element = gofakeit.RandomString(valids)
		}

		slice = append(slice, element)
	}

	return slice
}

func createTempUser(
	t *testing.T,
	user *core.User,
	db *sqlx.DB,
	validRoles []string,
) (model.ID, model.ID, model.UserPartial) {
	t.Helper()

	roles := randomSliceString(validRoles)

	id := model.NewID()
	input := model.UserPartial{
		Name:     gofakeit.Name(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
		Roles:    roles,
	}

	userID, err := user.Create(id, input)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DELETE FROM users WHERE username=$1", input.Username)
		require.NoError(t, err)
	})

	return userID, id, input
}

func TestUserCreate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_create")

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)

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

		for _, test := range invalidUserPartialInputs {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				userCreated, err := user.Create(model.NewID(), test.input)
				require.ErrorAs(t, err, &core.InvalidError{})
				require.Equal(t, userCreated, model.EmptyID)
			})
		}
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		t.Parallel()

		input := model.UserPartial{
			Name:     gofakeit.Name(),
			Username: gofakeit.Username(),
			Email:    gofakeit.Email(),
			Password: gofakeit.Password(true, true, true, true, true, 20),
			Roles:    []string{gofakeit.Name()},
		}

		userCreated, err := user.Create(model.NewID(), input)
		require.ErrorIs(t, err, errs.ErrRoleNotFound)
		require.Equal(t, userCreated, model.EmptyID)
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()

		_, _, input := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := input
			input.Email = gofakeit.Email()

			userCreated, err := user.Create(model.NewID(), input)
			require.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
			require.Equal(t, userCreated, model.EmptyID)
		})

		t.Run("Email", func(t *testing.T) {
			input := input
			input.Username = gofakeit.Username()

			userCreated, err := user.Create(model.NewID(), input)
			require.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
			require.Equal(t, userCreated, model.EmptyID)
		})
	})
}

type partialUser struct {
	userID    model.ID
	input     model.UserPartial
	createdBy model.ID
	deletedBy model.ID
}

func requireUser(t *testing.T, partial partialUser, userdb model.User, user *core.User) {
	t.Helper()

	require.Equal(t, partial.userID, userdb.ID)
	require.Equal(t, partial.input.Name, userdb.Name)
	require.Equal(t, partial.input.Username, userdb.Username)
	require.Equal(t, partial.input.Email, userdb.Email)
	require.Equal(t, partial.input.Roles, userdb.Roles)
	require.True(t, userdb.IsActive)
	require.LessOrEqual(t, time.Since(userdb.CreatedAt), time.Second*2)
	require.Equal(t, partial.createdBy, userdb.CreatedBy)
	require.True(t, time.Time{}.Equal(userdb.DeletedAt))
	require.Equal(t, model.EmptyID, userdb.DeletedBy)

	match, err := user.EqualPassword(partial.input.Password, userdb.Password)
	require.NoError(t, err)
	require.True(t, match)
}

func TestUserGet(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "user_get")

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)

	qtRoles := 10
	roles := make([]string, qtRoles)
	rolesSum := make(map[string]int, qtRoles)
	multiplesRolesSum := map[string]map[string]int{}

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
	}

	multiplesRolesSum[roles[0]] = map[string]int{}
	multiplesRolesSum[roles[5]] = map[string]int{}

	for _, role := range roles {
		rolesSum[role] = 0
		multiplesRolesSum[roles[0]][role] = 0
		multiplesRolesSum[roles[5]][role] = 0
	}

	qtUsers := 100
	usersTemp := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdBy, userTemp := createTempUser(t, user, db, roles)
		usersTemp[i] = partialUser{
			userID:    userid,
			input:     userTemp,
			createdBy: createdBy,
			deletedBy: model.NewID(),
		}

		for _, role := range userTemp.Roles {
			rolesSum[role]++
		}

		if slices.Contains(userTemp.Roles, roles[0]) {
			for _, role := range userTemp.Roles {
				multiplesRolesSum[roles[0]][role]++
			}
		}

		if slices.Contains(userTemp.Roles, roles[5]) {
			for _, role := range userTemp.Roles {
				multiplesRolesSum[roles[5]][role]++
			}
		}
	}

	for _, userTemp := range usersTemp {
		userTemp := userTemp

		t.Run("GetByID", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByID(userTemp.userID)
			require.NoError(t, err)

			requireUser(t, userTemp, userdb, user)
		})

		t.Run("GetByUsername", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByUsername(userTemp.input.Username)
			require.NoError(t, err)

			requireUser(t, userTemp, userdb, user)
		})

		t.Run("GetByEmail", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByEmail(userTemp.input.Email)
			require.NoError(t, err)

			requireUser(t, userTemp, userdb, user)
		})
	}

	t.Run("GetByID/NotFound", func(t *testing.T) {
		t.Parallel()

		userFound, err := user.GetByID(model.NewID())
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.Equal(t, userFound, model.EmptyUser)
	})

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetAll(0, qtUsers)
		require.NoError(t, err)

		require.Equal(t, qtUsers, len(usersdb))

		for _, userdb := range usersdb {
			index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			requireUser(t, usersTemp[index], userdb, user)
		}
	})

	for role, sum := range rolesSum {
		role := role
		sum := sum

		t.Run("GetByRole", func(t *testing.T) {
			t.Parallel()

			usersdb, err := user.GetByRole([]string{role}, 0, qtUsers)
			require.NoError(t, err)

			require.Equal(t, sum, len(usersdb))

			for _, userdb := range usersdb {
				index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
					return p.userID == userdb.ID
				})
				requireUser(t, usersTemp[index], userdb, user)
			}
		})
	}

	t.Run("GetByRole/Multiples", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		for role1, rolesSum := range multiplesRolesSum {
			role1 := role1
			rolesSum := rolesSum

			t.Run(role1, func(t *testing.T) {
				t.Parallel()

				for role2, sum := range rolesSum {
					role2 := role2
					sum := sum

					t.Run(role2, func(t *testing.T) {
						usersdb, err := user.GetByRole([]string{role1, role2}, 0, qtUsers)
						require.NoError(t, err)

						require.Equal(t, sum, len(usersdb))

						for _, userdb := range usersdb {
							index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
								return p.userID == userdb.ID
							})
							requireUser(t, usersTemp[index], userdb, user)
						}
					})
				}
			})
		}
	})

	t.Run("GetByRole/RoleNotFound", func(t *testing.T) {
		t.Parallel()

		userFound, err := user.GetByRole([]string{gofakeit.Name()}, 0, qtUsers)
		require.ErrorIs(t, err, errs.ErrRoleNotFound)
		require.Equal(t, userFound, model.EmptyUsers)
	})
}

func requireUserUpdate(
	t *testing.T,
	update model.UserUpdate,
	userTemp model.UserPartial,
	userid model.ID,
	user *core.User,
) {
	t.Helper()

	userdb, err := user.GetByID(userid)
	require.NoError(t, err)

	if update.Name != "" {
		require.Equal(t, update.Name, userdb.Name)
	} else {
		require.Equal(t, userTemp.Name, userdb.Name)
	}

	if update.Username != "" {
		require.Equal(t, update.Username, userdb.Username)
	} else {
		require.Equal(t, userTemp.Username, userdb.Username)
	}

	if update.Email != "" {
		require.Equal(t, update.Email, userdb.Email)
	} else {
		require.Equal(t, userTemp.Email, userdb.Email)
	}

	if update.Roles != nil {
		require.Equal(t, update.Roles, userdb.Roles)
	} else {
		require.Equal(t, userTemp.Roles, userdb.Roles)
	}

	if update.IsActive != nil {
		require.Equal(t, *update.IsActive, userdb.IsActive)
	}

	if update.Password != "" {
		match, err := user.EqualPassword(update.Password, userdb.Password)
		require.NoError(t, err)
		require.True(t, match)
	} else {
		match, err := user.EqualPassword(userTemp.Password, userdb.Password)
		require.NoError(t, err)
		require.True(t, match)
	}
}

func TestUserUpdate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_update")

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)

	qtRoles := 10
	roles := make([]string, qtRoles)

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
	}

	t.Run("ValidInputs", func(t *testing.T) {
		t.Parallel()

		for _, test := range validUserUpdateInputs {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				userid, _, userTemp := createTempUser(t, user, db, roles)

				err := user.Update(userid, test.input)
				require.NoError(t, err)

				requireUserUpdate(t, test.input, userTemp, userid, user)
			})
		}
	})

	t.Run("ValidInputs/OnlyRole", func(t *testing.T) {
		t.Parallel()

		userid, _, userTemp := createTempUser(t, user, db, roles)

		update := model.UserUpdate{Roles: randomSliceString(roles)} //nolint:exhaustruct

		err := user.Update(userid, update)
		require.NoError(t, err)

		requireUserUpdate(t, update, userTemp, userid, user)
	})

	t.Run("ValidInputs/All", func(t *testing.T) {
		t.Parallel()

		userid, _, userTemp := createTempUser(t, user, db, roles)

		update := model.UserUpdate{
			Name:     gofakeit.Name(),
			Username: gofakeit.Username(),
			Email:    gofakeit.Email(),
			Password: gofakeit.Password(true, true, true, true, true, 20),
			Roles:    randomSliceString(roles),
			IsActive: boolPointer(false),
		}

		err := user.Update(userid, update)
		require.NoError(t, err)

		requireUserUpdate(t, update, userTemp, userid, user)
	})

	t.Run("InvalidInputs", func(t *testing.T) {
		t.Parallel()

		for _, test := range invalidUserUpdateInputs {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				err := user.Update(model.NewID(), test.input)
				require.ErrorAs(t, err, &core.InvalidError{})
			})
		}
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		t.Parallel()

		userid, _, _ := createTempUser(t, user, db, roles)
		input := model.UserUpdate{Roles: []string{gofakeit.Name()}} //nolint:exhaustruct
		err := user.Update(userid, input)
		require.ErrorIs(t, err, errs.ErrRoleNotFound)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()

		input := model.UserUpdate{Name: gofakeit.Name()} //nolint:exhaustruct
		err := user.Update(model.NewID(), input)
		require.ErrorIs(t, err, errs.ErrUserNotFound)
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()

		id1, _, _ := createTempUser(t, user, db, roles)
		_, _, userTemp2 := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := model.UserUpdate{Username: userTemp2.Username} //nolint:exhaustruct
			err := user.Update(id1, input)
			require.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
		})

		t.Run("Email", func(t *testing.T) {
			input := model.UserUpdate{Email: userTemp2.Email} //nolint:exhaustruct
			err := user.Update(id1, input)
			require.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
		})
	})
}

func requireUserDelete(t *testing.T, partial partialUser, userdb model.User, user *core.User) {
	t.Helper()

	require.Equal(t, partial.userID, userdb.ID)
	require.Equal(t, partial.input.Name, userdb.Name)
	require.Equal(t, partial.input.Username, userdb.Username)
	require.Equal(t, partial.input.Email, userdb.Email)
	require.Equal(t, partial.input.Roles, userdb.Roles)
	require.True(t, userdb.IsActive)
	require.LessOrEqual(t, time.Since(userdb.CreatedAt), time.Minute)
	require.Equal(t, partial.createdBy, userdb.CreatedBy)
	require.LessOrEqual(t, time.Since(userdb.DeletedAt), time.Minute)
	require.Equal(t, partial.deletedBy, userdb.DeletedBy)

	match, err := user.EqualPassword(partial.input.Password, userdb.Password)
	require.NoError(t, err)
	require.True(t, match)
}

func TestUserDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "user_get")

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), false)

	qtRoles := 10
	roles := make([]string, qtRoles)
	rolesSum := make(map[string]int, qtRoles)
	multiplesRolesSum := map[string]map[string]int{}

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
	}

	multiplesRolesSum[roles[0]] = map[string]int{}
	multiplesRolesSum[roles[5]] = map[string]int{}

	for _, role := range roles {
		rolesSum[role] = 0
		multiplesRolesSum[roles[0]][role] = 0
		multiplesRolesSum[roles[5]][role] = 0
	}

	qtUsers := 100
	usersTemp := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdby, userTemp := createTempUser(t, user, db, roles)

		deletedby := model.NewID()

		err := user.Delete(userid, deletedby)
		require.NoError(t, err)

		usersTemp[i] = partialUser{
			userID:    userid,
			input:     userTemp,
			createdBy: createdby,
			deletedBy: deletedby,
		}

		for _, role := range userTemp.Roles {
			rolesSum[role]++
		}

		if slices.Contains(userTemp.Roles, roles[0]) {
			for _, role := range userTemp.Roles {
				multiplesRolesSum[roles[0]][role]++
			}
		}

		if slices.Contains(userTemp.Roles, roles[5]) {
			for _, role := range userTemp.Roles {
				multiplesRolesSum[roles[5]][role]++
			}
		}
	}

	for _, userTemp := range usersTemp {
		userTemp := userTemp

		t.Run("GetByID", func(t *testing.T) {
			t.Parallel()

			userFound, err := user.GetByID(userTemp.userID)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, userFound, model.EmptyUser)
		})

		t.Run("GetByUsername", func(t *testing.T) {
			t.Parallel()

			userFound, err := user.GetByUsername(userTemp.input.Username)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, userFound, model.EmptyUser)
		})

		t.Run("GetByEmail", func(t *testing.T) {
			t.Parallel()

			userDound, err := user.GetByEmail(userTemp.input.Email)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, userDound, model.EmptyUser)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetAll(0, qtUsers)
		require.NoError(t, err)

		require.Equal(t, qtUsers, len(usersdb))

		for _, userdb := range usersdb {
			index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			requireUserDelete(t, usersTemp[index], userdb, user)
		}
	})

	for role, sum := range rolesSum {
		role := role
		sum := sum

		t.Run("GetByRole", func(t *testing.T) {
			t.Parallel()

			usersdb, err := user.GetByRole([]string{role}, 0, qtUsers)
			require.NoError(t, err)

			require.Equal(t, sum, len(usersdb))

			for _, userdb := range usersdb {
				index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
					return p.userID == userdb.ID
				})
				requireUserDelete(t, usersTemp[index], userdb, user)
			}
		})
	}

	t.Run("GetByRole/NotExist", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetByRole([]string{gofakeit.Name()}, 0, qtRoles)
		require.ErrorIs(t, err, errs.ErrRoleNotFound)
		require.Equal(t, model.EmptyUsers, usersdb)
	})

	t.Run("GetByRole/Multiples", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		for role1, rolesSum := range multiplesRolesSum {
			role1 := role1
			rolesSum := rolesSum

			t.Run(role1, func(t *testing.T) {
				t.Parallel()

				for role2, sum := range rolesSum {
					role2 := role2
					sum := sum

					t.Run(role2, func(t *testing.T) {
						usersdb, err := user.GetByRole([]string{role1, role2}, 0, qtUsers)
						require.NoError(t, err)

						require.Equal(t, sum, len(usersdb))

						for _, userdb := range usersdb {
							index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
								return p.userID == userdb.ID
							})
							requireUserDelete(t, usersTemp[index], userdb, user)
						}
					})
				}
			})
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()

		err := user.Delete(model.NewID(), model.NewID())
		require.ErrorIs(t, err, errs.ErrUserNotFound)
	})
}

func TestUserWithArgon(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "user_get_argon")

	role := core.NewRole(data.NewRoleSQL(db), model.Validate())
	user := core.NewUser(data.NewUserSQL(db), role, model.Validate(), true)

	qtRoles := 5
	roles := make([]string, qtRoles)
	qtUsers := 5

	for i := range roles {
		_, role := createTempRole(t, role, db)
		roles[i] = role.Name
	}

	for i := 0; i < qtUsers; i++ {
		t.Run("Argon", func(t *testing.T) {
			t.Parallel()

			userid, createdBy, userTemp := createTempUser(t, user, db, roles)

			partial := partialUser{
				userID:    userid,
				input:     userTemp,
				createdBy: createdBy,
				deletedBy: model.NewID(),
			}

			userdb, err := user.GetByID(userid)
			require.NoError(t, err)
			requireUser(t, partial, userdb, user)

			update := model.UserUpdate{ //nolint:exhaustruct
				Password: gofakeit.Password(true, true, true, true, true, 20),
			}
			err = user.Update(userid, update)
			require.NoError(t, err)
			requireUserUpdate(t, update, userTemp, userid, user)

			err = user.Delete(userid, partial.deletedBy)
			require.NoError(t, err)

			userFound, err := user.GetByID(userid)
			require.ErrorIs(t, err, errs.ErrUserNotFound)
			require.Equal(t, userFound, model.EmptyUser)
		})
	}
}

func checkUserWrongDB(
	t *testing.T,
	user *core.User,
	input model.UserPartial,
	rolesValid []string,
	update model.UserUpdate,
) {
	t.Helper()

	id, err := user.Create(model.NewID(), input)
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyID, id)

	users, err := user.GetAll(0, 100)
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyUsers, users)

	users, err = user.GetByRole(rolesValid[0:2], 0, 100)
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyUsers, users)

	roleTemp, err := user.GetByID(model.NewID())
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyUser, roleTemp)

	roleTemp, err = user.GetByEmail(gofakeit.Email())
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyUser, roleTemp)

	roleTemp, err = user.GetByUsername(gofakeit.Username())
	require.ErrorContains(t, err, "no such host")
	require.Equal(t, model.EmptyUser, roleTemp)

	err = user.Update(model.NewID(), update)
	require.ErrorContains(t, err, "no such host")

	err = user.Delete(model.NewID(), model.NewID())
	require.ErrorContains(t, err, "no such host")

	equal, err := user.EqualPassword(gofakeit.Name(), "invalid-hash")
	require.ErrorContains(t, err, "error comparaing hash")
	require.False(t, equal)
}

func TestUserWrongDB(t *testing.T) {
	t.Parallel()

	db := createWrongDB(t)
	dbValid := createTempDB(t, "user_wrong_db")

	role1 := core.NewRole(data.NewRoleSQL(dbValid), model.Validate())
	role2 := core.NewRole(data.NewRoleSQL(db), model.Validate())

	qtRoles := 10
	rolesValid := make([]string, qtRoles)

	for i := range rolesValid {
		_, role := createTempRole(t, role1, dbValid)
		rolesValid[i] = role.Name
	}

	user1 := core.NewUser(data.NewUserSQL(db), role1, model.Validate(), true)
	user2 := core.NewUser(data.NewUserSQL(db), role2, model.Validate(), true)

	input := model.UserPartial{
		Name:     gofakeit.Name(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
		Roles:    rolesValid,
	}
	update := model.UserUpdate{
		Name:     gofakeit.Name(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
		Roles:    rolesValid,
		IsActive: boolPointer(false),
	}

	checkUserWrongDB(t, user1, input, rolesValid, update)
	checkUserWrongDB(t, user2, input, rolesValid, update)
}
