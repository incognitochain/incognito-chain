// Copyright (c) 2013-2014 The thaibaoautonomous developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"bytes"
	"errors"

	"github.com/ninjadotorg/constant/common"
)

// ErrChecksum indicates that the checksum of a check-encoded string does not verify against
// the checksum.
var ErrChecksum = errors.New("checksum error")

// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
var ErrInvalidFormat = errors.New("invalid format: version and/or checksum bytes missing")

// checksum: first four bytes of sha256^2
func ChecksumFirst4Bytes(input []byte) (cksum []byte) {
	cksum = make([]byte, 4)
	h2 := common.DoubleHashB(input)
	copy(cksum[:], h2[:4])
	return
}

type Base58Check struct {
}

// Encode prepends a version byte and appends a four byte checksum.
func (self Base58Check) Encode(input []byte, version byte) string {
	b := make([]byte, 0, 1+len(input)+4)
	b = append(b, version)
	b = append(b, input[:]...)
	cksum := ChecksumFirst4Bytes(b)
	b = append(b, cksum[:]...)
	return Base58{}.Encode(b)
}

// Decode decodes a string that was encoded with Encode and verifies the checksum.
func (self Base58Check) Decode(input string) (result []byte, version byte, err error) {
	decoded := Base58{}.Decode(input)
	if len(decoded) < 5 {
		return nil, 0, ErrInvalidFormat
	}
	version = decoded[0]
	// var cksum []byte
	cksum := make([]byte, 4)
	copy(cksum[:], decoded[len(decoded)-4:])
	if bytes.Compare(ChecksumFirst4Bytes(decoded[:len(decoded)-4]), cksum) != 0 {
		return nil, 0, ErrChecksum
	}
	payload := decoded[1: len(decoded)-4]
	result = append(result, payload...)
	return
}
