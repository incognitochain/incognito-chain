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

type usedInstData struct {
	tradeActivated map[string]bool
}

func (bc *BlockChain) verifyBuyFromGOVRequestTx(
	tx metadata.Transaction,
	insts [][]string,
	instUsed []int,
	shardID byte,
	accumulatedData usedInstData,
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
	accumulatedData.tradeActivated[string(meta.TradeID)] = true
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

func (bc *BlockChain) verifyStabilityTransactionsWithInstructions(
	insts [][]string,
	block *ShardBlock,
	instUsed []int,
	shardID byte,
	accumulatedData usedInstData,
) error {
	for _, tx := range block.Body.Transactions {
		if tx.GetMetadata() == nil {
			continue
		}

		var err error
		switch tx.GetMetadataType() {
		case metadata.BuyFromGOVRequestMeta:
			err = bc.verifyBuyFromGOVRequestTx(tx, insts, instUsed, shardID, accumulatedData)

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
	return nil
}

func (bc *BlockChain) verifyUnusedTradeActivationInst(inst []string, shardID byte, accumulatedData usedInstData) error {
	// TradeActivation failed either because it's activated, reqAmount too high or failed building Tx
	if inst[1] != strconv.Itoa(int(shardID)) {
		return nil
	}

	data, err := bc.calcTradeData(inst[2])
	if err != nil {
		return nil
	}

	if data.activated || data.reqAmount > data.amount || accumulatedData.tradeActivated[string(data.tradeID)] {
		return nil
	}

	if !data.buy {
		// Assume not enough bond to create BuyBackRequest tx
		return nil
	}
	return fmt.Errorf("invalid unused inst: %v, %d, %+v", inst, shardID, accumulatedData)
}

func (bc *BlockChain) verifyUnusedIssuingRequestInst(inst []string, shardID byte) error {
	// IssuingRequest inst unused either because type != accepted or failed building Tx
	if inst[1] != strconv.Itoa(int(shardID)) || inst[2] != "accepted" {
		return nil
	}

	_, err := parseIssuingInfo(inst[3])
	if err != nil {
		return nil
	}
	return fmt.Errorf("invalid unused inst: %v, %d, %+v", inst, shardID)
}

func (bc *BlockChain) verifyUnusedContractingRequestInst(inst []string, shardID byte) error {
	// ContractingRequest inst unused either because type != refund or failed building Tx
	if inst[1] != strconv.Itoa(int(shardID)) || inst[2] != "refund" {
		return nil
	}

	_, err := parseContractingInfo(inst[3])
	if err != nil {
		return nil
	}
	return fmt.Errorf("invalid unused inst: %v, %d, %+v", inst, shardID)
}

func (bc *BlockChain) verifyUnusedInstructions(
	insts [][]string,
	instUsed []int,
	shardID byte,
	accumulatedData usedInstData,
) error {
	for i, inst := range insts {
		if instUsed[i] > 0 {
			continue
		}

		var err error
		switch inst[0] {
		// TODO(@0xbunyip): review other insts
		case strconv.Itoa(metadata.TradeActivationMeta):
			err = bc.verifyUnusedTradeActivationInst(inst, shardID, accumulatedData)

		case strconv.Itoa(metadata.IssuingRequestMeta):
			err = bc.verifyUnusedIssuingRequestInst(inst, shardID)

		case strconv.Itoa(metadata.ContractingRequestMeta):
			err = bc.verifyUnusedContractingRequestInst(inst, shardID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockChain) verifyStabilityTransactionsForNewBlock(insts [][]string, block *ShardBlock) error {
	instUsed := make([]int, len(insts)) // Count how many times an inst is used by a tx
	accumulatedData := usedInstData{
		tradeActivated: map[string]bool{},
	}
	shardID := block.Header.ShardID

	err := bc.verifyStabilityTransactionsWithInstructions(insts, block, instUsed, shardID, accumulatedData)
	if err != nil {
		return err
	}

	return bc.verifyUnusedInstructions(insts, instUsed, shardID, accumulatedData)
}
