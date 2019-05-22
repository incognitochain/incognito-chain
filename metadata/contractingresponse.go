package metadata

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
)

type ContractingResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewContractingResponse(requestedTxID common.Hash, metaType int) *ContractingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &ContractingResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (cRes *ContractingResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (cRes *ContractingResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return false, nil
}

func (cRes *ContractingResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (cRes *ContractingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (cRes *ContractingResponse) CalculateSize() uint64 {
	return calculateSize(cRes)
}

func (cRes *ContractingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
) (bool, error) {
	return true, nil
}
