package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
)

type IssuingETHRequest struct {
	BlockHash  rCommon.Hash
	TxIndex    uint
	ProofStrs  []string
	IncTokenID common.Hash
	MetadataBase
}

type IssuingETHReqAction struct {
	Meta       IssuingETHRequest `json:"meta"`
	TxReqID    common.Hash       `json:"txReqId"`
	ETHReceipt *types.Receipt    `json:"ethReceipt"`
}

type IssuingETHAcceptedInst struct {
	ShardID         byte        `json:"shardId"`
	IssuingAmount   uint64      `json:"issuingAmount"`
	ReceiverAddrStr string      `json:"receiverAddrStr"`
	IncTokenID      common.Hash `json:"incTokenId"`
	TxReqID         common.Hash `json:"txReqId"`
	UniqETHTx       []byte      `json:"uniqETHTx"`
	ExternalTokenID []byte      `json:"externalTokenId"`
}

type GetBlockByNumberRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

func ParseETHIssuingInstContent(instContentStr string) (*IssuingETHReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEthRequestDecodeInstructionError, err)
	}
	var issuingETHReqAction IssuingETHReqAction
	err = json.Unmarshal(contentBytes, &issuingETHReqAction)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEthRequestUnmarshalJsonError, err)
	}
	return &issuingETHReqAction, nil
}

func NewIssuingETHRequest(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	incTokenID common.Hash,
	metaType int,
) (*IssuingETHRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingETHReq := &IssuingETHRequest{
		BlockHash:  blockHash,
		TxIndex:    txIndex,
		ProofStrs:  proofStrs,
		IncTokenID: incTokenID,
	}
	issuingETHReq.MetadataBase = metadataBase
	return issuingETHReq, nil
}

func NewIssuingETHRequestFromMap(
	data map[string]interface{},
) (*IssuingETHRequest, error) {
	blockHash := rCommon.HexToHash(data["BlockHash"].(string))
	txIdx := uint(data["TxIndex"].(float64))
	proofsRaw := data["ProofStrs"].([]interface{})
	proofStrs := []string{}
	for _, item := range proofsRaw {
		proofStrs = append(proofStrs, item.(string))
	}

	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, NewMetadataTxError(IssuingEthRequestNewIssuingETHRequestFromMapEror, errors.Errorf("TokenID incorrect"))
	}

	req, _ := NewIssuingETHRequest(
		blockHash,
		txIdx,
		proofStrs,
		*incTokenID,
		IssuingETHRequestMeta,
	)
	return req, nil
}

func (iReq IssuingETHRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	ethReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return false, NewMetadataTxError(IssuingEthRequestValidateTxWithBlockChainError, err)
	}
	if ethReceipt == nil {
		return false, errors.Errorf("The eth proof's receipt could not be null.")
	}
	return true, nil
}

func (iReq IssuingETHRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	if len(iReq.ProofStrs) == 0 {
		return false, false, NewMetadataTxError(IssuingEthRequestValidateSanityDataError, errors.New("Wrong request info's proof"))
	}
	return true, true, nil
}

func (iReq IssuingETHRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingETHRequestMeta {
		return false
	}
	return true
}

func (iReq IssuingETHRequest) Hash() *common.Hash {
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

func (iReq *IssuingETHRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	ethReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingEthRequestBuildReqActionsError, err)
	}
	if ethReceipt == nil {
		return [][]string{}, NewMetadataTxError(IssuingEthRequestBuildReqActionsError, errors.Errorf("The eth proof's receipt could not be null."))
	}
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":       *iReq,
		"txReqId":    txReqID,
		"ethReceipt": *ethReceipt,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingEthRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(IssuingETHRequestMeta), actionContentBase64Str}

	Logger.log.Debug("hahaha txreqid: ", txReqID)
	err = bcr.GetDatabase().TrackBridgeReqWithStatus(txReqID, byte(common.BridgeRequestProcessingStatus), nil)
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingEthRequestBuildReqActionsError, err)
	}
	return [][]string{action}, nil
}

func (iReq *IssuingETHRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}

func (iReq *IssuingETHRequest) verifyProofAndParseReceipt() (*types.Receipt, error) {
	ethHeader, err := GetETHHeader(iReq.BlockHash)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEthRequestVerifyProofAndParseReceipt, err)
	}
	if ethHeader == nil {
		Logger.log.Info("WARNING: Could not find out the ETH block header with the hash: ", iReq.BlockHash)
		return nil, NewMetadataTxError(IssuingEthRequestVerifyProofAndParseReceipt, errors.Errorf("WARNING: Could not find out the ETH block header with the hash: %s", iReq.BlockHash.String()))
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
	val, _, err := trie.VerifyProof(ethHeader.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		fmt.Printf("WARNING: ETH issuance proof verification failed: %v", err)
		return nil, NewMetadataTxError(IssuingEthRequestVerifyProofAndParseReceipt, err)
	}
	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, NewMetadataTxError(IssuingEthRequestVerifyProofAndParseReceipt, err)
	}
	return constructedReceipt, nil
}

func ParseETHLogData(data []byte) (map[string]interface{}, error) {
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

func IsETHTxHashUsedInBlock(uniqETHTx []byte, uniqETHTxsUsed [][]byte) bool {
	for _, item := range uniqETHTxsUsed {
		if bytes.Equal(uniqETHTx, item) {
			return true
		}
	}
	return false
}

func GetETHHeader(
	ethBlockHash rCommon.Hash,
) (*types.Header, error) {
	rpcClient := rpccaller.NewRPCClient()
	params := []interface{}{ethBlockHash, false}
	var getBlockByNumberRes GetBlockByNumberRes
	err := rpcClient.RPCCall(
		EthereumLightNodeProtocol,
		EthereumLightNodeHost,
		EthereumLightNodePort,
		"eth_getBlockByHash",
		params,
		&getBlockByNumberRes,
	)
	if err != nil {
		return nil, err
	}
	if getBlockByNumberRes.RPCError != nil {
		Logger.log.Debugf("WARNING: an error occured during calling eth_getBlockByHash: %s", getBlockByNumberRes.RPCError.Message)
		return nil, nil
	}
	return getBlockByNumberRes.Result, nil
}

func PickAndParseLogMapFromReceipt(constructedReceipt *types.Receipt, ethContractAddressStr string) (map[string]interface{}, error) {
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		Logger.log.Debug("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(rCommon.HexToAddress(ethContractAddressStr).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		return nil, nil
	}
	return ParseETHLogData(logData)
}
