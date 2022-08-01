package common

// for common
const (
	EmptyString       = ""
	ZeroByte          = byte(0x00)
	DateOutputFormat  = "2006-01-02T15:04:05.999999"
	BigIntSize        = 32 // bytes
	CheckSumLen       = 4  // bytes
	AESKeySize        = 32 // bytes
	Int32Size         = 4  // bytes
	Uint32Size        = 4  // bytes
	Uint64Size        = 8  // bytes
	HashSize          = 32 // bytes
	MaxHashStringSize = HashSize * 2
	Base58Version     = 0
)

// size data for incognito key and signature
const (
	// for key size
	PrivateKeySize      = 32  // bytes
	PublicKeySize       = 32  // bytes
	BLSPublicKeySize    = 128 // bytes
	BriPublicKeySize    = 33  // bytes
	TransmissionKeySize = 32  //bytes
	ReceivingKeySize    = 32  // bytes
	PaymentAddressSize  = 64  // bytes
	// for signature size
	// it is used for both privacy and no privacy
	SigPubKeySize    = 32
	SigNoPrivacySize = 64
	SigPrivacySize   = 96
	IncPubKeyB58Size = 51

	MaxPSMsgSize = 1 << 23 //4Mb
)

// For all Transaction information
const (
	TxNormalType          = "n"   // normal tx(send and receive coin)
	TxRewardType          = "s"   // reward tx
	TxReturnStakingType   = "rs"  //
	TxConversionType      = "cv"  // Convert 1 - 2 normal tx
	TxTokenConversionType = "tcv" // Convert 1 - 2 token tx
	//TxCustomTokenType        = "t"  // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "tp" // token  tx with supporting privacy
)

const (
	TxActTranfer = iota // Tx for tranfer PRV/Token
	TxActBurning        // Tx burn PRV/Token
	TxActInit           // Tx init PRV/Token
)

const (
	MaxTxRequestIssue = 400
)

var (
	MaxTxSize    = uint64(500)  // unit KB = 500KB
	MaxBlockSize = uint64(2000) //unit kilobytes = 2 Megabyte
)

// special token ids (aka. PropertyID in custom token)
var (
	PRVCoinID             = Hash{4} // To send PRV in custom token
	PRVCoinName           = "PRV"   // To send PRV in custom token
	ConfidentialAssetID   = Hash{5}
	ConfidentialAssetName = "CA"
	PDEXCoinID            = Hash{6}
	PDEXCoinName          = "PDEX"
	MaxShardNumber        = 0
)

// CONSENSUS
const (
	// NodeModeRelay  = "relay"
	// NodeModeShard  = "shard"
	// NodeModeAuto   = "auto"
	// NodeModeBeacon = "beacon"

	BeaconRole    = "beacon"
	ShardRole     = "shard"
	CommitteeRole = "committee"
	PendingRole   = "pending"
	SyncingRole   = "syncing" //this is for shard case - when beacon tell it is committee, but its state not
	WaitingRole   = "waiting"

	BlsConsensus    = "bls"
	BridgeConsensus = "dsa"
	IncKeyType      = "inc"
)

const (
	SALARY_VER_FIX_HASH = 1
)

const (
	BeaconChainKey = "beacon"
	ShardChainKey  = "shard"
)

const (
	BeaconChainID                = -1
	BeaconChainSyncID            = 255
	BeaconChainDatabaseDirectory = "beacon"
	ShardChainDatabaseDirectory  = "shard"
)

const (
	REPLACE_IN  = 0
	REPLACE_OUT = 1
)

