package rawdb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"sort"
)

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

// TODO - change json to CamelCase
type BridgeTokenInfo struct {
	TokenID         *common.Hash `json:"tokenId"`
	Amount          uint64       `json:"amount"`
	ExternalTokenID []byte       `json:"externalTokenId"`
	Network         string       `json:"network"`
	IsCentralized   bool         `json:"isCentralized"`
}

type PDEContribution struct {
	ContributorAddressStr string
	TokenIDStr            string
	Amount                uint64
	TxReqID               common.Hash
}

type PDEPoolForPair struct {
	Token1IDStr     string
	Token1PoolValue uint64
	Token2IDStr     string
	Token2PoolValue uint64
}

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

func getBridgePrefix(isCentralized bool) []byte {
	if isCentralized {
		return centralizedBridgePrefix
	}
	return decentralizedBridgePrefix
}

/**
 * NewKeyAddShardRewardRequest create a key for store reward of a shard X at epoch T in db.
 * @param epoch: epoch T
 * @param shardID: shard X
 * @param tokenID: currency unit
 * @return ([]byte, error): Key, error of this process
 */
func newKeyAddShardRewardRequest(
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
) []byte {
	res := []byte{}
	res = append(res, shardRequestRewardPrefix...)
	res = append(res, common.Uint64ToBytes(epoch)...)
	res = append(res, shardID)
	res = append(res, tokenID.GetBytes()...)
	return res
}

/**
 * NewKeyAddCommitteeReward create a key for store reward of a person P in committee in db.
 * @param committeeAddress: Public key of person P
 * @param tokenID: currency unit
 * @return ([]byte, error): Key, error of this process
 */
func newKeyAddCommitteeReward(
	committeeAddress []byte,
	tokenID common.Hash,
) []byte {
	res := []byte{}
	res = append(res, committeeRewardPrefix...)
	res = append(res, committeeAddress...)
	res = append(res, tokenID.GetBytes()...)
	return res
}

func getPrevPrefix(isBeacon bool, shardID byte) []byte {
	key := []byte{}
	if isBeacon {
		key = append(key, prevBeaconPrefix...)
	} else {
		key = append(key, append(prevShardPrefix, append([]byte{shardID}, byte('-'))...)...)
	}
	return key
}

func BuildPDEStatusKey(
	prefix []byte,
	suffix []byte,
) []byte {
	return append(prefix, suffix...)
}

func BuildPDESharesKey(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddressStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdeSharesByBCHeightPrefix := append(PDESharePrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdeSharesByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributedTokenIDStr+"-"+contributorAddressStr)...)
}

func BuildPDESharesKeyV2(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributorAddressStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdeSharesByBCHeightPrefix := append(PDESharePrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdeSharesByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributorAddressStr)...)
}

func BuildPDEPoolForPairKey(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdePoolForPairByBCHeightPrefix := append(PDEPoolPrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdePoolForPairByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1])...)
}

func BuildPDETradeFeesKey(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	tokenForFeeIDStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdeTradeFeesByBCHeightPrefix := append(PDETradeFeePrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdeTradeFeesByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+tokenForFeeIDStr)...)
}

func BuildWaitingPDEContributionKey(
	beaconHeight uint64,
	pairID string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	waitingPDEContribByBCHeightPrefix := append(WaitingPDEContributionPrefix, beaconHeightBytes...)
	return append(waitingPDEContribByBCHeightPrefix, []byte(pairID)...)
}
