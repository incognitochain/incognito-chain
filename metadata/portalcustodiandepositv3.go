package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"math/big"
	"sort"
	"strconv"
)

// PortalCustodianDepositV3 - portal custodian deposit collaterals on SC (ETH/ERC20)
// metadata - custodian deposit - create normal tx with this metadata
type PortalCustodianDepositV3 struct {
	MetadataBase
	IncAddressStr    string
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

	//IncAddressStr    string
	//DepositAmount    uint64
	//ExternalTokenID  string // collateral TokenID
	//UniqExternalTxID []byte
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
	custodianIncAddr string,
	remoteAddrs map[string]string,
	blockHash eCommon.Hash,
	txIndex uint,
	proofStrs []string) (*PortalCustodianDepositV3, error) {
	custodianDepositMetaV3 := &PortalCustodianDepositV3{
		MetadataBase: MetadataBase{
			Type: metaType,
		},
		RemoteAddresses: remoteAddrs,
		BlockHash:       blockHash,
		TxIndex:         txIndex,
		ProofStrs:       proofStrs,
		IncAddressStr: custodianIncAddr,
	}
	return custodianDepositMetaV3, nil
}

func NewPortalCustodianDepositV3FromMap(
	data map[string]interface{},
) (*PortalCustodianDepositV3, error) {
	incognitoAddress, ok := data["IncognitoAddress"].(string)
	if !ok {
		return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata IncognitoAddress is invalid"))
	}
	remoteAddressesMap, ok := data["RemoteAddresses"].(map[string]interface{})
	if !ok {
		return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata RemoteAddresses param is invalid"))
	}
	if len(remoteAddressesMap) < 1 {
		return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata RemoteAddresses must be at least one"))
	}
	remoteAddresses := make(map[string]string, 0)
	tokenIDKeys := make([]string, 0)
	for pTokenID, remoteAddress := range remoteAddressesMap {
		if !IsPortalToken(pTokenID) {
			return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata public token is not supported currently"))
		}
		_, ok := remoteAddress.(string)
		if !ok {
			return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata RemoteAddresses is invalid"))
		}
		tokenIDKeys = append(tokenIDKeys, pTokenID)
	}
	sort.Strings(tokenIDKeys)
	for _, pTokenID := range tokenIDKeys {
		remoteAddresses[pTokenID] = remoteAddressesMap[pTokenID].(string)
	}

	blockHashStr, ok := data["BlockHash"].(string)
	if !ok {
		return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata BlockHash should be a string"))
	}
	blockHash := eCommon.HexToHash(blockHashStr)

	txIndexFloat, ok := data["TxIndex"].(float64)
	if !ok {
		return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata TxIndex should be a number"))
	}
	txIdx := uint(txIndexFloat)

	proofsRaw, ok := data["ProofStrs"].([]interface{})
	if !ok {
		return nil, NewMetadataTxError(NewPortalCustodianDepositV3MetaFromMapError, errors.New("metadata ProofStrs should be an array of string"))
	}
	proofStrs := []string{}
	for _, item := range proofsRaw {
		proofStrs = append(proofStrs, item.(string))
	}

	meta, _ := NewPortalCustodianDepositV3(
		PortalCustodianDepositMetaV3,
		incognitoAddress,
		remoteAddresses,
		blockHash,
		txIdx,
		proofStrs,
	)
	return meta, nil
}

func (custodianDeposit PortalCustodianDepositV3) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalCustodianDepositV3) ValidateSanityData(
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("tx custodian deposit v3 must be TxNormalType"))
	}

	// validate remote addresses
	if len(custodianDeposit.RemoteAddresses) == 0 {
		return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("remote addresses should be at least one"))
	}
	for tokenID, remoteAddr := range custodianDeposit.RemoteAddresses {
		if !IsPortalToken(tokenID) {
			return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("TokenID in remote address is invalid"))
		}
		if len(remoteAddr) == 0 {
			return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("Remote address is invalid"))
		}
		chainID := GetChainIDByTokenID(tokenID, chainRetriever)
		if !IsValidRemoteAddress(chainRetriever, remoteAddr, tokenID, chainID) {
			return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError,
				fmt.Errorf("Remote address %v is not a valid address of tokenID %v", remoteAddr, tokenID))
		}
	}

	// validate deposit proof and receipt
	if len(custodianDeposit.ProofStrs) == 0 {
		return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("ProofStrs should be not empty"))
	}
	//ethReceipt, err := custodianDeposit.verifyProofAndParseReceipt()
	//if err != nil {
	//	return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, err)
	//}
	//if ethReceipt == nil {
	//	return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("The eth proof's receipt could not be null."))
	//}
	//
	//logMap, err := PickAndParseLogMapFromReceiptByContractAddr(ethReceipt, chainRetriever.GetPortalETHContractAddrStr(), "Deposit")
	//if err != nil {
	//	Logger.log.Info("WARNING: an error occured while parsing log map from receipt: ", err)
	//	return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, err)
	//}
	//if logMap == nil {
	//	Logger.log.Info("WARNING: could not find log map out from receipt")
	//	return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("log map is nil"))
	//}
	//
	//custodianIncAddr, externalTokenIDStr, amount, err := ParseInfoFromLogMap(logMap)
	//if err != nil {
	//	Logger.log.Info("WARNING: an error occured while parsing info from log map: ", err)
	//	return false, false, NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, err)
	//}

	// check sender's address
	keyWallet, err := wallet.Base58CheckDeserialize(custodianDeposit.IncAddressStr)
	if err != nil {
		return false, false,
			NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("could not decode CustodianIncAddressStr"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false,
			NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("CustodianIncAddressStr with empty public key"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false,
			NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("custodian incognito address is not signer tx"))
	}

	//// check externalTokenID should be one of supported collateral tokenIDs
	//if ok, err := common.SliceExists(chainRetriever.GetSupportedCollateralTokenIDs(beaconHeight), externalTokenIDStr); !ok || err != nil {
	//	return false, false,
	//		NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("external collateral tokenID is not supported on portal"))
	//}
	//
	//// check deposit amount
	//if amount <= 0 {
	//	return false, false,
	//		NewMetadataTxError(PortalCustodianDepositV3ValidateSanityDataError, errors.New("amount should be greater than zero"))
	//}

	return true, true, nil
}

func (custodianDeposit PortalCustodianDepositV3) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == PortalCustodianDepositMetaV3
}

func (custodianDeposit PortalCustodianDepositV3) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	record += custodianDeposit.IncAddressStr
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

func (custodianDeposit *PortalCustodianDepositV3) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte) ([][]string, error) {
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
	action := []string{strconv.Itoa(PortalCustodianDepositMetaV3), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalCustodianDepositV3) CalculateSize() uint64 {
	return calculateSize(custodianDeposit)
}

// ParseInfoFromLogMap parse info from log map
// Returns
// Custodian address string
// external TokenIDStr
// amount: for eth: unit is nanoETH
// todo: need to be updated to smart contract
func ParseInfoFromLogMap(logMap map[string]interface{}) (string, string, uint64, error) {
	// the token might be ETH/ERC20
	ethereumAddr, ok := logMap["token"].(eCommon.Address)
	if !ok {
		Logger.log.Info("WARNING: could not parse eth token id from log map.")
		return "", "", 0, errors.New("could not parse eth token id from log map")
	}
	ethereumAddrStr := ethereumAddr.String()

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
