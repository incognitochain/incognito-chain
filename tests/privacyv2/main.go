package main

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	shard0UrlListWithBeacon = []string{
		"http://localhost:9334",
		"http://localhost:9335",
		"http://localhost:9336",
		"http://localhost:9337",
		"http://localhost:9331",
		"http://localhost:9332",
		"http://localhost:9333",
		"http://localhost:9349",
		"http://localhost:9350",
		"http://localhost:9351",
		"http://localhost:9352",
		"http://localhost:9353",
	}
	shard1UrlListWithBeacon = []string{
		"http://localhost:9338",
		"http://localhost:9339",
		"http://localhost:9340",
		"http://localhost:9341",
		"http://localhost:9342",
		"http://localhost:9343",
		"http://localhost:9344",
		"http://localhost:9345",
		"http://localhost:9350",
		"http://localhost:9351",
		"http://localhost:9352",
		"http://localhost:9353",
	}
	shard0UrlList = []string{
		"http://localhost:9334",
		"http://localhost:9335",
		"http://localhost:9336",
		"http://localhost:9337",
		"http://localhost:9331",
		"http://localhost:9332",
		"http://localhost:9333",
		"http://localhost:9349",
	}
	shard1UrlList = []string{
		"http://localhost:9338",
		"http://localhost:9339",
		"http://localhost:9340",
		"http://localhost:9341",
		"http://localhost:9342",
		"http://localhost:9343",
		"http://localhost:9344",
		"http://localhost:9345",
	}
)

func main() {
	privateKeyShard = append(privateKeyShard, privateKeyShard0...)
	privateKeyShard = append(privateKeyShard, privateKeyShard1...)

	if os.Args[1] == "submit" || os.Args[1] == "all" {
		submitKeyShard0_0()
		submitKeyShard0_1()
		submitKeyShard0_2()
		submitKeyShard0_3()
		submitKeyShard1_1()
		submitKeyShard1_2()
		submitKeyShard1_3()
	}
	if os.Args[1] == "convert" || os.Args[1] == "all" {
		convertShard0_0()
		convertShard0_1()
		convertShard0_2()
		convertShard0_3()
		convertShard1_1()
		convertShard1_2()
		convertShard1_3()
	}
	if os.Args[1] == "send-tx" {
		sendTransactionFromTestnetGenesisKeyFromShard0_0()
		sendTransactionFromTestnetGenesisKeyFromShard0_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_0()
		sendTransactionFromTestnetGenesisKeyFromShard1_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_2()
		sendTransactionToShard1()
	}
	if os.Args[1] == "flush-tx" {
		flushTx()
	}

	if os.Args[1] == "stake" {
		var numberOfStaker int
		var err error
		if len(os.Args) == 2 {
			numberOfStaker = len(stakePayloads)
		} else {
			numberOfStaker, err = strconv.Atoi(os.Args[2])
			if err != nil {
				numberOfStaker = len(stakePayloads)
			}
		}
		stake(numberOfStaker)
	}
	if os.Args[1] == "stop-auto-stake" {
		var numberOfStaker int
		var err error
		if len(os.Args) == 2 {
			numberOfStaker = len(stopAutoStakePayloads)
		} else {
			numberOfStaker, err = strconv.Atoi(os.Args[2])
			if err != nil {
				numberOfStaker = len(stopAutoStakePayloads)
			}
		}
		stopAutoStake(numberOfStaker)
	}
	if os.Args[1] == "unstake" {
		var numberOfStaker int
		var err error
		if len(os.Args) == 2 {
			numberOfStaker = len(unstakePayloads)
		} else {
			numberOfStaker, err = strconv.Atoi(os.Args[2])
			if err != nil {
				numberOfStaker = len(unstakePayloads)
			}
		}
		unstake(numberOfStaker)
	}

	if os.Args[1] == "consensus_interval" {
		numberOfStaker := len(stakePayloads)
		stake(numberOfStaker)
		time.Sleep(5 * time.Minute)
		numberOfUnstake := len(unstakePayloads)
		unstake(numberOfUnstake)
		time.Sleep(1 * time.Minute)
		numberOfStopAutoStake := len(stopAutoStakePayloads)
		stopAutoStake(numberOfStopAutoStake)
		time.Sleep(5 * time.Minute)
	}
}

