package blockchain

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
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
		return errors.Errorf("error parsing TxCustomToken of tx %s", tx.Hash().String())
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
			fmt.Printf("[db] invalid bondID: %h %h\n", bondID, td.bondID)
			return errors.Errorf("invalid bondID for trade bond BuyBackRequest tx: got %h, expected %h", bondID, td.bondID)
		}

		instUsed[i] += 1
		fmt.Printf("[db] inst %d matched\n", i)
		return nil
	}

	return errors.Errorf("no instruction found for BuyBackRequest tx %s", tx.Hash().String())
}

func (bc *BlockChain) verifyCrowdsalePaymentTx(tx metadata.Transaction, insts [][]string, instUsed []int, shardID byte) error {
	fmt.Printf("[db] verifying crowdsale payment tx\n")
	for i, inst := range insts {
		// Find corresponding instruction in block
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(metadata.CrowdsalePaymentMeta) || inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		cpi, err := ParseCrowdsalePaymentInstruction(inst[2])
		if err != nil {
			continue
		}
		var assetID common.Hash
		amount := uint64(0)
		pk := []byte{}
		if common.IsConstantAsset(&cpi.AssetID) {
			if _, ok := tx.(*transaction.Tx); !ok {
				continue
			}
			assetID = cpi.AssetID
			unique := false
			unique, pk, amount = tx.GetUniqueReceiver()
			if !unique {
				continue
			}
		} else {
			var customTx *transaction.TxCustomToken
			ok := false
			if customTx, ok = tx.(*transaction.TxCustomToken); !ok {
				continue
			}
			unique := false
			unique, pk, amount = tx.GetTokenUniqueReceiver()
			assetID = customTx.TxTokenData.PropertyID
			if !unique {
				continue
			}
		}

		txData := CrowdsalePaymentInstruction{
			PaymentAddress: privacy.PaymentAddress{Pk: pk},
			Amount:         amount,
			AssetID:        assetID,
			SaleID:         nil, // no need to check these last fields
			SentAmount:     0,
			UpdateSale:     false,
		}
		if !txData.Compare(cpi) {
			fmt.Printf("[db] data mismatched: %+v\t%+v\n", txData, cpi)
			return errors.Errorf("invalid data for CrowdsalePayment tx: got %+v, expect %+v", txData, cpi)
		}

		instUsed[i] += 1
		fmt.Printf("[db] inst %d matched\n", i)
		return nil
	}

	return errors.Errorf("no instruction found for CrowdsalePayment tx %s", tx.Hash().String())
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

		case metadata.CrowdsalePaymentMeta:
			err = bc.verifyCrowdsalePaymentTx(tx, insts, instUsed, block.Header.ShardID)

			// TODO(@0xbunyip): IssuingResponseMeta
			// TODO(@0xbunyip): ContractingResponseMeta
		}

		if err != nil {
			return err
		}
	}

	// TODO(@0xbunyip): check if unused instructions (of the same shard) weren't ignored:
	// 1. TradeActivation failed either because it's activated, reqAmount too high or failed building Tx
	// 2. IssuingResponse: inst type == accepted or failed building Tx
	// 3. ContractingResponse: inst type == accepted or failed building Tx
	return nil
}
