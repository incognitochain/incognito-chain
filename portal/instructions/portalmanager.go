package instructions

import (
	bMeta "github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	portalMeta "github.com/incognitochain/incognito-chain/portal/metadata"
)

type relayingProcessor interface {
	getActions() [][]string
	putAction(action []string)
	buildRelayingInst(
		blockchain *BlockChain,
		relayingHeaderAction portalMeta.RelayingHeaderAction,
		relayingState *RelayingHeaderChainState,
	) [][]string
	buildHeaderRelayingInst(
		senderAddressStr string,
		header string,
		blockHeight uint64,
		metaType int,
		shardID byte,
		txReqID common.Hash,
		status string,
	) []string
}

type portalInstructionProcessor interface {
	getActions() map[byte][][]string
	putAction(action []string, shardID byte)
	prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error)
	buildNewInsts(
		bc bMeta.ChainRetriever,
		contentStr string,
		shardID byte,
		currentPortalState *CurrentPortalState,
		beaconHeight uint64,
		portalParams portal.PortalParams,
		optionalData map[string]interface{},
	) ([][]string, error)
}

type portalInstProcessor struct {
	actions map[byte][][]string
}

type portalManager struct {
	relayingChains     map[int]relayingProcessor
	portalInstructions map[int]portalInstructionProcessor
}

func NewPortalManager() *portalManager {
	rbnbChain := &relayingBNBChain{
		relayingChain: &relayingChain{
			actions: [][]string{},
		},
	}
	rbtcChain := &relayingBTCChain{
		relayingChain: &relayingChain{
			actions: [][]string{},
		},
	}

	relayingChainProcessor := map[int]relayingProcessor{
		bMeta.RelayingBNBHeaderMeta: rbnbChain,
		bMeta.RelayingBTCHeaderMeta: rbtcChain,
	}

	portalInstProcessor := map[int]portalInstructionProcessor{
		bMeta.PortalExchangeRatesMeta: &portalExchangeRateProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		bMeta.PortalCustodianDepositMeta: &portalCustodianDepositProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalCustodianDepositMetaV3: &portalCustodianDepositProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalCustodianWithdrawRequestMeta: &portalRequestWithdrawCollateralProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalCustodianWithdrawRequestMetaV3: &portalRequestWithdrawCollateralProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalRequestPortingMeta: &portalPortingRequestProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalUserRequestPTokenMeta: &portalRequestPTokenProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		bMeta.PortalRedeemRequestMeta: &portalRedeemRequestProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		bMeta.PortalReqMatchingRedeemMeta: &portalRequestMatchingRedeemProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalRequestUnlockCollateralMeta: &portalRequestUnlockCollateralProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalRequestUnlockCollateralMetaV3: &portalRequestUnlockCollateralProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalRequestWithdrawRewardMeta: &portalReqWithdrawRewardProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalRedeemFromLiquidationPoolMeta: &portalRedeemFromLiquidationPoolProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalCustodianTopupMetaV2: &portalCustodianTopupProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalTopUpWaitingPortingRequestMeta: &portalTopupWaitingPortingReqProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalRedeemFromLiquidationPoolMetaV3: &portalRedeemFromLiquidationPoolProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalCustodianTopupMetaV3: &portalCustodianTopupProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		bMeta.PortalTopUpWaitingPortingRequestMetaV3: &portalTopupWaitingPortingReqProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
	}

	return &portalManager{
		relayingChains:     relayingChainProcessor,
		portalInstructions: portalInstProcessor,
	}
}
