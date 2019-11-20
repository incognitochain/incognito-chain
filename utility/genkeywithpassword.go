package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	numberOfKey := 6                       // Number of keyset that you want to be generated
	randomString := []byte("abcde")        // A random string used to create keyset. The same string create the same keyset
	numberOfShard := common.MaxShardNumber // Number of Shard in Incognito Chain
	// numberOfShard := 2

	for j := 0; j < numberOfShard; j++ {
		privKeyLs := make([]string, 0)
		readOnlyKeyLs := make([]string, 0)
		paymentAddLs := make([]string, 0)
		miningSeedLs := make([]string, 0)
		publicKeyLs := make([]string, 0)
		committeeKeyLs := make([]string, 0)
		for i := 0; i < 1000; i++ {
			masterKey, _ := wallet.NewMasterKey(randomString)
			child, _ := masterKey.NewChildKey(uint32(i))

			privKeyB58 := child.Base58CheckSerialize(wallet.PriKeyType)
			readOnlyKeyB58 := child.Base58CheckSerialize(wallet.ReadonlyKeyType)
			paymentAddressB58 := child.Base58CheckSerialize(wallet.PaymentAddressType)
			shardID := common.GetShardIDFromLastByte(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])
			miningSeed := base58.Base58Check{}.Encode(common.HashB(common.HashB(child.KeySet.PrivateKey)), common.ZeroByte)
			publicKey := base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
			committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)

			if int(shardID) == j {

				privKeyLs = append(privKeyLs, (privKeyB58))
				readOnlyKeyLs = append(readOnlyKeyLs, (readOnlyKeyB58))
				paymentAddLs = append(paymentAddLs, (paymentAddressB58))
				miningSeedLs = append(miningSeedLs, (miningSeed))
				publicKeyLs = append(publicKeyLs, (publicKey))
				temp, _ := committeeKey.ToBase58()
				committeeKeyLs = append(committeeKeyLs, (temp))
				if len(privKeyLs) >= numberOfKey {
					break
				}
			}
		}

		fmt.Printf("\n\n\n ***** Shard %+v **** \n\n\n", j)
		for i := 0; i < len(privKeyLs); i++ {
			fmt.Println(i)
			fmt.Println("Payment Address  : " + paymentAddLs[i])
			fmt.Println("Private Key      : " + privKeyLs[i])
			fmt.Println("Public key       : " + publicKeyLs[i])
			fmt.Println("ReadOnly key     : " + readOnlyKeyLs[i])
			fmt.Println("Validator key    : " + miningSeedLs[i])
			fmt.Println("Committee key set: " + committeeKeyLs[i])
			fmt.Println("------------------------------------------------------------")
		}
	}
}
