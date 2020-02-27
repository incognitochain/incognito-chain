package relaying

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
	b64EncodePubKey := "RtEX+5PkniuWYH0rwr2NW742SXCoeqU3xvlzg7tmnVM="

	pubKeyBytes, _ := base64.StdEncoding.DecodeString(b64EncodePubKey)
	fmt.Printf("pubKeyBytes: %+v\n", pubKeyBytes)
}