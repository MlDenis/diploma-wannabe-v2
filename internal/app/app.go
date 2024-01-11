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
	logger  *zap.Logger
}

func (a *App) Run() error {

	go a.manager.ManageJobs(a.config.Accrual)

	err := a.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		a.logger.Info("Shutting down jobmanager")
		a.manager.Shutdown()
		a.logger.Info("Server error: %e")
		return err
	}
	return nil
}

func NewApp(config *config.Config) (*App, error) {
	l, err := logger.InitializeLogger(config.LogLevel)
	if err != nil {
		return nil, err
	}

	l.Info("Application is running on addr: ", zap.String("", config.Address))
	l.Info("Accrual addr is: ", zap.String("", config.Accrual))
	l.Info("DB addr is: ", zap.String("", config.DatabaseURI))

	ctx := context.Background()
	cursor, err := db.GetCursor(config.DatabaseURI, l)
	if err != nil {
		return nil, err
	}
	manager := jobmanager.NewJobmanager(cursor, config.Accrual, &ctx)
	handler := api.NewHandler(cursor, manager)
	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
	}
	return &App{
		config:  config,
		manager: manager,
		Server:  server,
		logger:  l,
	}, nil
}
