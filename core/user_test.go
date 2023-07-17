package core_test

import (
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
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

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New())

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
}

func assertUser(t *testing.T, partial partialUser, userdb model.User) {
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

	match, _, err := argon2id.CheckHash(partial.input.Password, userdb.Password)
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestUserGet(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New())

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
	users := make([]partialUser, qtUsers)

	for i := 0; i < qtUsers; i++ {
		userid, createdBy, user := createTempUser(t, user, db, roles)
		users[i] = partialUser{
			userID:    userid,
			input:     user,
			createdBy: createdBy,
		}

		for _, role := range user.Roles {
			rolesSum[role]++
		}

		if slices.Contains(user.Roles, roles[0]) {
			for _, role := range user.Roles {
				multiplesRolesSum[roles[0]][role]++
			}
		}

		if slices.Contains(user.Roles, roles[5]) {
			for _, role := range user.Roles {
				multiplesRolesSum[roles[5]][role]++
			}
		}
	}

	for _, userTemp := range users {
		userTemp := userTemp

		t.Run("GetByID", func(t *testing.T) {
			t.Parallel()

			userdb, err := user.GetByID(userTemp.userID)
			assert.NoError(t, err)

			assertUser(t, userTemp, *userdb)
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
			index := slices.IndexFunc(users, func(p partialUser) bool {
				return p.userID == userdb.ID
			})

			assertUser(t, users[index], userdb)
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
				index := slices.IndexFunc(users, func(p partialUser) bool {
					return p.userID == userdb.ID
				})
				assertUser(t, users[index], userdb)
			}
		})
	}

	t.Run("GetByRole/Multiples", func(t *testing.T) {
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
							index := slices.IndexFunc(users, func(p partialUser) bool {
								return p.userID == userdb.ID
							})
							assertUser(t, users[index], userdb)
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
	userTmp model.UserPartial,
	userid model.ID,
	user *core.User,
) {
	t.Helper()

	userdb, err := user.GetByID(userid)
	assert.NoError(t, err)

	if update.Name != "" {
		assert.Equal(t, update.Name, userdb.Name)
	} else {
		assert.Equal(t, userTmp.Name, userdb.Name)
	}

	if update.Username != "" {
		assert.Equal(t, update.Username, userdb.Username)
	} else {
		assert.Equal(t, userTmp.Username, userdb.Username)
	}

	if update.Email != "" {
		assert.Equal(t, update.Email, userdb.Email)
	} else {
		assert.Equal(t, userTmp.Email, userdb.Email)
	}

	if update.Roles != nil {
		assert.Equal(t, update.Roles, userdb.Roles)
	} else {
		assert.Equal(t, userTmp.Roles, userdb.Roles)
	}

	if update.IsActive != nil {
		assert.Equal(t, *update.IsActive, userdb.IsActive)
	}

	if update.Password != "" {
		match, _, err := argon2id.CheckHash(update.Password, userdb.Password)
		assert.NoError(t, err)
		assert.True(t, match)
	} else {
		match, _, err := argon2id.CheckHash(userTmp.Password, userdb.Password)
		assert.NoError(t, err)
		assert.True(t, match)
	}
}

func TestUserUpdate(t *testing.T) {
	t.Parallel()

	db := createTempDB(t, "role_get")

	role := core.NewRole(data.NewRoleSQL(db), validator.New())
	user := core.NewUser(data.NewUserSQL(db), role, validator.New())

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

				userid, _, userTmp := createTempUser(t, user, db, roles)

				err := user.Update(userid, test.input)
				assert.NoError(t, err)

				assertUserUpdate(t, test.input, userTmp, userid, user)
			})
		}
	})

	t.Run("ValidInputs/OnlyRole", func(t *testing.T) {
		t.Parallel()

		userid, _, userTmp := createTempUser(t, user, db, roles)

		update := model.UserUpdate{Roles: randomSliceString(roles)} //nolint:exhaustruct

		err := user.Update(userid, update)
		assert.NoError(t, err)

		assertUserUpdate(t, update, userTmp, userid, user)
	})

	t.Run("ValidInputs/All", func(t *testing.T) {
		t.Parallel()

		userid, _, userTmp := createTempUser(t, user, db, roles)

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

		assertUserUpdate(t, update, userTmp, userid, user)
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
		_, _, userTmp2 := createTempUser(t, user, db, roles)

		t.Run("Username", func(t *testing.T) {
			input := model.UserUpdate{Username: userTmp2.Username} //nolint:exhaustruct
			err := user.Update(id1, input)
			assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExist)
		})

		t.Run("Email", func(t *testing.T) {
			input := model.UserUpdate{Email: userTmp2.Email} //nolint:exhaustruct
			err := user.Update(id1, input)
			assert.ErrorIs(t, err, errs.ErrEmailAlreadyExist)
		})
	})
}
