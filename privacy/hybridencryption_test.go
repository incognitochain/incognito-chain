package privacy

import (
	"fmt"
	"math/big"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEncryptionBytes (t*testing.T){
	msg := RandBytes(100)
	privateKey := GeneratePrivateKey([]byte{123})
	publicKeyBytes := GeneratePublicKey(privateKey)

	privateKey := new(big.Int).SetBytes(privateKey)
	publicKey := new(EllipticPoint)
	err := publicKey.Decompress(publicKeyBytes)
	if err != nil{
		fmt.Printf("ERR: %v\n", err)
	}

	ciphertext, err := HybridEncrypt(msg, publicKey)
	if err != nil{
		fmt.Printf("ERR: %v\n", err)
	}

	ciphertextBytes := ciphertext.Bytes()
	ciphertext2 := new(Ciphertext)
	ciphertext2.SetBytes(ciphertextBytes)

	msg2, err := HybridDecrypt(ciphertext2, privateKey)
	if err != nil{
		fmt.Printf("ERR: %v\n", err)
	}

	fmt.Printf("msg: %v\n", msg)
	fmt.Printf("msg decypted: %v\n", msg2)

	assert.Equal(t, msg, msg2)


	// Test for JS
	// msg = [10, 20]
	privKey := big.NewInt(10)
	ciphertext3 := new(Ciphertext)
	ciphertext3.SetBytes([]byte{3, 69, 226, 202, 112, 192, 21, 80, 171, 139, 218, 70, 130, 163, 52, 40, 90, 127, 34, 118, 3, 123, 195, 92, 181, 130, 157, 210, 193, 53, 98, 214, 127, 3, 42, 7, 110, 114, 146, 181, 3, 168, 186, 99, 72, 178, 8, 112, 127, 221, 54, 218, 236, 52, 162, 228, 216, 81, 87, 34, 152, 219, 140, 250, 76, 151, 2, 247, 194, 71, 51, 195, 204, 139, 155, 189, 254, 110, 46, 76, 65, 35, 246, 227})

	msg3, err := HybridDecrypt(ciphertext3, privKey)
	if err != nil{
		fmt.Printf("ERR: %v\n", err)
	}

	assert.Equal(t, []byte{10, 20}, msg3)
}
