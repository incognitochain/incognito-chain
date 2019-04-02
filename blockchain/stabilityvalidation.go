package blockchain

import (
	"bytes"
	"strconv"

	"github.com/constant-money/constant-chain/metadata"
	"github.com/pkg/errors"
)

func (bc *BlockChain) verifyBuyFromGOVRequestTx(tx metadata.Transaction, insts [][]string, instUsed []int) error {
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
		tradeID, reqAmount, err := metadata.ParseTradeActivationActionValue(inst[2])
		if err != nil || !bytes.Equal(meta.TradeID, tradeID) {
			continue
		}

		// Check if amount is correct
		if meta.Amount != reqAmount {
			return errors.Errorf("invalid amount for trade bond BuySellRequest tx: got %d instead of %d", meta.Amount, reqAmount)
		}

		instUsed[i] += 1
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
		}

		if err != nil {
			return err
		}
	}
	return nil
}
