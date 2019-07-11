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
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/rpccaller"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type IssuingETHResponse struct {
	MetadataBase
	RequestedTxID   common.Hash
	UniqETHTx       []byte
	ExternalTokenID []byte
}

type IssuingETHReqAction struct {
	BridgeShardID byte              `json:"bridgeShardID"`
	Meta          IssuingETHRequest `json:"meta"`
	TxReqID       common.Hash       `json:"txReqId"`
}

type AccumulatedValues struct {
	UniqETHTxsUsed   [][]byte
	DBridgeTokenPair map[string][]byte
	CBridgeTokens    []*common.Hash
}

func NewIssuingETHResponse(
	requestedTxID common.Hash,
	uniqETHTx []byte,
	externalTokenID []byte,
	metaType int,
) *IssuingETHResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingETHResponse{
		RequestedTxID:   requestedTxID,
		UniqETHTx:       uniqETHTx,
		ExternalTokenID: externalTokenID,
		MetadataBase:    metadataBase,
	}
}

func (iRes *IssuingETHResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (iRes *IssuingETHResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes *IssuingETHResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes *IssuingETHResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes *IssuingETHResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += string(iRes.UniqETHTx)
	record += string(iRes.ExternalTokenID)
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingETHResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func IsETHHashUsedInBlock(uniqETHTx []byte, uniqETHTxsUsed [][]byte) bool {
	for _, item := range uniqETHTxsUsed {
		if bytes.Equal(uniqETHTx, item) {
			return true
		}
	}
	return false
}

func (ac *AccumulatedValues) CanProcessTokenPair(
	externalTokenID []byte,
	incTokenID common.Hash,
) (bool, error) {
	incTokenIDStr := incTokenID.String()
	for _, tokenID := range ac.CBridgeTokens {
		if bytes.Equal(tokenID[:], incTokenID[:]) {
			return false, nil
		}
	}
	bridgeTokenPair := ac.DBridgeTokenPair
	if existedExtTokenID, found := bridgeTokenPair[incTokenIDStr]; found {
		if bytes.Equal(existedExtTokenID, externalTokenID) {
			return true, nil
		}
		return false, nil
	}
	for _, existedExtTokenID := range bridgeTokenPair {
		if !bytes.Equal(existedExtTokenID, externalTokenID) {
			continue
		}
		return false, nil
	}
	return true, nil
}

func ParseETHLogData(data []byte) (map[string]interface{}, error) {
	abiIns, err := abi.JSON(strings.NewReader(common.ABIJSON))
	if err != nil {
		fmt.Println("haha err 1: ", err)
		return nil, err
	}
	dataMap := map[string]interface{}{}
	if err = abiIns.UnpackIntoMap(dataMap, "Deposit", data); err != nil {
		fmt.Println("haha err 2: ", err)
		return nil, err
	}
	fmt.Println("haha ngon lanh het")
	return dataMap, nil
}

func ParseInstContent(instContentStr string) (*IssuingETHReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, err
	}
	var issuingETHReqAction IssuingETHReqAction
	err = json.Unmarshal(contentBytes, &issuingETHReqAction)
	if err != nil {
		return nil, err
	}
	return &issuingETHReqAction, nil
}

type GetBlockByNumberRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

func GetETHHeader(
	bcr BlockchainRetriever,
	ethBlockHash rCommon.Hash,
) (*types.Header, error) {
	rpcClient := bcr.GetRPCClient()
	params := []interface{}{ethBlockHash, false}
	var getBlockByNumberRes GetBlockByNumberRes
	err := rpcClient.RPCCall(
		common.ETHERERUM_LIGHT_NODE_PROTOCOL,
		common.ETHERERUM_LIGHT_NODE_HOST,
		common.ETHERERUM_LIGHT_NODE_PORT,
		"eth_getBlockByHash",
		params,
		&getBlockByNumberRes,
	)
	if err != nil {
		return nil, err
	}
	if getBlockByNumberRes.RPCError != nil {
		fmt.Printf("WARNING: an error occured during calling eth_getBlockByHash: %s", getBlockByNumberRes.RPCError.Message)
		return nil, nil
	}
	return getBlockByNumberRes.Result, nil
}

func VerifyProofAndParseReceipt(
	issuingETHReqAction *IssuingETHReqAction,
	bcr BlockchainRetriever,
) (*types.Receipt, error) {
	md := issuingETHReqAction.Meta
	ethHeader, err := GetETHHeader(bcr, md.BlockHash)
	if err != nil {
		return nil, err
	}
	if ethHeader == nil {
		fmt.Println("WARNING: Could not find out the ETH block header with the hash: ", md.BlockHash)
		return nil, nil
	}
	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, md.TxIndex)

	nodeList := new(light.NodeList)
	for _, proofStr := range md.ProofStrs {
		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
		if err != nil {
			return nil, err
		}
		nodeList.Put([]byte{}, proofBytes)
	}
	proof := nodeList.NodeSet()
	val, _, err := trie.VerifyProof(ethHeader.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		fmt.Printf("ETH issuance proof verification failed: %v", err)
		return nil, err
	}
	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, err
	}
	return constructedReceipt, nil
}

