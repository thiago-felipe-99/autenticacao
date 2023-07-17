package core_test

import (
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
	assert.NoError(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DELETE FROM users WHERE username=$1", input.Username)
		assert.NoError(t, err)
	})

	return userID, id, input
}

func TestUserCreate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_create")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)

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

				_, err := user.Create(model.NewID(), test.input)
				assert.ErrorAs(t, err, &core.ModelInvalidError{})
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

		_, err := user.Create(model.NewID(), input)
		assert.ErrorIs(t, err, errs.ErrRoleNotFound)
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()

		_, _, input := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := input
			input.Email = gofakeit.Email()

			_, err := user.Create(model.NewID(), input)
			assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
		})

		t.Run("Email", func(t *testing.T) {
			input := input
			input.Username = gofakeit.Username()

			_, err := user.Create(model.NewID(), input)
			assert.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
		})
	})
}

type partialUser struct {
	userID    model.ID
	input     model.UserPartial
	createdBy model.ID
	deletedBy model.ID
}

func assertUser(t *testing.T, partial partialUser, userdb model.User, user *core.User) {
	t.Helper()

	assert.Equal(t, partial.userID, userdb.ID)
	assert.Equal(t, partial.input.Name, userdb.Name)
	assert.Equal(t, partial.input.Username, userdb.Username)
	assert.Equal(t, partial.input.Email, userdb.Email)
	assert.Equal(t, partial.input.Roles, userdb.Roles)
	assert.True(t, userdb.IsActive)
	assert.LessOrEqual(t, time.Since(userdb.CreatedAt), time.Minute)
	assert.Equal(t, partial.createdBy, userdb.CreatedBy)
	assert.True(t, time.Time{}.Equal(userdb.DeletedAt))
	assert.Equal(t, model.ID{}, userdb.DeletedBy)

	match, err := user.EqualPassword(partial.input.Password, userdb.Password)
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestUserGet(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)

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
			assert.NoError(t, err)

			assertUser(t, userTemp, *userdb, user)
		})

		t.Run("GetByUsername", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByUsername(userTemp.input.Username)
			assert.NoError(t, err)

			assertUser(t, userTemp, *userdb, user)
		})

		t.Run("GetByEmail", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByEmail(userTemp.input.Email)
			assert.NoError(t, err)

			assertUser(t, userTemp, *userdb, user)
		})
	}

	t.Run("GetByID/NotFound", func(t *testing.T) {
		t.Parallel()

		_, err := user.GetByID(model.NewID())
		assert.ErrorIs(t, err, errs.ErrUserNotFoud)
	})

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetAll(0, qtUsers)
		assert.NoError(t, err)

		assert.Equal(t, qtUsers, len(usersdb))

		for _, userdb := range usersdb {
			index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			assertUser(t, usersTemp[index], userdb, user)
		}
	})

	for role, sum := range rolesSum {
		role := role
		sum := sum

		t.Run("GetByRole", func(t *testing.T) {
			t.Parallel()

			usersdb, err := user.GetByRole([]string{role}, 0, qtUsers)
			assert.NoError(t, err)

			assert.Equal(t, sum, len(usersdb))

			for _, userdb := range usersdb {
				index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
					return p.userID == userdb.ID
				})
				assertUser(t, usersTemp[index], userdb, user)
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
						assert.NoError(t, err)

						assert.Equal(t, sum, len(usersdb))

						for _, userdb := range usersdb {
							index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
								return p.userID == userdb.ID
							})
							assertUser(t, usersTemp[index], userdb, user)
						}
					})
				}
			})
		}
	})

	t.Run("GetByRole/RoleNotFound", func(t *testing.T) {
		t.Parallel()

		_, err := user.GetByRole([]string{gofakeit.Name()}, 0, qtUsers)
		assert.ErrorIs(t, err, errs.ErrRoleNotFound)
	})
}

func assertUserUpdate(
	t *testing.T,
	update model.UserUpdate,
	userTemp model.UserPartial,
	userid model.ID,
	user *core.User,
) {
	t.Helper()

	userdb, err := user.GetByID(userid)
	assert.NoError(t, err)

	if update.Name != "" {
		assert.Equal(t, update.Name, userdb.Name)
	} else {
		assert.Equal(t, userTemp.Name, userdb.Name)
	}

	if update.Username != "" {
		assert.Equal(t, update.Username, userdb.Username)
	} else {
		assert.Equal(t, userTemp.Username, userdb.Username)
	}

	if update.Email != "" {
		assert.Equal(t, update.Email, userdb.Email)
	} else {
		assert.Equal(t, userTemp.Email, userdb.Email)
	}

	if update.Roles != nil {
		assert.Equal(t, update.Roles, userdb.Roles)
	} else {
		assert.Equal(t, userTemp.Roles, userdb.Roles)
	}

	if update.IsActive != nil {
		assert.Equal(t, *update.IsActive, userdb.IsActive)
	}

	if update.Password != "" {
		match, err := user.EqualPassword(update.Password, userdb.Password)
		assert.NoError(t, err)
		assert.True(t, match)
	} else {
		match, err := user.EqualPassword(userTemp.Password, userdb.Password)
		assert.NoError(t, err)
		assert.True(t, match)
	}
}

