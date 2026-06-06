package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devtracker/backend/internal/auth"
	"devtracker/backend/internal/config"
	"devtracker/backend/internal/database"
	appmiddleware "devtracker/backend/internal/middleware"
	projectmodule "devtracker/backend/internal/project"
	usermodule "devtracker/backend/internal/user"
	"devtracker/backend/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
)

func main() {
	cfg := config.Load()
	log := newLogger(cfg.App.Env)

	db, err := database.Connect(cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}

	if cfg.Database.RunMigrations {
		if err := database.RunMigrations(db, log); err != nil {
			log.Fatal().Err(err).Msg("failed to run database migrations")
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to access database handle")
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database connection")
		}
	}()

	userRepository := usermodule.NewRepository(db)
	projectRepository := projectmodule.NewRepository(db)

	userService := usermodule.NewService(userRepository)
	projectService := projectmodule.NewService(projectRepository)
	authService := auth.NewService(userRepository, cfg.JWT)

	authHandler := auth.NewHandler(authService)
	userHandler := usermodule.NewHandler(userService)
	projectHandler := projectmodule.NewHandler(projectService)

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		BodyLimit:    cfg.App.BodyLimit,
		ErrorHandler: response.ErrorHandler,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New())
	app.Use(appmiddleware.RequestLogger(log))

	api := app.Group(cfg.App.BasePath)
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			return err
		}

		return response.OK(c, "service is healthy", fiber.Map{
			"app": cfg.App.Name,
			"env": cfg.App.Env,
		})
	})

	authMiddleware := appmiddleware.JWTAuth(cfg.JWT, userRepository)
	adminOnly := appmiddleware.RequireRoles("admin")

	auth.RegisterRoutes(api, authHandler, authMiddleware)
	usermodule.RegisterRoutes(api, userHandler, authMiddleware, adminOnly)
	projectmodule.RegisterRoutes(api, projectHandler, authMiddleware)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.Listen(":" + cfg.App.Port)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatal().Err(err).Msg("server stopped unexpectedly")
		}
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Error().Err(err).Msg("graceful shutdown failed")
		}
	}
}

func newLogger(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().
			Timestamp().
			Logger()
	}

	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