// Ethereum Decentralized bridge
const (
	AbiJson                   = `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"string","name":"incognitoAddress","type":"string"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Deposit","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"string","name":"incognitoAddress","type":"string"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"depositID","type":"uint256"}],"name":"DepositV2","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"bytes","name":"redepositIncAddress","type":"bytes"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"bytes32","name":"itx","type":"bytes32"}],"name":"Redeposit","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"newIncognitoProxy","type":"address"}],"name":"UpdateIncognitoProxy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address[]","name":"assets","type":"address[]"},{"indexed":false,"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"name":"UpdateTokenTotal","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Withdraw","type":"event"},{"inputs":[],"name":"BURN_CALL_REQUEST_METADATA_TYPE","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"BURN_REQUEST_METADATA_TYPE","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"CURRENT_NETWORK_ID","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"ETH_TOKEN","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes","name":"externalCalldata","type":"bytes"},{"internalType":"address","name":"redepositToken","type":"address"}],"name":"_callExternal","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"addresspayable","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"_transferExternal","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"incognitoAddress","type":"string"}],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"incognitoAddress","type":"string"}],"name":"depositERC20","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"incognitoAddress","type":"string"}],"name":"depositERC20_V2","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"string","name":"incognitoAddress","type":"string"}],"name":"deposit_V2","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"address","name":"recipientToken","type":"address"},{"internalType":"address","name":"exchangeAddress","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"},{"internalType":"bytes","name":"timestamp","type":"bytes"},{"internalType":"bytes","name":"signData","type":"bytes"}],"name":"execute","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bytes","name":"inst","type":"bytes"},{"internalType":"uint256","name":"heights","type":"uint256"},{"internalType":"bytes32[]","name":"instPaths","type":"bytes32[]"},{"internalType":"bool[]","name":"instPathIsLefts","type":"bool[]"},{"internalType":"bytes32","name":"instRoots","type":"bytes32"},{"internalType":"bytes32","name":"blkData","type":"bytes32"},{"internalType":"uint256[]","name":"sigIdxs","type":"uint256[]"},{"internalType":"uint8[]","name":"sigVs","type":"uint8[]"},{"internalType":"bytes32[]","name":"sigRs","type":"bytes32[]"},{"internalType":"bytes32[]","name":"sigSs","type":"bytes32[]"}],"name":"executeWithBurnProof","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"getDecimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"}],"name":"getDepositedBalance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_prevVault","type":"address"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"isInitialized","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"hash","type":"bytes32"}],"name":"isSigDataUsed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"hash","type":"bytes32"}],"name":"isWithdrawed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],"name":"migration","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"notEntered","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes","name":"inst","type":"bytes"}],"name":"parseBurnInst","outputs":[{"components":[{"internalType":"uint8","name":"meta","type":"uint8"},{"internalType":"uint8","name":"shard","type":"uint8"},{"internalType":"address","name":"token","type":"address"},{"internalType":"addresspayable","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes32","name":"itx","type":"bytes32"}],"internalType":"structVault.BurnInstData","name":"","type":"tuple"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"bytes","name":"inst","type":"bytes"}],"name":"parseCalldataFromBurnInst","outputs":[{"components":[{"internalType":"uint8","name":"meta","type":"uint8"},{"internalType":"uint8","name":"shard","type":"uint8"},{"internalType":"address","name":"token","type":"address"},{"internalType":"addresspayable","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes32","name":"itx","type":"bytes32"}],"internalType":"structVault.BurnInstData","name":"","type":"tuple"},{"components":[{"internalType":"address","name":"redepositToken","type":"address"},{"internalType":"bytes","name":"redepositIncAddress","type":"bytes"},{"internalType":"addresspayable","name":"withdrawAddress","type":"address"}],"internalType":"structVault.RedepositOptions","name":"","type":"tuple"},{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"prevVault","outputs":[{"internalType":"contractWithdrawable","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"incognitoAddress","type":"string"},{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes","name":"signData","type":"bytes"},{"internalType":"bytes","name":"timestamp","type":"bytes"}],"name":"requestWithdraw","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"sigDataUsed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes","name":"signData","type":"bytes"},{"internalType":"bytes32","name":"hash","type":"bytes32"}],"name":"sigToAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"bytes","name":"inst","type":"bytes"},{"internalType":"uint256","name":"heights","type":"uint256"},{"internalType":"bytes32[]","name":"instPaths","type":"bytes32[]"},{"internalType":"bool[]","name":"instPathIsLefts","type":"bool[]"},{"internalType":"bytes32","name":"instRoots","type":"bytes32"},{"internalType":"bytes32","name":"blkData","type":"bytes32"},{"internalType":"uint256[]","name":"sigIdxs","type":"uint256[]"},{"internalType":"uint8[]","name":"sigVs","type":"uint8[]"},{"internalType":"bytes32[]","name":"sigRs","type":"bytes32[]"},{"internalType":"bytes32[]","name":"sigSs","type":"bytes32[]"}],"name":"submitBurnProof","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"totalDepositedToSCAmount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address[]","name":"assets","type":"address[]"},{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"name":"updateAssets","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes","name":"inst","type":"bytes"},{"internalType":"uint256","name":"heights","type":"uint256"},{"internalType":"bytes32[]","name":"instPaths","type":"bytes32[]"},{"internalType":"bool[]","name":"instPathIsLefts","type":"bool[]"},{"internalType":"bytes32","name":"instRoots","type":"bytes32"},{"internalType":"bytes32","name":"blkData","type":"bytes32"},{"internalType":"uint256[]","name":"sigIdxs","type":"uint256[]"},{"internalType":"uint8[]","name":"sigVs","type":"uint8[]"},{"internalType":"bytes32[]","name":"sigRs","type":"bytes32[]"},{"internalType":"bytes32[]","name":"sigSs","type":"bytes32[]"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],"name":"withdrawRequests","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"withdrawed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"stateMutability":"payable","type":"receive"}]`
	BridgeShardID             = 1
	EthAddrStr                = "0x0000000000000000000000000000000000000000"
	NativeToken               = "0x0000000000000000000000000000000000000000"
	BSCPrefix                 = "BSC"
	PLGPrefix                 = "PLG"
	FTMPrefix                 = "FTM"
	ExternalBridgeTokenLength = 20
	EVMAddressLength          = 40
	UnifiedTokenPrefix        = "UT"
	NEARPrefix                = "NAR"
)