func TestUserUpdate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_update")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)

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
				assert.NoError(t, err)

				assertUserUpdate(t, test.input, userTemp, userid, user)
			})
		}
	})

	t.Run("ValidInputs/OnlyRole", func(t *testing.T) {
		t.Parallel()

		userid, _, userTemp := createTempUser(t, user, db, roles)

		update := model.UserUpdate{Roles: randomSliceString(roles)} //nolint:exhaustruct

		err := user.Update(userid, update)
		assert.NoError(t, err)

		assertUserUpdate(t, update, userTemp, userid, user)
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
		assert.NoError(t, err)

		assertUserUpdate(t, update, userTemp, userid, user)
	})

	t.Run("InvalidInputs", func(t *testing.T) {
		t.Parallel()

		for _, test := range invalidUserUpdateInputs {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				err := user.Update(model.NewID(), test.input)
				assert.ErrorAs(t, err, &core.ModelInvalidError{})
			})
		}
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		t.Parallel()

		userid, _, _ := createTempUser(t, user, db, roles)
		input := model.UserUpdate{Roles: []string{gofakeit.Name()}} //nolint:exhaustruct
		err := user.Update(userid, input)
		assert.ErrorIs(t, err, errs.ErrRoleNotFound)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()

		input := model.UserUpdate{Name: gofakeit.Name()} //nolint:exhaustruct
		err := user.Update(model.NewID(), input)
		assert.ErrorIs(t, err, errs.ErrUserNotFoud)
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()

		id1, _, _ := createTempUser(t, user, db, roles)
		_, _, userTemp2 := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := model.UserUpdate{Username: userTemp2.Username} //nolint:exhaustruct
			err := user.Update(id1, input)
			assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
		})

		t.Run("Email", func(t *testing.T) {
			input := model.UserUpdate{Email: userTemp2.Email} //nolint:exhaustruct
			err := user.Update(id1, input)
			assert.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
		})
	})
}

func assertUserDelete(t *testing.T, partial partialUser, userdb model.User, user *core.User) {
	t.Helper()

	assert.Equal(t, partial.userID, userdb.ID)
	assert.Equal(t, partial.input.Name, userdb.Name)
	assert.Equal(t, partial.input.Username, userdb.Username)
	assert.Equal(t, partial.input.Email, userdb.Email)
	assert.Equal(t, partial.input.Roles, userdb.Roles)
	assert.True(t, userdb.IsActive)
	assert.LessOrEqual(t, time.Since(userdb.CreatedAt), time.Minute)
	assert.Equal(t, partial.createdBy, userdb.CreatedBy)
	assert.LessOrEqual(t, time.Since(userdb.DeletedAt), time.Minute)
	assert.Equal(t, partial.deletedBy, userdb.DeletedBy)

	match, err := user.EqualPassword(partial.input.Password, userdb.Password)
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestUserDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), false)

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
		assert.NoError(t, err)

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

			_, err := user.GetByID(userTemp.userID)
			assert.ErrorIs(t, err, errs.ErrUserNotFoud)
		})

		t.Run("GetByUsername", func(t *testing.T) {
			t.Parallel()

			_, err := user.GetByUsername(userTemp.input.Username)
			assert.ErrorIs(t, err, errs.ErrUserNotFoud)
		})

		t.Run("GetByEmail", func(t *testing.T) {
			t.Parallel()

			_, err := user.GetByEmail(userTemp.input.Email)
			assert.ErrorIs(t, err, errs.ErrUserNotFoud)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()

		usersdb, err := user.GetAll(0, qtUsers)
		assert.NoError(t, err)

		assert.Equal(t, qtUsers, len(usersdb))

		for _, userdb := range usersdb {
			index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			assertUserDelete(t, usersTemp[index], userdb, user)
		}
	})

	for role, sum := range rolesSum {
		role := role
		sum := sum

		t.Run("GetByRole", func(t *testing.T) {
			t.Parallel()

			usersdb, err := user.GetByRole([]string{role}, 0, qtUsers)
			assert.NoError(t, err)

			assert.Equal(t, sum, len(usersdb))

			for _, userdb := range usersdb {
				index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
					return p.userID == userdb.ID
				})
				assertUserDelete(t, usersTemp[index], userdb, user)
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
						assert.NoError(t, err)

						assert.Equal(t, sum, len(usersdb))

						for _, userdb := range usersdb {
							index := slices.IndexFunc(usersTemp, func(p partialUser) bool {
								return p.userID == userdb.ID
							})
							assertUserDelete(t, usersTemp[index], userdb, user)
						}
					})
				}
			})
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()

		err := user.Delete(model.NewID(), model.NewID())
		assert.ErrorIs(t, err, errs.ErrUserNotFoud)
	})
}

func TestUserWithArgon(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get_argon")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New(), true)

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
			assert.NoError(t, err)
			assertUser(t, partial, *userdb, user)

			update := model.UserUpdate{ //nolint:exhaustruct
				Password: gofakeit.Password(true, true, true, true, true, 20),
			}
			err = user.Update(userid, update)
			assert.NoError(t, err)
			assertUserUpdate(t, update, userTemp, userid, user)

			err = user.Delete(userid, partial.deletedBy)
			assert.NoError(t, err)

			_, err = user.GetByID(userid)
			assert.ErrorIs(t, err, errs.ErrUserNotFoud)
		})
	}
}