func PickNParseLogMapFromReceipt(constructedReceipt *types.Receipt) (map[string]interface{}, error) {
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		fmt.Println("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(rCommon.HexToAddress(common.ETH_CONTRACT_ADDR_STR).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		return nil, nil
	}
	return ParseETHLogData(logData)
}

func (iRes *IssuingETHResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ac *AccumulatedValues,
) (bool, error) {
	db := bcr.GetDatabase()
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not IssuingETHRequest instruction
			continue
		}
		instMetaType := inst[0]
		issuingETHReqAction, err := ParseInstContent(inst[3])
		if err != nil {
			fmt.Println("WARNING: an error occured during parsing instruction content: ", err)
			continue
		}

		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(IssuingETHRequestMeta) ||
			!bytes.Equal(iRes.RequestedTxID[:], issuingETHReqAction.TxReqID[:]) {
			continue
		}

		md := issuingETHReqAction.Meta
		uniqETHTx := append(md.BlockHash[:], []byte(strconv.Itoa(int(md.TxIndex)))...)
		if !bytes.Equal(uniqETHTx, iRes.UniqETHTx) {
			continue
		}

		constructedReceipt, err := VerifyProofAndParseReceipt(issuingETHReqAction, bcr)
		if err != nil {
			fmt.Println("WARNING: an error occured during verifying proof & parsing receipt: ", err)
			continue
		}

		isUsedInBlock := IsETHHashUsedInBlock(uniqETHTx, ac.UniqETHTxsUsed)
		if isUsedInBlock {
			fmt.Println("WARNING: already issued for the hash in current block: ", uniqETHTx)
			continue
		}

		isIssued, err := db.IsETHTxHashIssued(uniqETHTx)
		if err != nil {
			continue
		}
		if isIssued {
			fmt.Println("WARNING: already issued for the hash in previous block: ", uniqETHTx)
			continue
		}

		logMap, err := PickNParseLogMapFromReceipt(constructedReceipt)
		if err != nil {
			fmt.Println("WARNING: an error occured during parsing log map from receipt: ", err)
			continue
		}
		if logMap == nil {
			fmt.Println("WARNING: could not find log map out from receipt")
			continue
		}

		// the token might be ETH/ERC20
		ethereumAddr, ok := logMap["_token"].(rCommon.Address)
		if !ok {
			continue
		}
		ethereumToken := ethereumAddr[:]
		if !bytes.Equal(ethereumToken[:], iRes.ExternalTokenID[:]) {
			fmt.Println("WARNING: ethereumToken is not matched to metadata's ExternalTokenID")
			continue
		}

		canProcess, err := ac.CanProcessTokenPair(ethereumToken, md.IncTokenID)
		if err != nil {
			continue
		}
		if !canProcess {
			fmt.Println("WARNING: pair of incognito token id & ethereum's id is invalid in current block")
			continue
		}

		isValid, err := db.CanProcessTokenPair(ethereumToken, md.IncTokenID)
		if err != nil {
			fmt.Println("WARNING: An error occured as checking pair of tokens (incognito's & ethereum's): ", err)
			continue
		}
		if !isValid {
			fmt.Println("WARNING: pair of incognito token id & ethereum's id is invalid with previous blocks")
			continue
		}

		addressStr := logMap["_incognito_address"].(string)
		key, err := wallet.Base58CheckDeserialize(addressStr)
		if err != nil {
			continue
		}
		amt := logMap["_amount"].(*big.Int)
		amount := uint64(0)
		if bytes.Equal(rCommon.HexToAddress(common.ETH_ADDR_STR).Bytes(), ethereumToken) {
			// convert amt from wei (10^18) to nano eth (10^9)
			amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
		} else { // ERC20
			amount = amt.Uint64()
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			amount != paidAmount ||
			!bytes.Equal(md.IncTokenID[:], assetID[:]) {
			continue
		}
		ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqETHTx)
		ac.DBridgeTokenPair[md.IncTokenID.String()] = ethereumToken
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no IssuingETHRequest tx found for IssuingETHResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
