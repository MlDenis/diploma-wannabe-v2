package configuration

type Config struct {
	Address     string
	DatabaseURI string
	Accrual     string
}

func NewConfig(flags *CLIOptions, envs *EnvConfig) *Config {
	result := &Config{
		Address:     flags.Address,
		Accrual:     flags.Accrual,
		DatabaseURI: flags.DatabaseURI,
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
	return result
}
