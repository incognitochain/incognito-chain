package blockchain

import (
	"github.com/constant-money/constant-chain/metadata"
)

func (bc *BlockChain) verifyUnusedInstructions(
	insts [][]string,
	instUsed []int,
	shardID byte,
) error {
	for i, inst := range insts {
		if instUsed[i] > 0 {
			continue
		}

		var err error
		switch inst[0] {
		// case strconv.Itoa(metadata.IssuingRequestMeta):
		// 	err = bc.verifyUnusedIssuingRequestInst(inst, shardID)

		// case strconv.Itoa(metadata.ContractingRequestMeta):
		// 	err = bc.verifyUnusedContractingRequestInst(inst, shardID)
		default:
			return nil
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockChain) verifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	txs []metadata.Transaction,
	shardID byte,
) ([]metadata.Transaction, error) {

	instUsed := make([]int, len(insts))
	invalidTxs := []metadata.Transaction{}
	for _, tx := range txs {
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(insts, instUsed, shardID, bc)
		if err != nil {
			return nil, err
		}
		if !ok {
			invalidTxs = append(invalidTxs, tx)
		}
	}
	if len(invalidTxs) > 0 {
		return invalidTxs, nil
	}
	err := bc.verifyUnusedInstructions(insts, instUsed, shardID)
	if err != nil {
		return nil, err
	}
	return invalidTxs, nil
}
