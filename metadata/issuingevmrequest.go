package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/incognitochain/incognito-chain/common/base58"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"math/big"
	"strconv"
	"strings"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/pkg/errors"
)

// IssuingEVMRequest represents an EVM shielding request. Users create transactions with this metadata after
// sending public tokens to the corresponding smart contract. There are two ways to use this metadata,
// depending on which data has been enclosed with the depositing transaction:
// 	- payment address: Receiver and Signature must be empty;
//	- using one-time depositing public key: Receiver must be an OTAReceiver, a signature is required.
type IssuingEVMRequest struct {
	// BlockHash is the hash of the block where the public depositing transaction resides in.
	BlockHash rCommon.Hash

	// TxIndex is the index of the public transaction in the BlockHash.
	TxIndex uint

	// ProofStrs is the generated proof for this shielding request.
	ProofStrs []string

	// IncTokenID is the Incognito tokenID of the shielding token.
	IncTokenID common.Hash

	// Signature is the signature for validating the authenticity of the request. This signature is different from a
	// MetadataBaseWithSignature type since it is signed with the tx privateKey.
	Signature []byte `json:"Signature,omitempty"`

	// Receiver is the recipient of this shielding request. It is an OTAReceiver if OTDepositPubKey is not empty.
	Receiver string `json:"Receiver,omitempty"`

	MetadataBase
}

type IssuingEVMReqAction struct {
	Meta       IssuingEVMRequest `json:"meta"`
	TxReqID    common.Hash       `json:"txReqId"`
	EVMReceipt *types.Receipt    `json:"ethReceipt"` // don't update the jsontag to make it compatible with the old shielding eth tx
}

type IssuingEVMAcceptedInst struct {
	ShardID         byte        `json:"shardId"`
	IssuingAmount   uint64      `json:"issuingAmount"`
	Receiver        string      `json:"receiverAddrStr"`
	OTDepositKey    []byte      `json:"OTDepositKey,omitempty"`
	IncTokenID      common.Hash `json:"incTokenId"`
	TxReqID         common.Hash `json:"txReqId"`
	UniqTx          []byte      `json:"uniqETHTx"` // don't update the jsontag to make it compatible with the old shielding eth tx
	ExternalTokenID []byte      `json:"externalTokenId"`
}

type GetEVMHeaderByHashRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetEVMHeaderByNumberRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetEVMBlockNumRes struct {
	rpccaller.RPCBaseRes
	Result string `json:"result"`
}

const (
	LegacyTxType = iota
	AccessListTxType
	DynamicFeeTxType
)

func ParseEVMIssuingInstContent(instContentStr string) (*IssuingEVMReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestDecodeInstructionError, err)
	}
	var issuingEVMReqAction IssuingEVMReqAction
	err = json.Unmarshal(contentBytes, &issuingEVMReqAction)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestUnmarshalJsonError, err)
	}
	return &issuingEVMReqAction, nil
}

func ParseEVMIssuingInstAcceptedContent(instAcceptedContentStr string) (*IssuingEVMAcceptedInst, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instAcceptedContentStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestDecodeInstructionError, err)
	}
	var issuingETHAcceptedInst IssuingEVMAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestUnmarshalJsonError, err)
	}
	return &issuingETHAcceptedInst, nil
}

func NewIssuingEVMRequest(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	incTokenID common.Hash,
	receiver string,
	signature []byte,
	metaType int,
) (*IssuingEVMRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingEVMReq := &IssuingEVMRequest{
		BlockHash:  blockHash,
		TxIndex:    txIndex,
		ProofStrs:  proofStrs,
		IncTokenID: incTokenID,
		Receiver:   receiver,
		Signature:  signature,
	}
	issuingEVMReq.MetadataBase = metadataBase
	return issuingEVMReq, nil
}

func NewIssuingEVMRequestFromMap(
	data map[string]interface{},
	metaType int,
) (*IssuingEVMRequest, error) {
	blockHash := rCommon.HexToHash(data["BlockHash"].(string))
	txIdx := uint(data["TxIndex"].(float64))
	proofsRaw := data["ProofStrs"].([]interface{})
	proofStrs := []string{}
	for _, item := range proofsRaw {
		proofStrs = append(proofStrs, item.(string))
	}

	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestNewIssuingEVMRequestFromMapError, fmt.Errorf("TokenID incorrect"))
	}

	var sig []byte
	tmpSig, ok := data["Signature"]
	if ok {
		sigStr, ok := tmpSig.(string)
		if ok {
			sig, _, err = base58.Base58Check{}.Decode(sigStr)
			if err != nil {
				return nil, NewMetadataTxError(IssuingEvmRequestNewIssuingEVMRequestFromMapError, fmt.Errorf("invalid base58-encoded signature"))
			}
		}
	}

	tmpReceiver, ok := data["Receiver"]
	var receiver string
	if ok {
		receiver, _ = tmpReceiver.(string)
	}

	if _, ok := data["MetadataType"]; ok {
		tmpMdType, ok := data["MetadataType"].(float64)
		if ok {
			metaType = int(tmpMdType)
		}
	}

	req, _ := NewIssuingEVMRequest(
		blockHash,
		txIdx,
		proofStrs,
		*incTokenID,
		receiver,
		sig,
		metaType,
	)
	return req, nil
}

