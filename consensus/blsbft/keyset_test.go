package blsbft

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"testing"
)

func TestMiningKey_GetKeyTuble(t *testing.T) {
	lenOutput := 500
	for j := 0; j < common.MaxShardNumber; j++ {
		privKeyLs := make([]string, 0)
		paymentAddLs := make([]string, 0)
		for i := 0; i < 10000; i++ {
			seed := privacy.RandomScalar().ToBytesS()
			masterKey, _ := wallet.NewMasterKey(seed)
			child, _ := masterKey.NewChildKey(uint32(i))
			privKeyB58 := child.Base58CheckSerialize(wallet.PriKeyType)
			paymentAddressB58 := child.Base58CheckSerialize(wallet.PaymentAddressType)
			shardID := common.GetShardIDFromLastByte(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])

			//viewingKeyB58 := child.Base58CheckSerialize(wallet.ReadonlyKeyType)
			//publicKeyB58 := child.KeySet.GetPublicKeyInBase58CheckEncode()

			//fmt.Println("privKeyB58: ", privKeyB58)
			//fmt.Println("publicKeyB58: ", publicKeyB58)
			//fmt.Println("paymentAddressB58: ", paymentAddressB58)
			//fmt.Println("viewingKeyB58: ", viewingKeyB58)

			//blsBft := BLSBFT{}
			//privateSeed, _ := blsBft.LoadUserKeyFromIncPrivateKey(privKeyB58)

			//fmt.Println("privateSeed: ", privateSeed)
			//fmt.Println()
			if int(shardID) == j {

				privKeyLs = append(privKeyLs, strconv.Quote(privKeyB58))
				paymentAddLs = append(paymentAddLs, strconv.Quote(paymentAddressB58))
				if len(privKeyLs) >= lenOutput {
					break
				}
			}
		}
		fmt.Println("privKeyLs"+ strconv.Itoa(j)," = [", strings.Join(privKeyLs, ", "), "]")
		fmt.Println("paymentAddLs" + strconv.Itoa(j), " = [",  strings.Join(paymentAddLs, ", "), "]")
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
