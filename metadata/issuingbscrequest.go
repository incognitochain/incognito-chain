package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/pkg/errors"
)

type IssuingBSCRequest struct {
	BlockHash  rCommon.Hash
	TxIndex    uint
	ProofStrs  []string
	IncTokenID common.Hash
	MetadataBase
}

type IssuingBSCReqAction struct {
	Meta       IssuingBSCRequest `json:"meta"`
	TxReqID    common.Hash       `json:"txReqId"`
	BSCReceipt *types.Receipt    `json:"bscReceipt"`
}

type IssuingBSCAcceptedInst struct {
	ShardID         byte        `json:"shardId"`
	IssuingAmount   uint64      `json:"issuingAmount"`
	ReceiverAddrStr string      `json:"receiverAddrStr"`
	IncTokenID      common.Hash `json:"incTokenId"`
	TxReqID         common.Hash `json:"txReqId"`
	UniqBSCTx       []byte      `json:"uniqBSCTx"`
	ExternalTokenID []byte      `json:"externalTokenId"`
}

func ParseBSCIssuingInstContent(instContentStr string) (*IssuingBSCReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingBSCRequestDecodeInstructionError, err)
	}
	var issuingBSCReqAction IssuingBSCReqAction
	err = json.Unmarshal(contentBytes, &issuingBSCReqAction)
	if err != nil {
		return nil, NewMetadataTxError(IssuingBSCRequestUnmarshalJsonError, err)
	}
	return &issuingBSCReqAction, nil
}

func NewIssuingBSCRequest(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	incTokenID common.Hash,
	metaType int,
) (*IssuingBSCRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingBSCReq := &IssuingBSCRequest{
		BlockHash:  blockHash,
		TxIndex:    txIndex,
		ProofStrs:  proofStrs,
		IncTokenID: incTokenID,
	}
	issuingBSCReq.MetadataBase = metadataBase
	return issuingBSCReq, nil
}

func NewIssuingBSCRequestFromMap(
	data map[string]interface{},
) (*IssuingBSCRequest, error) {
	blockHash := rCommon.HexToHash(data["BlockHash"].(string))
	txIdx := uint(data["TxIndex"].(float64))
	proofsRaw := data["ProofStrs"].([]interface{})
	proofStrs := []string{}
	for _, item := range proofsRaw {
		proofStrs = append(proofStrs, item.(string))
	}

	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, NewMetadataTxError(IssuingBscRequestNewIssuingBSCRequestFromMapError, errors.Errorf("TokenID incorrect"))
	}

	req, _ := NewIssuingBSCRequest(
		blockHash,
		txIdx,
		proofStrs,
		*incTokenID,
		IssuingBSCRequestMeta,
	)
	return req, nil
}

func (iReq IssuingBSCRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	bscReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return false, NewMetadataTxError(IssuingBSCRequestValidateTxWithBlockChainError, err)
	}
	if bscReceipt == nil {
		return false, errors.Errorf("The bsc proof's receipt could not be null.")
	}
	return true, nil
}

func (iReq IssuingBSCRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if len(iReq.ProofStrs) == 0 {
		return false, false, NewMetadataTxError(IssuingBSCRequestValidateSanityDataError, errors.New("Wrong request info's proof"))
	}
	return true, true, nil
}

func (iReq IssuingBSCRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingBSCRequestMeta {
		return false
	}
	return true
}

func (iReq IssuingBSCRequest) Hash() *common.Hash {
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

func (iReq *IssuingBSCRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	bscReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingBSCRequestBuildReqActionsError, err)
	}
	if bscReceipt == nil {
		return [][]string{}, NewMetadataTxError(IssuingBSCRequestBuildReqActionsError, errors.Errorf("The bsc proof's receipt could not be null."))
	}
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":       *iReq,
		"txReqId":    txReqID,
		"bscReceipt": *bscReceipt,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingBSCRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(IssuingBSCRequestMeta), actionContentBase64Str}

	return [][]string{action}, nil
}

func (iReq *IssuingBSCRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}

func (iReq *IssuingBSCRequest) verifyProofAndParseReceipt() (*types.Receipt, error) {
	bscParam := config.Param().BSCParam
	bscParam.GetFromEnv()
	bscHeader, err := GetETHHeader(iReq.BlockHash, bscParam.Protocol, bscParam.Host, bscParam.Port)
	if err != nil {
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, err)
	}
	if bscHeader == nil {
		Logger.log.Info("WARNING: Could not find out the BSC block header with the hash: ", iReq.BlockHash)
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, errors.Errorf("WARNING: Could not find out the BSC block header with the hash: %s", iReq.BlockHash.String()))
	}

	mostRecentBlkNum, err := GetMostRecentETHBlockHeight(bscParam.Protocol, bscParam.Host, bscParam.Port)
	if err != nil {
		Logger.log.Info("WARNING: Could not find the most recent block height on Binance smart chain")
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, err)
	}

	if mostRecentBlkNum.Cmp(big.NewInt(0).Add(bscHeader.Number, big.NewInt(BSCConfirmationBlocks))) == -1 {
		errMsg := fmt.Sprintf("WARNING: It needs 15 confirmation blocks for the process, the requested block (%s) but the latest block (%s)", bscHeader.Number.String(), mostRecentBlkNum.String())
		Logger.log.Info(errMsg)
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, errors.New(errMsg))
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
	val, _, err := trie.VerifyProof(bscHeader.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		fmt.Printf("WARNING: BSC issuance proof verification failed: %v", err)
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, err)
	}
	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, err)
	}

	if constructedReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, errors.New("The constructedReceipt's status is not success"))
	}

	logMap, err := PickAndParseLogMapFromReceipt(constructedReceipt, config.Param().BscContractAddressStr)
	if err != nil || logMap == nil {
		Logger.log.Warn("WARNING: an error occured while parsing log map from receipt: ", err)
		return nil, NewMetadataTxError(IssuingBscRequestVerifyProofAndParseReceipt, errors.New("The constructedReceipt's status is not success"))
	}

	return constructedReceipt, nil
}
