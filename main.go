package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/thiago-felipe-99/autenticacao/core"
	"github.com/thiago-felipe-99/autenticacao/data"
	"github.com/thiago-felipe-99/autenticacao/errs"
	"github.com/thiago-felipe-99/autenticacao/model"
	"github.com/thiago-felipe-99/autenticacao/server"
)

func createFirst(configurations *configurations, cores *core.Cores) error {
	roleAdmin := model.RolePartial{Name: configurations.Role.Name}

	err := cores.Role.Create(model.EmptyID, roleAdmin)
	if err != nil {
		if errors.Is(err, errs.ErrRoleAlreadyExist) {
			log.Printf("[INFO] - Role '%s' already exist", roleAdmin.Name)
		} else {
			return err
		}
	} else {
		log.Printf("[INFO] - Role '%s' created", roleAdmin.Name)
	}

	userAdmin := model.UserPartial{
		Name:     configurations.User.Name,
		Username: configurations.User.Username,
		Email:    configurations.User.Email,
		Password: configurations.User.Password,
		Roles:    []string{roleAdmin.Name},
	}

	_, err = cores.User.Create(model.EmptyID, userAdmin)
	if err != nil {
		if errors.Is(err, errs.ErrUsernameAlreadyExist) { //nolint:gocritic
			log.Printf("[INFO] - User with username '%s' already exist", userAdmin.Username)
		} else if errors.Is(err, errs.ErrEmailAlreadyExist) {
			log.Printf("[INFO] - User with email '%s' already exist", userAdmin.Email)
		} else {
			return err
		}
	} else {
		log.Printf("[INFO] - User with username '%s' created", userAdmin.Username)
	}

	return nil
}

func noError(err error, msg string) {
	if err != nil {
		log.Panicf("[ERROR] - %s: %s", msg, err)
	}
}

// Authorization main function
//
//	@title						Authorization
//	@version					1.0
//	@host						localhost:8080
//	@BasePath					/
//	@description				This is an api that make authorization
//	@securityDefinitions.apikey	BasicAuth
//	@in							header
//	@name						Session
func main() {
	validate := model.Validate()

	configurations, err := getConfigurations(validate)
	noError(err, "Error getting server configurations")

	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		configurations.Postgres.Username,
		configurations.Postgres.Password,
		configurations.Postgres.Host,
		configurations.Postgres.Port,
		configurations.Postgres.DB,
	)
	if configurations.Postgres.SSLDisable {
		url += "?sslmode=disable"
	}

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
		Addr:     fmt.Sprintf("%s:%d", configurations.Redis.Host, configurations.Redis.Port),
		Password: configurations.Redis.Password,
		DB:       configurations.Redis.DB,
	})

	expires := time.Minute * 15 //nolint:gomnd
	if configurations.DevMode {
		expires = time.Hour
	}

	data, err := data.NewDataSQLRedis(db, redisClient, expires, 2000, 1000) //nolint:gomnd
	noError(err, "Error starting data")

	cores := core.NewCore(data, validate, true, time.Hour)

	err = createFirst(configurations, cores)
	noError(err, "Erro creating initial resources")

	server, err := server.CreateHTTPServer(validate, cores, configurations.DevMode)
	noError(err, "Error creating server")

	err = server.Listen(":8080")
	noError(err, "Error listening server")
}