func (iReq IssuingEVMRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingEVMRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if len(iReq.ProofStrs) == 0 {
		return false, false, NewMetadataTxError(IssuingEvmRequestValidateSanityDataError, fmt.Errorf("wrong request info's proof"))
	}

	if (iReq.Type == IssuingPRVBEP20RequestMeta || iReq.Type == IssuingPRVERC20RequestMeta) && iReq.IncTokenID.String() != common.PRVIDStr {
		return false, false, NewMetadataTxError(IssuingEvmRequestValidateSanityDataError, fmt.Errorf("invalid token id"))
	}

	evmReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return false, false, NewMetadataTxError(IssuingEvmRequestValidateSanityDataError, err)
	}
	if evmReceipt == nil {
		return false, false, NewMetadataTxError(IssuingEvmRequestValidateSanityDataError, fmt.Errorf("the evm proof's receipt could not be null"))
	}

	if iReq.Receiver != "" {
		otaReceiver := new(privacy.OTAReceiver)
		err = otaReceiver.FromString(iReq.Receiver)
		if err != nil {
			return false, false, NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, fmt.Errorf("invalid OTAReceiver"))
		}
		if !otaReceiver.IsValid() {
			return false, false, NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, fmt.Errorf("invalid OTAReceiver"))
		}
	}

	if iReq.Signature != nil {
		schnorrSig := new(schnorr.SchnSignature)
		err = schnorrSig.SetBytes(iReq.Signature)
		if err != nil {
			return false, false, NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, fmt.Errorf("invalid signature %v", iReq.Signature))
		}
	}

	return true, true, nil
}

func (iReq IssuingEVMRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingETHRequestMeta && iReq.Type != IssuingBSCRequestMeta &&
		iReq.Type != IssuingPRVERC20RequestMeta && iReq.Type != IssuingPRVBEP20RequestMeta &&
		iReq.Type != IssuingPLGRequestMeta && iReq.Type != IssuingFantomRequestMeta {
		return false
	}
	return true
}

func (iReq IssuingEVMRequest) Hash() *common.Hash {
	record := iReq.BlockHash.String()
	record += string(iReq.TxIndex)
	proofStrs := iReq.ProofStrs
	for _, proofStr := range proofStrs {
		record += proofStr
	}
	record += iReq.MetadataBase.Hash().String()
	record += iReq.IncTokenID.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *IssuingEVMRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	evmReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingEvmRequestBuildReqActionsError, err)
	}
	if evmReceipt == nil {
		return [][]string{}, NewMetadataTxError(IssuingEvmRequestBuildReqActionsError, errors.Errorf("The evm proof's receipt could not be null."))
	}
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":       *iReq,
		"txReqId":    txReqID,
		"ethReceipt": *evmReceipt,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingEvmRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(iReq.Type), actionContentBase64Str}

	return [][]string{action}, nil
}

func (iReq *IssuingEVMRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}

