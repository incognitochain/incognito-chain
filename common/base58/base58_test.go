package base58

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
		Unit test for Encode function
 */
func TestBase58Encode(t *testing.T){
	data := [][]byte{
		{1},
		{1,2,3},
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 25 bytes
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 30 bytes
	}

	base58 := new(Base58)
	for _, item := range data {
		encodedData := base58.Encode(item)
		assert.Greater(t, len(encodedData), 0)
	}
}

func TestBase58EncodeWithEmptyData(t *testing.T){
	base58 := new(Base58)
	encodedData := base58.Encode([]byte{})
	assert.Equal(t,0,  len(encodedData))
}

/*
		Unit test for Decode function
 */

func TestBase58Decode(t *testing.T){
	data := [][]byte{
		{1},
		{1,2,3},
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 25 bytes
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 30 bytes
	}

	base58 := new(Base58)
	for _, item := range data {
		encodedData := base58.Encode(item)

		decodedData := base58.Decode(encodedData)
		assert.Equal(t, item, decodedData)
	}
}

func TestBase58DecodeWithEmptyData(t *testing.T){
	base58 := new(Base58)
	decodedData := base58.Decode("")
	assert.Equal(t,0, len(decodedData))
}