package blsbft

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"testing"
)

func TestMiningKey_GetPublicKey(t *testing.T) {
	for i := 0; i < 20; i++ {
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

func newMiningKey(privateSeed string) (*MiningKey, error) {
	var miningKey MiningKey
	privateSeedBytes, _, err := base58.Base58Check{}.Decode(privateSeed)
	if err != nil {
		return nil, consensus.NewConsensusError(consensus.LoadKeyError, err)
	}

	blsPriKey, blsPubKey := blsmultisig.KeyGen(privateSeedBytes)

	// privateKey := blsmultisig.B2I(privateKeyBytes)
	// publicKeyBytes := blsmultisig.PKBytes(blsmultisig.PKGen(privateKey))
	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[common.BlsConsensus] = blsmultisig.SKBytes(blsPriKey)
	miningKey.PubKey[common.BlsConsensus] = blsmultisig.PKBytes(blsPubKey)
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateSeedBytes)
	miningKey.PriKey[common.BridgeConsensus] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[common.BridgeConsensus] = bridgesig.PKBytes(&bridgePubKey)
	return &miningKey, nil
}

func Test_newMiningKey(t *testing.T) {
	type args struct {
		privateSeed string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Get mining key from private seed",
			args: args{
				privateSeed: "1Md5Jd3syKLygiphTyXZGLQFswsbgPpVfchYfiVrHX86A6Zsyn",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if key, err := newMiningKey(tt.args.privateSeed); (err != nil) != tt.wantErr {
				t.Errorf("newMiningKey() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				fmt.Println("BLS Key:", base58.Base58Check{}.Encode(key.PubKey[common.BlsConsensus], common.Base58Version))
				fmt.Println("BRI Key:", base58.Base58Check{}.Encode(key.PubKey[common.BridgeConsensus], common.Base58Version))
			}
		})
	}
}
