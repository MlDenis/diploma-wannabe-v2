package configuration

import (
	"github.com/caarlos0/env/v6"
)

type EnvConfig struct {
	Address     string `env:"RUN_ADDRESS,required" envDefault:"localhost:8080"`
	DatabaseURI string `env:"DATABASE_URI,required" envDefault:"localhost:5432"`
	Accrual     string `env:"ACCRUAL_SYSTEM_ADDRESS,required" envDefault:"localhost:8081"`
	LogLevel    string `env:"LOG_LEVEL,required" envDefault:"info"`
}

func NewEnvConfig() (*EnvConfig, error) {
	var config EnvConfig
	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
