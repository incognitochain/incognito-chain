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
	b64EncodePubKey := "98r+D2A3M3E8sQvYobUkx3gEIMaJhc23md9cUXbtNhs="

	pubKeyBytes, _ := base64.StdEncoding.DecodeString(b64EncodePubKey)
	fmt.Printf("pubKeyBytes: %#v\n", pubKeyBytes)
}