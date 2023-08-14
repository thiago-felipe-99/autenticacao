package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type admin struct {
	Name     string `config:"name"     validate:"required"`
	Username string `config:"username" validate:"required"`
	Password string `config:"password" validate:"required"`
	Email    string `config:"email"    validate:"required"`
}

type roleAdmin struct {
	Name string `config:"name" validate:"required"`
}

type postgresConfig struct {
	Host       string `config:"host"        validate:"required"`
	Port       int    `config:"port"        validate:"required"`
	Username   string `config:"username"    validate:"required"`
	Password   string `config:"password"    validate:"required"`
	DB         string `config:"db"          validate:"required"`
	SSLDisable bool   `config:"ssl_disable" validate:""`
}

type redisConfig struct {
	Host     string `config:"host"     validate:"required"`
	Port     int    `config:"port"     validate:"required"`
	Password string `config:"password" validate:"required"`
	DB       int    `config:"db"       validate:"min=0"`
}

type configurations struct {
	User     admin          `config:"user"     validate:"required"`
	Role     roleAdmin      `config:"role"     validate:"required"`
	Postgres postgresConfig `config:"postgres" validate:"required"`
	Redis    redisConfig    `config:"redis"    validate:"required"`
}

//nolint:gomnd
func defaultConfigurations() configurations {
	return configurations{
		User: admin{
			Name:     "First Admin",
			Username: "admin",
			Password: "admin",
			Email:    "admin@local.com",
		},
		Role: roleAdmin{
			Name: "role",
		},
		Postgres: postgresConfig{
			Host:       "localhost",
			Port:       5432,
			Username:   "postgres",
			Password:   "postgres",
			DB:         "postgres",
			SSLDisable: false,
		},
		Redis: redisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "redis",
			DB:       0,
		},
	}
}

func parseEnv(env string) string {
	keys := strings.SplitN(env, "_", 3) //nolint:gomnd
	size := len(keys)

	var key string

	switch size {
	case 0:
		return ""
	case 1:
		return ""
	case 2: //nolint:gomnd
		key = keys[1]
	default:
		key = keys[1] + "__" + strings.Join(keys[2:], "_")
	}

	return strings.ToLower(key)
}

func getConfigurations(validate *validator.Validate) (*configurations, error) {
	koanfConfig := koanf.Conf{
		Delim:       "__",
		StrictMerge: false,
	}

	configRaw := koanf.NewWithConf(koanfConfig)

	err := configRaw.Load(structs.Provider(defaultConfigurations(), "config"), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get default configurations: %w", err)
	}

	err = configRaw.Load(file.Provider(".env"), dotenv.ParserEnv("AUTENTICACAO", "__", parseEnv))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("unable to get .env file: %w", err)
	}

	err = configRaw.Load(env.Provider("AUTENTICACAO", "__", parseEnv), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get environment vaiables: %w", err)
	}

	config := &configurations{} //nolint:exhaustruct

	err = configRaw.UnmarshalWithConf("", config, koanf.UnmarshalConf{ //nolint:exhaustruct
		Tag:       "config",
		FlatPaths: false,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal configurations: %w", err)
	}

	err = validate.Struct(config)
	if err != nil {
		return nil, fmt.Errorf("error validating configurations: %w", err)
	}

	return config, nil
}
