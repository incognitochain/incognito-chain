package main

import (
	"fmt"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"os"
	"strconv"
)

func main() {
	param, _ := strconv.Atoi(os.Args[1])
	// dcb: 100000000
	// gov: 10000000
	// burn: 1000000
	//param = 10000000
	burnPubKeyE := privacy.PedCom.G[0].Hash(param)
	burnPubKey := burnPubKeyE.Compress()
	burnKey := wallet.KeyWallet{
		KeySet: cashec.KeySet{
			PaymentAddress: privacy.PaymentAddress{
				Pk: burnPubKey,
			},
		},
	}
	burnPaymentAddress := burnKey.Base58CheckSerialize(wallet.PaymentAddressType)
	fmt.Printf("Burn payment address : %s \n", burnPaymentAddress)
	os.Exit(0)

	/*keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	fmt.Println("======================================")
	fmt.Println(keyWalletBurningAdd.KeySet.PaymentAddress.Pk)
	fmt.Println("======================================")*/
}
