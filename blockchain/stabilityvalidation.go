package blockchain

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/metadata"
	"github.com/pkg/errors"
)

func (bc *BlockChain) verifyBuyFromGOVRequestTx(tx metadata.Transaction, insts [][]string, instUsed []int) error {
	fmt.Printf("[db] verifying buy form GOV Request tx\n")
	meta, ok := tx.GetMetadata().(*metadata.BuySellRequest)
	if !ok {
		return errors.Errorf("error parsing metadata BuySellRequest of tx %s", tx.Hash().String())
	}
	if len(meta.TradeID) == 0 {
		return nil
	}

	for i, inst := range insts {
		// Find corresponding instruction in block
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(metadata.TradeActivationMeta) {
			continue
		}
		td, err := bc.calcTradeData(inst[2])
		if err != nil && !bytes.Equal(meta.TradeID, td.tradeID) {
			continue
		}

		fmt.Printf("[db] found inst: %s\n", inst[2])

		txData := &tradeData{
			tradeID:   meta.TradeID,
			bondID:    &meta.TokenID,
			buy:       true,
			activated: false,
			amount:    td.amount, // no need to check
			reqAmount: meta.Amount,
		}

		if !td.Compare(txData) {
			fmt.Printf("[db] data mismatched: %+v\t %+v", td, txData)
			return errors.Errorf("invalid data for trade bond BuySellRequest tx: got %+v, expect %+v", td, txData)
		}

		instUsed[i] += 1
		fmt.Printf("[db] inst %d matched\n", i)
		return nil
	}

	return errors.Errorf("no instruction found for BuySellRequest tx %s", tx.Hash().String())
}

func (bc *BlockChain) VerifyStabilityTransactionsForNewBlock(insts [][]string, block *ShardBlock) error {
	instUsed := make([]int, len(insts)) // Count how many times an inst is used by a tx
	for _, tx := range block.Body.Transactions {
		if tx.GetMetadata() == nil {
			continue
		}

		var err error
		switch tx.GetMetadataType() {
		case metadata.BuyFromGOVRequestMeta:
			err = bc.verifyBuyFromGOVRequestTx(tx, insts, instUsed)
		case metadata.ShardBlockSalaryResponseMeta:
			err = bc.verifyShardBlockSalaryResTx(tx, insts, instUsed, block.Header.ShardID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
