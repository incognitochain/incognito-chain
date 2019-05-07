package blockchain

import (
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/metadata"
)

func (bc *BlockChain) verifyUnusedTradeActivationInst(inst []string, shardID byte, accumulatedData *component.UsedInstData) error {
	// TradeActivation failed either because it's activated, reqAmount too high or failed building Tx
	if inst[1] != strconv.Itoa(int(shardID)) {
		return nil
	}

	data, err := bc.CalcTradeData(inst[2])
	if err != nil {
		return nil
	}

	if data.Activated || data.ReqAmount > data.Amount || accumulatedData.TradeActivated[string(data.TradeID)] {
		return nil
	}

	if !data.Buy {
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

	_, err := component.ParseIssuingInfo(inst[3])
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

	_, err := component.ParseContractingInfo(inst[3])
	if err != nil {
		return nil
	}
	return fmt.Errorf("invalid unused inst: %v, %d, %+v", inst, shardID)
}

func (bc *BlockChain) verifyUnusedCrowdsalePaymentInst(inst []string, shardID byte) error {
	// ContractingRequest inst unused either because invalid inst data or failed building Tx
	if inst[1] != strconv.Itoa(int(shardID)) {
		return nil
	}

	_, err := component.ParseContractingInfo(inst[3])
	if err != nil {
		return nil
	}

	// Asumme Constant and bonds are always enough so unused inst is unacceptable
	return fmt.Errorf("invalid unused inst: %v, %d, %+v", inst, shardID)
}

func (bc *BlockChain) verifyUnusedInstructions(
	insts [][]string,
	instUsed []int,
	shardID byte,
	accumulatedData *component.UsedInstData,
) error {
	for i, inst := range insts {
		if instUsed[i] > 0 {
			continue
		}

		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.TradeActivationMeta):
			err = bc.verifyUnusedTradeActivationInst(inst, shardID, accumulatedData)

		case strconv.Itoa(metadata.IssuingRequestMeta):
			err = bc.verifyUnusedIssuingRequestInst(inst, shardID)

		case strconv.Itoa(metadata.ContractingRequestMeta):
			err = bc.verifyUnusedContractingRequestInst(inst, shardID)

		case strconv.Itoa(metadata.CrowdsalePaymentMeta):
			err = bc.verifyUnusedCrowdsalePaymentInst(inst, shardID)
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
	accumulatedData := component.UsedInstData{
		TradeActivated: map[string]bool{},
	}

	instUsed := make([]int, len(insts))
	invalidTxs := []metadata.Transaction{}
	for _, tx := range txs {
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(insts, instUsed, shardID, bc, &accumulatedData)
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
	err := bc.verifyUnusedInstructions(insts, instUsed, shardID, &accumulatedData)
	if err != nil {
		return nil, err
	}
	return invalidTxs, nil
}
