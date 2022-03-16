package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/pkg/errors"
)

type IssuingEVMRequest struct {
	BlockHash  rCommon.Hash
	TxIndex    uint
	ProofStrs  []string
	IncTokenID common.Hash
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
	ReceiverAddrStr string      `json:"receiverAddrStr"`
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
	}
	issuingEVMReq.MetadataBase = metadataBase
	return issuingEVMReq, nil
}

func NewIssuingEVMRequestFromMap(
	data map[string]interface{},
	metatype int,
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
		return nil, NewMetadataTxError(IssuingEvmRequestNewIssuingEVMRequestFromMapError, errors.Errorf("TokenID incorrect"))
	}

	req, _ := NewIssuingEVMRequest(
		blockHash,
		txIdx,
		proofStrs,
		*incTokenID,
		metatype,
	)
	return req, nil
}

func (iReq IssuingEVMRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingEVMRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if len(iReq.ProofStrs) == 0 {
		return false, false, NewMetadataTxError(IssuingEvmRequestValidateSanityDataError, errors.New("Wrong request info's proof"))
	}

	if (iReq.Type == IssuingPRVBEP20RequestMeta || iReq.Type == IssuingPRVERC20RequestMeta) && iReq.IncTokenID.String() != common.PRVIDStr {
		return false, false, NewMetadataTxError(IssuingEvmRequestValidateSanityDataError, errors.New("Invalid token id"))
	}

	return true, true, nil
}

func (iReq IssuingEVMRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingETHRequestMeta && iReq.Type != IssuingBSCRequestMeta &&
		iReq.Type != IssuingPRVERC20RequestMeta && iReq.Type != IssuingPRVBEP20RequestMeta &&
		iReq.Type != IssuingPLGRequestMeta && iReq.Type != IssuingFantomRequestMeta {
		return false
	}
	evmReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		Logger.log.Error(NewMetadataTxError(IssuingEvmRequestValidateTxWithBlockChainError, err))
		return false
	}
	if evmReceipt == nil {
		Logger.log.Error(errors.Errorf("The evm proof's receipt could not be null."))
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
	var protocol, host, port string
	if iReq.Type == IssuingBSCRequestMeta || iReq.Type == IssuingPRVBEP20RequestMeta {
		evmParam := config.Param().BSCParam
		evmParam.GetFromEnv()
		host = evmParam.Host
	} else if iReq.Type == IssuingETHRequestMeta || iReq.Type == IssuingPRVERC20RequestMeta {
		evmParam := config.Config().GethParam
		evmParam.GetFromEnv()
		protocol = evmParam.Protocol
		host = evmParam.Host
		port = evmParam.Port
	} else if iReq.Type == IssuingPLGRequestMeta {
		evmParam := config.Param().PLGParam
		evmParam.GetFromEnv()
		host = evmParam.Host
	} else if iReq.Type == IssuingFantomRequestMeta {
		evmParam := config.Param().FTMParam
		evmParam.GetFromEnv()
		host = evmParam.Host
	} else {
		return nil, errors.New("[verifyProofAndParseReceipt] invalid metatype")
	}
	evmHeader, err := GetEVMHeader(iReq.BlockHash, protocol, host, port)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}
	if evmHeader == nil {
		Logger.log.Warn("WARNING: Could not find out the EVM block header with the hash: ", iReq.BlockHash)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, errors.Errorf("WARNING: Could not find out the EVM block header with the hash: %s", iReq.BlockHash.String()))
	}

	mostRecentBlkNum, err := GetMostRecentEVMBlockHeight(protocol, host, port)
	if err != nil {
		Logger.log.Warn("WARNING: Could not find the most recent block height on Ethereum")
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	minEVMConfirmationBlocks := EVMConfirmationBlocks
	if iReq.Type == IssuingPLGRequestMeta {
		minEVMConfirmationBlocks = PLGConfirmationBlocks
	} else if iReq.Type == IssuingFantomRequestMeta {
		minEVMConfirmationBlocks = FantomConfirmationBlocks
	}
	if mostRecentBlkNum.Cmp(big.NewInt(0).Add(evmHeader.Number, big.NewInt(int64(minEVMConfirmationBlocks)))) == -1 {
		errMsg := fmt.Sprintf("WARNING: It needs %v confirmation blocks for the process, "+
			"the requested block (%s) but the latest block (%s)", minEVMConfirmationBlocks,
			evmHeader.Number.String(), mostRecentBlkNum.String())
		Logger.log.Warn(errMsg)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, errors.New(errMsg))
	}

	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, iReq.TxIndex)

	nodeList := new(light.NodeList)
	for _, proofStr := range iReq.ProofStrs {
		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
		if err != nil {
			return nil, err
		}
		nodeList.Put([]byte{}, proofBytes)
	}
	proof := nodeList.NodeSet()
	val, _, err := trie.VerifyProof(evmHeader.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		errMsg := fmt.Sprintf("WARNING: EVM issuance proof verification failed: %v", err)
		Logger.log.Warn(errMsg)
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	if iReq.Type == IssuingETHRequestMeta || iReq.Type == IssuingPRVERC20RequestMeta ||
		iReq.Type == IssuingPLGRequestMeta || iReq.Type == IssuingFantomRequestMeta {
		if len(val) == 0 {
			return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, errors.New("the encoded receipt is empty"))
		}

		// hardfork london with new transaction type => 0x02 || RLP([...SenderPayload, ...SenderSignature, ...GasPayerPayload, ...GasPayerSignature])
		if val[0] == AccessListTxType || val[0] == DynamicFeeTxType {
			val = val[1:]
		}
	}

	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	if constructedReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, NewMetadataTxError(IssuingEvmRequestVerifyProofAndParseReceipt, errors.New("The constructedReceipt's status is not success"))
	}

	return constructedReceipt, nil
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
