package main

import (
	"github.com/MlDenis/diploma-wannabe-v2/internal/app"
	config "github.com/MlDenis/diploma-wannabe-v2/internal/configuration"
	"github.com/MlDenis/diploma-wannabe-v2/internal/logger"
)

func main() {
	flags := config.NewCliOptions()
	envs, err := config.NewEnvConfig()
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	newApp, _err := app.NewApp(config.NewConfig(flags, envs))
	if _err != nil {
		logger.ErrorLog.Fatal(_err)
	}
	newApp.Run()

}
