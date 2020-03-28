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
	b64EncodePubKey := "uND4Li1FIzpmjmEe9RZGZlKr53zLP8ZHUP8DSQCZpN4="

	pubKeyBytes, _ := base64.StdEncoding.DecodeString(b64EncodePubKey)
	fmt.Printf("pubKeyBytes: %#v\n", pubKeyBytes)
}