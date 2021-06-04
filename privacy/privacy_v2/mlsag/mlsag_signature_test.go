package mlsag

import (
	"bytes"
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

	bRingTemp, err := ringTemp.ToBytes()
	assert.Equal(t, nil, err, "There should not be any error when ring.ToBytes")

	assert.Equal(t, true, bytes.Equal(bRingTemp, bRing))
}

func TestSignatureHexBytesConversion(t *testing.T) {
	signer := InitializeSignatureForTest()
	s := common.HashH([]byte("Test"))
	signature, err_sig := signer.Sign(s[:])
	assert.Equal(t, err_sig, nil, "Signing signature should not have error")
	sig_byte, err_byte := signature.ToBytes()
	assert.Equal(t, err_byte, nil, "Error of byte should be nil")
	temp_sig_byte, err_from_bytes := new(MlsagSig).FromBytes(sig_byte)
	assert.Equal(t, err_from_bytes, nil, "Bytes to signature should not have errors")
	assert.Equal(t, signature, temp_sig_byte, "Bytes to signature should be correct")
}

func removeLastElement(s []*operation.Point) []*operation.Point {
	return s[:len(s)-1]
}

func TestErrorBrokenRealSignature(t *testing.T) {
	signer := InitializeSignatureForTest()

	s := common.HashH([]byte("Test"))
	signature, err_sig := signer.Sign(s[:])
	assert.Equal(t, err_sig, nil, "Signing signature should not have error")
	assert.NotEqual(t, nil, signature)
}

func TestErrorBrokenHexByteSignature(t *testing.T) {
	signer := InitializeSignatureForTest()

	s := common.HashH([]byte("Test"))
	signature, err_sig := signer.Sign(s[:])
	assert.Equal(t, err_sig, nil, "Signing signature should not have error")

	// Make signature byte broken
	sig_byte, _ := signature.ToBytes()
	sig_byte = sig_byte[:len(sig_byte)-1]

	tmp_sig, byte_err := new(MlsagSig).FromBytes(sig_byte)
	assert.Equal(t, nil == tmp_sig, true, "FromByte of broken signature should return empty signature")
	assert.NotEqual(t, nil, byte_err)

	sig_byte = sig_byte[:len(sig_byte)-31]
	tmp_sig, byte_err = new(MlsagSig).FromBytes(sig_byte)
	assert.Equal(t, tmp_sig == nil, true, "FromByte of broken signature should return empty signature")
	assert.NotEqual(t, nil, byte_err)
}
