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

func (bc *BlockChain) verifyBuyFromGOVRequestTx(
	tx metadata.Transaction,
	insts [][]string,
	instUsed []int,
	shardID byte,
) error {
	meta, ok := tx.GetMetadata().(*metadata.BuySellRequest)
	if !ok {
		return errors.Errorf("error parsing metadata BuySellRequest of tx %s", tx.Hash().String())
	}
	if len(meta.TradeID) == 0 {
		return nil
	}

	fmt.Printf("[db] verifying buy from GOV Request tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(metadata.TradeActivationMeta) || inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		td, err := bc.calcTradeData(inst[2])
		if err != nil || !bytes.Equal(meta.TradeID, td.tradeID) {
			continue
		}

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
		if td.Compare(txData) && meta.BuyPrice == buyPrice {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.Errorf("no instruction found for BuySellRequest tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return nil
}

func (bc *BlockChain) verifyBuyBackRequestTx(
	tx metadata.Transaction,
	insts [][]string,
	instUsed []int,
	shardID byte,
) error {
	meta, ok := tx.GetMetadata().(*metadata.BuyBackRequest)
	if !ok {
		return errors.Errorf("error parsing metadata BuyBackRequest of tx %s", tx.Hash().String())
	}
	if len(meta.TradeID) == 0 {
		return nil
	}

	fmt.Printf("[db] verifying buy back GOV Request tx\n")

	txToken, ok := tx.(*transaction.TxCustomToken)
	if !ok {
		return errors.Errorf("error parsing TxCustomToken of tx %s", tx.Hash().String())
	}
	bondID := &txToken.TxTokenData.PropertyID

	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(metadata.TradeActivationMeta) || inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		td, err := bc.calcTradeData(inst[2])
		if err != nil || !bytes.Equal(meta.TradeID, td.tradeID) {
			continue
		}

		// PaymentAddress is validated in metadata's ValidateWithBlockChain
		txData := &tradeData{
			tradeID:   meta.TradeID,
			bondID:    td.bondID, // not available for BuyBackRequest meta
			buy:       false,
			activated: false,
			amount:    td.amount, // no need to check
			reqAmount: meta.Amount,
		}

		if td.Compare(txData) && bondID.IsEqual(td.bondID) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.Errorf("no instruction found for BuyBackRequest tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return nil
}

func (bc *BlockChain) verifyCrowdsalePaymentTx(
	tx metadata.Transaction,
	insts [][]string,
	instUsed []int,
	shardID byte,
) error {
	fmt.Printf("[db] verifying crowdsale payment tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(metadata.CrowdsalePaymentMeta) || inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		cpi, err := ParseCrowdsalePaymentInstruction(inst[2])
		if err != nil {
			continue
		}
		unique, pk, amount, assetID := tx.GetTransferData()
		txData := CrowdsalePaymentInstruction{
			PaymentAddress: privacy.PaymentAddress{Pk: pk},
			Amount:         amount,
			AssetID:        *assetID,
			SaleID:         nil, // no need to check these last fields
			SentAmount:     0,
			UpdateSale:     false,
		}
		if unique && txData.Compare(cpi) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.Errorf("no instruction found for CrowdsalePayment tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return nil
}

func (bc *BlockChain) verifyIssuingResponseTx(
	tx metadata.Transaction,
	insts [][]string,
	instUsed []int,
	shardID byte,
) error {
	fmt.Printf("[db] verifying issuing response tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 ||
			inst[0] != strconv.Itoa(metadata.IssuingRequestMeta) ||
			inst[1] != strconv.Itoa(int(shardID)) ||
			inst[2] != "accepted" {
			continue
		}
		issuingInfo, err := parseIssuingInfo(inst[3])
		if err != nil {
			continue
		}
		unique, pk, amount, assetID := tx.GetTransferData()
		txData := &IssuingInfo{
			ReceiverAddress: privacy.PaymentAddress{Pk: pk},
			Amount:          amount,
			TokenID:         *assetID,
		}

		if unique && txData.Compare(issuingInfo) {
			idx = i
			break
		}

	}

	if idx == -1 {
		return errors.Errorf("no instruction found for IssuingResponse tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return nil
}

func (bc *BlockChain) verifyContractingResponseTx(
	tx metadata.Transaction,
	insts [][]string,
	instUsed []int,
	shardID byte,
) error {
	fmt.Printf("[db] verifying Contracting response tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 ||
			inst[0] != strconv.Itoa(metadata.ContractingRequestMeta) ||
			inst[1] != strconv.Itoa(int(shardID)) ||
			inst[2] != "refund" {
			continue
		}
		contractingInfo, err := parseContractingInfo(inst[3])
		if err != nil {
			continue
		}

		unique, pk, amount, assetID := tx.GetTransferData()
		txData := &ContractingInfo{
			BurnerAddress:     privacy.PaymentAddress{Pk: pk},
			BurnedConstAmount: amount,
		}

		if unique && txData.Compare(contractingInfo) && assetID.IsEqual(&common.ConstantID) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.Errorf("no instruction found for ContractingResponse tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return nil
}

func (bc *BlockChain) VerifyStabilityTransactionsForNewBlock(insts [][]string, block *ShardBlock) error {
	instUsed := make([]int, len(insts)) // Count how many times an inst is used by a tx
	shardID := block.Header.ShardID
	for _, tx := range block.Body.Transactions {
		if tx.GetMetadata() == nil {
			continue
		}

		var err error
		switch tx.GetMetadataType() {
		case metadata.BuyFromGOVRequestMeta:
			err = bc.verifyBuyFromGOVRequestTx(tx, insts, instUsed, shardID)

		case metadata.BuyBackRequestMeta:
			err = bc.verifyBuyBackRequestTx(tx, insts, instUsed, shardID)

		case metadata.ShardBlockSalaryResponseMeta:
			err = bc.verifyShardBlockSalaryResTx(tx, insts, instUsed, shardID)

		case metadata.CrowdsalePaymentMeta:
			err = bc.verifyCrowdsalePaymentTx(tx, insts, instUsed, shardID)

		case metadata.IssuingResponseMeta:
			err = bc.verifyIssuingResponseTx(tx, insts, instUsed, shardID)

		case metadata.ContractingResponseMeta:
			err = bc.verifyContractingResponseTx(tx, insts, instUsed, shardID)
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
