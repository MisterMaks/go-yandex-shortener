package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	expectedConfig := &Config{
		ServerAddress:   "localhost:8080",
		GRPCAddress:     "localhost:8081",
		BaseURL:         "http://localhost:8080/",
		LogLevel:        "INFO",
		FileStoragePath: "/tmp/short-url-db.json",
		DatabaseDSN:     "",
		EnableHTTPS:     false,
		Config:          "",
	}

	config, err := NewConfig()

	require.NoError(t, err)
	assert.Equal(t, expectedConfig, config)

	t.TempDir()
}
