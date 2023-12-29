package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvConfig(t *testing.T) {
	testConfig, _ := NewEnvConfig()
	assert.Equal(t, testConfig.Address, "localhost:8080")
	assert.Equal(t, testConfig.DatabaseURI, "localhost:5432")
	assert.Equal(t, testConfig.Accrual, "localhost:8081")
}
