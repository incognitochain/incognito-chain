package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/basemeta"

	"math/big"
	"sort"
	"strconv"
	"strings"
)

// PortalCustodianDepositV3 - portal custodian deposit collaterals on SC (ETH/ERC20)
// metadata - custodian deposit - create normal tx with this metadata
type PortalCustodianDepositV3 struct {
	basemeta.MetadataBase
	RemoteAddresses map[string]string // tokenID: remote address

	// ETH proof
	BlockHash eCommon.Hash
	TxIndex   uint
	ProofStrs []string
}

// PortalCustodianDepositActionV3 - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
type PortalCustodianDepositActionV3 struct {
	Meta    PortalCustodianDepositV3
	TxReqID common.Hash
	ShardID byte
}

// PortalCustodianDepositContentV3 - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and refund status
type PortalCustodianDepositContentV3 struct {
	IncAddressStr    string
	RemoteAddresses  map[string]string // tokenID: remote address
	DepositAmount    uint64
	ExternalTokenID  string // collateral TokenID
	UniqExternalTxID []byte
	TxReqID          common.Hash
	ShardID          byte
}

// PortalCustodianDepositStatus - Beacon tracks status of custodian deposit tx into db
type PortalCustodianDepositStatusV3 struct {
	Status           byte
	IncAddressStr    string
	RemoteAddresses  map[string]string // tokenID: remote address
	DepositAmount    uint64
	ExternalTokenID  string // collateral TokenID
	UniqExternalTxID []byte
}

func NewPortalCustodianDepositV3(
	metaType int,
	remoteAddrs map[string]string,
	blockHash eCommon.Hash,
	txIndex uint,
	proofStrs []string) (*PortalCustodianDepositV3, error) {
	custodianDepositMetaV3 := &PortalCustodianDepositV3{
		MetadataBase: basemeta.MetadataBase{
			Type: metaType,
		},
		RemoteAddresses: remoteAddrs,
		BlockHash:       blockHash,
		TxIndex:         txIndex,
		ProofStrs:       proofStrs,
	}
	return custodianDepositMetaV3, nil
}

func NewPortalCustodianDepositV3FromMap(
	data map[string]interface{},
) (*PortalCustodianDepositV3, error) {
	remoteAddressesMap, ok := data["RemoteAddresses"].(map[string]interface{})
	if !ok {
		return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata RemoteAddresses param is invalid"))
	}
	if len(remoteAddressesMap) < 1 {
		return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata RemoteAddresses must be at least one"))
	}
	remoteAddresses := make(map[string]string, 0)
	tokenIDKeys := make([]string, 0)
	for pTokenID, remoteAddress := range remoteAddressesMap {
		if !IsPortalToken(pTokenID) {
			return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata public token is not supported currently"))
		}
		_, ok := remoteAddress.(string)
		if !ok {
			return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata RemoteAddresses is invalid"))
		}
		tokenIDKeys = append(tokenIDKeys, pTokenID)
	}
	sort.Strings(tokenIDKeys)
	for _, pTokenID := range tokenIDKeys {
		remoteAddresses[pTokenID] = remoteAddressesMap[pTokenID].(string)
	}

	blockHashStr, ok := data["BlockHash"].(string)
	if !ok {
		return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata BlockHash should be a string"))
	}
	blockHash := eCommon.HexToHash(blockHashStr)

	txIndexFloat, ok := data["TxIndex"].(float64)
	if !ok {
		return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata TxIndex should be a number"))
	}
	txIdx := uint(txIndexFloat)

	proofsRaw, ok := data["ProofStrs"].([]interface{})
	if !ok {
		return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata ProofStrs should be an array of string"))
	}
	proofStrs := []string{}
	for _, item := range proofsRaw {
		pItem, ok := item.(string)
		if !ok {
			return nil, basemeta.NewMetadataTxError(basemeta.NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata ProofStrs should be an array of string"))
		}
		proofStrs = append(proofStrs, pItem)
	}

	meta, _ := NewPortalCustodianDepositV3(
		basemeta.PortalCustodianDepositMetaV3,
		remoteAddresses,
		blockHash,
		txIdx,
		proofStrs,
	)
	return meta, nil
}

func (custodianDeposit PortalCustodianDepositV3) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever,
	shardViewRetriever basemeta.ShardViewRetriever,
	beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalCustodianDepositV3) ValidateSanityData(
	chainRetriever basemeta.ChainRetriever,
	shardViewRetriever basemeta.ShardViewRetriever,
	beaconViewRetriever basemeta.BeaconViewRetriever,
	beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalCustodianDepositV3ValidateSanityDataError, errors.New("tx custodian deposit v3 must be TxNormalType"))
	}

	// validate remote addresses
	isValid, err := ValidatePortalRemoteAddresses(custodianDeposit.RemoteAddresses, chainRetriever, beaconHeight)
	if !isValid || err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalCustodianDepositV3ValidateSanityDataError, err)
	}

	// validate deposit proof and receipt
	if len(custodianDeposit.ProofStrs) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalCustodianDepositV3ValidateSanityDataError, errors.New("ProofStrs should be not empty"))
	}
	if len(custodianDeposit.BlockHash.Bytes()) == 0 || bytes.Equal(custodianDeposit.BlockHash.Bytes(), eCommon.HexToHash("").Bytes()) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalCustodianDepositV3ValidateSanityDataError, errors.New("BlockHash should be not empty"))
	}

	return true, true, nil
}

