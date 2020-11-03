package devframework

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Account struct {
	PublicKey           string
	PrivateKey          string
	MiningKey           string
	MiningPubkey        string
	PaymentAddress      string
	keyset              *incognitokey.KeySet
	SelfCommitteePubkey string
}

func newAccountFromShard(sid int, keyID int) *Account {
	acc, _ := GenerateAddressByShard(sid, keyID)
	return acc
}

func GetShardIDFromPubkey(pk string, args ...interface{}) (int, error) {
	numShard := 256
	for _, arg := range args {
		switch arg.(type) {
		case int:
			numShard = arg.(int)
		}
	}

	sid, _, err := base58.Base58Check{}.Decode(pk)
	if err != nil {
		return -1, err
	}
	return int(sid[len(sid)-1]) % numShard, nil
}

func (acc *Account) ShardID(_numShard ...interface{}) int {
	numShard := 256
	for _, arg := range _numShard {
		switch arg.(type) {
		case int:
			numShard = arg.(int)
		}
	}
	sid, _, err := base58.Base58Check{}.Decode(acc.PublicKey)
	if err != nil {
		return -1
	}
	return int(sid[len(sid)-1]) % numShard
}

func NewAccountFromCommitteePublicKey(committeePK string) (*Account, error) {
	cpk := incognitokey.CommitteePublicKey{}
	err := cpk.FromBase58(committeePK)
	if err != nil {
		return nil, err
	}
	acc := &Account{
		PublicKey:    cpk.GetIncKeyBase58(),
		MiningPubkey: cpk.GetMiningKeyBase58("bls"),
	}
	return acc, nil
}

func NewAccountFromPrivatekey(privateKey string) (*Account, error) {
	acc := &Account{}
	wl, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, err
	}
	acc.PrivateKey = privateKey
	acc.PublicKey = wl.KeySet.GetPublicKeyInBase58CheckEncode()
	acc.PaymentAddress = wl.Base58CheckSerialize(wallet.PaymentAddressType)
	validatorKeyBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	acc.MiningKey = base58.Base58Check{}.Encode(validatorKeyBytes, common.ZeroByte)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(wl.KeySet.PrivateKey)), wl.KeySet.PaymentAddress.Pk)
	acc.MiningPubkey = committeeKey.GetMiningKeyBase58("bls")
	acc.SelfCommitteePubkey, _ = committeeKey.ToBase58()
	acc.keyset = &wl.KeySet
	return acc, nil
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
			acc.keyset = &child.KeySet
			committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)
			acc.SelfCommitteePubkey, _ = committeeKey.ToBase58()
			if k == keyID {
				break
			}
			k++
		}
		i++
	}
	return acc, nil
}

func (acc *Account) BuildCommitteePubkey(rewardAddr string) (*incognitokey.CommitteePublicKey, error) {
	pma := privacy.PublicKey{}
	if rewardAddr == "" {
		pma = acc.keyset.PaymentAddress.Pk
	} else {
		k, err := wallet.Base58CheckDeserialize(rewardAddr)
		if err != nil {
			return nil, err
		}
		pma = k.KeySet.PaymentAddress.Pk
	}

	committeeKey, err := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(acc.keyset.PrivateKey)), pma)
	return &committeeKey, err
}

func (acc *Account) ParseCommitteePubkey(committeeKey string) (*incognitokey.CommitteePublicKey, error) {
	cm := &incognitokey.CommitteePublicKey{}
	err := cm.FromBase58(committeeKey)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func IncPubkeyFromPaymentAddr(paymentAddr string) string {
	wl, _ := wallet.Base58CheckDeserialize(paymentAddr)
	return base58.Base58Check{}.Encode(wl.KeySet.PaymentAddress.Pk, common.ZeroByte)
}
