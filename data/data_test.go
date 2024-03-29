package data_test

import (
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/thiago-felipe-99/autenticacao/data"
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

	migrations, err := migrate.New("file://migrations", urldb)
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

func TestNewDataSQLRedis(t *testing.T) {
	t.Parallel()

	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})

	Data, err := data.NewDataSQLRedis(createTempDB(t, "data"), redisClient, time.Second, 200, 100)
	require.NoError(t, err)
	require.NotNil(t, Data)
	require.IsType(t, &data.RoleSQL{}, Data.Role)
	require.IsType(t, &data.UserSQL{}, Data.User)
	require.IsType(t, &data.UserSessionRedis{}, Data.UserSession)
}
