package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/pkg/errors"
)

func (bc *BlockChain) StoreMetadataInstructions(inst []string, shardID byte) error {
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	case strconv.Itoa(metadata.IssuingRequestMeta):
		return bc.storeIssuingResponseInstruction(inst, shardID)
	case strconv.Itoa(metadata.ContractingRequestMeta):
		return bc.storeContractingResponseInstruction(inst, shardID)
	}
	return nil
}

func (bc *BlockChain) storeIssuingResponseInstruction(inst []string, shardID byte) error {
	// fmt.Printf("[db] store meta inst: %+v\n", inst)
	if strconv.Itoa(int(shardID)) != inst[1] {
		return nil
	}

	issuingInfo := &component.IssuingInfo{}
	err := json.Unmarshal([]byte(inst[3]), issuingInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	return bc.config.DataBase.StoreIssuingInfo(issuingInfo.RequestedTxID, issuingInfo.Amount, instType)
}

func (bc *BlockChain) storeContractingResponseInstruction(inst []string, shardID byte) error {
	// fmt.Printf("[db] store meta inst: %+v\n", inst)
	if strconv.Itoa(int(shardID)) != inst[1] {
		return nil
	}

	contractingInfo := &component.ContractingInfo{}
	err := json.Unmarshal([]byte(inst[3]), contractingInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	return bc.config.DataBase.StoreContractingInfo(contractingInfo.RequestedTxID, contractingInfo.BurnedConstAmount, contractingInfo.RedeemAmount, instType)
}

func (bc *BlockChain) processConfirmBuyBackInst(inst []string) error {
	var buyBackInfo BuyBackInfo
	json.Unmarshal([]byte(inst[3]), &buyBackInfo)
	bondID, buy, _, amount, err := bc.config.DataBase.GetTradeActivation(buyBackInfo.TradeID)
	if err != nil {
		return err
	}

	// Update activation status to false to retry later
	activated := false
	if inst[2] == "refund" {
		amount += buyBackInfo.Value
	}
	fmt.Printf("[db] processBuyBack update: %x %h %t %t %d\n", buyBackInfo.TradeID, bondID, buy, activated, amount)
	return bc.config.DataBase.StoreTradeActivation(buyBackInfo.TradeID, bondID, buy, activated, amount)
}

func (bc *BlockChain) processConfirmBuySellInst(inst []string) error {
	//fmt.Printf("[db] processBuyFromGOV inst: %s\n", inst)
	contentBytes, _ := base64.StdEncoding.DecodeString(inst[3])
	var buySellReqAction BuySellReqAction
	json.Unmarshal(contentBytes, &buySellReqAction)
	meta := buySellReqAction.Meta
	bondID, buy, _, amount, err := bc.config.DataBase.GetTradeActivation(meta.TradeID)
	if err != nil {
		return err
	}

	// Update activation status to false to retry later
	activated := false
	if inst[2] == "refund" {
		amount += meta.Amount
	}
	fmt.Printf("[db] processBuyFromGOV update: %x %h %t %t %d\n", meta.TradeID, bondID, buy, activated, amount)
	return bc.config.DataBase.StoreTradeActivation(meta.TradeID, bondID, buy, activated, amount)
}

// ProcessStandAloneInstructions processes all stand-alone instructions in block (created by both producer and validators)
func (bc *BlockChain) ProcessStandAloneInstructions(block *ShardBlock) error {
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(component.ConfirmBuySellRequestMeta):
			err = bc.processConfirmBuySellInst(inst)
		case strconv.Itoa(component.ConfirmBuyBackRequestMeta):
			err = bc.processConfirmBuyBackInst(inst)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockChain) updateTradeActivation(tradeID []byte, reqAmount uint64) error {
	// Use balance left from previous activation if it exist
	bondID, buy, _, amount, err := bc.GetLatestTradeActivation(tradeID)
	if err != nil {
		return err
	}
	if amount < reqAmount {
		return errors.Errorf("trade bond requested amount too high, %d > %d\n", reqAmount, amount)
	}

	activated := true
	fmt.Printf("[db] updating trade bond status: %v %h %t %t %d\n", tradeID, bondID, buy, activated, reqAmount)
	return bc.config.DataBase.StoreTradeActivation(tradeID, bondID, buy, activated, amount-reqAmount)
}

func (bc *BlockChain) processBuyBondTx(tx metadata.Transaction) error {
	meta := tx.GetMetadata().(*metadata.BuySellRequest)
	tradeID := meta.TradeID
	if len(tradeID) == 0 {
		return nil
	}
	fmt.Printf("[db] process buy bond tx: %x %d\n", tradeID, meta.Amount)
	return bc.updateTradeActivation(tradeID, meta.Amount)
}

func (bc *BlockChain) processSellBondTx(tx metadata.Transaction) error {
	meta := tx.GetMetadata().(*metadata.BuyBackRequest)
	tradeID := meta.TradeID
	if len(tradeID) == 0 {
		return nil
	}
	fmt.Printf("[db] process sell bond tx: %x %d\n", tradeID, meta.Amount)
	return bc.updateTradeActivation(tradeID, meta.Amount)
}

func (bc *BlockChain) processTradeBondTx(block *ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		var err error
		switch tx.GetMetadataType() {
		case metadata.BuyFromGOVRequestMeta:
			err = bc.processBuyBondTx(tx)

		case metadata.BuyBackRequestMeta:
			err = bc.processSellBondTx(tx)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
