package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	wl, _ := wallet.Base58CheckDeserialize("112t8rnmy64Qp7prg3QebvFq8mSYEYFu5Q8seqjYKKoBdsSeUeqC5YhxZUFbMkKtpjit78kdxeVZKHUTnhBf3oDpnGe4TmoTDz861iS36DRy")
	//privKeyB58 := wl.Base58CheckSerialize(wallet.PriKeyType)
	wl.KeySet.InitFromPrivateKey(&wl.KeySet.PrivateKey)
	//readOnlyKeyB58 := wl.Base58CheckSerialize(wallet.ReadonlyKeyType)
	paymentAddressB58 := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	//shardID := common.GetShardIDFromLastByte(wl.KeySet.PaymentAddress.Pk[len(wl.KeySet.PaymentAddress.Pk)-1])
	miningSeed := base58.Base58Check{}.Encode(common.HashB(common.HashB(wl.KeySet.PrivateKey)), common.ZeroByte)
	readOnlyKeyBytes, _ := wl.Serialize(wallet.ReadonlyKeyType)
	readOnlyKey := base58.Base58Check{}.Encode(readOnlyKeyBytes, common.ZeroByte)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(wl.KeySet.PrivateKey)), wl.KeySet.PaymentAddress.Pk)
	res, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committeeKey})
	fmt.Println("Payment Address", paymentAddressB58)
	fmt.Println("ShardID", wl.KeySet.PaymentAddress.Pk[len(wl.KeySet.PaymentAddress.Pk)-1])
	fmt.Println("Mining Seed", miningSeed)
	fmt.Println("Read Only Key: ", readOnlyKey)
	fmt.Println(res)
	fmt.Println("-----------------------")
	str := base58.Base58Check{}.Encode([]byte{}, common.Base58Version)
	fmt.Println(str)
	fmt.Println(base58.Base58Check{}.Decode(str))
}
