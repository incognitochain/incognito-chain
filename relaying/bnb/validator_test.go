package relaying

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodePublicKeyValidator(t *testing.T) {
	err := DecodePublicKeyValidator()
	assert.Equal(t, nil, err)
}
