package base58

import (
	"bytes"
	"github.com/btcsuite/btcutil/base58"
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
		isNewCheckSum := (common.RandInt() % 2) == 1
		checkSum := ChecksumFirst4Bytes(item, isNewCheckSum)
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
		encodedData := base58.NewEncode(item.input, item.version)

		data, version, err := base58.Decode(encodedData)
		assert.Equal(t, item.input, data)
		assert.Equal(t, item.version, version)
		assert.Equal(t, nil, err)
	}
}

func TestNewEncodingAndCheckSum(t *testing.T) {
	for i := 0; i < 100; i++ {
		data := common.RandBytes(common.RandIntInterval(0, 1000))
		expectedEncoding := base58.CheckEncode(data, 0)

		actualEncoded := Base58Check{}.Encode(data, common.ZeroByte)

		assert.Equal(t, expectedEncoding, actualEncoded, "encodings not equals: %v, %v", expectedEncoding, actualEncoded)
	}
}

func TestNewOldEncodeDecode(t *testing.T) {
	for i := 0; i < 100; i++ {
		data := common.RandBytes(common.RandIntInterval(0, 1000))
		oldEncoding := Base58Check{}.Encode(data, 0x00)
		newEncoding := Base58Check{}.NewEncode(data, 0x00)

		oldDecode, _, err := Base58Check{}.Decode(oldEncoding)
		assert.Equal(t, nil, err, "base58Check old-decode returns an error: %v\n", err)

		newDecode, _, err := Base58Check{}.Decode(newEncoding)
		assert.Equal(t, nil, err, "base58Check new-decode returns an error: %v\n", err)

		assert.Equal(t, true, bytes.Equal(oldDecode, data), "encodings not equals: %v, %v\n", oldDecode, data)
		assert.Equal(t, true, bytes.Equal(newDecode, data), "encodings not equals: %v, %v\n", newDecode, data)
	}
}
