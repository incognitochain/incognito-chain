package devframework

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Account struct {
	PublicKey       string
	PrivateKey      string
	MiningKey       string
	PaymentAddress  string
	CommitteePubkey string
}

func newAccountFromShard(sid int, keyID int) *Account {
	acc, _ := GenerateAddressByShard(sid, keyID)
	return acc
}

func GenerateAddressByShard(shardID int, keyID int) (*Account, error) {
	acc := &Account{}
	key, _ := wallet.NewMasterKey([]byte(fmt.Sprintf("masterkey-%v", shardID)))
	var i int
	var k = 0
	for {
		i++
		child, _ := key.NewChildKey(uint32(i))
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		if child.KeySet.PaymentAddress.Pk[common.PublicKeySize-1] == byte(shardID) {
			acc.PublicKey = base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
			acc.PrivateKey = privAddr
			acc.PaymentAddress = paymentAddress
			validatorKeyBytes := common.HashB(common.HashB(child.KeySet.PrivateKey))
			acc.MiningKey = base58.Base58Check{}.Encode(validatorKeyBytes, common.ZeroByte)
			committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)
			res, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committeeKey})
			acc.CommitteePubkey = res[0]
			if k == keyID {
				break
			}
			k++
		}
		i++
	}
	return acc, nil
}
