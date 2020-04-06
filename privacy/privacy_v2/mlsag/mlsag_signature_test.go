package mlsag

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func InitializeSignatureForTest() (mlsag *Mlsag) {
	keyInputs := []*operation.Scalar{}
	for i := 0; i < 3; i += 1 {
		privateKey := operation.RandomScalar()
		keyInputs = append(keyInputs, privateKey)
	}
	numFake := 8
	pi := common.RandInt() % numFake
	ring := NewRandomRing(keyInputs, numFake, pi)
	return NewMlsag(keyInputs, ring, pi)
}

func TestRing(t *testing.T) {
	keyInputs := []*operation.Scalar{}
	for i := 0; i < 8; i += 1 {
		privateKey := operation.RandomScalar()
		keyInputs = append(keyInputs, privateKey)
	}
	numFake := 5
	pi := common.RandInt() % numFake
	ring := NewRandomRing(keyInputs, numFake, pi)
	bRing, err := ring.ToBytes()
	assert.Equal(t, nil, err, "There should not be any error when ring.ToBytes")

	ringTemp, err := new(Ring).FromBytes(bRing)
	assert.Equal(t, nil, err, "There should not be any error when ring.FromBytes")

	fmt.Println(ring.keys)
	fmt.Println(ringTemp.keys)

	bRingTemp, err := ringTemp.ToBytes()
	assert.Equal(t, nil, err, "There should not be any error when ring.ToBytes")

	assert.Equal(t, true, bytes.Equal(bRingTemp, bRing))
}

func TestSignatureHexBytesConversion(t *testing.T) {
	signer := InitializeSignatureForTest()

	signature, err_sig := signer.Sign([]byte("Test"))
	sig_hex, err_hex := signature.ToHex()
	sig_byte, err_byte := signature.ToBytes()

	assert.Equal(t, err_sig, nil, "Signing signature should not have error")
	assert.Equal(t, err_hex, nil, "Error of hex should be nil")
	assert.Equal(t, err_byte, nil, "Error of byte should be nil")
	assert.Equal(t, hex.EncodeToString(sig_byte), sig_hex, "Hex encoding signature should be correct")

	temp_sig_byte, err_from_bytes := new(MlsagSig).FromBytes(sig_byte)
	assert.Equal(t, err_from_bytes, nil, "Bytes to signature should not have errors")
	assert.Equal(t, signature, temp_sig_byte, "Bytes to signature should be correct")

	temp_sig_hex, err_from_hex := new(MlsagSig).FromHex(sig_hex)
	assert.Equal(t, err_from_hex, nil, "Hex to signature should not have errors")
	assert.Equal(t, signature, temp_sig_hex, "Hex to signature should be correct")
}

func removeLastElement(s []*operation.Point) []*operation.Point {
	return s[:len(s)-1]
}

func TestErrorBrokenRealSignature(t *testing.T) {
	signer := InitializeSignatureForTest()

	signature, err_sig := signer.Sign([]byte("Test"))
	assert.Equal(t, err_sig, nil, "Signing signature should not have error")

	// Make the signature broken
	signature.keyImages = removeLastElement(signature.keyImages)

	// Test
	hx, err_hex := signature.ToHex()
	assert.Equal(t, hx, "", "Hex of broken signature should be empty")
	assert.Equal(
		t, err_hex.Error(),
		"Error in MLSAG MlsagSig ToHex: the signature is broken (size of keyImages and r differ",
		"ToHex of broken signature should return error",
	)
	b, err_byte := signature.ToBytes()
	assert.Equal(t, len(b), 0, "Byte of broken signature should be empty")
	assert.Equal(
		t, err_byte.Error(),
		"Error in MLSAG MlsagSig ToBytes: the signature is broken (size of keyImages and r differ)",
		"ToByte of broken signature should return error",
	)
}

func TestErrorBrokenHexByteSignature(t *testing.T) {
	signer := InitializeSignatureForTest()

	signature, err_sig := signer.Sign([]byte("Test"))
	assert.Equal(t, err_sig, nil, "Signing signature should not have error")

	// Make signature hex broken
	sig_hex, _ := signature.ToHex()
	sig_hex = sig_hex[:len(sig_hex)-1]

	sig, hex_err := new(MlsagSig).FromHex(sig_hex)
	assert.Equal(t, sig == nil, true, "FromHex of broken signature should return empty signature")
	assert.Equal(
		t, hex_err.Error(),
		"Error in MLSAG MlsagSig FromHex: the signature hex is broken",
		"FromHex of broken signature should return error",
	)

	// Make signature byte broken
	sig_byte, _ := signature.ToBytes()
	sig_byte = sig_byte[:len(sig_byte)-1]

	tmp_sig, byte_err := new(MlsagSig).FromBytes(sig_byte)
	assert.Equal(t, tmp_sig == nil, true, "FromByte of broken signature should return empty signature")
	assert.Equal(
		t, byte_err.Error(),
		"Error in MLSAG MlsagSig FromBytes: the signature byte is broken (missing byte)",
		"Broken byte signature should return error",
	)

	sig_byte = sig_byte[:len(sig_byte)-31]
	tmp_sig, byte_err = new(MlsagSig).FromBytes(sig_byte)
	assert.Equal(t, tmp_sig == nil, true, "FromByte of broken signature should return empty signature")
	assert.Equal(
		t, byte_err.Error(),
		"Error in MLSAG MlsagSig FromBytes: the signature byte is broken (some scalar is missing)",
		"Broken byte signature should return error",
	)

	// TODO
	// I did not test signature with broken keyImages because I don't know how to create broken byte of edwards point.
	// Will do later
}
