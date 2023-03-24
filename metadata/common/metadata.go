package common

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/utils"
)

// Interface for all types of metadata in tx
type Metadata interface {
	GetType() int
	Sign(*privacy.PrivateKey, Transaction) error
	VerifyMetadataSignature([]byte, Transaction) (bool, error)
	Hash() *common.Hash
	HashWithoutSig() *common.Hash
	CheckTransactionFee(Transaction, uint64, int64, *statedb.StateDB) bool
	ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error)
	ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error)
	ValidateMetadataByItself() bool
	BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error)
	CalculateSize() uint64
	VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error)
	IsMinerCreatedMetaType() bool
	SetSharedRandom([]byte)
}

// This is tx struct which is really saved in tx mempool
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx Transaction

	// Height is the best block's height when the entry was added to the the source pool.
	Height uint64

	// Fee is the total fee the transaction associated with the entry pays.
	Fee uint64

	// FeeToken is the total token fee the transaction associated with the entry pays.
	// FeeToken is zero if tx is PRV transaction
	FeeToken uint64

	// FeePerKB is the fee the transaction pays in coin per 1000 bytes.
	FeePerKB int32
}

// Interface for mempool which is used in metadata
type MempoolRetriever interface {
	GetSerialNumbersHashH() map[common.Hash][]common.Hash
	GetTxsInMem() map[common.Hash]TxDesc
	GetOTAHashH() map[common.Hash][]common.Hash
}

type ChainRetriever interface {
	GetCentralizedWebsitePaymentAddress(uint64) string
	GetBurningAddress(blockHeight uint64) string
	GetTransactionByHash(common.Hash) (byte, common.Hash, uint64, int, Transaction, error)
	ListPrivacyTokenAndBridgeTokenAndPRVByShardID(byte) ([]common.Hash, error)
	GetBNBChainID() string
	GetBTCChainID() string
	GetBTCHeaderChain() *btcrelaying.BlockChain
	GetBTCChainParams() *chaincfg.Params
	GetShardStakingTx(shardID byte, beaconHeight uint64) (map[string]string, error)
	IsAfterNewZKPCheckPoint(beaconHeight uint64) bool
	IsAfterPrivacyV2CheckPoint(beaconHeight uint64) bool
	IsAfterPdexv3CheckPoint(beaconHeight uint64) bool
	GetPortalFeederAddress(beaconHeight uint64) string
	IsSupportedTokenCollateralV3(beaconHeight uint64, externalTokenID string) bool
	GetPortalETHContractAddrStr(beaconHeight uint64) string
	GetLatestBNBBlkHeight() (int64, error)
	GetBNBDataHash(blockHeight int64) ([]byte, error)
	CheckBlockTimeIsReached(recentBeaconHeight, beaconHeight, recentShardHeight, shardHeight uint64, duration time.Duration) bool
	IsPortalExchangeRateToken(beaconHeight uint64, tokenIDStr string) bool
	GetMinAmountPortalToken(tokenIDStr string, beaconHeight uint64, version uint) (uint64, error)
	IsPortalToken(beaconHeight uint64, tokenIDStr string, version uint) (bool, error)
	IsValidPortalRemoteAddress(tokenIDStr string, remoteAddr string, beaconHeight uint64, version uint) (bool, error)
	ValidatePortalRemoteAddresses(remoteAddresses map[string]string, beaconHeight uint64, version uint) (bool, error)
	IsEnableFeature(featureFlag string, epoch uint64) bool
	GetPortalV4MinUnshieldAmount(tokenIDStr string, beaconHeight uint64) uint64
	GetPortalV4GeneralMultiSigAddress(tokenIDStr string, beaconHeight uint64) string
	GetPortalReplacementAddress(beaconHeight uint64) string
	CheckBlockTimeIsReachedByBeaconHeight(recentBeaconHeight, beaconHeight uint64, duration time.Duration) bool
	GetPortalV4MultipleTokenAmount(tokenIDStr string, beaconHeight uint64) uint64
	GetFinalBeaconHeight() uint64
	GetPdexv3Cached(common.Hash) interface{}
	GetBeaconChainDatabase() incdb.Database
	GetDelegationRewardAmount(stateDB *statedb.StateDB, pk key.PublicKey) (uint64, error)
}

