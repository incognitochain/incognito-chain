package bnb

import (
	"encoding/base64"
	"fmt"
	"testing"
)

//func TestDecodePublicKeyValidator(t *testing.T) {
//	err := DecodePublicKeyValidator()
//	assert.Equal(t, nil, err)
//}


func TestDecodePubKeyValidator(t *testing.T){
	b64EncodePubKey := "3bWVA8EcEN9ZfvcWL6kV29Sfa50O58jHInmg4AkfkxU="

	pubKeyBytes, _ := base64.StdEncoding.DecodeString(b64EncodePubKey)
	fmt.Printf("pubKeyBytes: %#v\n", pubKeyBytes)
}