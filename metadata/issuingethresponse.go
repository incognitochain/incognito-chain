package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/ethrelaying/accounts/abi"
	rCommon "github.com/incognitochain/incognito-chain/ethrelaying/common"
	"github.com/incognitochain/incognito-chain/ethrelaying/core/types"
	"github.com/incognitochain/incognito-chain/ethrelaying/light"
	"github.com/incognitochain/incognito-chain/ethrelaying/rlp"
	"github.com/incognitochain/incognito-chain/ethrelaying/trie"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type IssuingETHResponse struct {
	MetadataBase
	RequestedTxID common.Hash
	ETHTxHash     rCommon.Hash
}

type IssuingETHReqAction struct {
	BridgeShardID byte              `json:"bridgeShardID"`
	Meta          IssuingETHRequest `json:"meta"`
	TxReqID       common.Hash       `json:"txReqId"`
}

func ParseETHLogData(data []byte) (map[string]interface{}, error) {
	fmt.Println("haha log data: ", data)

	abiIns, err := abi.JSON(strings.NewReader(common.ABIJSON))
	if err != nil {
		return nil, err
	}
	dataMap := map[string]interface{}{}
	if err = abiIns.UnpackIntoMap(dataMap, "Deposit", data); err != nil {
		return nil, err
	}
	return dataMap, nil
}

func NewIssuingETHResponse(
	requestedTxID common.Hash,
	ethTxHash rCommon.Hash,
	metaType int,
) *IssuingETHResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingETHResponse{
		RequestedTxID: requestedTxID,
		ETHTxHash:     ethTxHash,
		MetadataBase:  metadataBase,
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
	record += iRes.ETHTxHash.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingETHResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func IsETHHashUsedInBlock(ethTxHash rCommon.Hash, ethTxHashesUsed []rCommon.Hash) bool {
	for _, hash := range ethTxHashesUsed {
		if bytes.Equal(ethTxHash[:], hash[:]) {
			return true
		}
	}
	return false
}

func (iRes *IssuingETHResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ethTxHashesUsed []rCommon.Hash,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not IssuingETHRequest instruction
			continue
		}
		instMetaType := inst[0]
		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			continue
		}
		var issuingETHReqAction IssuingETHReqAction
		err = json.Unmarshal(contentBytes, &issuingETHReqAction)
		if err != nil {
			continue
		}
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(IssuingETHRequestMeta) ||
			!bytes.Equal(iRes.RequestedTxID[:], issuingETHReqAction.TxReqID[:]) {
			continue
		}

		md := issuingETHReqAction.Meta
		ethHeader := bcr.GetLightEthereum().GetLightChain().GetHeaderByHash(md.BlockHash)
		keybuf := new(bytes.Buffer)
		keybuf.Reset()
		rlp.Encode(keybuf, md.TxIndex)

		nodeList := new(light.NodeList)
		for _, proofStr := range md.ProofStrs {
			proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
			if err != nil {
				continue
			}
			nodeList.Put([]byte{}, proofBytes)
		}
		proof := nodeList.NodeSet()
		val, _, err := trie.VerifyProof(ethHeader.ReceiptHash, keybuf.Bytes(), proof)
		if err != nil {
			fmt.Printf("ETH issuance proof verification failed: %v", err)
			continue
		}

		// Decode value from VerifyProof into Receipt
		constructedReceipt := new(types.Receipt)
		err = rlp.DecodeBytes(val, constructedReceipt)
		if err != nil {
			continue
		}
		ethTxHash := constructedReceipt.TxHash
		if !bytes.Equal(ethTxHash[:], iRes.ETHTxHash[:]) {
			continue
		}
		isUsedInBlock := IsETHHashUsedInBlock(ethTxHash, ethTxHashesUsed)
		if isUsedInBlock {
			fmt.Println("WARNING: already issued for the hash in current block: ", ethTxHash)
			continue
		}

		isIssued, err := bcr.GetDatabase().IsETHTxHashIssued(ethTxHash)
		if err != nil {
			continue
		}
		if isIssued {
			fmt.Println("WARNING: already issued for the hash in previous block: ", ethTxHash)
			continue
		}
		if len(constructedReceipt.Logs) == 0 {
			continue
		}
		logData := constructedReceipt.Logs[0].Data
		logMap, err := ParseETHLogData(logData)
		if err != nil {
			continue
		}
		addressStr := logMap["_incognito_address"].(string)
		key, err := wallet.Base58CheckDeserialize(addressStr)
		if err != nil {
			continue
		}
		amt := logMap["_amount"].(*big.Int)
		// convert amt from wei (10^18) to nano eth (10^9)
		amount := big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
		tokenID, err := common.NewHashFromStr(common.PETHTokenID)
		if err != nil {
			continue
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			amount != paidAmount ||
			!bytes.Equal(tokenID[:], assetID[:]) {
			continue
		}
		ethTxHashesUsed = append(ethTxHashesUsed, ethTxHash)
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no IssuingETHRequest tx found for IssuingETHResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