// Bridge, PDE & Portal statuses for RPCs
const (
	BridgeRequestNotFoundStatus   = 0
	BridgeRequestProcessingStatus = 1
	BridgeRequestAcceptedStatus   = 2
	BridgeRequestRejectedStatus   = 3

	PDENotFoundStatus = 0

	PDEContributionWaitingStatus          = 1
	PDEContributionAcceptedStatus         = 2
	PDEContributionRefundStatus           = 3
	PDEContributionMatchedNReturnedStatus = 4

	PDETradeAcceptedStatus = 1
	PDETradeRefundStatus   = 2

	PDECrossPoolTradeAcceptedStatus = 1
	PDECrossPoolTradeRefundStatus   = 2

	PDEWithdrawalAcceptedStatus = 1
	PDEWithdrawalRejectedStatus = 2

	PDEFeeWithdrawalAcceptedStatus = 1
	PDEFeeWithdrawalRejectedStatus = 2

	MinTxFeesOnTokenRequirement                             = 10000000000000 // 10000 prv, this requirement is applied from beacon height 87301 mainnet
	BeaconBlockHeighMilestoneForMinTxFeesOnTokenRequirement = 87301          // milestone of beacon height, when apply min fee on token requirement
)

// PDE statuses for chain
const (
	PDEContributionWaitingChainStatus          = "waiting"
	PDEContributionMatchedChainStatus          = "matched"
	PDEContributionRefundChainStatus           = "refund"
	PDEContributionMatchedNReturnedChainStatus = "matchedNReturned"

	PDETradeAcceptedChainStatus = "accepted"
	PDETradeRefundChainStatus   = "refund"

	PDEWithdrawalAcceptedChainStatus = "accepted"
	PDEWithdrawalRejectedChainStatus = "rejected"

	PDEFeeWithdrawalAcceptedChainStatus = "accepted"
	PDEFeeWithdrawalRejectedChainStatus = "rejected"

	PDEWithdrawalOnFeeAcceptedChainStatus      = "onFeeAccepted"
	PDEWithdrawalOnPoolPairAcceptedChainStatus = "onPoolPairAccepted"
	PDEWithdrawalWithPRVFeeRejectedChainStatus = "withPRVFeeRejected"

	PDECrossPoolTradeFeeRefundChainStatus          = "xPoolTradeRefundFee"
	PDECrossPoolTradeSellingTokenRefundChainStatus = "xPoolTradeRefundSellingToken"
	PDECrossPoolTradeAcceptedChainStatus           = "xPoolTradeAccepted"
)

const (
	Pdexv3RejectUserMintNftStatus = "rejected"
	Pdexv3AcceptUserMintNftStatus = "accept"
	Pdexv3RejectStakingStatus     = "rejected"
	Pdexv3AcceptStakingStatus     = "accept"
	Pdexv3RejectUnstakingStatus   = "rejected"
	Pdexv3AcceptUnstakingStatus   = "accept"

	Pdexv3AcceptStatus = byte(1)
	Pdexv3RejectStatus = byte(2)
)

const (
	RejectedStatusStr  = "rejected"
	AcceptedStatusStr  = "accepted"
	WaitingStatusStr   = "waiting"
	FilledStatusStr    = "filled"
	RejectedStatusByte = byte(0)
	AcceptedStatusByte = byte(1)
	WaitingStatusByte  = byte(2)
	FilledStatusByte   = byte(3)
)

const PRVIDStr = "0000000000000000000000000000000000000000000000000000000000000004"
const PDEXIDStr = "0000000000000000000000000000000000000000000000000000000000000006"
const PDEXDenominatingDecimal = 9

const ETHChainName = "eth"

const (
	HexEmptyRoot = "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
)

// burning addresses
const (
	BurningAddress  = "15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs"
	BurningAddress2 = "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
)

var (
	EmptyRoot = HexToHash(HexEmptyRoot)
)

/* ================ Feature Flags ================ */
const (
	PortalRelayingFlag = "PortalRelaying"
	PortalV3Flag       = "PortalV3"
	PortalV4Flag       = "PortalV4"
)
const (
	PortalVersion3 = 3
	PortalVersion4 = 4
)

var TIMESLOT = uint64(0) //need to be set when init chain

var (
	BurningAddressByte  []byte
	BurningAddressByte2 []byte
)

// Add to the end of the list. DO NOT edit the order
const (
	DefaultNetworkID = iota
	ETHNetworkID
	BSCNetworkID
	PLGNetworkID
	FTMNetworkID
)

const (
	EVMNetworkType = iota
)

const (
	SubOperator = 0
	AddOperator = 1
)
