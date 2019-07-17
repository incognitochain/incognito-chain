package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type BurningReqAction struct {
	Meta          metadata.BurningRequest `json:"meta"`
	RequestedTxID *common.Hash            `json:"RequestedTxID"`
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
		case strconv.Itoa(metadata.BurningRequestMeta):
			updatingInfoByTokenID, err = chain.processBurningReq(inst, updatingInfoByTokenID)
		case strconv.Itoa(metadata.IssuingResponseMeta):
			err = chain.processIssuingRes(inst)
		case strconv.Itoa(metadata.IssuingETHResponseMeta):
			err = chain.processIssuingETHRes(inst)
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

func (chain *BlockChain) processIssuingRes(inst []string) error {
	contentBytes, err := base64.StdEncoding.DecodeString(inst[1])
	if err != nil {
		return err
	}
	var issuingResAction metadata.IssuingResAction
	err = json.Unmarshal(contentBytes, &issuingResAction)
	if err != nil {
		return err
	}
	return chain.GetDatabase().UpdateBridgeTokenInfo(
		*issuingResAction.IncTokenID,
		[]byte{},
		true,
	)
}

func (chain *BlockChain) processIssuingETHRes(inst []string) error {
	contentBytes, err := base64.StdEncoding.DecodeString(inst[1])
	if err != nil {
		return err
	}
	var issuingETHResAction metadata.IssuingETHResAction
	err = json.Unmarshal(contentBytes, &issuingETHResAction)
	if err != nil {
		return err
	}
	db := chain.GetDatabase()
	err = db.InsertETHTxHashIssued(issuingETHResAction.Meta.UniqETHTx)
	if err != nil {
		return err
	}
	return db.UpdateBridgeTokenInfo(
		*issuingETHResAction.IncTokenID,
		issuingETHResAction.Meta.ExternalTokenID,
		false,
	)
}

func decodeContent(content string, action interface{}) error {
	contentBytes, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return err
	}
	return json.Unmarshal(contentBytes, &action)
}

func (bc *BlockChain) processBurningReq(
	inst []string,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo,
) (map[common.Hash]UpdatingInfo, error) {
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
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

func (bc *BlockChain) storeBurningConfirm(block *ShardBlock) error {
	if len(block.Body.Instructions) > 0 {
		fmt.Printf("[db] storeBurningConfirm for block %d %v\n", block.Header.Height, block.Body.Instructions)
	}
	for _, inst := range block.Body.Instructions {
		if inst[0] != strconv.Itoa(metadata.BurningConfirmMeta) {
			continue
		}
		fmt.Printf("[db] storeBurning: %s\n", inst)

		txID, err := common.Hash{}.NewHashFromStr(inst[5])
		if err != nil {
			fmt.Printf("[db] storeBurning err: %v\n", err)
			return err
		}
		fmt.Printf("[db] storing BurningConfirm inst with txID: %x\n", txID)
		if err := bc.config.DataBase.StoreBurningConfirm(txID[:], block.Header.Height); err != nil {
			return err
		}
	}
	return nil
}
