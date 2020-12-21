package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

func (blockchain *BlockChain) processBridgeInstructions(bridgeStateDB *statedb.StateDB, block *BeaconBlock) error {
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.InitPTokenRequestMeta):
			err = blockchain.processPTokenInitReq(bridgeStateDB, inst)

		if err != nil {
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) processPTokenInitReq(bridgeStateDB *statedb.StateDB, instruction []string) error {
	if len(instruction) != 4 {
		return nil // skip the instruction
	}
	if instruction[2] == "rejected" {
		txReqID, err := common.Hash{}.NewHashFromStr(instruction[3])
		if err != nil {
			fmt.Println("WARNING: an error occured while building tx request id in bytes from string: ", err)
			return updatingInfoByTokenID, nil
		}
		err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, *txReqID, common.BridgeRequestRejectedStatus)
		if err != nil {
			fmt.Println("WARNING: an error occured while tracking bridge request with rejected status to leveldb: ", err)
		}
		return updatingInfoByTokenID, nil
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		fmt.Println("WARNING: an error occured while decoding content string of accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	var issuingETHAcceptedInst metadata.IssuingETHAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
	if err != nil {
		fmt.Println("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	err = statedb.InsertETHTxHashIssued(bridgeStateDB, issuingETHAcceptedInst.UniqETHTx)
	if err != nil {
		fmt.Println("WARNING: an error occured while inserting ETH tx hash issued to leveldb: ", err)
		return updatingInfoByTokenID, nil
	}
	updatingInfo, found := updatingInfoByTokenID[issuingETHAcceptedInst.IncTokenID]
	if found {
		updatingInfo.countUpAmt += issuingETHAcceptedInst.IssuingAmount
	} else {
		updatingInfo = UpdatingInfo{
			countUpAmt:      issuingETHAcceptedInst.IssuingAmount,
			deductAmt:       0,
			tokenID:         issuingETHAcceptedInst.IncTokenID,
			externalTokenID: issuingETHAcceptedInst.ExternalTokenID,
			isCentralized:   false,
		}
	}
	updatingInfoByTokenID[issuingETHAcceptedInst.IncTokenID] = updatingInfo
	return updatingInfoByTokenID, nil
}
