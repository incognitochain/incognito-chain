package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/metadata"
)

type IssuingReqAction struct {
	Meta metadata.IssuingRequest `json:"meta"`
}

type ContractingReqAction struct {
	Meta metadata.ContractingRequest `json:"meta"`
}

func (chain *BlockChain) processBridgeInstructions(block *BeaconBlock) error {
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		switch inst[0] {
		case strconv.Itoa(metadata.IssuingRequestMeta):
			return chain.processIssuingReq(inst)

		case strconv.Itoa(metadata.ContractingRequestMeta):
			return chain.processContractingReq(inst)
		}
	}
	return nil
}

func (bc *BlockChain) processIssuingReq(inst []string) error {
	actionContentStr := inst[1]
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return err
	}
	var issuingReqAction IssuingReqAction
	err = json.Unmarshal(contentBytes, &issuingReqAction)
	if err != nil {
		return err
	}
	md := issuingReqAction.Meta
	err = bc.GetDatabase().CountUpDepositedAmtByTokenID(&md.TokenID, md.DepositedAmount)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) processContractingReq(inst []string) error {
	actionContentStr := inst[1]
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return err
	}
	var contractingReqAction ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		return err
	}
	md := contractingReqAction.Meta
	err = bc.GetDatabase().DeductAmtByTokenID(&md.TokenID, md.BurnedAmount)
	if err != nil {
		return err
	}
	return nil
}
