package statedb

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	ErrInvalidByteArrayType                   = "invalid byte array type"
	ErrInvalidHashType                        = "invalid hash type"
	ErrInvalidBigIntType                      = "invalid big int type"
	ErrInvalidCommitteeStateType              = "invalid committee state type"
	ErrInvalidStakerInfoType                  = "invalid staker info type"
	ErrInvalidCommitteeTermType               = "invalid committee term type"
	ErrInvalidPaymentAddressType              = "invalid payment address type"
	ErrInvalidIncognitoPublicKeyType          = "invalid incognito public key type"
	ErrInvalidCommitteeRewardStateType        = "invalid reward receiver state type "
	ErrInvalidRewardRequestStateType          = "invalid reward request state type"
	ErrInvalidBlackListProducerStateType      = "invalid black list producer state type"
	ErrInvalidSerialNumberStateType           = "invalid serial number state type"
	ErrInvalidCommitmentStateType             = "invalid commitment state type"
	ErrInvalidSNDerivatorStateType            = "invalid snderivator state type"
	ErrInvalidOutputCoinStateType             = "invalid output coin state type"
	ErrInvalidOTACoinStateType                = "invalid ota coin state type"
	ErrInvalidOnetimeAddressStateType         = "invalid onetime address state type"
	ErrInvalidTokenStateType                  = "invalid token state type"
	ErrInvalidWaitingPDEContributionStateType = "invalid waiting pde contribution state type"
	ErrInvalidPDEPoolPairStateType            = "invalid pde pool pair state type"
	ErrInvalidPDEShareStateType               = "invalid pde shard state type"
	ErrInvalidPDEStatusStateType              = "invalid pde status state type"
	ErrInvalidBridgeEthTxStateType            = "invalid bridge eth tx state type"
	ErrInvalidBridgeTokenInfoStateType        = "invalid bridge token info state type"
	ErrInvalidBridgeStatusStateType           = "invalid bridge status state type"
	ErrInvalidBurningConfirmStateType         = "invalid burning confirm state type"
	ErrInvalidTokenTransactionStateType       = "invalid token transaction state type"
	ErrInvalidBridgeBSCTxStateType            = "invalid bridge bsc tx state type"
	ErrInvalidBridgePRVEVMStateType           = "invalid bridge prv evm tx state type"
	ErrInvalidBridgePLGTxStateType            = "invalid bridge polygon tx state type"
	ErrInvalidBridgeFTMTxStateType            = "invalid bridge fantom tx state type"
	ErrInvalidBridgeAURORATxStateType         = "invalid bridge aurora tx state type"
	ErrInvalidBridgeAVAXTxStateType           = "invalid bridge avax tx state type"
	// A
	ErrInvalidFinalExchangeRatesStateType  = "invalid final exchange rates state type"
	ErrInvalidLiquidationExchangeRatesType = "invalid liquidation exchange rates type"
	ErrInvalidWaitingPortingRequestType    = "invalid waiting porting request type"
	// B
	ErrInvalidPortalStatusStateType              = "invalid portal status state type"
	ErrInvalidPortalCustodianStateType           = "invalid portal custodian state type"
	ErrInvalidPortalWaitingRedeemRequestType     = "invalid portal waiting redeem request type"
	ErrInvalidPortalRewardInfoStateType          = "invalid portal reward info state type"
	ErrInvalidPortalLockedCollateralStateType    = "invalid portal locked collateral state type"
	ErrInvalidRewardFeatureStateType             = "invalid feature reward state type"
	ErrInvalidPDETradingFeeStateType             = "invalid pde trading fee state type"
	ErrInvalidUnlockOverRateCollateralsStateType = "invalid unlock over rate collaterals state type"
	ErrInvalidSlasingCommitteeStateType          = "invalid slashing committee state type"
	ErrInvalidPortalV4StatusStateType            = "invalid portal v4 status state type"
	ErrInvalidPortalExternalTxStateType          = "invalid portal external tx state type"
	ErrInvalidPortalConfirmProofStateType        = "invalid portal confirm proof state type"
	ErrInvalidPortalUTXOType                     = "invalid portal utxo state type"
	ErrInvalidPortalShieldingRequestType         = "invalid portal shielding request type"
	ErrInvalidPortalV4WaitingUnshieldRequestType = "invalid portal waiting unshielding request type"
	ErrInvalidPortalV4BatchUnshieldRequestType   = "invalid portal batch unshielding request type"
	// pDex v3
	ErrInvalidPdexv3StatusStateType                    = "invalid pdex v3 status state type"
	ErrInvalidPdexv3ParamsStateType                    = "invalid pdex v3 params state type"
	ErrInvalidPdexv3ContributionStateType              = "invalid pdex v3 contribution state type"
	ErrInvalidPdexv3PoolPairStateType                  = "invalid pdex v3 pool pair state type"
	ErrInvalidPdexv3ShareStateType                     = "invalid pdex v3 share state type"
	ErrInvalidPdexv3NftStateType                       = "invalid pdex v3 nft state type"
	ErrInvalidPdexv3OrderStateType                     = "invalid pdex v3 order state type"
	ErrInvalidPdexv3StakingPoolStateType               = "invalid pdex v3 staking pool state type"
	ErrInvalidPdexv3StakerStateType                    = "invalid pdex v3 staker state type"
	ErrInvalidPdexv3PoolPairLpFeePerShareStateType     = "invalid pdex v3 pool pair lp fee per share state type"
	ErrInvalidPdexv3PoolPairLmRewardPerShareStateType  = "invalid pdex v3 pool pair lm reward per share state type"
	ErrInvalidPdexv3PoolPairProtocolFeeStateType       = "invalid pdex v3 pool pair protocol fee state type"
	ErrInvalidPdexv3PoolPairStakingPoolFeeStateType    = "invalid pdex v3 pool pair staking pool fee state type"
	ErrInvalidPdexv3ShareTradingFeeStateType           = "invalid pdex v3 share trading fee state type"
	ErrInvalidPdexv3LastLPFeesPerShareStateType        = "invalid pdex v3 share last lp fees per share state type"
	ErrInvalidPdexv3LastLmRewardPerShareStateType      = "invalid pdex v3 share last lm reward per share state type"
	ErrInvalidPdexv3StakingPoolRewardPerShareStateType = "invalid pdex v3 staking pool reward per share state type"
	ErrInvalidPdexv3StakerRewardStateType              = "invalid pdex v3 staker reward state type"
	ErrInvalidPdexv3StakerLastRewardPerShareStateType  = "invalid pdex v3 staker last reward per share state type"
	ErrInvalidPdexv3PoolPairMakingVolumeStateType      = "invalid pdex v3 pool pair making voulme state type"
	ErrInvalidPdexv3PoolPairOrderRewardStateType       = "invalid pdex v3 pool pair order reward state type"
	ErrInvalidPdexv3PoolPairLmLockedShareStateType     = "invalid pdex v3 pool pair lm locked share state type"
	// bridge agg
	ErrInvalidBridgeAggStatusStateType         = "invalid bridge agg status state type"
	ErrInvalidBridgeAggUnifiedTokenStateType   = "invalid bridge agg unified token state type"
	ErrInvalidBridgeAggConvertedTokenStateType = "invalid bridge agg converted token state type"
	ErrInvalidBridgeAggVaultStateType          = "invalid bridge agg vault state type"
	ErrInvalidBridgeAggWaitingUnshieldReqType  = "invalid bridge agg waiting unshield request state type"
	ErrInvalidBridgeAggParamStateType          = "invalid bridge agg param state type"

	ErrInvalidBridgeNEARTxStateType = "invalid bridge near tx state type"

	// Decentralized Bridge Hub
	ErrInvalidBridgeHubParamStateType      = "invalid bridge hub param state type"
	ErrInvalidBridgeHubPTokenStateType     = "invalid bridge hub pToken state type"
	ErrInvalidBridgeHubBridgeInfoStateType = "invalid bridge hub bridge info state type"
	ErrInvalidBridgeHubStatusStateType     = "invalid bridge hub status state type"
)
const (
	InvalidByteArrayTypeError = iota
	InvalidHashTypeError
	InvalidBigIntTypeError
	InvalidCommitteeStateTypeError
	InvalidPaymentAddressTypeError
	InvalidIncognitoPublicKeyTypeError
	InvalidCommitteeRewardStateTypeError
	InvalidRewardRequestStateTypeError
	InvalidBlackListProducerStateTypeError
	InvalidSerialNumberStateTypeError
	InvalidCommitmentStateTypeError
	InvalidSNDerivatorStateTypeError
	InvalidOutputCoinStateTypeError
	// general error
	MethodNotSupportError
	// transaction related error
	StoreSerialNumberError
	GetSerialNumberError
	StoreCommitmentError
	GetCommitmentError
	StoreCommitmentIndexError
	GetCommitmentIndexError
	StoreCommitmentLengthError
	GetCommitmentLengthError
	StoreOutputCoinError
	GetOutputCoinError
	StoreOTACoinError
	GetOTACoinIndexError
	StoreOTACoinIndexError
	StoreOTACoinLengthError
	GetOTACoinLengthError
	StoreOnetimeAddressError
	StoreSNDerivatorError
	GetSNDerivatorError
	StorePrivacyTokenError
	StorePrivacyTokenTransactionError
	GetPrivacyTokenError
	GetPrivacyTokenTxsError
	PrivacyTokenIDExistedError
	// Consensus Related Error
	StoreBlockHashError
	GetBlockHashError
	StoreBeaconCommitteeError
	GetBeaconCommitteeError
	StoreShardCommitteeError
	GetShardCommitteeError
	StoreAllShardCommitteeError
	StoreNextEpochCandidateError
	StoreOneShardSubstitutesValidatorError
	StoreBeaconSubstitutesValidatorError
	StoreCurrentEpochCandidateError
	StoreRewardRequestError
	GetRewardRequestError
	StoreCommitteeRewardError
	GetCommitteeRewardError
	ListCommitteeRewardError
	RemoveCommitteeRewardError
	StoreBlackListProducersError
	StoreSyncingValidatorsError
	SaveStopAutoStakerInfoError

	DeleteBeaconCommitteeError
	DeleteOneShardCommitteeError
	DeleteAllShardCommitteeError
	DeleteNextEpochShardCandidateError
	DeleteCurrentEpochShardCandidateError
	DeleteNextEpochBeaconCandidateError
	DeleteCurrentEpochBeaconCandidateError
	DeleteAllShardSubstitutesValidatorError
	DeleteBeaconSubstituteValidatorError
	DeleteBeaconWaitingError
	DeleteBeaconLockingError
	// pdex error
	StoreWaitingPDEContributionError
	StorePDEPoolPairError
	StorePDEShareError
	GetPDEPoolForPairError
	TrackPDEStatusError
	GetPDEStatusError
	// bridge error
	BridgeInsertETHTxHashIssuedError
	IsETHTxHashIssuedError
	IsBridgeTokenExistedByTypeError
	CanProcessCIncTokenError
	CanProcessTokenPairError
	UpdateBridgeTokenInfoError
	GetAllBridgeTokensError
	TrackBridgeReqWithStatusError
	GetBridgeReqWithStatusError
	// burning confirm
	StoreBurningConfirmError
	GetBurningConfirmError

	// portal
	StoreCustodianStateError
	StoreWaitingRedeemRequestError
	StorePortalRewardError
	StorePortalStatusError
	GetPortalStatusError
	GetPortalStatusNotFoundError
	GetPortalRedeemRequestStatusError
	StorePortalRedeemRequestStatusError
	StorePortalCustodianDepositStatusError
	GetPortalCustodianDepositStatusError
	StorePortalRequestPTokenStatusError
	GetPortalRequestPTokenStatusError
	GetPortalRedeemRequestByTxIDStatusError
	StorePortalRedeemRequestByTxIDStatusError
	GetPortalRequestUnlockCollateralStatusError
	StorePortalRequestUnlockCollateralStatusError
	GetPortalLiquidationCustodianRunAwayStatusError
	StorePortalLiquidationCustodianRunAwayStatusError
	GetPortalExpiredPortingReqStatusError
	StorePortalExpiredPortingReqStatusError
	GetPortalRequestWithdrawRewardStatusError
	StorePortalRequestWithdrawRewardStatusError
	StoreLockedCollateralStateError
	GetLockedCollateralStateError
	StorePortalReqMatchingRedeemByTxIDStatusError
	GetPortalReqMatchingRedeemByTxIDStatusError
	GetPortalRedeemRequestFromLiquidationByTxIDStatusError
	StorePortalRedeemRequestFromLiquidationByTxIDStatusError
	GetPortalCustodianWithdrawCollateralStatusError
	StorePortalCustodianWithdrawCollateralStatusError
	GetPortalPortingRequestStatusError
	StorePortalPortingRequestStatusError
	GetPortalPortingRequestByTxIDStatusError
	StorePortalPortingRequestByTxIDStatusError
	StoreWaitingPortingRequestError
	StorePortalExchangeRatesStatusError
	GetPortalExchangeRatesStatusError
	StoreFinalExchangeRatesStateError
	GetPortalLiquidationExchangeRatesPoolError
	GetLiquidationByExchangeRatesStatusError
	StoreLiquidationByExchangeRatesStatusError
	StoreLiquidateExchangeRatesPoolError
	GetCustodianTopupStatusError
	StoreCustodianTopupStatusError
	StoreRewardFeatureError
	GetRewardFeatureError
	GetAllRewardFeatureError
	GetRewardFeatureAmountByTokenIDError

	// PDEX v2
	StorePDETradingFeeError

	InvalidStakerInfoTypeError
	StoreMemberCommonShardPoolError
	StoreMemberShardPoolError
	StoreBeaconWaitingError
	StoreBeaconLockingError
	DeleteMemberCommonShardPoolError
	DeleteMemberShardPoolError
	DeleteMemberCommonBeaconPoolError
	DeleteMemberBeaconPoolError

	// Portal v3
	IsPortalExternalTxHashSubmittedError
	InsertPortalExternalTxHashSubmittedError
	StoreWithdrawCollateralConfirmError
	GetWithdrawCollateralConfirmError
	StorePortalUnlockOverRateCollateralsError
	GetPortalUnlockOverRateCollateralsStatusError

	// portal v4
	StorePortalV4StatusError
	GetPortalV4StatusError
	StorePortalV4UTXOsError
	StorePortalV4ShieldingRequestStatusError
	GetPortalV4ShieldingRequestStatusError
	StorePortalShieldingRequestsError
	GetPortalShieldingRequestsError
	GetPortalUnshieldRequestStatusError
	StorePortalUnshieldRequestStatusError
	GetPortalBatchUnshieldRequestStatusError
	StorePortalBatchUnshieldRequestStatusError
	StorePortalListWaitingUnshieldRequestError
	StorePortalListProcessedBatchUnshieldRequestError
	GetPortalUnshieldBatchFeeReplacementRequestStatusError
	StorePortalUnshieldBatchFeeReplacementRequestStatusError
	GetPortalSubmitConfirmedTxRequestStatusError
	StorePortalSubmitConfirmedTxRequestStatusError
	StorePortalV4ConvertVaultRequestStatusError
	GetPortalV4ConvertVaultRequestStatusError

	// bsc bridge
	BridgeInsertBSCTxHashIssuedError
	IsBSCTxHashIssuedError

	// prv pegging erc20/bep20
	BridgeInsertPRVEVMTxHashIssuedError
	IsPRVEVMTxHashIssuedError

	// pDex v3
	GetPdexv3StatusError
	StorePdexv3StatusError
	GetPdexv3ParamsError
	StorePdexv3ParamsError
	StorePdexv3ContributionError
	StorePdexv3PoolPairError
	StorePdexv3ShareError
	StorePdexv3TradingFeesError
	StorePdexv3NftsError
	GetPdexv3PoolPairError

	// Polygon bridge
	BridgeInsertPLGTxHashIssuedError
	IsPLGTxHashIssuedError

	// Fantom bridge
	BridgeInsertFTMTxHashIssuedError
	IsFTMTxHashIssuedError

	// Aurora bridge
	BridgeInsertAURORATxHashIssuedError
	IsAURORATxHashIssuedError

	// Avalanche bridge
	BridgeInsertAVAXTxHashIssuedError
	IsAVAXTxHashIssuedError

	// Bridge Agg
	GetBridgeAggStatusError
	StoreBridgeAggStatusError

	// Near bridge
	BridgeInsertNEARTxHashIssuedError
	IsNEARTxHashIssuedError

	// Bridge Hub
	GetBridgeHubStatusError
	StoreBridgeHubStatusError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	MethodNotSupportError: {-1, "Method Not Support"},
	// -1xxx reposistory level
	InvalidByteArrayTypeError:              {-1000, "invalid byte array type"},
	InvalidHashTypeError:                   {-1001, "invalid hash type"},
	InvalidBigIntTypeError:                 {-1002, "invalid big int type"},
	InvalidCommitteeStateTypeError:         {-1003, "invalid committee state type"},
	InvalidPaymentAddressTypeError:         {-1004, "invalid payment address type"},
	InvalidIncognitoPublicKeyTypeError:     {-1005, "invalid incognito public key type"},
	InvalidCommitteeRewardStateTypeError:   {-1006, "invalid reward receiver state type "},
	InvalidRewardRequestStateTypeError:     {-1007, "invalid reward request state type"},
	InvalidBlackListProducerStateTypeError: {-1008, "invalid black list producer state type"},
	InvalidSerialNumberStateTypeError:      {-1009, "invalid serial number state type"},
	InvalidCommitmentStateTypeError:        {-1010, "invalid commitment state type"},
	InvalidSNDerivatorStateTypeError:       {-1011, "invalid snderivator state type"},
	InvalidOutputCoinStateTypeError:        {-1011, "invalid output coin state type"},
	// -2xxx transaction error
	StoreSerialNumberError:            {-2000, "Store Serial Number Error"},
	GetSerialNumberError:              {-2001, "Get Serial Number Error"},
	StoreCommitmentError:              {-2002, "Store Commitment Error"},
	GetCommitmentError:                {-2003, "Get Commitment Error"},
	StoreCommitmentIndexError:         {-2004, "Store Commitment Index Error"},
	GetCommitmentIndexError:           {-2005, "Get Commitment Index Error"},
	StoreCommitmentLengthError:        {-2006, "Store Commitment Length Error"},
	GetCommitmentLengthError:          {-2007, "Get Commitment Length Error"},
	StoreOutputCoinError:              {-2008, "Store Output Coin Error"},
	GetOutputCoinError:                {-2009, "Get Output Coin Error"},
	StoreSNDerivatorError:             {-2010, "Store SNDeriavator Error"},
	GetSNDerivatorError:               {-2011, "Get SNDeriavator Error"},
	StorePrivacyTokenError:            {-2012, "Store Privacy Token Error"},
	StorePrivacyTokenTransactionError: {-2013, "Store Privacy Token Transaction Error"},
	GetPrivacyTokenError:              {-2014, "Get Privacy Token Error"},
	GetPrivacyTokenTxsError:           {-2015, "Get Privacy Token Txs Error"},
	PrivacyTokenIDExistedError:        {-2016, "Privacy Token ID Existed Error"},
	// -3xxx: consensus error
	StoreBeaconCommitteeError:              {-3000, "Store Beacon Committee Error"},
	GetBeaconCommitteeError:                {-3001, "Get Beacon Committee Error"},
	StoreShardCommitteeError:               {-3002, "Store Shard Committee Error"},
	GetShardCommitteeError:                 {-3003, "Get Shard Committee Error"},
	StoreAllShardCommitteeError:            {-3004, "Store All Shard Committee Error"},
	StoreNextEpochCandidateError:           {-3005, "Store Next Epoch Candidate Error"},
	StoreCurrentEpochCandidateError:        {-3006, "Store Next Current Candidate Error"},
	StoreRewardRequestError:                {-3007, "Store Reward Request Error"},
	GetRewardRequestError:                  {-3008, "Get Reward Request Error"},
	StoreCommitteeRewardError:              {-3009, "Store Committee Reward Error"},
	GetCommitteeRewardError:                {-3010, "Get Committee Reward Error"},
	ListCommitteeRewardError:               {-3011, "List Committee Reward Error"},
	RemoveCommitteeRewardError:             {-3012, "Remove Committee Reward Error"},
	StoreBlackListProducersError:           {-3013, "Store Black List Producers Error"},
	StoreOneShardSubstitutesValidatorError: {-3014, "Store One Shard Substitutes Validator Error"},
	StoreBeaconSubstitutesValidatorError:   {-3014, "Store Beacon Substitutes Validator Error"},
	StoreSyncingValidatorsError:            {-3015, "Store Syncing Validators Error"},
	SaveStopAutoStakerInfoError:            {-3016, "Store Stop Autostake Info Error"},
	// -4xxx: pdex error
	StoreWaitingPDEContributionError: {-4000, "Store Waiting PDEX Contribution Error"},
	StorePDEPoolPairError:            {-4001, "Store PDEX Pool Pair Error"},
	StorePDEShareError:               {-4002, "Store PDEX Share Error"},
	GetPDEPoolForPairError:           {-4003, "Get PDEX Pool Pair Error"},
	TrackPDEStatusError:              {-4004, "Track PDEX Status Error"},
	GetPDEStatusError:                {-4005, "Get PDEX Status Error"},
	// -5xxx: bridge error
	BridgeInsertETHTxHashIssuedError: {-5000, "Bridge Insert ETH Tx Hash Issued Error"},
	IsETHTxHashIssuedError:           {-5001, "Is ETH Tx Hash Issued Error"},
	IsBridgeTokenExistedByTypeError:  {-5002, "Is Bridge Token Existed By Type Error"},
	CanProcessCIncTokenError:         {-5003, "Can Process Centralized Inc Token Error"},
	CanProcessTokenPairError:         {-5004, "Can Process Token Pair Error"},
	UpdateBridgeTokenInfoError:       {-5005, "Update Bridge Token Info Error"},
	GetAllBridgeTokensError:          {-5006, "Get All Bridge Tokens Error"},
	TrackBridgeReqWithStatusError:    {-5007, "Track Bridge Request With Status Error"},
	GetBridgeReqWithStatusError:      {-5008, "Get Bridge Request With Status Error"},
	// -6xxx: burning confirm
	StoreBurningConfirmError: {-6000, "Store Burning Confirm Error"},
	GetBurningConfirmError:   {-6001, "Get Burning Confirm Error"},

	// portal
	StorePortalStatusError:       {-14000, "Store portal status error"},
	GetPortalStatusError:         {-14001, "Get portal status error"},
	GetPortalStatusNotFoundError: {-14002, "Get portal status not found error"},
	// custodian
	StoreCustodianStateError:                          {-14003, "Store custodian state error"},
	GetPortalCustodianDepositStatusError:              {-14004, "Get portal custodian deposit status error"},
	StorePortalCustodianDepositStatusError:            {-14005, "Store portal custodian deposit status error"},
	GetPortalCustodianWithdrawCollateralStatusError:   {-14006, "Get portal custodian withdraw collateral by txID status error"},
	StorePortalCustodianWithdrawCollateralStatusError: {-14007, "Store portal custodian withdraw collateral by txID status error"},
	// porting
	StoreWaitingPortingRequestError:            {-14008, "Store waiting porting requests error"},
	StorePortalRequestPTokenStatusError:        {-14009, "Store portal request ptoken status error"},
	GetPortalRequestPTokenStatusError:          {-14010, "Get portal request ptoken status error"},
	GetPortalPortingRequestStatusError:         {-14011, "Get portal porting request status error"},
	StorePortalPortingRequestStatusError:       {-14012, "Store portal porting request status error"},
	GetPortalPortingRequestByTxIDStatusError:   {-14013, "Get portal porting request by txID status error"},
	StorePortalPortingRequestByTxIDStatusError: {-14014, "Store portal porting request by txID status error"},
	// redeem
	StoreWaitingRedeemRequestError:                {-14015, "Store waiting redeem requests error"},
	GetPortalRedeemRequestStatusError:             {-14016, "Get portal redeem request status error"},
	StorePortalRedeemRequestStatusError:           {-14017, "Store portal redeem request status error"},
	GetPortalRedeemRequestByTxIDStatusError:       {-14018, "Get portal redeem request by txid status error"},
	StorePortalRedeemRequestByTxIDStatusError:     {-14019, "Store portal redeem request by txid status error"},
	GetPortalRequestUnlockCollateralStatusError:   {-14020, "Get portal request unlock collateral status error"},
	StorePortalRequestUnlockCollateralStatusError: {-14021, "Store portal request unlock collateral status error"},
	StorePortalReqMatchingRedeemByTxIDStatusError: {-14022, "Store req matching redeem request error"},
	GetPortalReqMatchingRedeemByTxIDStatusError:   {-14023, "Get req matching redeem request error"},
	// liquidation
	StoreLiquidationByExchangeRatesStatusError:               {-14024, "Store liquidation by exchange rates status error"},
	GetLiquidationByExchangeRatesStatusError:                 {-14025, "Get liquidation by exchange rates status error"},
	StoreLiquidateExchangeRatesPoolError:                     {-14026, "Store liquidation pool error"},
	GetPortalLiquidationExchangeRatesPoolError:               {-14027, "Get liquidation pool error"},
	GetPortalLiquidationCustodianRunAwayStatusError:          {-14028, "Get portal liquidation custodian run away status error"},
	StorePortalLiquidationCustodianRunAwayStatusError:        {-14029, "Store portal liquidation custodian run away status error"},
	GetPortalExpiredPortingReqStatusError:                    {-14030, "Get portal expired porting request status error"},
	StorePortalExpiredPortingReqStatusError:                  {-14031, "Store portal expired porting request status error"},
	GetPortalRedeemRequestFromLiquidationByTxIDStatusError:   {-14032, "Get portal redeem req from liquidation pool status error"},
	StorePortalRedeemRequestFromLiquidationByTxIDStatusError: {-14033, "Store portal redeem req from liquidation pool status error"},
	GetCustodianTopupStatusError:                             {-14034, "Get custodian topup status error"},
	StoreCustodianTopupStatusError:                           {-14035, "Store custodian topup status error"},
	// exchange rate
	StoreFinalExchangeRatesStateError:   {-14036, "Store final exchange rates request error"},
	StorePortalExchangeRatesStatusError: {-14037, "Store portal exchange rates status error"},
	GetPortalExchangeRatesStatusError:   {-14038, "Get portal exchange rates status error"},
	// reward
	StorePortalRewardError:                      {-14039, "Store portal reward error"},
	GetPortalRequestWithdrawRewardStatusError:   {-14040, "Get portal request withdraw reward status error"},
	StorePortalRequestWithdrawRewardStatusError: {-14041, "Store portal request withdraw reward status error"},
	StoreLockedCollateralStateError:             {-14042, "Store locked collateral state error"},
	GetLockedCollateralStateError:               {-14043, "Get locked collateral state error"},
	// external unique txID
	IsPortalExternalTxHashSubmittedError:     {-14044, "Portal check external tx hash submitted error"},
	InsertPortalExternalTxHashSubmittedError: {-14045, "Portal insert external tx hash submitted error"},
	// portal proof
	StoreWithdrawCollateralConfirmError: {-14046, "Store portal withdraw collateral confirm proof error"},
	GetWithdrawCollateralConfirmError:   {-14047, "Get portal withdraw collateral confirm proof error"},
	// portal unlock over rate collaterals
	StorePortalUnlockOverRateCollateralsError:     {-14048, "Store portal unlock over rate collaterals error"},
	GetPortalUnlockOverRateCollateralsStatusError: {-14049, "Get portal unlock over rate collaterals error"},
	// feature reward
	StoreRewardFeatureError:              {-15000, "Store reward feature state error"},
	GetRewardFeatureError:                {-15001, "Get reward feature state error"},
	GetAllRewardFeatureError:             {-15002, "Get all reward feature state error"},
	GetRewardFeatureAmountByTokenIDError: {-15004, "Get reward feature amount by tokenID error"},
	InvalidStakerInfoTypeError:           {-15005, "Staker info invalid"},

	// Portal v4
	StorePortalV4UTXOsError:                                  {-15006, "Store portal v4 list uxtos error"},
	StorePortalV4ShieldingRequestStatusError:                 {-15007, "Store portal v4 shielding request status error"},
	GetPortalV4ShieldingRequestStatusError:                   {-15008, "Get portal v4 shielding request status error"},
	StorePortalShieldingRequestsError:                        {-15009, "Store portal v4 list shielding requests error"},
	GetPortalShieldingRequestsError:                          {-15010, "Get portal v4 list shielding requests error"},
	GetPortalUnshieldRequestStatusError:                      {-15011, "Get portal v4 unshielding request status error"},
	StorePortalUnshieldRequestStatusError:                    {-15012, "Store portal v4 unshielding request status error"},
	GetPortalBatchUnshieldRequestStatusError:                 {-15013, "Get portal v4 batching unshield request status error"},
	StorePortalBatchUnshieldRequestStatusError:               {-15014, "Store portal v4 batching unshield request status error"},
	StorePortalListWaitingUnshieldRequestError:               {-15015, "Store portal v4 list waiting unshield request error"},
	StorePortalListProcessedBatchUnshieldRequestError:        {-15016, "Store portal v4 list processed batch unshield request error"},
	GetPortalUnshieldBatchFeeReplacementRequestStatusError:   {-15017, "Get portal unshield batch replacement request status error"},
	StorePortalUnshieldBatchFeeReplacementRequestStatusError: {-15018, "Store portal unshield batch replacement request status error"},
	GetPortalSubmitConfirmedTxRequestStatusError:             {-15019, "Get portal submit confirmed tx request status error"},
	StorePortalSubmitConfirmedTxRequestStatusError:           {-15020, "Store portal submit confirmed tx request status error"},
	StorePortalV4StatusError:                                 {-15021, "Store portal v4 status error"},
	GetPortalV4StatusError:                                   {-15022, "Get portal v4 status error"},

	// bsc bridge
	BridgeInsertBSCTxHashIssuedError: {-15100, "Bridge Insert BSC Tx Hash Issued Error"},
	IsBSCTxHashIssuedError:           {-15101, "Is BSC Tx Hash Issued Error"},

	// prv pegging erc20/bep20
	BridgeInsertPRVEVMTxHashIssuedError: {-15102, "Bridge Insert PRV pegging evm Tx Hash Issued Error"},
	IsPRVEVMTxHashIssuedError:           {-15103, "Is PRV pegging evm Tx Hash Issued Error"},

	// polygon bridge
	BridgeInsertPLGTxHashIssuedError: {-15104, "Bridge Insert PLG Tx Hash Issued Error"},
	IsPLGTxHashIssuedError:           {-15105, "Is Polygon Tx Hash Issued Error"},

	// fantom bridge
	BridgeInsertFTMTxHashIssuedError: {-15106, "Bridge Insert Fantom Tx Hash Issued Error"},
	IsFTMTxHashIssuedError:           {-15107, "Is Fantom Tx Hash Issued Error"},

	// aurora bridge
	BridgeInsertAURORATxHashIssuedError: {-15110, "Bridge Insert Aurora Tx Hash Issued Error"},
	IsAURORATxHashIssuedError:           {-15111, "Is Aurora Tx Hash Issued Error"},

	// avax bridge
	BridgeInsertAVAXTxHashIssuedError: {-15112, "Bridge Insert Avax Tx Hash Issued Error"},
	IsAVAXTxHashIssuedError:           {-15113, "Is Avax Tx Hash Issued Error"},

	// bridge agg
	GetBridgeAggStatusError:   {-15108, "Get bridge agg status error"},
	StoreBridgeAggStatusError: {-15109, "Store bridge agg status Error"},

	// near bridge
	BridgeInsertNEARTxHashIssuedError: {-15110, "Insert near shield transaction error"},
	IsNEARTxHashIssuedError:           {-15111, "Is Near Tx Hash Issued Error"},

	// bridge hub
	GetBridgeHubStatusError:   {-15112, "Get bridge hub status error"},
	StoreBridgeHubStatusError: {-15113, "Store bridge hub status Error"},
}

type StatedbError struct {
	err     error
	Code    int
	Message string
}

func (e StatedbError) GetErrorCode() int {
	return e.Code
}
func (e StatedbError) GetError() error {
	return e.err
}
func (e StatedbError) GetMessage() string {
	return e.Message
}

func (e StatedbError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewStatedbError(key int, err error, params ...interface{}) *StatedbError {
	return &StatedbError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
