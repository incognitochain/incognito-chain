package metadata

import (
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
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
	accumulatedData *component.UsedInstData,
) (bool, error) {
	fmt.Printf("[db] verifying Contracting response tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 ||
			inst[0] != strconv.Itoa(ContractingRequestMeta) ||
			inst[1] != strconv.Itoa(int(shardID)) ||
			inst[2] != "refund" {
			continue
		}
		contractingInfo, err := component.ParseContractingInfo(inst[3])
		if err != nil {
			continue
		}

		unique, pk, amount, assetID := tx.GetTransferData()
		txData := &component.ContractingInfo{
			BurnerAddress:     privacy.PaymentAddress{Pk: pk},
			BurnedConstAmount: amount,
		}

		if unique && txData.Compare(contractingInfo) && assetID.IsEqual(&common.ConstantID) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false, errors.Errorf("no instruction found for ContractingResponse tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return true, nil
}
