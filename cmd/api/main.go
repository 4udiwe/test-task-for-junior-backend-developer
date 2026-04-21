package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	taskscheduler "example.com/taskservice/internal/scheduler"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Repositories
	taskRepo := postgresrepo.New(pool)
	recurrenceRuleRepo := postgresrepo.NewRecurrenceRuleRepository(pool)
	jobRepo := postgresrepo.NewJobRepository(pool)

	// Usecase with recurrence support
	taskUsecase := task.NewServiceWithRecurrence(taskRepo, recurrenceRuleRepo, jobRepo)

	// HTTP Handlers
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	recurrenceHandler := httphandlers.NewRecurrenceHandler(logger, taskUsecase)
	docsHandler := swaggerdocs.NewHandler()

	// Router with recurrence support
	router := transporthttp.NewRouterWithRecurrence(taskHandler, recurrenceHandler, docsHandler)

	// Scheduler for recurring tasks
	scheduler := taskscheduler.NewScheduler(
		logger,
		jobRepo,
		recurrenceRuleRepo,
		taskRepo,
		1*time.Hour,  // Poll interval
		5*time.Minute, // Lock duration
	)

	// Start scheduler
	scheduler.Start(ctx)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Graceful scheduler shutdown
		scheduler.Stop()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
