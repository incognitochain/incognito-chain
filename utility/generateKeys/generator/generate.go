package generator

import (
	"crypto/rand"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"strconv"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
)

func GenerateAddress(seeds []string) ([]string, error) {
	var pubAddresses []string
	for _, seed := range seeds {
		keySet := (&cashec.KeySet{}).GenerateKey([]byte(seed))
		pubAddress := base58.Base58Check{}.Encode(keySet.PaymentAddress.Pk, common.ZeroByte)
		pubAddresses = append(pubAddresses, pubAddress)
	}
	return pubAddresses, nil
}
func GenerateAddressByShard(shardID byte) ([]string, error) {
	var privateKeys []string
	for i := 200000; i < 230000; i++ {
		seed := strconv.Itoa(i)
		key, _ := wallet.NewMasterKey([]byte(seed))
		var i int
		var k = 0
		for {
			i++
			child, _ := key.NewChildKey(uint32(i))
			privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
			paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
			if child.KeySet.PaymentAddress.Pk[32] == byte(shardID) {
				fmt.Printf("Acc %+v : \n", i)
				fmt.Printf("PublicKey %+v \n ", base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte))
				fmt.Printf("PaymentAddress: %+v \n", paymentAddress)
				fmt.Printf("PrivateKey: %+v \n ", privAddr)
				privateKeys = append(privateKeys, privAddr)
				k++
				if k == 3 {
					break
				}
			}
			i++
		}
	}
	return privateKeys, nil
}
func GenerateAddressByte(seeds [][]byte) ([]string, []string, error) {
	var privateKeys []string
	var pubAddresses []string
	for _, seed := range seeds {
		privateKey := base58.Base58Check{}.Encode(seed, common.ZeroByte)
		privateKeys = append(privateKeys, privateKey)
		keySet := (&cashec.KeySet{}).GenerateKey(seed)
		pubAddress := base58.Base58Check{}.Encode(keySet.PaymentAddress.Pk, common.ZeroByte)
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