func flushTx() {
	ticker := time.Tick(500 * time.Millisecond)
	for _ = range ticker {
		sendTransactionFromTestnetGenesisKeyFromShard0_0()
		sendTransactionFromTestnetGenesisKeyFromShard0_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_0()
		sendTransactionFromTestnetGenesisKeyFromShard1_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_2()
		sendTransactionToShard1()
	}
}

var (
	privateKeyShard0 = []string{
		"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
		"112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
		"112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX",
		"112t8sSj637mhpaJUboUEjkXsEUQm8q82T6kND3mWtNwig71qX2aFeZegWYsLVtyxBWdiZMBoNkdJ1MZYAcWetUP8DjYFnUac4vW7kzHfYsc",
		"112t8sSjSEck5J5RKWGWurVfigruYDxEzjjVPqHaTRJ57YFNo7gXBH8onUQxtdpoyFnBZrLhfGWQ4k4MNadwa6F7qYwcuFLW9R1VxTfN7q4d",
		"112t8sSijugr8azxAiHMWS9rA22grKDv5o7AEXQ9datpT1V7N5FLHiJMjvfVnXcitL3fpj35Xt5DNnBq8iFq618X31nCgn2RjrYx5tZZWCtj",
	}
	privateKeyShard1 = []string{
		"112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL",
		"112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC",
		"112t8rtt9Kd5LUcfXNmd7aMnQehCnKabArVB3BUk2RHVjeh88x5MJnJY4okB8JdFm4JNm4A2WjSe58qWNVkJPEFjpLHNYfKHpWfRdqyfDD9f",
		"112t8sSjkAqVJi4KkbCS75GYrsag7QZYP7FTPRfZ63D1AJgzfmdHnE9sbpdJV4Kx5tN9MgbqbRYDgzER2xpgsxrHvWxNgTHHrghYwLJLfe2R",
	}
	privateKeyShard = []string{}
)

// PrivateKeyToPaymentAddress returns the payment address for its private key corresponding to the key type.
// KeyType should be -1, 0, 1 where
//	- -1: payment address of version 2
//	- 0: payment address of version 1 with old encoding
//	- 1: payment address of version 1 with new encoding
func PrivateKeyToPaymentAddress(privateKey string, keyType int) string {
	keyWallet, _ := wallet.Base58CheckDeserialize(privateKey)
	err := keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		return ""
	}
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	switch keyType {
	case 0: //Old address, old encoding
		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, false)
		return addr
	case 1:
		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, true)
		return addr
	default:
		return paymentAddStr
	}
}

// PrivateKeyToPublicKey returns the public key of a private key.
//
// If the private key is invalid, it returns nil.
func PrivateKeyToPublicKey(privateKey string) []byte {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		return nil
	}
	return keyWallet.KeySet.PaymentAddress.Pk
}

// PrivateKeyToPrivateOTAKey returns the private OTA key of a private key.
//
// If the private key is invalid, it returns an empty string.
func PrivateKeyToPrivateOTAKey(privateKey string) string {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		log.Println("no private key found")
		return ""
	}
	return keyWallet.Base58CheckSerialize(wallet.OTAKeyType)
}

// PrivateKeyToReadonlyKey returns the readonly key of a private key.
//
// If the private key is invalid, it returns an empty string.
func PrivateKeyToReadonlyKey(privateKey string) string {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		log.Println("no private key found")
		return ""
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	return keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
}

// PrivateKeyToMiningKey returns the mining key of a private key.
func PrivateKeyToMiningKey(privateKey string) string {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return ""
	}
	miningKey := base58.Base58Check{}.Encode(common.HashB(common.HashB(keyWallet.KeySet.PrivateKey)), common.ZeroByte)
	return miningKey
}
