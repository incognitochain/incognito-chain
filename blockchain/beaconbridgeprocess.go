package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

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
		case strconv.Itoa(metadata.IssuingResponseMeta):
			err = chain.processIssuingRes(inst)
		case strconv.Itoa(metadata.IssuingETHResponseMeta):
			err = chain.processIssuingETHRes(inst)
		}
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

func (bc *BlockChain) storeBurningConfirm(block *ShardBlock) error {
	if len(block.Body.Instructions) > 0 {
		BLogger.log.Debugf("storeBurningConfirm for block %d %v", block.Header.Height, block.Body.Instructions)
	}
	for _, inst := range block.Body.Instructions {
		if inst[0] != strconv.Itoa(metadata.BurningConfirmMeta) {
			continue
		}

		txID, err := common.Hash{}.NewHashFromStr(inst[5])
		if err != nil {
			return errors.Wrap(err, "txid invalid")
		}
		if err := bc.config.DataBase.StoreBurningConfirm(txID[:], block.Header.Height); err != nil {
			return errors.Wrapf(err, "store failed, txID: %x", txID)
		}
	}
	return nil
}
