package blockchain

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/pkg/errors"
)

func (bc *BlockChain) verifyBuyFromGOVRequestTx(tx metadata.Transaction, insts [][]string, instUsed []int) error {
	fmt.Printf("[db] verifying buy from GOV Request tx\n")
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
		if err != nil || !bytes.Equal(meta.TradeID, td.tradeID) {
			continue
		}

		fmt.Printf("[db] found inst: %s\n", inst[2])

		// PaymentAddress is validated in metadata's ValidateWithBlockChain
		txData := &tradeData{
			tradeID:   meta.TradeID,
			bondID:    &meta.TokenID,
			buy:       true,
			activated: false,
			amount:    td.amount, // no need to check
			reqAmount: meta.Amount,
		}

		buyPrice := bc.getSellBondPrice(txData.bondID)
		if !td.Compare(txData) || meta.BuyPrice != buyPrice {
			fmt.Printf("[db] data mismatched: %+v %d\t %+v %d\n", txData, meta.BuyPrice, td, buyPrice)
			return errors.Errorf("invalid data for trade bond BuySellRequest tx: got %+v %d, expect %+v %d", txData, meta.BuyPrice, td, buyPrice)
		}

		instUsed[i] += 1
		fmt.Printf("[db] inst %d matched\n", i)
		return nil
	}

	return errors.Errorf("no instruction found for BuySellRequest tx %s", tx.Hash().String())
}

func (bc *BlockChain) verifyBuyBackRequestTx(tx metadata.Transaction, insts [][]string, instUsed []int) error {
	fmt.Printf("[db] verifying buy back GOV Request tx\n")
	meta, ok := tx.GetMetadata().(*metadata.BuyBackRequest)
	if !ok {
		return errors.Errorf("error parsing metadata BuyBackRequest of tx %s", tx.Hash().String())
	}
	if len(meta.TradeID) == 0 {
		return nil
	}

	txToken, ok := tx.(*transaction.TxCustomToken)
	if !ok {
		return errors.Errorf("erroor parsing TxCustomToken of tx %s", tx.Hash().String())
	}

	for i, inst := range insts {
		// Find corresponding instruction in block
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(metadata.TradeActivationMeta) {
			continue
		}
		td, err := bc.calcTradeData(inst[2])
		if err != nil || !bytes.Equal(meta.TradeID, td.tradeID) {
			continue
		}

		fmt.Printf("[db] found inst: %s\n", inst[2])

		// PaymentAddress is validated in metadata's ValidateWithBlockChain
		txData := &tradeData{
			tradeID:   meta.TradeID,
			bondID:    td.bondID, // not available for BuyBackRequest meta
			buy:       false,
			activated: false,
			amount:    td.amount, // no need to check
			reqAmount: meta.Amount,
		}

		if !td.Compare(txData) {
			fmt.Printf("[db] data mismatched: %+v\t%+v\n", txData, td)
			return errors.Errorf("invalid data for trade bond BuyBackRequest tx: got %+v, expect %+v", txData, td)
		}

		bondID := &txToken.TxTokenData.PropertyID
		if !bondID.IsEqual(td.bondID) {
			fmt.Printf("[db] invalid bondID: %s %s\n", bondID.Short(), td.bondID.Short())
			return errors.Errorf("invalid bondID for trade bond BuyBackRequest tx: got %s, expected %s", bondID.Short(), td.bondID.Short())
		}

		instUsed[i] += 1
		fmt.Printf("[db] inst %d matched\n", i)
		return nil
	}

	return errors.Errorf("no instruction found for BuyBackRequest tx %s", tx.Hash().String())
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

		case metadata.BuyBackRequestMeta:
			err = bc.verifyBuyBackRequestTx(tx, insts, instUsed)

		case metadata.ShardBlockSalaryResponseMeta:
			err = bc.verifyShardBlockSalaryResTx(tx, insts, instUsed, block.Header.ShardID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
