package main

import (
	"errors"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/server"
)

func noError(err error, msg string) {
	if err != nil {
		log.Panicf("[ERROR] - %s: %s", msg, err)
	}
}

func main() {
	url := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	validate := validator.New()

	migrations, err := migrate.New("file://data/migrations", url)
	noError(err, "Error starting migrations")

	err = migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		noError(err, "Error migrating database")
	}

	sourcerr, err := migrations.Close()
	noError(err, "Error closing migration")
	noError(sourcerr, "Error in migration source")

	db, err := sqlx.Connect("postgres", url)
	noError(err, "Error opening database")

	redisClient := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       0,
	})

	data, err := data.NewDataSQLRedis(db, redisClient, time.Hour, 2000, 1000) //nolint:gomnd
	noError(err, "Error starting data")

	cores := core.NewCore(data, validate, true, time.Hour)

	server, err := server.CreateHTTPServer(validate, cores)
	noError(err, "Error creating server")

	err = server.Listen(":8080")
	noError(err, "Error listening server")
}
