package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	wl, _ := wallet.Base58CheckDeserialize("112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f")
	privKeyB58 := wl.Base58CheckSerialize(wallet.PriKeyType)
	wl.KeySet.InitFromPrivateKey(&wl.KeySet.PrivateKey)
	readOnlyKeyB58 := wl.Base58CheckSerialize(wallet.ReadonlyKeyType)
	paymentAddressB58 := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	shardID := common.GetShardIDFromLastByte(wl.KeySet.PaymentAddress.Pk[len(wl.KeySet.PaymentAddress.Pk)-1])
	miningSeed := base58.Base58Check{}.Encode(common.HashB(common.HashB(wl.KeySet.PrivateKey)), common.ZeroByte)
	publicKey := base58.Base58Check{}.Encode(wl.KeySet.PaymentAddress.Pk, common.ZeroByte)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(wl.KeySet.PrivateKey)), wl.KeySet.PaymentAddress.Pk)
	fmt.Println(privKeyB58, readOnlyKeyB58, paymentAddressB58, shardID, miningSeed, publicKey, committeeKey)
	res, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committeeKey})
	fmt.Println(res)

}
