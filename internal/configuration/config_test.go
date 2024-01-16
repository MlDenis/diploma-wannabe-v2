package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFlagsProvided(t *testing.T) {
	envs := EnvConfig{}
	flags := CLIOptions{
		Address:     "localhost",
		DatabaseURI: "localhost",
		Accrual:     "localhost",
	}
	config := NewConfig(&flags, &envs)
	assert.Equal(t, &Config{
		Address:     "localhost",
		DatabaseURI: "localhost",
		Accrual:     "localhost",
	}, config)
}

func TestConfigFlagsNotProvided(t *testing.T) {
	envs := EnvConfig{
		Address:     "env",
		DatabaseURI: "env",
		Accrual:     "env",
	}
	flags := CLIOptions{
		Address:     "cli",
		DatabaseURI: "cli",
	}
	config := NewConfig(&flags, &envs)
	assert.Equal(t, &Config{
		Address:     "cli",
		DatabaseURI: "cli",
		Accrual:     "env",
	}, config)
}

func TestConfigBothNotProvided(t *testing.T) {
	envs, _ := NewEnvConfig()
	flags := NewCliOptions()
	config := NewConfig(flags, envs)
	assert.Equal(t, &Config{
		Address:     "localhost:8080",
		DatabaseURI: "localhost:5432",
		Accrual:     "localhost:8081",
	}, config)
}
