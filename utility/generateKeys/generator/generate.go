package generator

import (
	"crypto/rand"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common/base58"
)

func GenerateAddress(seeds []string) ([]string, error) {
	var pubAddresses []string
	for _, seed := range seeds {
		keySet := (&cashec.KeySet{}).GenerateKey([]byte(seed))
		pubAddress := base58.Base58Check{}.Encode(keySet.PaymentAddress.Pk, byte(0x00))
		pubAddresses = append(pubAddresses, pubAddress)
	}
	return pubAddresses, nil
}
func GenerateAddressByte(seeds [][]byte) ([]string, []string, error) {
	var privateKeys []string
	var pubAddresses []string
	for _, seed := range seeds {
		privateKey := base58.Base58Check{}.Encode(seed, byte(0x00))
		privateKeys = append(privateKeys, privateKey)
		keySet := (&cashec.KeySet{}).GenerateKey(seed)
		pubAddress := base58.Base58Check{}.Encode(keySet.PaymentAddress.Pk, byte(0x00))
		pubAddresses = append(pubAddresses, pubAddress)
	}
	return privateKeys, pubAddresses, nil
}
func GenerateKeyPair() [][]byte {
	seed := [][]byte{}
	i := 0
	for i < 1024 {
		token := make([]byte, 33)
		rand.Read(token)
		seed = append(seed, token[:])
		i++
	}
	return seed
}
