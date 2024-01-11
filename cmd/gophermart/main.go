package main

import (
	"log"

	"github.com/MlDenis/diploma-wannabe-v2/internal/app"
	config "github.com/MlDenis/diploma-wannabe-v2/internal/configuration"
)

func main() {
	flags := config.NewCliOptions()
	envs, err := config.NewEnvConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.NewConfig(flags, envs)

	gophermart, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := gophermart.Run(); err != nil {
		log.Fatalln(err)
	}
}
