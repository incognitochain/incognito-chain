package metadata

import (
	"bytes"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
)

type IssuingResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewIssuingResponse(requestedTxID common.Hash, metaType int) *IssuingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes *IssuingResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (iRes *IssuingResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes *IssuingResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes *IssuingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes *IssuingResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes *IssuingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
) (bool, error) {

	fmt.Println("hahaha - len(txsInBlock): ", len(txsInBlock))
	idx := -1
	for i, txInBlock := range txsInBlock {
		if txsUsed[i] > 0 ||
			txInBlock.GetMetadataType() != IssuingRequestMeta ||
			!bytes.Equal(iRes.RequestedTxID[:], txInBlock.Hash()[:]) {
			fmt.Println("iRes.RequestedTxID: ", iRes.RequestedTxID[:])
			fmt.Println("txInBlock.Hash(): ", txInBlock.Hash())
			fmt.Println("hahaha - chet dk 1")
			continue
		}
		issuingReqRaw := txInBlock.GetMetadata()
		issuingReq, ok := issuingReqRaw.(*IssuingRequest)
		if !ok {
			fmt.Println("hahaha - chet dk 2")
			continue
		}
		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(issuingReq.ReceiverAddress.Pk[:], pk[:]) ||
			issuingReq.DepositedAmount != amount ||
			!bytes.Equal(issuingReq.TokenID[:], assetID[:]) {
			fmt.Println("hahaha - chet dk 3")
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no IssuingRequest tx found for IssuingResponse tx %s", tx.Hash().String())
	}
	txsUsed[idx] = 1
	return true, nil
}
