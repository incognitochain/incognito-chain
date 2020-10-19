package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type relayingProcessor interface {
	getActions() [][]string
	putAction(action []string)
	buildRelayingInst(
		blockchain *BlockChain,
		relayingHeaderAction metadata.RelayingHeaderAction,
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
		bc *BlockChain,
		contentStr string,
		shardID byte,
		currentPortalState *CurrentPortalState,
		beaconHeight uint64,
		portalParams PortalParams,
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
		metadata.RelayingBNBHeaderMeta: rbnbChain,
		metadata.RelayingBTCHeaderMeta: rbtcChain,
	}

	portalInstProcessor := map[int]portalInstructionProcessor{
		metadata.PortalExchangeRatesMeta: &portalExchangeRateProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		metadata.PortalCustodianDepositMeta: &portalCustodianDepositProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianDepositMetaV3: &portalCustodianDepositProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianWithdrawRequestMeta: &portalRequestWithdrawCollateralProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianWithdrawRequestMetaV3: &portalRequestWithdrawCollateralProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalUserRegisterMeta: &portalPortingRequestProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalUserRequestPTokenMeta: &portalRequestPTokenProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		metadata.PortalRedeemRequestMeta: &portalRedeemRequestProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		metadata.PortalRedeemRequestMetaV3: &portalRedeemRequestProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},

		metadata.PortalReqMatchingRedeemMeta: &portalRequestMatchingRedeemProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalRequestUnlockCollateralMeta: &portalRequestUnlockCollateralProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalRequestWithdrawRewardMeta: &portalReqWithdrawRewardProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalRedeemFromLiquidationPoolMeta: &portalRedeemFromLiquidationPoolProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianTopupMetaV2: &portalCustodianTopupProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalTopUpWaitingPortingRequestMeta: &portalTopupWaitingPortingReqProcessor{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalRedeemFromLiquidationPoolMetaV3: &portalRedeemFromLiquidationPoolProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianTopupMetaV3: &portalCustodianTopupProcessorV3{
			portalInstProcessor: &portalInstProcessor{
				actions: map[byte][][]string{},
			},
		},
		metadata.PortalTopUpWaitingPortingRequestMetaV3: &portalTopupWaitingPortingReqProcessorV3{
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
