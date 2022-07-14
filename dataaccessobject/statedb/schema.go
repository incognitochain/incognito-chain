package statedb

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
)

var (
	blockHashByIndexPrefix             = []byte("block-hash-by-index-")
	committeePrefix                    = []byte("shard-com-")
	substitutePrefix                   = []byte("shard-sub-")
	nextShardCandidatePrefix           = []byte("next-sha-cand-")
	currentShardCandidatePrefix        = []byte("cur-sha-cand-")
	nextBeaconCandidatePrefix          = []byte("next-bea-cand-")
	currentBeaconCandidatePrefix       = []byte("cur-bea-cand-")
	committeeRewardPrefix              = []byte("committee-reward-")
	slashingCommitteePrefix            = []byte("slashing-committee-")
	rewardRequestPrefix                = []byte("reward-request-")
	blackListProducerPrefix            = []byte("black-list-")
	serialNumberPrefix                 = []byte("serial-number-")
	commitmentPrefix                   = []byte("com-value-")
	commitmentIndexPrefix              = []byte("com-index-")
	commitmentLengthPrefix             = []byte("com-length-")
	snDerivatorPrefix                  = []byte("sn-derivator-")
	outputCoinPrefix                   = []byte("output-coin-")
	otaCoinPrefix                      = []byte("ota-coin-")
	otaCoinIndexPrefix                 = []byte("ota-index-")
	otaCoinLengthPrefix                = []byte("ota-length-")
	onetimeAddressPrefix               = []byte("onetime-address-")
	tokenPrefix                        = []byte("token-")
	tokenTransactionPrefix             = []byte("token-transaction-")
	waitingPDEContributionPrefix       = []byte("waitingpdecontribution-")
	pdePoolPrefix                      = []byte("pdepool-")
	pdeSharePrefix                     = []byte("pdeshare-")
	pdeTradingFeePrefix                = []byte("pdetradingfee-")
	pdeTradeFeePrefix                  = []byte("pdetradefee-")
	pdeContributionStatusPrefix        = []byte("pdecontributionstatus-")
	pdeTradeStatusPrefix               = []byte("pdetradestatus-")
	pdeWithdrawalStatusPrefix          = []byte("pdewithdrawalstatus-")
	pdeStatusPrefix                    = []byte("pdestatus-")
	bridgeEthTxPrefix                  = []byte("bri-eth-tx-")
	bridgeBSCTxPrefix                  = []byte("bri-bsc-tx-")
	bridgePLGTxPrefix                  = []byte("bri-plg-tx-")
	bridgeFTMTxPrefix                  = []byte("bri-ftm-tx-")
	bridgePRVEVMPrefix                 = []byte("bri-prv-evm-tx-")
	bridgeCentralizedTokenInfoPrefix   = []byte("bri-cen-token-info-")
	bridgeDecentralizedTokenInfoPrefix = []byte("bri-de-token-info-")
	bridgeStatusPrefix                 = []byte("bri-status-")
	burnPrefix                         = []byte("burn-")
	syncingValidatorsPrefix            = []byte("syncing-validators-")
	stakerInfoPrefix                   = common.HashB([]byte("stk-info-"))[:prefixHashKeyLength]

	// pdex v3
	pdexv3StatusPrefix                      = []byte("pdexv3-status-")
	pdexv3ParamsModifyingPrefix             = []byte("pdexv3-paramsmodifyingstatus-")
	pdexv3TradeStatusPrefix                 = []byte("pdexv3-trade-status-")
	pdexv3AddOrderStatusPrefix              = []byte("pdexv3-addorder-status-")
	pdexv3WithdrawOrderStatusPrefix         = []byte("pdexv3-withdraworder-status-")
	pdexv3ParamsPrefix                      = []byte("pdexv3-params-")
	pdexv3WaitingContributionsPrefix        = []byte("pdexv3-waitingContributions-")
	pdexv3DeletedWaitingContributionsPrefix = []byte("pdexv3-deletedwaitingContributions-")
	pdexv3PoolPairsPrefix                   = []byte("pdexv3-poolpairs-")
	pdexv3SharesPrefix                      = []byte("pdexv3-shares-")
	pdexv3WithdrawalLPFeePrefix             = []byte("pdexv3-withdrawallpfeestatus-")
	pdexv3WithdrawalProtocolFeePrefix       = []byte("pdexv3-withdrawalprotocolfeestatus-")
	pdexv3WithdrawalStakingRewardPrefix     = []byte("pdexv3-withdrawalstakingrewardstatus-")
	pdexv3OrdersPrefix                      = []byte("pdexv3-orders-")
	pdexv3MintNftPrefix                     = []byte("pdexv3-nfts-")
	pdexv3WithdrawLiquidityStatusPrefix     = []byte("pdexv3-withdrawliquidity-statuses-")
	pdexv3WaitingContributionStatusPrefix   = []byte("pdexv3-waitingContribution-statuses-")
	pdexv3StakingPoolsPrefix                = []byte("pdexv3-stakingpools-")
	pdexv3StakerPrefix                      = []byte("pdexv3-staker-")
	pdexv3StakingStatusPrefix               = []byte("pdexv3-staking-status-")
	pdexv3UnstakingStatusPrefix             = []byte("pdexv3-unstaking-status-")
	pdexv3UserMintNftStatusPrefix           = []byte("pdexv3-usermintnft-status-")
	pdexv3PoolPairLpFeePerSharePrefix       = []byte("pdexv3-poolpair-lpfeepershare-")
	pdexv3PoolPairLmRewardPerSharePrefix    = []byte("pdexv3-poolpair-lmewardpershare-")
	pdexv3PoolPairProtocolFeePrefix         = []byte("pdexv3-poolpair-protocolfee-")
	pdexv3PoolPairStakingPoolFeePrefix      = []byte("pdexv3-poolpair-stakingpoolfee-")
	pdexv3ShareTradingFeePrefix             = []byte("pdexv3-share-tradingfee-")
	pdexv3ShareLastLpFeesPerSharePrefix     = []byte("pdexv3-share-lastlpfeespershare-")
	pdexv3ShareLastLmRewardPerSharePrefix   = []byte("pdexv3-share-lastlmrewardspershare-")
	pdexv3StakingPoolRewardPerSharePrefix   = []byte("pdexv3-stakingpool-rewardpershare-")
	pdexv3StakerRewardPrefix                = []byte("pdexv3-staker-reward-")
	pdexv3StakerLastRewardPerSharePrefix    = []byte("pdexv3-staker-lastrewardpershare-")
	pdexv3PoolPairMakingVolumePrefix        = []byte("pdexv3-poolpair-makingvolume-")
	pdexv3PoolPairOrderRewardPrefix         = []byte("pdexv3-poolpair-orderreward-")
	pdexv3PoolPairLmLockedSharePrefix       = []byte("pdexv3-poolpair-lmlockedshare-")

	// bridge agg
	bridgeAggStatusPrefix            = []byte("bridgeagg-status-")
	bridgeAggModifyParamStatusPrefix = []byte("bridgeagg-modifyparamstatus-")
	bridgeAggConvertStatusPrefix     = []byte("bridgeagg-convertstatus-")
	bridgeAggShieldStatusPrefix      = []byte("bridgeagg-shieldStatus-")
	bridgeAggUnshieldStatusPrefix    = []byte("bridgeagg-unshieldStatus-")
	bridgeAggUnifiedTokenprefix      = []byte("bridgeagg-unifiedtoken-")
	bridgeAggConvertedTokenPrefix    = []byte("bridgeagg-convertedtoken-")
	bridgeAggVaultPrefix             = []byte("bridgeagg-vault-")
	bridgeAggWaitUnshieldReqPrefix   = []byte("bridgeagg-waitUnshield-")
	bridgeAggParamPrefix             = []byte("bridgeagg-param-")

	// portal
	portalFinaExchangeRatesStatePrefix                   = []byte("portalfinalexchangeratesstate-")
	portalExchangeRatesRequestStatusPrefix               = []byte("portalexchangeratesrequeststatus-")
	portalUnlockOverRateCollateralsRequestStatusPrefix   = []byte("portalunlockoverratecollateralsstatus-")
	portalUnlockOverRateCollateralsRequestTxStatusPrefix = []byte("portalunlockoverratecollateralstxstatus-")
	portalPortingRequestStatusPrefix                     = []byte("portalportingrequeststatus-")
	portalPortingRequestTxStatusPrefix                   = []byte("portalportingrequesttxstatus-")
	portalCustodianWithdrawStatusPrefix                  = []byte("portalcustodianwithdrawstatus-")
	portalCustodianWithdrawStatusPrefixV3                = []byte("portalcustodianwithdrawstatusv3-")
	portalLiquidationTpExchangeRatesStatusPrefix         = []byte("portalliquidationtpexchangeratesstatus-")
	portalLiquidationTpExchangeRatesStatusPrefixV3       = []byte("portalliquidationbyratesstatusv3-")
	portalLiquidationExchangeRatesPoolPrefix             = []byte("portalliquidationexchangeratespool-")
	portalLiquidationCustodianDepositStatusPrefix        = []byte("portalliquidationcustodiandepositstatus-")
	portalLiquidationCustodianDepositStatusPrefixV3      = []byte("portalliquidationcustodiandepositstatusv3-")
	portalTopUpWaitingPortingStatusPrefix                = []byte("portaltopupwaitingportingstatus-")
	portalTopUpWaitingPortingStatusPrefixV3              = []byte("portaltopupwaitingportingstatusv3-")
	portalLiquidationRedeemRequestStatusPrefix           = []byte("portalliquidationredeemrequeststatus-")
	portalLiquidationRedeemRequestStatusPrefixV3         = []byte("portalliquidationredeemrequeststatusv3-")
	portalWaitingPortingRequestPrefix                    = []byte("portalwaitingportingrequest-")
	portalCustodianStatePrefix                           = []byte("portalcustodian-")
	portalWaitingRedeemRequestsPrefix                    = []byte("portalwaitingredeemrequest-")
	portalMatchedRedeemRequestsPrefix                    = []byte("portalmatchedredeemrequest-")

	portalStatusPrefix                           = []byte("portalstatus-")
	portalCustodianDepositStatusPrefix           = []byte("custodiandeposit-")
	portalCustodianDepositStatusPrefixV3         = []byte("custodiandepositv3-")
	portalRequestPTokenStatusPrefix              = []byte("requestptoken-")
	portalRedeemRequestStatusPrefix              = []byte("redeemrequest-")
	portalRedeemRequestStatusByTxReqIDPrefix     = []byte("redeemrequestbytxid-")
	portalRequestUnlockCollateralStatusPrefix    = []byte("requestunlockcollateral-")
	portalRequestWithdrawRewardStatusPrefix      = []byte("requestwithdrawportalreward-")
	portalReqMatchingRedeemStatusByTxReqIDPrefix = []byte("reqmatchredeembytxid-")

	// liquidation for portal
	portalLiquidateCustodianRunAwayPrefix = []byte("portalliquidaterunaway-")
	portalExpiredPortingReqPrefix         = []byte("portalexpiredportingreq-")

	// reward for portal
	portalRewardInfoStatePrefix       = []byte("portalreward-")
	portalLockedCollateralStatePrefix = []byte("portallockedcollateral-")

	// reward for features in network (such as portal, pdex, etc)
	rewardFeatureStatePrefix = []byte("rewardfeaturestate-")
	// feature names
	PortalRewardName = "portal"

	portalExternalTxPrefix      = []byte("portalexttx-")
	portalConfirmProofPrefix    = []byte("portalproof-")
	withdrawCollateralProofType = []byte("0-")

	// portal v4
	portalV4StatusPrefix                         = []byte("portalv4status-")
	portalUTXOStatePrefix                        = []byte("portalutxo-")
	portalShieldRequestPrefix                    = []byte("portalshieldrequest-")
	portalWaitingUnshieldRequestsPrefix          = []byte("portalwaitingunshieldrequest-")
	portalUnshieldRequestsProcessedPrefix        = []byte("portalprocessingbatchunshield-")
	portalUnshieldRequestStatusPrefix            = []byte("unshieldrequest-")
	portalBatchUnshieldRequestStatusPrefix       = []byte("batchunshield-")
	portalUnshielFeeReplacementBatchStatusPrefix = []byte("unshieldrequestbatchfeereplacementprocessed-")
	portalUnshielSubmitConfirmedTxStatusPrefix   = []byte("unshieldrequestsubmitconfirmedtx-")
	portalConvertVaultRequestPrefix              = []byte("portalconvertvaultrequest-")
)

