package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestElGamalEncryption(t *testing.T) {
	privKey := new(ElGamalPrivKey)
	privKey.X = RandScalar()

	pubKey := privKey.GenPubKey()

	message := new(EllipticPoint)
	message.Randomize()

	ciphertext1 := pubKey.Encrypt(message)

	ciphertext2 := new(ElGamalCiphertext)
	ciphertext2.SetBytes(ciphertext1.Bytes())

	decryptedCiphertext, err := privKey.Decrypt(ciphertext2)
	if err != nil{
		fmt.Printf("Err: %v\n", err)
	}

	privateKey := GeneratePrivateKey(new(big.Int).SetInt64(123).Bytes())
	receivingKey := GenerateReceivingKey(privateKey)
	myprivKey := new(ElGamalPrivKey)
	myprivKey.X = new(big.Int).SetBytes(receivingKey)

	ciphertext3 := new(ElGamalCiphertext)
	bytes := []byte{3, 87, 168, 233, 184, 99, 27, 41, 20, 102, 254, 249, 199, 143, 50, 16, 225, 202, 172, 93, 198, 244, 175, 216, 135, 5, 72, 219, 103, 155, 157, 123, 208, 2, 106, 157, 179, 6, 237, 201, 102, 143, 176, 234, 237, 65, 194, 79, 123, 4, 199, 143, 253, 94, 223, 95, 51, 125, 252, 211, 109, 62, 224, 60, 209, 207}
	//err := ciphertext3.SetBytes([]byte{3, 71, 59, 137, 15, 206, 74, 44, 104, 244, 133, 79, 59, 133, 32, 187, 246, 115, 102, 165, 150, 64, 15, 222, 172, 244, 16, 62, 105, 29, 61, 75, 100, 3, 21, 78, 180, 205, 172, 216, 8, 39, 71, 156, 154, 122, 166, 165, 150, 33, 10, 168, 81, 224, 37, 180, 207, 36, 5, 29, 151, 84, 238, 226, 67, 37})
	err = ciphertext3.SetBytes(bytes)
	fmt.Printf("error: %v\n", err)
	fmt.Printf("ciphertext after set byte C1: %+v\n", ciphertext3.C1)
	fmt.Printf("ciphertext after set byte C2: %+v\n", ciphertext3.C2)

	myplaintext, err := myprivKey.Decrypt(ciphertext3)
	if err != nil{
		fmt.Printf("Err: %v\n", err)
	}
	fmt.Printf("my plain text: %+v\n", myplaintext)

	assert.Equal(t, message, decryptedCiphertext)
}
