package core_test

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func createTempDB(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	url := "postgres://postgres:postgres@localhost:5432/?sslmode=disable"
	dbname := "test_" + name + "_" + strings.ToLower(gofakeit.LetterN(20))
	urldb := url + "&dbname=" + dbname

	db, err := sqlx.Connect("postgres", url)
	assert.Nil(t, err)

	_, err = db.Exec("CREATE DATABASE " + dbname)
	assert.Nil(t, err)

	err = db.Close()
	assert.Nil(t, err)

	migrations, err := migrate.New("file://../data/migrations", urldb)
	assert.Nil(t, err)

	err = migrations.Up()
	assert.Nil(t, err)

	sourceerr, err := migrations.Close()
	assert.Nil(t, sourceerr)
	assert.Nil(t, err)

	db, err = sqlx.Connect("postgres", urldb)
	assert.Nil(t, err)

	t.Cleanup(func() {
		err = db.Close()
		assert.Nil(t, err)

		// não é possível fazer drop do database conectado, precisa entrar no database padrão
		db, err = sqlx.Connect("postgres", url)
		assert.Nil(t, err)

		_, err = db.Exec("DROP DATABASE " + dbname)
		assert.Nil(t, err)

		err = db.Close()
		assert.Nil(t, err)
	})

	return db
}
