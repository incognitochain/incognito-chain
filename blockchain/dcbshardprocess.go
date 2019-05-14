package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/metadata"
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
