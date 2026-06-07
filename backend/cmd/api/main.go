package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	auditmodule "devtracker/backend/internal/audit"
	"devtracker/backend/internal/auth"
	"devtracker/backend/internal/config"
	dashboardmodule "devtracker/backend/internal/dashboard"
	"devtracker/backend/internal/database"
	docsmodule "devtracker/backend/internal/docs"
	kpimodule "devtracker/backend/internal/kpi"
	appmiddleware "devtracker/backend/internal/middleware"
	notificationmodule "devtracker/backend/internal/notification"
	projectmodule "devtracker/backend/internal/project"
	sprintmodule "devtracker/backend/internal/sprint"
	statusmodule "devtracker/backend/internal/status"
	taskmodule "devtracker/backend/internal/task"
	usermodule "devtracker/backend/internal/user"
	workloadmodule "devtracker/backend/internal/workload"
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
	sprintRepository := sprintmodule.NewRepository(db)
	statusRepository := statusmodule.NewRepository(db)
	taskRepository := taskmodule.NewRepository(db)
	dashboardRepository := dashboardmodule.NewRepository(db)
	kpiRepository := kpimodule.NewRepository(db)
	auditRepository := auditmodule.NewRepository(db)
	notificationRepository := notificationmodule.NewRepository(db)
	workloadRepository := workloadmodule.NewRepository(db)

	userService := usermodule.NewService(userRepository)
	projectService := projectmodule.NewService(projectRepository)
	statusService := statusmodule.NewService(statusRepository)
	taskService := taskmodule.NewService(taskRepository, userRepository, projectRepository, sprintRepository, statusRepository)
	dashboardService := dashboardmodule.NewService(dashboardRepository, sprintRepository)
	kpiService := kpimodule.NewService(kpiRepository, sprintRepository)
	sprintService := sprintmodule.NewService(sprintRepository, projectRepository, kpiService)
	authService := auth.NewService(userRepository, cfg.JWT)
	auditService := auditmodule.NewService(auditRepository)
	notificationService := notificationmodule.NewService(notificationRepository)
	workloadService := workloadmodule.NewService(workloadRepository, sprintRepository, projectRepository)

	authHandler := auth.NewHandler(authService, auditService)
	userHandler := usermodule.NewHandler(userService, auditService)
	projectHandler := projectmodule.NewHandler(projectService, auditService)
	sprintHandler := sprintmodule.NewHandler(sprintService, auditService)
	statusHandler := statusmodule.NewHandler(statusService, auditService)
	taskHandler := taskmodule.NewHandler(taskService, auditService, notificationService)
	dashboardHandler := dashboardmodule.NewHandler(dashboardService)
	kpiHandler := kpimodule.NewHandler(kpiService)
	auditHandler := auditmodule.NewHandler(auditService)
	notificationHandler := notificationmodule.NewHandler(notificationService)
	workloadHandler := workloadmodule.NewHandler(workloadService)

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
	docsmodule.RegisterRoutes(app, cfg.App.BasePath)

	api := app.Group(cfg.App.BasePath)
	// @Summary Health check
	// @Tags Health
	// @Success 200 {object} response.Body
	// @Router /health [get]
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

	auth.RegisterRoutes(api, authHandler, authMiddleware)
	usermodule.RegisterRoutes(api, userHandler, authMiddleware, appmiddleware.RequirePermission)
	projectmodule.RegisterRoutes(api, projectHandler, authMiddleware, appmiddleware.RequirePermission)
	sprintmodule.RegisterRoutes(api, sprintHandler, authMiddleware, appmiddleware.RequirePermission)
	statusmodule.RegisterRoutes(api, statusHandler, authMiddleware, appmiddleware.RequirePermission)
	taskmodule.RegisterRoutes(api, taskHandler, authMiddleware, appmiddleware.RequirePermission)
	dashboardmodule.RegisterRoutes(api, dashboardHandler, authMiddleware, appmiddleware.RequirePermission)
	kpimodule.RegisterRoutes(api, kpiHandler, authMiddleware, appmiddleware.RequirePermission)
	auditmodule.RegisterRoutes(api, auditHandler, authMiddleware, appmiddleware.RequireRole("admin", "project_manager"))
	notificationmodule.RegisterRoutes(api, notificationHandler, authMiddleware, appmiddleware.RequireRole)
	workloadmodule.RegisterRoutes(api, workloadHandler, authMiddleware, appmiddleware.RequirePermission)

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