func (custodianDeposit PortalCustodianDepositV3) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == basemeta.PortalCustodianDepositMetaV3
}

func (custodianDeposit PortalCustodianDepositV3) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	tokenIDKeys := make([]string, 0)
	for tokenID := range custodianDeposit.RemoteAddresses {
		tokenIDKeys = append(tokenIDKeys, tokenID)
	}
	sort.Strings(tokenIDKeys)
	for _, tokenID := range tokenIDKeys {
		record += custodianDeposit.RemoteAddresses[tokenID]
	}

	record += custodianDeposit.BlockHash.String()
	record += strconv.FormatUint(uint64(custodianDeposit.TxIndex), 10)
	for _, proofStr := range custodianDeposit.ProofStrs {
		record += proofStr
	}

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (custodianDeposit *PortalCustodianDepositV3) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalCustodianDepositActionV3{
		Meta:             *custodianDeposit,
		TxReqID:          *tx.Hash(),
		ShardID:          shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalCustodianDepositMetaV3), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalCustodianDepositV3) CalculateSize() uint64 {
	return basemeta.CalculateSize(custodianDeposit)
}

// ParseInfoFromLogMap parse info from log map
// Returns
// Custodian address string
// external TokenIDStr
// amount: for eth: unit is nanoETH
func ParseInfoFromLogMap(logMap map[string]interface{}) (string, string, uint64, error) {
	// the token might be ETH/ERC20
	ethereumAddr, ok := logMap["token"].(eCommon.Address)
	if !ok {
		Logger.log.Info("WARNING: could not parse eth token id from log map.")
		return "", "", 0, errors.New("could not parse eth token id from log map")
	}
	ethereumAddrStr := strings.ToLower(ethereumAddr.String())

	// custodian incognito address
	addressStr, ok := logMap["incognitoAddress"].(string)
	if !ok {
		Logger.log.Info("WARNING: could not parse incognito address from eth log map.")
		return "", "", 0, errors.New("could not parse incognito address from eth log map")
	}
	amt, ok := logMap["amount"].(*big.Int)
	if !ok {
		Logger.log.Info("WARNING: could not parse amount from eth log map.")
		return "", "", 0, errors.New("could not parse amount from eth log map")
	}
	amount := uint64(0)
	if common.EthAddrStr == ethereumAddrStr {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20
		amount = amt.Uint64()
	}

	return addressStr, ethereumAddrStr, amount, nil
}
