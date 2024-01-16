package configuration

import "flag"

type CLIOptions struct {
	Address     string
	DatabaseURI string
	Accrual     string
	LogLevel    string
}

func NewCliOptions() *CLIOptions {
	var address = flag.String("a", "", "server address")
	var accrual = flag.String("r", "", "accrual address")
	var database = flag.String("d", "", "database address")
	var logLevel = flag.String("l", "", "log level")
	flag.Parse()

	return &CLIOptions{
		Address:     *address,
		DatabaseURI: *accrual,
		Accrual:     *database,
		LogLevel:    *logLevel,
	}
}
