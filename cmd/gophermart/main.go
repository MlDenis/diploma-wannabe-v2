package main

import (
	"context"
	"github.com/MlDenis/diploma-wannabe-v2/internal/app"
	config "github.com/MlDenis/diploma-wannabe-v2/internal/configuration"
	"log"
	"net/http"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flags := config.NewCliOptions()
	envs, err := config.NewEnvConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.NewConfig(flags, envs)

	wg.Add(1)

	gophermart, err := app.NewApp(cfg, ctx)
	if err != nil {
		log.Fatal(err)
	}

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := gophermart.Run(ctx)
		if err != nil && err != http.ErrServerClosed {
			gophermart.Logger.Error(err.Error())
		}
	}(wg)
	<-ctx.Done()

	gophermart.Server.Shutdown(ctx)
	wg.Wait()
}
