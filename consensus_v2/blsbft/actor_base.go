package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

//Actor
type Actor interface {
	// GetConsensusName - retrieve consensus name
	GetConsensusName() string
	GetChainKey() string
	GetChainID() int
	// GetUserPublicKey - get user public key of loaded mining key
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	// Start - start consensus
	Start() error
	// Stop - stop consensus
	Stop() error
	Destroy()
	// IsOngoing - check whether consensus is currently voting on a block
	IsStarted() bool
	// ProcessBFTMsg - process incoming BFT message
	ProcessBFTMsg(msg *wire.MessageBFT)
	// LoadUserKey - load user mining key
	LoadUserKeys(miningKey []signatureschemes2.MiningKey)
	// ValidateData - validate data with this consensus signature scheme
	ValidateData(data []byte, sig string, publicKey string) error
	// SignData - sign data with this consensus signature scheme
	SignData(data []byte) (string, error)
	BlockVersion() int
	SetBlockVersion(version int)
}

func NewActorWithValue(
	chain Chain, committeeChain CommitteeChainHandler, version int,
	chainID, blockVersion int, chainName string,
	node NodeInterface, logger common.Logger,
) Actor {
	var res Actor = nil
	if version == types.BFT_VERSION {
		res = NewActorV1WithValue(chain, chainName, chainID, node, logger)
	}
	if version >= types.MULTI_VIEW_VERSION {
		res = NewActorV2WithValue(chain, committeeChain, chainName, chainID, blockVersion, node, logger)
<<<<<<< HEAD
	case types.INSTANT_FINALITY_VERSION_V2:
		if chain.IsBeaconChain() {
			res = NewActorV3WithValue(chain, committeeChain, chainName, chainID, blockVersion, node, logger)
		} else {
			res = NewActorV2WithValue(chain, committeeChain, chainName, chainID, blockVersion, node, logger)
		}

	default:
		panic("Bft version is not valid")
=======
	}
	if res == nil {
		panic("Version not recognized")
>>>>>>> reduceblocktime
	}
	return res
}
