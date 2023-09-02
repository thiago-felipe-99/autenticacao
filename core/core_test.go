package core_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func createTempDB(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	url := "postgres://postgres:postgres@localhost:5432/?sslmode=disable"
	dbname := "test_" + name + "_" + strings.ToLower(gofakeit.LetterN(20))
	urldb := url + "&dbname=" + dbname

	db, err := sqlx.Connect("postgres", url)
	require.NoError(t, err)

	_, err = db.Exec("CREATE DATABASE " + dbname)
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)

	migrations, err := migrate.New("file://../data/migrations", urldb)
	require.NoError(t, err)

	err = migrations.Up()
	require.NoError(t, err)

	sourceerr, err := migrations.Close()
	require.NoError(t, sourceerr)
	require.NoError(t, err)

	db, err = sqlx.Connect("postgres", urldb)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = db.Close()
		require.NoError(t, err)

		// não é possível fazer drop do database conectado, precisa entrar no database padrão
		db, err = sqlx.Connect("postgres", url)
		require.NoError(t, err)

		_, err = db.Exec("DROP DATABASE " + dbname)
		require.NoError(t, err)

		err = db.Close()
		require.NoError(t, err)
	})

	return db
}

func createWrongDB(t *testing.T) *sqlx.DB {
	t.Helper()

	url := "postgres://wrong:wrong@wrong:5432/?sslmode=disable"

	db, err := sqlx.Open("postgres", url)
	require.NoError(t, err)

	return db
}

func boolPointer(b bool) *bool {
	return &b
}

func TestValidate(t *testing.T) {
	t.Parallel()

	validate := model.Validate()
	errMsgs := []string{
		"Key: 'UserPartial.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		"Key: 'UserPartial.Username' Error:Field validation for 'Username' failed on the 'required' tag",
		"Key: 'UserPartial.Email' Error:Field validation for 'Email' failed on the 'required' tag",
		"Key: 'UserPartial.Password' Error:Field validation for 'Password' failed on the 'required' tag",
	}

	errOrig := validate.Struct(model.UserPartial{})     //nolint:exhaustruct
	err := core.Validate(validate, model.UserPartial{}) //nolint:exhaustruct

	require.Equal(t, errOrig.Error(), err.Error())
	require.ErrorAs(t, err, &core.InvalidError{})

	validationErrs := core.InvalidError{}

	okay := errors.As(err, &validationErrs)
	if !okay {
		t.Fatal()
	}

	trans, _ := ut.New(en.New()).GetTranslator("en")
	receivedMsg := validationErrs.Translate(trans)

	for _, errMsg := range errMsgs {
		require.Contains(t, receivedMsg, errMsg)
	}

	err = core.Validate(validate, nil)
	require.ErrorIs(t, err, errs.ErrBodyValidate)
}

func TestNewCore(t *testing.T) {
	t.Parallel()

	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})

	Data, err := data.NewDataSQLRedis(createTempDB(t, "data"), redisClient, time.Second, 200, 100)
	require.NoError(t, err)

	Core := core.NewCore(Data, model.Validate(), false, time.Second)
	require.NotNil(t, Core)
	require.NotNil(t, Core.Role)
	require.NotNil(t, Core.User)
	require.NotNil(t, Core.UserSession)
}
