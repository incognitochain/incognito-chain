package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

// NOTE: for whole bridge's deposit process, anytime an error occurs an error will be logged for debugging and the request will be skipped for retry later. No error will be returned so that the network can still continue to process others.

type BurningReqAction struct {
	Meta          metadata.BurningRequest `json:"meta"`
	RequestedTxID *common.Hash            `json:"RequestedTxID"`
}

func (chain *BlockChain) processBridgeInstructions(block *BeaconBlock) error {
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.IssuingETHRequestMeta):
			err = chain.processIssuingETHReq(inst)
		case strconv.Itoa(metadata.IssuingRequestMeta):
			err = chain.processIssuingReq(inst)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (chain *BlockChain) processIssuingETHReq(inst []string) error {
	if len(inst) != 4 {
		return nil // skip the instruction
	}
	db := chain.GetDatabase()
	contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
	if err != nil {
		fmt.Println("WARNING: an error occured while decoding content string of accepted issuance instruction: ", err)
		return nil
	}
	var issuingETHAcceptedInst metadata.IssuingETHAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
	if err != nil {
		fmt.Println("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return nil
	}
	err = db.InsertETHTxHashIssued(issuingETHAcceptedInst.UniqETHTx)
	if err != nil {
		fmt.Println("WARNING: an error occured while inserting ETH tx hash issued to leveldb: ", err)
		return nil
	}
	err = db.UpdateBridgeTokenInfo(
		issuingETHAcceptedInst.IncTokenID,
		issuingETHAcceptedInst.ExternalTokenID,
		false,
	)
	if err != nil {
		fmt.Println("WARNING: an error occured while updating bridge token info to leveldb: ", err)
		return nil
	}
	return nil
}

func (chain *BlockChain) processIssuingReq(inst []string) error {
	if len(inst) != 4 {
		return nil // skip the instruction
	}
	db := chain.GetDatabase()
	contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
	if err != nil {
		fmt.Println("WARNING: an error occured while decoding content string of accepted issuance instruction: ", err)
		return nil
	}
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		fmt.Println("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return nil
	}
	err = db.UpdateBridgeTokenInfo(
		issuingAcceptedInst.IncTokenID,
		[]byte{},
		true,
	)
	if err != nil {
		fmt.Println("WARNING: an error occured while updating bridge token info to leveldb: ", err)
		return nil
	}
	return nil
}

func decodeContent(content string, action interface{}) error {
	contentBytes, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return err
	}
	return json.Unmarshal(contentBytes, &action)
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
