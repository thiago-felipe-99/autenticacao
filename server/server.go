package server

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/pt"
	"github.com/go-playground/locales/pt_BR"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	ptTranslations "github.com/go-playground/validator/v10/translations/pt"
	pt_br_translations "github.com/go-playground/validator/v10/translations/pt_BR"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/thiago-felipe-99/autenticacao/core"
	_ "github.com/thiago-felipe-99/autenticacao/docs" // importing docs for swagger
)

const defaultQtResults = 100

type sent struct {
	Message string `json:"message" bson:"message"`
}

type expectError struct {
	err    error
	status int
}

type okay struct {
	message string
	status  int
}

func callingCore(
	coreFunc func() error,
	expectErrors []expectError,
	unexpectMessageError string,
	okay okay,
	language ut.Translator,
	handler *fiber.Ctx,
) error {
	err := coreFunc()
	if err != nil {
		modelInvalid := core.InvalidError{}
		if okay := errors.As(err, &modelInvalid); okay {
			return handler.Status(fiber.StatusBadRequest).
				JSON(sent{modelInvalid.Translate(language)})
		}

		for _, expectError := range expectErrors {
			if errors.Is(err, expectError.err) {
				return handler.Status(expectError.status).JSON(sent{expectError.err.Error()})
			}
		}

		log.Printf("[ERROR] - %s: %s", unexpectMessageError, err)

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{unexpectMessageError})
	}

	return handler.Status(okay.status).JSON(sent{okay.message})
}

func callingCoreWithReturn[T any](
	coreFunc func() (T, error),
	expectErrors []expectError,
	unexpectMessageError string,
	language ut.Translator,
	handler *fiber.Ctx,
) error {
	data, err := coreFunc()
	if err != nil {
		modelInvalid := core.InvalidError{}
		if okay := errors.As(err, &modelInvalid); okay {
			return handler.Status(fiber.StatusBadRequest).
				JSON(sent{modelInvalid.Translate(language)})
		}

		for _, expectError := range expectErrors {
			if errors.Is(err, expectError.err) {
				return handler.Status(expectError.status).JSON(sent{expectError.err.Error()})
			}
		}

		log.Printf("[ERROR] - %s: %s", unexpectMessageError, err)

		return handler.Status(fiber.StatusInternalServerError).
			JSON(sent{unexpectMessageError})
	}

	return handler.JSON(data)
}

func createTranslator(validate *validator.Validate) (*ut.UniversalTranslator, error) {
	translator := ut.New(en.New(), pt.New(), pt_BR.New())

	enTrans, _ := translator.GetTranslator("en")

	err := en_translations.RegisterDefaultTranslations(validate, enTrans)
	if err != nil {
		return nil, fmt.Errorf("error register 'en' translation: %w", err)
	}

	ptTrans, _ := translator.GetTranslator("pt")

	err = ptTranslations.RegisterDefaultTranslations(validate, ptTrans)
	if err != nil {
		return nil, fmt.Errorf("error register 'pt' translation: %w", err)
	}

	ptBRTrans, _ := translator.GetTranslator("pt_BR")

	err = pt_br_translations.RegisterDefaultTranslations(validate, ptBRTrans)
	if err != nil {
		return nil, fmt.Errorf("error register 'pt_BR' translation: %w", err)
	}

	return translator, nil
}

func registerDefaultMiddlewares(app *fiber.App) {
	prometheus := fiberprometheus.New("publisher")
	prometheus.RegisterAt(app, "/metrics")

	app.Use(logger.New(logger.Config{
		//nolint:lll
		Format:        "${time} [INFO] - Finished request | ${ip} | ${status} | ${latency} | ${method} | ${path} | ${bytesSent} | ${bytesReceived} | ${error}\n",
		TimeFormat:    "2006/01/02 15:04:05",
		Next:          nil,
		Done:          nil,
		TimeZone:      "Local",
		Output:        os.Stdout,
		DisableColors: false,
		TimeInterval:  500 * time.Millisecond, //nolint:gomnd
		CustomTags:    logger.ConfigDefault.CustomTags,
	}))

	app.Use(recover.New(recover.Config{
		EnableStackTrace:  true,
		Next:              nil,
		StackTraceHandler: recover.ConfigDefault.StackTraceHandler,
	}))

	app.Use(prometheus.Middleware)

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowCredentials: true,
		MaxAge:           10, //nolint:gomnd
		ExposeHeaders:    "session",
		Next:             nil,
		AllowOriginsFunc: nil,
	}))

	swaggerConfig := swagger.Config{ //nolint:exhaustruct
		Title:                  "Autenticação",
		WithCredentials:        true,
		DisplayRequestDuration: true,
	}

	app.Get("/swagger/*", swagger.New(swaggerConfig))
}

func CreateHTTPServer(validate *validator.Validate, cores *core.Cores) (*fiber.App, error) {
	app := fiber.New()

	registerDefaultMiddlewares(app)

	translator, err := createTranslator(validate)
	if err != nil {
		return nil, err
	}

	languages := []string{"en", "pt_BR", "pt"}

	role := Role{
		core:       cores.Role,
		translator: translator,
		languages:  languages,
	}

	app.Get("/role", role.GetAll)
	app.Post("/role", role.Create)
	app.Get("/role/:name", role.GetByName)
	app.Delete("/role/:name", role.Delete)

	return app, nil
}
