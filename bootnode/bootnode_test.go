package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBootNodeLoadConfig(t *testing.T) {
	config, err := loadConfig()
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, config)
	assert.Equal(t, DefaultRPCServerPort, config.RPCPort)
}
