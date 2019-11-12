package lvdb

import "github.com/incognitochain/incognito-chain/common"

// key prefix
var (
	prevShardPrefix          = []byte("prevShd-")
	prevBeaconPrefix         = []byte("prevBea-")
	beaconPrefix             = []byte("bea-")
	beaconBestBlockkeyPrefix = []byte("bea-bestBlock")
	committeePrefix          = []byte("com-")
	rewardReceiverPrefix     = []byte("rewardreceiver-")
	heightPrefix             = []byte("height-")
	shardIDPrefix            = []byte("s-")
	blockKeyPrefix           = []byte("b-")
	blockHeaderKeyPrefix     = []byte("bh-")
	blockKeyIdxPrefix        = []byte("i-")
	crossShardKeyPrefix      = []byte("csh-")
	nextCrossShardKeyPrefix  = []byte("ncsh-")
	shardPrefix              = []byte("shd-")
	autoStakingPrefix        = []byte("aust-")

	shardToBeaconKeyPrefix       = []byte("stb-")
	transactionKeyPrefix         = []byte("tx-")
	privateKeyPrefix             = []byte("prk-")
	serialNumbersPrefix          = []byte("serinalnumbers-")
	commitmentsPrefix            = []byte("commitments-")
	outcoinsPrefix               = []byte("outcoins-")
	snderivatorsPrefix           = []byte("snderivators-")
	bestBlockKeyPrefix           = []byte("bestBlock")
	feeEstimatorPrefix           = []byte("feeEstimator")
	tokenPrefix                  = []byte("token-")
	privacyTokenPrefix           = []byte("privacy-token-")
	privacyTokenCrossShardPrefix = []byte("privacy-cross-token-")
	tokenInitPrefix              = []byte("token-init-")
	privacyTokenInitPrefix       = []byte("privacy-token-init-")

	// multisigs
	multisigsPrefix = []byte("multisigs")

	// centralized bridge
	bridgePrefix              = []byte("bridge-")
	centralizedBridgePrefix   = []byte("centralizedbridge-")
	decentralizedBridgePrefix = []byte("decentralizedbridge-")
	ethTxHashIssuedPrefix     = []byte("ethtxhashissued-")

	// Incognito -> Ethereum relayer
	burnConfirmPrefix = []byte("burnConfirm-")

	//epoch reward
	shardRequestRewardPrefix = []byte("shardrequestreward-")
	committeeRewardPrefix    = []byte("committee-reward-")

	// public variable
	TokenPaymentAddressPrefix = []byte("token-paymentaddress-")
	Splitter                  = []byte("-[-]-")

	// slash
	producersBlackListPrefix = []byte("producersblacklist-")

	// PDE
	WaitingPDEContributionPrefix = []byte("waitingpdecontribution-")
	PDEPoolPrefix                = []byte("pdepool-")
	PDESharePrefix               = []byte("pdeshare-")
	PDETradeFeePrefix            = []byte("pdetradefee-")
	PDEContributionStatusPrefix  = []byte("pdecontributionstatus-")
	PDETradeStatusPrefix         = []byte("pdetradestatus-")
	PDEWithdrawalStatusPrefix    = []byte("pdewithdrawalstatus-")
)

// value
var (
	Spent   = []byte("spent")
	Unspent = []byte("unspent")
)

func addPrefixToKeyHash(keyType string, keyHash common.Hash) []byte {
	var dbkey []byte
	switch keyType {
	case string(blockKeyPrefix):
		dbkey = append(blockKeyPrefix, keyHash[:]...)
	case string(blockKeyIdxPrefix):
		dbkey = append(blockKeyIdxPrefix, keyHash[:]...)
	case string(serialNumbersPrefix):
		dbkey = append(serialNumbersPrefix, keyHash[:]...)
	case string(commitmentsPrefix):
		dbkey = append(commitmentsPrefix, keyHash[:]...)
	case string(outcoinsPrefix):
		dbkey = append(outcoinsPrefix, keyHash[:]...)
	case string(snderivatorsPrefix):
		dbkey = append(snderivatorsPrefix, keyHash[:]...)
	case string(tokenPrefix):
		dbkey = append(tokenPrefix, keyHash[:]...)
	case string(privacyTokenPrefix):
		dbkey = append(privacyTokenPrefix, keyHash[:]...)
	case string(privacyTokenCrossShardPrefix):
		dbkey = append(privacyTokenCrossShardPrefix, keyHash[:]...)
	case string(tokenInitPrefix):
		dbkey = append(tokenInitPrefix, keyHash[:]...)
	case string(privacyTokenInitPrefix):
		dbkey = append(privacyTokenInitPrefix, keyHash[:]...)
	}
	return dbkey
}
