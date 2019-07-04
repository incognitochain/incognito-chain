package common

import (
	"encoding/hex"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for NewHash function
 */

func TestHashNewHash(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},                // 32 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, // 32 bytes
	}

	for _, item := range data {
		hashObj, err := Hash{}.NewHash(item)

		assert.Equal(t, nil, err)
		assert.Equal(t, item, hashObj[:])
	}
}

func TestHashNewHashWithInvalidData(t *testing.T) {
	data := [][]byte{
		{}, // empty
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},                        // 31 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}, // 33 bytes
	}

	for _, item := range data {
		hashObj, err := Hash{}.NewHash(item)

		assert.Equal(t, InvalidHashSizeErr, err)
		assert.Equal(t, (*Hash)(nil), hashObj)
	}
}

/*
	Unit test for SetBytes function
 */

func TestHashSetBytes(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},                // 32 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, // 32 bytes
	}

	hashObj := new(Hash)

	for _, item := range data {
		err := hashObj.SetBytes(item)

		assert.Equal(t, nil, err)
		assert.Equal(t, item, hashObj[:])
	}
}

func TestHashSetBytesWithInvalidData(t *testing.T) {
	data := [][]byte{
		{}, // empty
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},                        // 31 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}, // 33 bytes
	}

	hashObj := new(Hash)

	for _, item := range data {
		err := hashObj.SetBytes(item)
		assert.Equal(t, InvalidHashSizeErr, err)
	}
}

/*
	Unit test for GetBytes function
 */

func TestHashGetBytes(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},                // 32 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, // 32 bytes
	}

	hashObj := new(Hash)

	for _, item := range data {
		hashObj.SetBytes(item)

		bytes := hashObj.GetBytes()
		assert.Equal(t, item, bytes)
	}
}

/*
	Unit test for String function
 */

func TestHashString(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},                // 32 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, // 32 bytes
	}
	hashObj := new(Hash)

	for _, item := range data {
		hashObj.SetBytes(item)
		str := hashObj.String()
		assert.Equal(t, hex.EncodedLen(len(hashObj[:])), len(str))
	}
}

/*
	Unit test for NewHashFromStr function
 */

func TestHashNewHashFromStr(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},                // 32 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, // 32 bytes
	}
	hashObj := new(Hash)

	for _, item := range data {
		hashObj.SetBytes(item)
		str := hashObj.String()

		newHash, err := Hash{}.NewHashFromStr(str)
		assert.Equal(t, nil, err)
		assert.Equal(t, item, newHash[:])
	}
}

func TestHashNewHashFromStrWithInvalidString(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},                // 32 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, // 32 bytes
	}
	hashObj := new(Hash)

	for _, item := range data {
		hashObj.SetBytes(item)
		str := hashObj.String()

		// edit string
		str = str + "abc"
		_, err := Hash{}.NewHashFromStr(str)
		assert.Equal(t, InvalidMaxHashSizeErr, err)
	}
}

/*
	Unit test for IsEqual function
 */

func TestHashIsEqual(t *testing.T) {
	hash1, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32})
	hash2, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32})
	hash3, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33})
	hash4, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34})
	data := []struct {
		hash1   *Hash
		hash2   *Hash
		isEqual bool
	}{
		{hash1, hash2, true},
		{nil, nil, true},
		{nil, hash3, false},
		{hash3, hash4, false},
	}

	for _, item := range data {
		isEqual := item.hash1.IsEqual(item.hash2)
		assert.Equal(t, item.isEqual, isEqual)
	}
}

/*
	Unit test for Cmp function
 */

func TestHashCmp(t *testing.T) {
	hash1, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32})
	hash2, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32})
	hash3, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33})
	hash4, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34})
	data := []struct {
		hash1 *Hash
		hash2 *Hash
		cmp   int
	}{
		{hash1, hash2, 0},
		{hash3, hash2, 1},
		{hash1, hash4, -1},
	}

	for _, item := range data {
		cmp, err := item.hash1.Cmp(item.hash2)
		assert.Equal(t, nil, err)
		assert.Equal(t, item.cmp, cmp)
	}
}

func TestHashCmpWithNilHash(t *testing.T) {
	hash1, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32})
	hash2, _ := Hash{}.NewHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33})
	data := []struct {
		hash1 *Hash
		hash2 *Hash
	}{
		{nil, hash2},
		{hash1, nil},
	}

	for _, item := range data {
		_, err := item.hash1.Cmp(item.hash2)
		assert.Equal(t, NilHashErr, err)
	}
}

/*
	Unit test for HashArrayInterface function
 */

func TestHashHashArrayInterface(t *testing.T) {
	data := []interface{}{
		[]byte{1, 2, 3, 4},
		[]string{"1", "2", "3", "4"},
		[]string{"a", "b", "c", "d"},
	}

	for _, item := range data {
		hash, err := HashArrayInterface(item)
		assert.Equal(t, nil, err)
		assert.Equal(t, HashSize, len(hash[:]))
	}
}

func TestHashHashArrayInterfaceWithInvalidInterface(t *testing.T) {
	data := []interface{}{
		"abc",
		123,
	}

	for _, item := range data {
		_, err := HashArrayInterface(item)
		assert.Equal(t, errors.New("interface input is not an array"), err)
	}
}