type BeaconViewRetriever interface {
	GetHeight() uint64
	GetAllStakers() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey)
	GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error)
	GetAutoStakingList() map[string]bool
	// GetAllBridgeTokens() ([]common.Hash, error)
	GetBeaconFeatureStateDB() *statedb.StateDB
	GetBeaconRewardStateDB() *statedb.StateDB
	GetBeaconSlashStateDB() *statedb.StateDB
	GetStakerInfo(string) (*statedb.StakerInfo, bool, error)
	GetBeaconStakerInfo(stakerPubkey string) (*statedb.BeaconStakerInfo, bool, error)
	GetShardCommitteeFlattenList() []string
	GetShardPendingFlattenList() []string
	GetBeaconConsensusStateDB() *statedb.StateDB
	CandidateWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	IsValidNftID(incdb.Database, interface{}, string) error
	IsValidPoolPairID(incdb.Database, interface{}, string) error
	IsValidMintNftRequireAmount(incdb.Database, interface{}, uint64) error
	IsValidPdexv3StakingPool(incdb.Database, interface{}, string) error
	IsValidPdexv3UnstakingAmount(incdb.Database, interface{}, string, string, uint64) error
	IsValidPdexv3ShareAmount(incdb.Database, interface{}, string, string, uint64) error
	BlockHash() common.Hash
	GetBeaconLocking() []incognitokey.CommitteePublicKey
}

type ShardViewRetriever interface {
	GetEpoch() uint64
	GetBeaconHeight() uint64
	GetShardID() byte
	GetStakingTx() map[string]string
	ListShardPrivacyTokenAndPRV() []common.Hash
	GetShardRewardStateDB() *statedb.StateDB
	GetCopiedFeatureStateDB() *statedb.StateDB
	GetCopiedTransactionStateDB() *statedb.StateDB
	GetHeight() uint64
	GetBlockVersion() int
	GetTriggeredFeature() map[string]uint64
	GetShardCommittee() []incognitokey.CommitteePublicKey
	GetShardPendingValidator() []incognitokey.CommitteePublicKey
}

type ValidationEnviroment interface {
	IsPrivacy() bool
	IsConfimed() bool
	TxType() string
	TxAction() int
	ShardID() int
	ShardHeight() uint64
	BeaconHeight() uint64
	ConfirmedTime() int64
	Version() int
	SigPubKey() []byte
	HasCA() bool
	TokenID() common.Hash
	DBData() [][]byte
}

