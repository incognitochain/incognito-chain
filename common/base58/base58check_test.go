package base58

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for ChecksumFirst4Bytes function
*/

func TestBase58CheckChecksumFirst4Bytes(t *testing.T) {
	data := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5},                // 25 bytes
		{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, // 30 bytes
	}

	for _, item := range data {
		checkSum := ChecksumFirst4Bytes(item)
		assert.Equal(t, common.CheckSumLen, len(checkSum))
	}
}

/*
	Unit test for Encode Base58Check function
*/

func TestBase58CheckEncode(t *testing.T) {
	data := []struct {
		input   []byte
		version byte
	}{
		{[]byte{1}, byte(0)},
		{[]byte{1, 2, 3}, byte(1)},
		{[]byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, byte(2)},
		{[]byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, byte(3)},
	}

	base58 := new(Base58Check)
	for _, item := range data {
		encodedData := base58.Encode(item.input, item.version)
		assert.Greater(t, len(encodedData), 0)
	}
}

/*
	Unit test for Decode Base58Check function
*/

func TestBase58CheckDecode(t *testing.T) {
	data := []struct {
		input   []byte
		version byte
	}{
		{[]byte{1}, byte(0)},
		{[]byte{1, 2, 3}, byte(1)},
		{[]byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, byte(2)},
		{[]byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, byte(3)},
	}

	base58 := new(Base58Check)
	for _, item := range data {
		encodedData := base58.Encode(item.input, item.version)

		data, version, err := base58.Decode(encodedData)
		assert.Equal(t, item.input, data)
		assert.Equal(t, item.version, version)
		assert.Equal(t, nil, err)
	}
}
