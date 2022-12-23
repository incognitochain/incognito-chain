package consensus

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type MiningState struct {
	Role             string
	Layer            string
	ChainID          int
	IsBeaconFullnode bool
}

type Validator struct {
	MiningKey   signatureschemes.MiningKey
	PrivateSeed string
	State       MiningState
}

func (validator *Validator) IncMiningPublicKey() *incognitokey.CommitteePublicKey {
	committeePublicKey := new(incognitokey.CommitteePublicKey)
	committeePublicKey.IncPubKey = []byte{}
	committeePublicKey.MiningPubKey = map[string][]byte{}
	_, blsPubKey := blsmultisig.KeyGen([]byte(validator.PrivateSeed))
	blsPubKeyBytes := blsmultisig.PKBytes(blsPubKey)
	committeePublicKey.MiningPubKey[common.BlsConsensus] = blsPubKeyBytes
	_, briPubKey := bridgesig.KeyGen([]byte(validator.PrivateSeed))
	briPubKeyBytes := bridgesig.PKBytes(&briPubKey)
	committeePublicKey.MiningPubKey[common.BridgeConsensus] = briPubKeyBytes
	return committeePublicKey
}
