package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCmdLoadParams(t *testing.T) {
	params, err := loadParams()
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, params)
	assert.Equal(t, false, params.TestNet)
}