// Interface for all type of transaction
type Transaction interface {
	// GET/SET FUNCTION
	GetVersion() int8
	SetVersion(int8)
	GetMetadataType() int
	GetType() string
	SetType(string)
	GetLockTime() int64
	SetLockTime(int64)
	GetSenderAddrLastByte() byte
	SetGetSenderAddrLastByte(byte)
	GetTxFee() uint64
	SetTxFee(uint64)
	GetTxFeeToken() uint64
	GetInfo() []byte
	SetInfo([]byte)
	GetSigPubKey() []byte
	SetSigPubKey([]byte)
	GetSig() []byte
	SetSig([]byte)
	GetProof() privacy.Proof
	SetProof(privacy.Proof)
	GetTokenID() *common.Hash
	GetMetadata() Metadata
	SetMetadata(Metadata)

	// =================== FUNCTIONS THAT GET STUFF AND REQUIRE SOME CODING ===================
	GetTxActualSize() uint64
	GetReceivers() ([][]byte, []uint64)
	GetTransferData() (bool, []byte, uint64, *common.Hash)
	GetReceiverData() ([]coin.Coin, error)
	GetTxMintData() (bool, coin.Coin, *common.Hash, error)
	GetTxBurnData() (bool, coin.Coin, *common.Hash, error)
	GetTxFullBurnData() (bool, coin.Coin, coin.Coin, *common.Hash, error)
	ListOTAHashH() []common.Hash
	ListSerialNumbersHashH() []common.Hash
	String() string
	Hash() *common.Hash
	HashWithoutMetadataSig() *common.Hash
	CalculateTxValue() uint64

	// =================== FUNCTION THAT CHECK STUFFS  ===================
	CheckTxVersion(int8) bool
	IsSalaryTx() bool
	IsPrivacy() bool
	IsCoinsBurning(ChainRetriever, ShardViewRetriever, BeaconViewRetriever, uint64) bool

	// =================== FUNCTIONS THAT VALIDATE STUFFS ===================
	ValidateTxSalary(*statedb.StateDB) (bool, error)
	ValidateTxWithCurrentMempool(MempoolRetriever) error
	ValidateSanityData(ChainRetriever, ShardViewRetriever, BeaconViewRetriever, uint64) (bool, error)

	ValidateTxWithBlockChain(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error
	ValidateDoubleSpendWithBlockchain(byte, *statedb.StateDB, *common.Hash) error
	ValidateTxByItself(map[string]bool, *statedb.StateDB, *statedb.StateDB, ChainRetriever, byte, ShardViewRetriever, BeaconViewRetriever) (bool, error)
	ValidateType() bool
	ValidateTransaction(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash) (bool, []privacy.Proof, error)
	VerifyMinerCreatedTxBeforeGettingInBlock(*MintData, byte, ChainRetriever, *AccumulatedValues, ShardViewRetriever, BeaconViewRetriever) (bool, error)

	// Init Transaction, the input should be params such as: TxPrivacyInitParams
	Init(interface{}) error
	// Verify the init function above, which verify zero knowledge proof and signatures
	Verify(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash) (bool, error)

	GetValidationEnv() ValidationEnviroment
	SetValidationEnv(ValidationEnviroment)
	UnmarshalJSON(data []byte) error

	// VerifySigTx() (bool, error)
	ValidateSanityDataByItSelf() (bool, error)
	ValidateTxCorrectness(db *statedb.StateDB) (bool, error)
	LoadData(db *statedb.StateDB) error
	CheckData(db *statedb.StateDB) error
	ValidateSanityDataWithBlockchain(
		chainRetriever ChainRetriever,
		shardViewRetriever ShardViewRetriever,
		beaconViewRetriever BeaconViewRetriever,
		beaconHeight uint64,
	) (
		bool,
		error,
	)
}

type MintData struct {
	ReturnStaking  map[string]bool
	WithdrawReward map[string]bool
	Txs            []Transaction
	TxsUsed        []int
	Insts          [][]string
	InstsUsed      []int
}

func getPDEPoolPair(
	prvIDStr, tokenIDStr string,
	beaconHeight int64,
	stateDB *statedb.StateDB,
) (*rawdbv2.PDEPoolForPair, error) {
	var pdePoolForPair rawdbv2.PDEPoolForPair
	var err error
	poolPairBytes := []byte{}
	if beaconHeight == -1 {
		poolPairBytes, err = statedb.GetLatestPDEPoolForPair(stateDB, prvIDStr, tokenIDStr)
	} else {
		poolPairBytes, err = statedb.GetPDEPoolForPair(stateDB, uint64(beaconHeight), prvIDStr, tokenIDStr)
	}
	if err != nil {
		return nil, err
	}
	if len(poolPairBytes) == 0 {
		return nil, NewMetadataTxError(CouldNotGetExchangeRateError, fmt.Errorf("Could not find out pdePoolForPair with token ids: %s & %s", prvIDStr, tokenIDStr))
	}
	err = json.Unmarshal(poolPairBytes, &pdePoolForPair)
	if err != nil {
		return nil, err
	}
	return &pdePoolForPair, nil
}

func isPairValid(poolPair *rawdbv2.PDEPoolForPair, beaconHeight int64) bool {
	if poolPair == nil {
		return false
	}
	prvIDStr := common.PRVCoinID.String()
	if poolPair.Token1IDStr == prvIDStr &&
		poolPair.Token1PoolValue < uint64(common.MinTxFeesOnTokenRequirement) &&
		beaconHeight >= common.BeaconBlockHeighMilestoneForMinTxFeesOnTokenRequirement {
		return false
	}
	if poolPair.Token2IDStr == prvIDStr &&
		poolPair.Token2PoolValue < uint64(common.MinTxFeesOnTokenRequirement) &&
		beaconHeight >= common.BeaconBlockHeighMilestoneForMinTxFeesOnTokenRequirement {
		return false
	}
	return true
}

func convertValueBetweenCurrencies(
	amount uint64,
	currentCurrencyIDStr string,
	tokenID *common.Hash,
	beaconHeight int64,
	stateDB *statedb.StateDB,
) (float64, error) {
	prvIDStr := common.PRVCoinID.String()
	tokenIDStr := tokenID.String()
	pdePoolForPair, err := getPDEPoolPair(prvIDStr, tokenIDStr, beaconHeight, stateDB)
	if err != nil {
		return 0, NewMetadataTxError(CouldNotGetExchangeRateError, err)
	}
	if !isPairValid(pdePoolForPair, beaconHeight) {
		return 0, NewMetadataTxError(CouldNotGetExchangeRateError, errors.New("PRV pool size on pdex is smaller minimum initial adding liquidity amount"))
	}
	invariant := float64(0)
	invariant = float64(pdePoolForPair.Token1PoolValue) * float64(pdePoolForPair.Token2PoolValue)
	if invariant == 0 {
		return 0, NewMetadataTxError(CouldNotGetExchangeRateError, err)
	}
	if pdePoolForPair.Token1IDStr == currentCurrencyIDStr {
		remainingValue := invariant / (float64(pdePoolForPair.Token1PoolValue) + float64(amount))
		if float64(pdePoolForPair.Token2PoolValue) <= remainingValue {
			return 0, NewMetadataTxError(CouldNotGetExchangeRateError, err)
		}
		return float64(pdePoolForPair.Token2PoolValue) - remainingValue, nil
	}
	remainingValue := invariant / (float64(pdePoolForPair.Token2PoolValue) + float64(amount))
	if float64(pdePoolForPair.Token1PoolValue) <= remainingValue {
		return 0, NewMetadataTxError(CouldNotGetExchangeRateError, err)
	}
	return float64(pdePoolForPair.Token1PoolValue) - remainingValue, nil
}

// return error if there is no exchange rate between native token and privacy token
// beaconHeight = -1: get the latest beacon height
func ConvertNativeTokenToPrivacyToken(
	nativeTokenAmount uint64,
	tokenID *common.Hash,
	beaconHeight int64,
	stateDB *statedb.StateDB,
) (float64, error) {
	return convertValueBetweenCurrencies(
		nativeTokenAmount,
		common.PRVCoinID.String(),
		tokenID,
		beaconHeight,
		stateDB,
	)
}

// return error if there is no exchange rate between native token and privacy token
// beaconHeight = -1: get the latest beacon height
func ConvertPrivacyTokenToNativeToken(
	privacyTokenAmount uint64,
	tokenID *common.Hash,
	beaconHeight int64,
	stateDB *statedb.StateDB,
) (float64, error) {
	return convertValueBetweenCurrencies(
		privacyTokenAmount,
		tokenID.String(),
		tokenID,
		beaconHeight,
		stateDB,
	)
}

type Action struct {
	Meta      Metadata    `json:"Meta"`
	TxReqID   common.Hash `json:"TxReqID"`
	ExtraData [][]byte    `json:"ExtraData,omitempty"`
}

func NewAction() *Action {
	return &Action{}
}

func NewActionWithValue(
	meta Metadata, txReqID common.Hash, extraData [][]byte,
) *Action {
	return &Action{
		Meta:      meta,
		TxReqID:   txReqID,
		ExtraData: extraData,
	}
}

func (a *Action) FromString(source string) error {
	contentBytes, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(contentBytes, &a)
	if err != nil {
		return err
	}
	return nil
}

func (a *Action) StringSlice(metaType int) ([]string, error) {
	contentBytes, err := json.Marshal(a)
	if err != nil {
		return []string{}, err
	}
	contentStr := base64.StdEncoding.EncodeToString(contentBytes)
	action := []string{strconv.Itoa(metaType), contentStr}
	return action, nil
}

type Instruction struct {
	MetaType int    `json:"MetaType"`
	Status   string `json:"Status"`
	ShardID  byte   `json:"ShardID"`
	Content  string `json:"Content,omitempty"`
}

func NewInstruction() *Instruction {
	return &Instruction{}
}

func NewInstructionWithValue(
	metaType int, status string, shardID byte, content string,
) *Instruction {
	return &Instruction{
		MetaType: metaType,
		Status:   status,
		ShardID:  shardID,
		Content:  content,
	}
}

func (i *Instruction) StringSlice() []string {
	return []string{
		strconv.Itoa(i.MetaType),
		strconv.Itoa(int(i.ShardID)),
		i.Status,
		i.Content,
	}
}

func (i *Instruction) FromStringSlice(source []string) error {
	if len(source) != 4 {
		return fmt.Errorf("Expect length of instructions is %d but get %d", 4, len(source))
	}
	metaType, err := strconv.Atoi(source[0])
	if err != nil {
		return err
	}
	i.MetaType = metaType
	shardID, err := strconv.Atoi(source[1])
	if err != nil {
		return err
	}
	i.ShardID = byte(shardID)
	i.Status = source[2]
	i.Content = source[3]
	return nil
}

func (i *Instruction) StringSliceWithRejectContent(rejectContent *RejectContent) ([]string, error) {
	str, err := rejectContent.String()
	i.Content = str
	i.Status = common.RejectedStatusStr
	return i.StringSlice(), err
}

type RejectContent struct {
	TxReqID   common.Hash `json:"TxReqID"`
	ErrorCode int         `json:"ErrorCode,omitempty"`
	Data      []byte      `json:"Data,omitempty"`
}

func NewRejectContent() *RejectContent { return &RejectContent{} }

func NewRejectContentWithValue(txReqID common.Hash, errorCode int, data []byte) *RejectContent {
	return &RejectContent{TxReqID: txReqID, ErrorCode: errorCode, Data: data}
}

func (r *RejectContent) String() (string, error) {
	contentBytes, err := json.Marshal(r)
	if err != nil {
		return utils.EmptyString, err
	}
	return base64.StdEncoding.EncodeToString(contentBytes), nil
}

func (r *RejectContent) FromString(source string) error {
	contentBytes, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(contentBytes, &r)
	if err != nil {
		return err
	}
	return nil
}
