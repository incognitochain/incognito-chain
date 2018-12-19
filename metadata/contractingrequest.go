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
	chainID byte,
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
	if cReq.Type != ContractingRequestMeta {
		return false
	}
	return true
}

func (cReq *ContractingRequest) Hash() *common.Hash {
	record := string(cReq.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
