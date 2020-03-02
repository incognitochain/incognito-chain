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
	b64EncodePubKey := "O6ZcNqhg97e2rAL873tUm9c66RTdXh+423O+C1B8Kgc="

	pubKeyBytes, _ := base64.StdEncoding.DecodeString(b64EncodePubKey)
	fmt.Printf("pubKeyBytes: %#v\n", pubKeyBytes)
}