package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type IssuingReqAction struct {
	Meta metadata.IssuingRequest `json:"meta"`
}

type ContractingReqAction struct {
	Meta metadata.ContractingRequest `json:"meta"`
}

type BurningReqAction struct {
	Meta metadata.BurningRequest `json:"meta"`
}

type UpdatingInfo struct {
	countUpAmt uint64
	deductAmt  uint64
}

func (chain *BlockChain) processBridgeInstructions(block *BeaconBlock) error {
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.IssuingRequestMeta):
			updatingInfoByTokenID, err = chain.processIssuingReq(inst, updatingInfoByTokenID)
		case strconv.Itoa(metadata.ContractingRequestMeta):
			updatingInfoByTokenID, err = chain.processContractingReq(inst, updatingInfoByTokenID)
		case strconv.Itoa(metadata.BurningRequestMeta):
			updatingInfoByTokenID, err = chain.processBurningReq(inst, updatingInfoByTokenID)
		}
		if err != nil {
			return err
		}
	}
	for tokenID, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.countUpAmt > updatingInfo.deductAmt {
			updatingAmt = updatingInfo.countUpAmt - updatingInfo.deductAmt
			updatingType = "+"
		}
		if updatingInfo.countUpAmt < updatingInfo.deductAmt {
			updatingAmt = updatingInfo.deductAmt - updatingInfo.countUpAmt
			updatingType = "-"
		}
		err := chain.GetDatabase().UpdateAmtByTokenID(tokenID, updatingAmt, updatingType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockChain) processIssuingReq(
	inst []string,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo,
) (map[common.Hash]UpdatingInfo, error) {
	actionContentStr := inst[1]
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return nil, err
	}
	var issuingReqAction IssuingReqAction
	err = json.Unmarshal(contentBytes, &issuingReqAction)
	if err != nil {
		return nil, err
	}
	md := issuingReqAction.Meta
	updatingInfo, found := updatingInfoByTokenID[md.TokenID]
	if found {
		updatingInfo.countUpAmt += md.DepositedAmount
	} else {
		updatingInfo = UpdatingInfo{
			countUpAmt: md.DepositedAmount,
			deductAmt:  0,
		}
	}
	updatingInfoByTokenID[md.TokenID] = updatingInfo

	return updatingInfoByTokenID, nil
}

func (bc *BlockChain) processContractingReq(
	inst []string,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo,
) (map[common.Hash]UpdatingInfo, error) {
	actionContentStr := inst[1]
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return nil, err
	}
	var contractingReqAction ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		return nil, err
	}
	md := contractingReqAction.Meta
	updatingInfo, found := updatingInfoByTokenID[md.TokenID]
	if found {
		updatingInfo.deductAmt -= md.BurnedAmount
	} else {
		updatingInfo = UpdatingInfo{
			countUpAmt: 0,
			deductAmt:  md.BurnedAmount,
		}
	}
	updatingInfoByTokenID[md.TokenID] = updatingInfo
	return updatingInfoByTokenID, nil
}

func (bc *BlockChain) processBurningReq(
	inst []string,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo,
) (map[common.Hash]UpdatingInfo, error) {
	actionContentStr := inst[1]
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return nil, err
	}
	var burningReqAction BurningReqAction
	err = json.Unmarshal(contentBytes, &burningReqAction)
	if err != nil {
		return nil, err
	}
	md := burningReqAction.Meta
	updatingInfo, found := updatingInfoByTokenID[md.TokenID]
	if found {
		updatingInfo.deductAmt += md.BurningAmount
	} else {
		updatingInfo = UpdatingInfo{
			countUpAmt: 0,
			deductAmt:  md.BurningAmount,
		}
	}
	updatingInfoByTokenID[md.TokenID] = updatingInfo

	return updatingInfoByTokenID, nil
}
