package common

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for GZipFromBytes function
 */

func TestGzipGZipFromBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}

	compressedData, err := GZipFromBytes(data)
	assert.Equal(t, nil, err)
	assert.Greater(t, len(data), len(compressedData))
}

func TestGzipGZipFromBytesWithEmptyInput(t *testing.T) {
	data := []byte{}

	compressedData, err := GZipFromBytes(data)
	assert.Equal(t, errors.New("input to gzip compress is empty"), err)
	assert.Equal(t, 0, len(compressedData))
}

/*
	Unit test for GZipToBytes function
 */

func TestGzipGZipToBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}

	compressedData, err := GZipFromBytes(data)
	assert.Equal(t, nil, err)
	assert.Greater(t, len(data), len(compressedData))

	decompressedData, err := GZipToBytes(compressedData)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, decompressedData)
}

func TestGzipGZipToBytesWithInvalidData(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}
	compressedData, _ := GZipFromBytes(data)

	// edit compressedData
	compressedData = append([]byte{1}, compressedData...)

	decompressedData, err := GZipToBytes(compressedData)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 0, len(decompressedData))
}
