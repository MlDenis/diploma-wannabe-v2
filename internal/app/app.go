package app

import (
	"context"
	"go.uber.org/zap"
	"net/http"

	"github.com/MlDenis/diploma-wannabe-v2/internal/api"
	config "github.com/MlDenis/diploma-wannabe-v2/internal/configuration"
	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/jobmanager"
	"github.com/MlDenis/diploma-wannabe-v2/internal/logger"
)

type App struct {
	config  *config.Config
	manager *jobmanager.Jobmanager
	Server  *http.Server
	Logger  *zap.Logger
}

func (a *App) Run(ctx context.Context) error {

	done := make(chan bool)

	go a.manager.ManageJobs(ctx, a.config.Accrual, a.Logger)

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Info("Shutting down jobmanager")
			a.manager.Shutdown()
			a.Logger.Info("Server error: %e")
		}
	}()

	select {
	case <-ctx.Done():
		// Если контекст был отменен, остановите сервер и закройте канал done
		a.Server.Shutdown(ctx)
		close(done)
	case <-done:
		// Если работа была завершена, просто верните nil
		return nil
	}

	return ctx.Err()
}

func NewApp(config *config.Config, ctx context.Context) (*App, error) {
	l, err := logger.InitializeLogger(config.LogLevel)
	if err != nil {
		return nil, err
	}

	l.Info("Application is running on addr: ", zap.String("", config.Address))
	l.Info("Accrual addr is: ", zap.String("", config.Accrual))
	l.Info("DB addr is: ", zap.String("", config.DatabaseURI))

	cursor, err := db.GetCursor(config.DatabaseURI, l)
	if err != nil {
		return nil, err
	}
	manager := jobmanager.NewJobmanager(cursor, config.Accrual, &ctx)
	handler := api.NewHandler(cursor, manager, l)
	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
	}
	return &App{
		config:  config,
		manager: manager,
		Server:  server,
		Logger:  l,
	}, nil
}
