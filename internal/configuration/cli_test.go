package configuration

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentCLI(t *testing.T) {
	os.Args = []string{"main.go", "-a", "localhost:4444", "-d=11s", "-r=5m"}
	var address = flag.String("a", "", "server address")
	var accrual = flag.String("r", "", "accrual address")
	var database = flag.String("d", "", "database address")
	flag.Parse()

	args := &CLIOptions{
		Address:     *address,
		DatabaseURI: *database,
		Accrual:     *accrual,
	}

	assert.Equal(t, "localhost:4444", args.Address)
	assert.Equal(t, "11s", args.DatabaseURI)
	assert.Equal(t, "5m", args.Accrual)
}
