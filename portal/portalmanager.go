package portal

import (
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

type PortalManager struct {
	RelayingChainsProcessors map[int]portalrelaying.RelayingProcessor
	PortalInstProcessorsV3   map[int]portalprocessv3.PortalInstructionProcessorV3
	PortalInstProcessorsV4   map[int]portalprocessv4.PortalInstructionProcessorV4
}

func NewPortalManager() *PortalManager {
	rbnbChain := &portalrelaying.RelayingBNBChain{
		RelayingChain: &portalrelaying.RelayingChain{
			Actions: [][]string{},
		},
	}
	rbtcChain := &portalrelaying.RelayingBTCChain{
		RelayingChain: &portalrelaying.RelayingChain{
			Actions: [][]string{},
		},
	}

	relayingChainProcessor := map[int]portalrelaying.RelayingProcessor{
		metadata.RelayingBNBHeaderMeta: rbnbChain,
		metadata.RelayingBTCHeaderMeta: rbtcChain,
	}

	portalInstProcessorV3 := map[int]portalprocessv3.PortalInstructionProcessorV3{
		metadata.PortalExchangeRatesMeta: &portalprocessv3.PortalExchangeRateProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianDepositMeta: &portalprocessv3.PortalCustodianDepositProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianDepositMetaV3: &portalprocessv3.PortalCustodianDepositProcessorV3{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianWithdrawRequestMeta: &portalprocessv3.PortalRequestWithdrawCollateralProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianWithdrawRequestMetaV3: &portalprocessv3.PortalRequestWithdrawCollateralProcessorV3{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalRequestPortingMetaV3: &portalprocessv3.PortalPortingRequestProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalUserRequestPTokenMeta: &portalprocessv3.PortalRequestPTokenProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},

		metadata.PortalRedeemRequestMetaV3: &portalprocessv3.PortalRedeemRequestProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},

		metadata.PortalReqMatchingRedeemMeta: &portalprocessv3.PortalRequestMatchingRedeemProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalRequestUnlockCollateralMetaV3: &portalprocessv3.PortalRequestUnlockCollateralProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalRequestWithdrawRewardMeta: &portalprocessv3.PortalReqWithdrawRewardProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalRedeemFromLiquidationPoolMeta: &portalprocessv3.PortalRedeemFromLiquidationPoolProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianTopupMetaV2: &portalprocessv3.PortalCustodianTopupProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalTopUpWaitingPortingRequestMeta: &portalprocessv3.PortalTopupWaitingPortingReqProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalRedeemFromLiquidationPoolMetaV3: &portalprocessv3.PortalRedeemFromLiquidationPoolProcessorV3{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalCustodianTopupMetaV3: &portalprocessv3.PortalCustodianTopupProcessorV3{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalTopUpWaitingPortingRequestMetaV3: &portalprocessv3.PortalTopupWaitingPortingReqProcessorV3{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalPickMoreCustodianForRedeemMeta: &portalprocessv3.PortalPickMoreCustodianForRedeemProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalLiquidateCustodianMetaV3: &portalprocessv3.PortalLiquidationCustodianRunAwayProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalExpiredWaitingPortingReqMeta: &portalprocessv3.PortalExpiredWaitingPortingProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalLiquidateTPExchangeRatesMeta: &portalprocessv3.PortalLiquidationByRatesProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalLiquidateByRatesMetaV3: &portalprocessv3.PortalLiquidationByRatesV3Processor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalUnlockOverRateCollateralsMeta: &portalprocessv3.PortalCusUnlockOverRateCollateralsProcessor{
			PortalInstProcessorV3: &portalprocessv3.PortalInstProcessorV3{
				Actions: map[byte][][]string{},
			},
		},
	}

	portalInstProcessorV4 := map[int]portalprocessv4.PortalInstructionProcessorV4{
		metadata.PortalV4ShieldingRequestMeta: &portalprocessv4.PortalShieldingRequestProcessor{
			PortalInstProcessorV4: &portalprocessv4.PortalInstProcessorV4{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalV4UnshieldingRequestMeta: &portalprocessv4.PortalUnshieldRequestProcessor{
			PortalInstProcessorV4: &portalprocessv4.PortalInstProcessorV4{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalV4UnshieldBatchingMeta: &portalprocessv4.PortalUnshieldBatchingProcessor{
			PortalInstProcessorV4: &portalprocessv4.PortalInstProcessorV4{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalV4FeeReplacementRequestMeta: &portalprocessv4.PortalFeeReplacementRequestProcessor{
			PortalInstProcessorV4: &portalprocessv4.PortalInstProcessorV4{
				Actions: map[byte][][]string{},
			},
		},
		metadata.PortalV4SubmitConfirmedTxMeta: &portalprocessv4.PortalSubmitConfirmedTxProcessor{
			PortalInstProcessorV4: &portalprocessv4.PortalInstProcessorV4{
				Actions: map[byte][][]string{},
			},
		},
	}

	return &PortalManager{
		RelayingChainsProcessors: relayingChainProcessor,
		PortalInstProcessorsV3:   portalInstProcessorV3,
		PortalInstProcessorsV4:   portalInstProcessorV4,
	}
}