func GetCommitteePrefixWithRole(role int, shardID int) []byte {
	switch role {
	case NextEpochShardCandidate:
		temp := []byte(string(nextShardCandidatePrefix))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case CurrentEpochShardCandidate:
		temp := []byte(string(currentShardCandidatePrefix))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case NextEpochBeaconCandidate:
		temp := []byte(string(nextBeaconCandidatePrefix))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case CurrentEpochBeaconCandidate:
		temp := []byte(string(currentBeaconCandidatePrefix))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case SubstituteValidator:
		temp := []byte(string(substitutePrefix) + strconv.Itoa(shardID))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case CurrentValidator:
		temp := []byte(string(committeePrefix) + strconv.Itoa(shardID))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case SyncingValidators:
		temp := []byte(string(syncingValidatorsPrefix) + strconv.Itoa(shardID))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	default:
		panic("role not exist: " + strconv.Itoa(role))
	}
	return []byte{}
}

func GetStakerInfoPrefix() []byte {
	h := common.HashH(stakerInfoPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetCommitteeTermKey(stakerPublicKey []byte) common.Hash {
	h := common.HashH(stakerInfoPrefix)
	final := append(h[:][:prefixHashKeyLength], common.HashH(stakerPublicKey).Bytes()[:prefixKeyLength]...)
	finalHash, err := common.Hash{}.NewHash(final)
	if err != nil {
		panic("Create key fail1")
	}
	return *finalHash
}

func GetStakerInfoKey(stakerPublicKey []byte) common.Hash {
	h := common.HashH(stakerInfoPrefix)
	final := append(h[:][:prefixHashKeyLength], common.HashH(stakerPublicKey).Bytes()[:prefixKeyLength]...)
	finalHash, err := common.Hash{}.NewHash(final)
	if err != nil {
		panic("Create key fail1")
	}
	return *finalHash
}

func GetSlashingCommitteePrefix(epoch uint64) []byte {
	buf := common.Uint64ToBytes(epoch)
	temp := append(slashingCommitteePrefix, buf...)
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}

func GetCommitteeRewardPrefix() []byte {
	h := common.HashH(committeeRewardPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetRewardRequestPrefix(epoch uint64) []byte {
	buf := common.Uint64ToBytes(epoch)
	temp := append(rewardRequestPrefix, buf...)
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}

func GetBlackListProducerPrefix() []byte {
	h := common.HashH(blackListProducerPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetSerialNumberPrefix(tokenID common.Hash, shardID byte) []byte {
	h := common.HashH(append(serialNumberPrefix, append(tokenID[:], shardID)...))
	return h[:][:prefixHashKeyLength]
}

func GetCommitmentPrefix(tokenID common.Hash, shardID byte) []byte {
	h := common.HashH(append(commitmentPrefix, append(tokenID[:], shardID)...))
	return h[:][:prefixHashKeyLength]
}

func GetCommitmentIndexPrefix(tokenID common.Hash, shardID byte) []byte {
	h := common.HashH(append(commitmentIndexPrefix, append(tokenID[:], shardID)...))
	return h[:][:prefixHashKeyLength]
}

func GetCommitmentLengthPrefix() []byte {
	h := common.HashH(commitmentLengthPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetSNDerivatorPrefix(tokenID common.Hash) []byte {
	h := common.HashH(append(snDerivatorPrefix, tokenID[:]...))
	return h[:][:prefixHashKeyLength]
}

func GetOutputCoinPrefix(tokenID common.Hash, shardID byte, publicKey []byte) []byte {
	h := common.HashH(append(outputCoinPrefix, append(tokenID[:], append(publicKey, shardID)...)...))
	return h[:][:prefixHashKeyLength]
}

func GetOTACoinPrefix(tokenID common.Hash, shardID byte, height []byte) []byte {
	// non-PRV coins will be indexed together
	if tokenID != common.PRVCoinID {
		tokenID = common.ConfidentialAssetID
	}
	h := common.HashH(append(otaCoinPrefix, append(tokenID[:], append(height, shardID)...)...))
	return h[:][:prefixHashKeyLength]
}

func GetOTACoinIndexPrefix(tokenID common.Hash, shardID byte) []byte {
	h := common.HashH(append(otaCoinIndexPrefix, append(tokenID[:], shardID)...))
	return h[:][:prefixHashKeyLength]
}

func GetOTACoinLengthPrefix() []byte {
	h := common.HashH(otaCoinLengthPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetOnetimeAddressPrefix(tokenID common.Hash) []byte {
	h := common.HashH(append(onetimeAddressPrefix, tokenID[:]...))
	return h[:][:prefixHashKeyLength]
}

func GetTokenPrefix() []byte {
	h := common.HashH(tokenPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetTokenTransactionPrefix(tokenID common.Hash) []byte {
	h := common.HashH(append(tokenTransactionPrefix, tokenID[:]...))
	return h[:][:prefixHashKeyLength]
}

func GetWaitingPDEContributionPrefix() []byte {
	h := common.HashH(waitingPDEContributionPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPDEPoolPairPrefix() []byte {
	h := common.HashH(pdePoolPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPDESharePrefix() []byte {
	h := common.HashH(pdeSharePrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPDETradingFeePrefix() []byte {
	h := common.HashH(pdeTradingFeePrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPDEStatusPrefix() []byte {
	h := common.HashH(pdeStatusPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBridgeEthTxPrefix() []byte {
	h := common.HashH(bridgeEthTxPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBridgeBSCTxPrefix() []byte {
	h := common.HashH(bridgeBSCTxPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBridgePRVEVMPrefix() []byte {
	h := common.HashH(bridgePRVEVMPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBridgePLGTxPrefix() []byte {
	h := common.HashH(bridgePLGTxPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBridgeFTMTxPrefix() []byte {
	h := common.HashH(bridgeFTMTxPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBridgeTokenInfoPrefix(isCentralized bool) []byte {
	if isCentralized {
		h := common.HashH(bridgeCentralizedTokenInfoPrefix)
		return h[:][:prefixHashKeyLength]
	} else {
		h := common.HashH(bridgeDecentralizedTokenInfoPrefix)
		return h[:][:prefixHashKeyLength]
	}
}

func GetBridgeStatusPrefix() []byte {
	h := common.HashH(bridgeStatusPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetBurningConfirmPrefix() []byte {
	h := common.HashH(burnPrefix)
	return h[:][:prefixHashKeyLength]
}
func WaitingPDEContributionPrefix() []byte {
	return waitingPDEContributionPrefix
}
func PDEPoolPrefix() []byte {
	return pdePoolPrefix
}
func PDESharePrefix() []byte {
	return pdeSharePrefix
}
func PDETradeFeePrefix() []byte {
	return pdeTradeFeePrefix
}
func PDEContributionStatusPrefix() []byte {
	return pdeContributionStatusPrefix
}
func PDETradeStatusPrefix() []byte {
	return pdeTradeStatusPrefix
}
func PDEWithdrawalStatusPrefix() []byte {
	return pdeWithdrawalStatusPrefix
}

// GetWaitingPDEContributionKey: WaitingPDEContributionPrefix - beacon height - pairid
func GetWaitingPDEContributionKey(beaconHeight uint64, pairID string) []byte {
	prefix := append(waitingPDEContributionPrefix, []byte(fmt.Sprintf("%d-", beaconHeight))...)
	return append(prefix, []byte(pairID)...)
}

// GetPDEPoolForPairKey: PDEPoolPrefix - beacon height - token1ID - token2ID
func GetPDEPoolForPairKey(beaconHeight uint64, token1ID string, token2ID string) []byte {
	prefix := append(pdePoolPrefix, []byte(fmt.Sprintf("%d-", beaconHeight))...)
	tokenIDs := []string{token1ID, token2ID}
	sort.Strings(tokenIDs)
	return append(prefix, []byte(tokenIDs[0]+"-"+tokenIDs[1])...)
}

// GetPDEShareKey: PDESharePrefix + beacon height + token1ID + token2ID + contributor address
func GetPDEShareKey(beaconHeight uint64, token1ID string, token2ID string, contributorAddress string) ([]byte, error) {
	prefix := append(pdeSharePrefix, []byte(fmt.Sprintf("%d-", beaconHeight))...)
	tokenIDs := []string{token1ID, token2ID}
	sort.Strings(tokenIDs)

	var keyAddr string
	var err error
	if len(contributorAddress) == 0 {
		keyAddr = contributorAddress
	} else {
		//Always parse the contributor address into the oldest version for compatibility
		keyAddr, err = wallet.GetPaymentAddressV1(contributorAddress, false)
		if err != nil {
			return nil, err
		}
	}
	return append(prefix, []byte(tokenIDs[0]+"-"+tokenIDs[1]+"-"+keyAddr)...), nil
}

// GetPDETradingFeeKey: PDETradingFeePrefix + beacon height + token1ID + token2ID
func GetPDETradingFeeKey(beaconHeight uint64, token1ID string, token2ID string, contributorAddress string) ([]byte, error) {
	prefix := append(pdeTradingFeePrefix, []byte(fmt.Sprintf("%d-", beaconHeight))...)
	tokenIDs := []string{token1ID, token2ID}
	sort.Strings(tokenIDs)

	var keyAddr string
	var err error
	if len(contributorAddress) == 0 {
		keyAddr = contributorAddress
	} else {
		//Always parse the contributor address into the oldest version for compatibility
		keyAddr, err = wallet.GetPaymentAddressV1(contributorAddress, false)
		if err != nil {
			return nil, err
		}
	}
	return append(prefix, []byte(tokenIDs[0]+"-"+tokenIDs[1]+"-"+keyAddr)...), nil
}

func GetPDEStatusKey(prefix []byte, suffix []byte) []byte {
	return append(prefix, suffix...)
}

// Portal
func GetFinalExchangeRatesStatePrefix() []byte {
	h := common.HashH(portalFinaExchangeRatesStatePrefix)
	return h[:][:prefixHashKeyLength]
}

func PortalPortingRequestStatusPrefix() []byte {
	return portalPortingRequestStatusPrefix
}

func PortalPortingRequestTxStatusPrefix() []byte {
	return portalPortingRequestTxStatusPrefix
}

func PortalExchangeRatesRequestStatusPrefix() []byte {
	return portalExchangeRatesRequestStatusPrefix
}

func PortalUnlockOverRateCollateralsRequestStatusPrefix() []byte {
	return portalUnlockOverRateCollateralsRequestStatusPrefix
}

func PortalCustodianWithdrawStatusPrefix() []byte {
	return portalCustodianWithdrawStatusPrefix
}

func PortalCustodianWithdrawStatusPrefixV3() []byte {
	return portalCustodianWithdrawStatusPrefixV3
}

func PortalLiquidationTpExchangeRatesStatusPrefix() []byte {
	return portalLiquidationTpExchangeRatesStatusPrefix
}

func PortalLiquidationTpExchangeRatesStatusPrefixV3() []byte {
	return portalLiquidationTpExchangeRatesStatusPrefixV3
}

func PortalLiquidationCustodianDepositStatusPrefix() []byte {
	return portalLiquidationCustodianDepositStatusPrefix
}
func PortalLiquidationCustodianDepositStatusPrefixV3() []byte {
	return portalLiquidationCustodianDepositStatusPrefixV3
}

func PortalTopUpWaitingPortingStatusPrefix() []byte {
	return portalTopUpWaitingPortingStatusPrefix
}

func PortalTopUpWaitingPortingStatusPrefixV3() []byte {
	return portalTopUpWaitingPortingStatusPrefixV3
}

func PortalLiquidationRedeemRequestStatusPrefix() []byte {
	return portalLiquidationRedeemRequestStatusPrefix
}
func PortalLiquidationRedeemRequestStatusPrefixV3() []byte {
	return portalLiquidationRedeemRequestStatusPrefixV3
}

func GetPortalUnlockOverRateCollateralsPrefix() []byte {
	h := common.HashH(portalUnlockOverRateCollateralsRequestTxStatusPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPortalWaitingPortingRequestPrefix() []byte {
	h := common.HashH(portalWaitingPortingRequestPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPortalLiquidationPoolPrefix() []byte {
	h := common.HashH(portalLiquidationExchangeRatesPoolPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPortalCustodianStatePrefix() []byte {
	h := common.HashH(portalCustodianStatePrefix)
	return h[:][:prefixHashKeyLength]
}

func GetWaitingRedeemRequestPrefix() []byte {
	h := common.HashH(portalWaitingRedeemRequestsPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetMatchedRedeemRequestPrefix() []byte {
	h := common.HashH(portalMatchedRedeemRequestsPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPortalRewardInfoStatePrefix(beaconHeight uint64) []byte {
	h := common.HashH(append(portalRewardInfoStatePrefix, []byte(fmt.Sprintf("%d-", beaconHeight))...))
	return h[:][:prefixHashKeyLength]
}

func GetPortalStatusPrefix() []byte {
	h := common.HashH(portalStatusPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetLockedCollateralStatePrefix() []byte {
	h := common.HashH(portalLockedCollateralStatePrefix)
	return h[:][:prefixHashKeyLength]
}

func GetRewardFeatureStatePrefix(epoch uint64) []byte {
	h := common.HashH(append(rewardFeatureStatePrefix, []byte(fmt.Sprintf("%d-", epoch))...))
	return h[:][:prefixHashKeyLength]
}

func GetPortalExternalTxPrefix() []byte {
	h := common.HashH(portalExternalTxPrefix)
	return h[:][:prefixHashKeyLength]
}

func GetPortalConfirmProofPrefixV3(proofType []byte) []byte {
	h := common.HashH(append(portalConfirmProofPrefix, proofType...))
	return h[:][:prefixHashKeyLength]
}

func PortalWithdrawCollateralProofType() []byte {
	return withdrawCollateralProofType
}

func PortalCustodianDepositStatusPrefix() []byte {
	return portalCustodianDepositStatusPrefix
}

func PortalCustodianDepositStatusPrefixV3() []byte {
	return portalCustodianDepositStatusPrefixV3
}

func PortalRequestPTokenStatusPrefix() []byte {
	return portalRequestPTokenStatusPrefix
}

func PortalRedeemRequestStatusPrefix() []byte {
	return portalRedeemRequestStatusPrefix
}

func PortalRedeemRequestStatusByTxReqIDPrefix() []byte {
	return portalRedeemRequestStatusByTxReqIDPrefix
}

func PortalRequestUnlockCollateralStatusPrefix() []byte {
	return portalRequestUnlockCollateralStatusPrefix
}

func PortalRequestWithdrawRewardStatusPrefix() []byte {
	return portalRequestWithdrawRewardStatusPrefix
}

func PortalLiquidateCustodianRunAwayPrefix() []byte {
	return portalLiquidateCustodianRunAwayPrefix
}

func PortalExpiredPortingReqPrefix() []byte {
	return portalExpiredPortingReqPrefix
}

func PortalReqMatchingRedeemStatusByTxReqIDPrefix() []byte {
	return portalReqMatchingRedeemStatusByTxReqIDPrefix
}

// pDex v3 prefix for status
func Pdexv3ParamsModifyingStatusPrefix() []byte {
	return pdexv3ParamsModifyingPrefix
}

func Pdexv3TradeStatusPrefix() []byte {
	return pdexv3TradeStatusPrefix
}

func Pdexv3WithdrawalLPFeeStatusPrefix() []byte {
	return pdexv3WithdrawalLPFeePrefix
}

func Pdexv3WithdrawalProtocolFeeStatusPrefix() []byte {
	return pdexv3WithdrawalProtocolFeePrefix
}

func Pdexv3WithdrawalStakingRewardStatusPrefix() []byte {
	return pdexv3WithdrawalStakingRewardPrefix
}

func Pdexv3AddOrderStatusPrefix() []byte {
	return pdexv3AddOrderStatusPrefix
}

func Pdexv3WithdrawOrderStatusPrefix() []byte {
	return pdexv3WithdrawOrderStatusPrefix
}

// pDex v3 prefix hash of the key
func GetPdexv3StatusPrefix(statusType []byte) []byte {
	h := common.HashH(append(pdexv3StatusPrefix, statusType...))
	return h[:][:prefixHashKeyLength]
}

func GetPdexv3ParamsPrefix() []byte {
	return pdexv3ParamsPrefix
}

func GetPdexv3WaitingContributionsPrefix() []byte {
	hash := common.HashH(pdexv3WaitingContributionsPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairsPrefix() []byte {
	hash := common.HashH(pdexv3PoolPairsPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3SharesPrefix() []byte {
	hash := common.HashH(pdexv3SharesPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3OrdersPrefix() []byte {
	return pdexv3OrdersPrefix
}

func GetPdexv3NftPrefix() []byte {
	hash := common.HashH(pdexv3MintNftPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3StakingPoolsPrefix() []byte {
	hash := common.HashH(pdexv3StakingPoolsPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3StakersPrefix() []byte {
	hash := common.HashH(pdexv3StakerPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairLpFeePerSharesPrefix() []byte {
	hash := common.HashH(pdexv3PoolPairLpFeePerSharePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairLmRewardPerSharesPrefix() []byte {
	hash := common.HashH(pdexv3PoolPairLmRewardPerSharePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairLmLockedSharePrefix() []byte {
	hash := common.HashH(pdexv3PoolPairLmLockedSharePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairProtocolFeesPrefix() []byte {
	hash := common.HashH(pdexv3PoolPairProtocolFeePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairStakingPoolFeesPrefix() []byte {
	hash := common.HashH(pdexv3PoolPairStakingPoolFeePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3ShareTradingFeesPrefix() []byte {
	hash := common.HashH(pdexv3ShareTradingFeePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3ShareLastLpFeePerSharesPrefix() []byte {
	hash := common.HashH(pdexv3ShareLastLpFeesPerSharePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3ShareLastLmRewardPerSharesPrefix() []byte {
	hash := common.HashH(pdexv3ShareLastLmRewardPerSharePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3StakingPoolRewardPerSharePrefix() []byte {
	hash := common.HashH(pdexv3StakingPoolRewardPerSharePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairMakingVolumePrefix() []byte {
	hash := common.HashH(pdexv3PoolPairMakingVolumePrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3PoolPairOrderRewardPrefix() []byte {
	hash := common.HashH(pdexv3PoolPairOrderRewardPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3StakerReward() []byte {
	hash := common.HashH(pdexv3StakerRewardPrefix)
	return hash[:prefixHashKeyLength]
}

func GetPdexv3StakerLastRewardPerShare() []byte {
	hash := common.HashH(pdexv3StakerLastRewardPerSharePrefix)
	return hash[:prefixHashKeyLength]
}

//
func Pdexv3WithdrawLiquidityStatusPrefix() []byte {
	return pdexv3WithdrawLiquidityStatusPrefix
}

// pDex v3 prefix for mintnft status
func Pdexv3UserMintNftStatusPrefix() []byte {
	return pdexv3UserMintNftStatusPrefix
}

// pDex v3 prefix for contribution status
func Pdexv3ContributionStatusPrefix() []byte {
	return pdexv3WaitingContributionStatusPrefix
}

// pDex v3 prefix for staking status
func Pdexv3StakingStatusPrefix() []byte {
	return pdexv3StakingStatusPrefix
}

// pDex v3 prefix for unstaking status
func Pdexv3UnstakingStatusPrefix() []byte {
	return pdexv3UnstakingStatusPrefix
}

// TODO: rename
// PORTAL V4
// Portal v4 prefix for portal v4 status
func PortalShieldingRequestStatusPrefix() []byte {
	return portalShieldRequestPrefix
}

func PortalUnshieldRequestStatusPrefix() []byte {
	return portalUnshieldRequestStatusPrefix
}

func PortalBatchUnshieldRequestStatusPrefix() []byte {
	return portalBatchUnshieldRequestStatusPrefix
}

func PortaConvertVaultRequestStatusPrefix() []byte {
	return portalConvertVaultRequestPrefix
}

// Portal v4 prefix hash of the key

func GetPortalV4StatusPrefix(statusType []byte) []byte {
	h := common.HashH(append(portalV4StatusPrefix, statusType...))
	return h[:][:prefixHashKeyLength]
}

func GetPortalUTXOStatePrefix(tokenID string) []byte {
	h := common.HashH(append(portalUTXOStatePrefix, []byte(tokenID)...))
	return h[:][:prefixHashKeyLength]
}

func GetShieldingRequestPrefix(tokenID string) []byte {
	h := common.HashH(append(portalShieldRequestPrefix, []byte(tokenID)...))
	return h[:][:prefixHashKeyLength]
}

func GetWaitingUnshieldRequestPrefix(tokenID string) []byte {
	h := common.HashH(append(portalWaitingUnshieldRequestsPrefix, []byte(tokenID)...))
	return h[:][:prefixHashKeyLength]
}

func GetProcessedUnshieldRequestBatchPrefix(tokenID string) []byte {
	h := common.HashH(append(portalUnshieldRequestsProcessedPrefix, []byte(tokenID)...))
	return h[:][:prefixHashKeyLength]
}

func PortalUnshielFeeReplacementBatchStatusPrefix() []byte {
	return portalUnshielFeeReplacementBatchStatusPrefix
}

func PortalSubmitConfirmedTxStatusPrefix() []byte {
	return portalUnshielSubmitConfirmedTxStatusPrefix
}

func GetBridgeAggStatusPrefix(statusType []byte) []byte {
	h := common.HashH(append(bridgeAggStatusPrefix, statusType...))
	return h[:][:prefixHashKeyLength]
}

func BridgeAggModifyParamStatusPrefix() []byte {
	return bridgeAggModifyParamStatusPrefix
}

func BridgeAggConvertStatusPrefix() []byte {
	return bridgeAggConvertStatusPrefix
}

func BridgeAggShieldStatusPrefix() []byte {
	return bridgeAggShieldStatusPrefix
}

func BridgeAggUnshieldStatusPrefix() []byte {
	return bridgeAggUnshieldStatusPrefix
}

func GetBridgeAggUnifiedTokenPrefix() []byte {
	hash := common.HashH(bridgeAggUnifiedTokenprefix)
	return hash[:prefixHashKeyLength]
}

func GetBridgeAggConvertedTokenPrefix() []byte {
	hash := common.HashH(bridgeAggConvertedTokenPrefix)
	return hash[:prefixHashKeyLength]
}

func GetBridgeAggVaultPrefix() []byte {
	hash := common.HashH(bridgeAggVaultPrefix)
	return hash[:prefixHashKeyLength]
}

func GetBridgeAggWaitingUnshieldReqPrefix(unifiedTokenID []byte) []byte {
	h := common.HashH(append(bridgeAggWaitUnshieldReqPrefix, unifiedTokenID...))
	return h[:][:prefixHashKeyLength]
}

func GetBridgeAggParamPrefix() []byte {
	h := common.HashH(bridgeAggParamPrefix)
	return h[:][:prefixHashKeyLength]
}

var _ = func() (_ struct{}) {
	m := make(map[string]string)
	prefixs := [][]byte{}
	// Current validator
	for i := -1; i < 256; i++ {
		temp := GetCommitteePrefixWithRole(CurrentValidator, i)
		prefixs = append(prefixs, temp)
		if v, ok := m[string(temp)]; ok {
			panic("shard-com-" + strconv.Itoa(i) + " same prefix " + v)
		}
		m[string(temp)] = "shard-com-" + strconv.Itoa(i)
	}
	// Substitute validator
	for i := -1; i < 256; i++ {
		temp := GetCommitteePrefixWithRole(SubstituteValidator, i)
		prefixs = append(prefixs, temp)
		if v, ok := m[string(temp)]; ok {
			panic("shard-sub-" + strconv.Itoa(i) + " same prefix " + v)
		}
		m[string(temp)] = "shard-sub-" + strconv.Itoa(i)
	}
	// Current Candidate
	tempCurrentCandidate := GetCommitteePrefixWithRole(CurrentEpochShardCandidate, -2)
	prefixs = append(prefixs, tempCurrentCandidate)
	if v, ok := m[string(tempCurrentCandidate)]; ok {
		panic("cur-cand-" + " same prefix " + v)
	}
	m[string(tempCurrentCandidate)] = "cur-cand-"
	// Next candidate
	tempNextCandidate := GetCommitteePrefixWithRole(NextEpochShardCandidate, -2)
	prefixs = append(prefixs, tempNextCandidate)
	if v, ok := m[string(tempNextCandidate)]; ok {
		panic("next-cand-" + " same prefix " + v)
	}
	m[string(tempNextCandidate)] = "next-cand-"
	// reward receiver
	tempRewardReceiver := GetCommitteeRewardPrefix()
	prefixs = append(prefixs, tempRewardReceiver)
	if v, ok := m[string(tempRewardReceiver)]; ok {
		panic("committee-reward-" + " same prefix " + v)
	}
	m[string(tempRewardReceiver)] = "committee-reward-"
	// black list producer
	tempBlackListProducer := GetBlackListProducerPrefix()
	prefixs = append(prefixs, tempBlackListProducer)
	if v, ok := m[string(tempBlackListProducer)]; ok {
		panic("black-list-" + " same prefix " + v)
	}
	m[string(tempBlackListProducer)] = "black-list-"
	for i, v1 := range prefixs {
		for j, v2 := range prefixs {
			if i == j {
				continue
			}
			if bytes.HasPrefix(v1, v2) || bytes.HasPrefix(v2, v1) {
				panic("(prefix: " + fmt.Sprintf("%+v", v1) + ", value: " + m[string(v1)] + ")" + " is prefix or being prefix of " + " (prefix: " + fmt.Sprintf("%+v", v1) + ", value: " + m[string(v2)] + ")")
			}
		}
	}
	return
}()
