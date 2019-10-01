package blsbft

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"testing"
)

func TestMiningKey_GetPublicKey(t *testing.T) {
	for i:= 0; i <20; i++ {
		seed := privacy.RandomScalar().ToBytesS()
		masterKey, _ := wallet.NewMasterKey(seed)

		child, _ := masterKey.NewChildKey(uint32(i))
		privKeyB58 := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddressB58 := child.Base58CheckSerialize(wallet.PaymentAddressType)
		publicKeyB58 := child.KeySet.GetPublicKeyInBase58CheckEncode()

		fmt.Println(privKeyB58)
		fmt.Println(publicKeyB58)
		fmt.Println(paymentAddressB58)

		blsBft := BLSBFT{}
		privateSeed, _ := blsBft.LoadUserKeyFromIncPrivateKey(privKeyB58)

		fmt.Println(privateSeed)
		fmt.Println()
	}
}
