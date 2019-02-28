package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type ContractingRequest struct {
	MetadataBase
}

func NewContractingRequest(
	metaType int,
) *ContractingRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	contractingReq := &ContractingRequest{}
	contractingReq.MetadataBase = metadataBase
	return contractingReq
}

func (cReq *ContractingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (cReq *ContractingRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if !txr.IsCoinsBurning() {
		return false, false, nil
	}
	return true, true, nil
}

func (cReq *ContractingRequest) ValidateMetadataByItself() bool {
	return cReq.Type != ContractingRequestMeta
}

func (cReq *ContractingRequest) Hash() *common.Hash {
	record := cReq.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cReq *ContractingRequest) CalculateSize() uint64 {
	return calculateSize(cReq)
}