func (iReq *IssuingEVMRequest) verifyProofAndParseReceipt() (*types.Receipt, error) {
	// get hosts, minEVMConfirmationBlocks, networkPrefix depend iReq.Type
	hosts, networkPrefix, minEVMConfirmationBlocks, checkEVMHardFork, err := GetEVMInfoByMetadataType(iReq.Type)
	if err != nil {
		Logger.log.Errorf("Can not get EVM info - Error: %+v", err)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	return VerifyProofAndParseEVMReceipt(iReq.BlockHash, iReq.TxIndex, iReq.ProofStrs, hosts, minEVMConfirmationBlocks, networkPrefix, checkEVMHardFork)
}

func (iReq *IssuingEVMRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration

	if iReq.Receiver != "" {
		otaReceiver := privacy.OTAReceiver{}
		_ = otaReceiver.FromString(iReq.Receiver)
		otaDecl := metadataCommon.OTADeclaration{
			PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
		}
		if iReq.Type == IssuingPRVERC20RequestMeta || iReq.Type == IssuingPRVBEP20RequestMeta {
			otaDecl.TokenID = common.PRVCoinID
		}
		result = append(result, otaDecl)
	}

	return result
}

func ParseEVMLogData(data []byte) (map[string]interface{}, error) {
	abiIns, err := abi.JSON(strings.NewReader(common.AbiJson))
	if err != nil {
		return nil, err
	}
	dataMap := map[string]interface{}{}
	if err = abiIns.UnpackIntoMap(dataMap, "Deposit", data); err != nil {
		return nil, NewMetadataTxError(UnexpectedError, err)
	}
	return dataMap, nil
}

func GetEVMHeader(
	evmBlockHash rCommon.Hash,
	protocol string,
	host string,
	port string,
) (*types.Header, error) {
	rpcClient := rpccaller.NewRPCClient()
	getEVMHeaderByHashParams := []interface{}{evmBlockHash, false}
	var getEVMHeaderByHashRes GetEVMHeaderByHashRes
	err := rpcClient.RPCCall(
		protocol,
		host,
		port,
		"eth_getBlockByHash",
		getEVMHeaderByHashParams,
		&getEVMHeaderByHashRes,
	)
	if err != nil {
		return nil, err
	}
	if getEVMHeaderByHashRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: %s", getEVMHeaderByHashRes.RPCError.Message)
		return nil, errors.New(fmt.Sprintf("An error occured during calling eth_getBlockByHash: %s", getEVMHeaderByHashRes.RPCError.Message))
	}

	if getEVMHeaderByHashRes.Result == nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: result is nil")
		return nil, errors.New(fmt.Sprintf("An error occured during calling eth_getBlockByHash: result is nil"))
	}

	evmHeaderByHash := getEVMHeaderByHashRes.Result
	headerNum := evmHeaderByHash.Number

	getEVMHeaderByNumberParams := []interface{}{fmt.Sprintf("0x%x", headerNum), false}
	var getEVMHeaderByNumberRes GetEVMHeaderByNumberRes
	err = rpcClient.RPCCall(
		protocol,
		host,
		port,
		"eth_getBlockByNumber",
		getEVMHeaderByNumberParams,
		&getEVMHeaderByNumberRes,
	)
	if err != nil {
		return nil, err
	}
	if getEVMHeaderByNumberRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: %s", getEVMHeaderByNumberRes.RPCError.Message)
		return nil, errors.New(fmt.Sprintf("An error occured during calling eth_getBlockByNumber: %s", getEVMHeaderByNumberRes.RPCError.Message))
	}

	if getEVMHeaderByNumberRes.Result == nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: result is nil")
		return nil, errors.New(fmt.Sprintf("An error occured during calling eth_getBlockByNumber: result is nil"))
	}

	evmHeaderByNum := getEVMHeaderByNumberRes.Result
	if evmHeaderByNum.Hash().String() != evmHeaderByHash.Hash().String() {
		return nil, errors.New(fmt.Sprintf("The requested eth BlockHash is being on fork branch, rejected!"))
	}
	return evmHeaderByHash, nil
}

// GetMostRecentEVMBlockHeight get most recent block height on Ethereum/BSC
func GetMostRecentEVMBlockHeight(protocol string, host string, port string) (*big.Int, error) {
	rpcClient := rpccaller.NewRPCClient()
	params := []interface{}{}
	var getEVMBlockNumRes GetEVMBlockNumRes
	err := rpcClient.RPCCall(
		protocol,
		host,
		port,
		"eth_blockNumber",
		params,
		&getEVMBlockNumRes,
	)
	if err != nil {
		return nil, err
	}
	if getEVMBlockNumRes.RPCError != nil {
		return nil, errors.New(fmt.Sprintf("an error occured during calling eth_blockNumber: %s", getEVMBlockNumRes.RPCError.Message))
	}

	if len(getEVMBlockNumRes.Result) < 2 {
		return nil, errors.New(fmt.Sprintf("invalid block height number eth_blockNumber: %s", getEVMBlockNumRes.Result))
	}

	blockNumber := new(big.Int)
	_, ok := blockNumber.SetString(getEVMBlockNumRes.Result[2:], 16)
	if !ok {
		return nil, errors.New("Cannot convert blockNumber into integer")
	}
	return blockNumber, nil
}

func PickAndParseLogMapFromReceipt(constructedReceipt *types.Receipt, contractAddressStr string) (map[string]interface{}, error) {
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		Logger.log.Warn("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(rCommon.HexToAddress(contractAddressStr).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		return nil, nil
	}
	return ParseEVMLogData(logData)
}
