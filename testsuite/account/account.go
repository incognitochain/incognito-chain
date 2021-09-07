package account

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus_v2"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Account struct {
	Name                string
	PublicKey           string
	PrivateKey          string
	MiningKey           string
	MiningPubkey        string
	PaymentAddress      string
	Keyset              *incognitokey.KeySet
	SelfCommitteePubkey string
	MiningKeySet        *signatureschemes.MiningKey
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

func (acc *Account) SetName(n string) {
	acc.Name = n
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

func NewAccountFromPrivatekey(privateKey string) (Account, error) {
	acc := Account{}
	wl, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return Account{}, err
	}
	acc.PrivateKey = privateKey
	acc.PublicKey = wl.KeySet.GetPublicKeyInBase58CheckEncode()
	paymentv2 := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	//paymentv1, err := wallet.GetPaymentAddressV1(paymentv2, false)
	if err != nil {
		return Account{}, err
	}
	acc.PaymentAddress = paymentv2

	validatorKeyBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	acc.MiningKey = base58.Base58Check{}.Encode(validatorKeyBytes, common.ZeroByte)
	privateSeed, err := consensus_v2.GenMiningKeyFromPrivateKey(privateKey)
	acc.MiningKeySet, _ = consensus_v2.GetMiningKeyFromPrivateSeed(privateSeed)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(wl.KeySet.PrivateKey)), wl.KeySet.PaymentAddress.Pk)
	acc.MiningPubkey = committeeKey.GetMiningKeyBase58("bls")
	acc.SelfCommitteePubkey, _ = committeeKey.ToBase58()
	acc.Keyset = &wl.KeySet
	return acc, nil
}

func GenerateAccountByShard(shardID int, keyID int, seed string) (acc Account, err error) {
	key, _ := wallet.NewMasterKey([]byte(fmt.Sprintf("%v-%v", seed, shardID)))
	var i int
	var k = 0
	for {
		i++
		child, _ := key.NewChildKey(uint32(i))
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		if child.KeySet.PaymentAddress.Pk[common.PublicKeySize-1] == byte(shardID) {
			acc, err = NewAccountFromPrivatekey(privAddr)
			if err != nil {
				panic(err)
			}
			if k == keyID {
				break
			}
			k++
		}
		i++
	}
	return acc, nil
}

func BuildCommitteePubkey(minerPrivateKey string, stakerAddr string) (*incognitokey.CommitteePublicKey, error) {
	wl, err := wallet.Base58CheckDeserialize(minerPrivateKey)
	pma := privacy.PublicKey{}
	if stakerAddr == "" {
		pma = wl.KeySet.PaymentAddress.Pk
	} else {
		k, err := wallet.Base58CheckDeserialize(stakerAddr)
		if err != nil {
			return nil, err
		}
		pma = k.KeySet.PaymentAddress.Pk
	}

	committeeKey, err := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(wl.KeySet.PrivateKey)), pma)
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
