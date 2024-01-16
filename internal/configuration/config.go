package configuration

type Config struct {
	Address     string
	DatabaseURI string
	Accrual     string
	LogLevel    string
}

func NewConfig(flags *CLIOptions, envs *EnvConfig) *Config {
	result := &Config{
		Address:     flags.Address,
		Accrual:     flags.Accrual,
		DatabaseURI: flags.DatabaseURI,
		LogLevel:    flags.LogLevel,
	}
	if flags.Address == "" {
		result.Address = envs.Address
	}
	if flags.Accrual == "" {
		result.Accrual = envs.Accrual
	}
	if flags.DatabaseURI == "" {
		result.DatabaseURI = envs.DatabaseURI
	}
	if flags.LogLevel == "" {
		result.LogLevel = envs.LogLevel
	}
	return result
}
