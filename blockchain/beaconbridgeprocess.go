package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

// NOTE: for whole bridge's deposit process, anytime an error occurs it will be logged for debugging and the request will be skipped for retry later. No error will be returned so that the network can still continue to process others.

type BurningReqAction struct {
	Meta          metadata.BurningRequest `json:"meta"`
	RequestedTxID *common.Hash            `json:"RequestedTxID"`
}

func (blockchain *BlockChain) processBridgeInstructions(bridgeStateDB *statedb.StateDB, block *BeaconBlock) error {
	updatingInfoByTokenID := map[common.Hash]basemeta.UpdatingInfo{}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(basemeta.IssuingETHRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingETHReq(bridgeStateDB, inst, updatingInfoByTokenID)

		case strconv.Itoa(basemeta.IssuingRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingReq(bridgeStateDB, inst, updatingInfoByTokenID)

		case strconv.Itoa(basemeta.ContractingRequestMeta):
			updatingInfoByTokenID, err = blockchain.processContractingReq(inst, updatingInfoByTokenID)

		case strconv.Itoa(basemeta.BurningConfirmMeta), strconv.Itoa(basemeta.BurningConfirmForDepositToSCMeta), strconv.Itoa(basemeta.BurningConfirmMetaV2), strconv.Itoa(basemeta.BurningConfirmForDepositToSCMetaV2):
			updatingInfoByTokenID, err = blockchain.processBurningReq(inst, updatingInfoByTokenID)

		}
		if err != nil {
			return err
		}
	}
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		}
		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			bridgeStateDB,
			updatingInfo.TokenID,
			updatingInfo.ExternalTokenID,
			updatingInfo.IsCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) processIssuingETHReq(bridgeStateDB *statedb.StateDB, instruction []string, updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo) (map[common.Hash]basemeta.UpdatingInfo, error) {
	if len(instruction) != 4 {
		return updatingInfoByTokenID, nil // skip the instruction
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
		updatingInfo.CountUpAmt += issuingETHAcceptedInst.IssuingAmount
	} else {
		updatingInfo = basemeta.UpdatingInfo{
			CountUpAmt:      issuingETHAcceptedInst.IssuingAmount,
			DeductAmt:       0,
			TokenID:         issuingETHAcceptedInst.IncTokenID,
			ExternalTokenID: issuingETHAcceptedInst.ExternalTokenID,
			IsCentralized:   false,
		}
	}
	updatingInfoByTokenID[issuingETHAcceptedInst.IncTokenID] = updatingInfo
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) processIssuingReq(bridgeStateDB *statedb.StateDB, instruction []string, updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo) (map[common.Hash]basemeta.UpdatingInfo, error) {
	if len(instruction) != 4 {
		return updatingInfoByTokenID, nil // skip the instruction
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
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		fmt.Println("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	updatingInfo, found := updatingInfoByTokenID[issuingAcceptedInst.IncTokenID]
	if found {
		updatingInfo.CountUpAmt += issuingAcceptedInst.DepositedAmount
	} else {
		updatingInfo = basemeta.UpdatingInfo{
			CountUpAmt:    issuingAcceptedInst.DepositedAmount,
			DeductAmt:     0,
			TokenID:       issuingAcceptedInst.IncTokenID,
			IsCentralized: true,
		}
	}
	updatingInfoByTokenID[issuingAcceptedInst.IncTokenID] = updatingInfo
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) processContractingReq(instruction []string, updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo) (map[common.Hash]basemeta.UpdatingInfo, error) {
	if len(instruction) != 4 {
		return updatingInfoByTokenID, nil // skip the instruction
	}
	if instruction[2] == "rejected" {
		return updatingInfoByTokenID, nil // skip the instruction
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		fmt.Println("WARNING: an error occured while decoding content string of accepted contracting instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	var contractingReqAction metadata.ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		fmt.Println("WARNING: an error occured while unmarshaling accepted contracting instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	md := contractingReqAction.Meta
	updatingInfo, found := updatingInfoByTokenID[md.TokenID]
	if found {
		updatingInfo.DeductAmt += md.BurnedAmount
	} else {
		updatingInfo = basemeta.UpdatingInfo{
			CountUpAmt:    0,
			DeductAmt:     md.BurnedAmount,
			TokenID:       md.TokenID,
			IsCentralized: true,
		}
	}
	updatingInfoByTokenID[md.TokenID] = updatingInfo
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) processBurningReq(instruction []string, updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo) (map[common.Hash]basemeta.UpdatingInfo, error) {
	if len(instruction) < 8 {
		return updatingInfoByTokenID, nil // skip the instruction
	}

	externalTokenID, _, errExtToken := base58.Base58Check{}.Decode(instruction[2])
	incTokenIDBytes, _, errIncToken := base58.Base58Check{}.Decode(instruction[6])
	amountBytes, _, errAmount := base58.Base58Check{}.Decode(instruction[4])
	if err := common.CheckError(errExtToken, errIncToken, errAmount); err != nil {
		BLogger.log.Error(errors.WithStack(err))
		return updatingInfoByTokenID, nil
	}
	amt := big.NewInt(0).SetBytes(amountBytes)
	amount := uint64(0)
	if bytes.Equal(externalTokenID, rCommon.HexToAddress(common.EthAddrStr).Bytes()) {
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else {
		amount = amt.Uint64()
	}

	incTokenID := &common.Hash{}
	incTokenID, _ = (*incTokenID).NewHash(incTokenIDBytes)
	updatingInfo, found := updatingInfoByTokenID[*incTokenID]
	if found {
		updatingInfo.DeductAmt += amount
	} else {
		updatingInfo = basemeta.UpdatingInfo{
			CountUpAmt:      0,
			DeductAmt:       amount,
			TokenID:         *incTokenID,
			ExternalTokenID: externalTokenID,
			IsCentralized:   false,
		}
	}
	updatingInfoByTokenID[*incTokenID] = updatingInfo
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) storeBurningConfirm(stateDB *statedb.StateDB, instructions [][]string, blockHeight uint64, metas []string) error {
	for _, inst := range instructions {
		found := false
		for _, meta := range metas {
			if inst[0] == meta {
				found = true
			}
		}

		if !found {
			continue
		}

		BLogger.log.Infof("storeBurningConfirm for block %d, inst %v, meta type %v", blockHeight, inst, inst[0])

		txID, err := common.Hash{}.NewHashFromStr(inst[5])
		if err != nil {
			return errors.Wrap(err, "txid invalid")
		}
		if err := statedb.StoreBurningConfirm(stateDB, *txID, blockHeight); err != nil {
			return errors.Wrapf(err, "store failed, txID: %x", txID)
		}
	}
	return nil
}

func (blockchain *BlockChain) updateBridgeIssuanceStatus(bridgeStateDB *statedb.StateDB, block *ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		metaType := tx.GetMetadataType()
		var err error
		var reqTxID common.Hash
		if metaType == basemeta.IssuingETHRequestMeta || metaType == basemeta.IssuingRequestMeta {
			reqTxID = *tx.Hash()
			err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, reqTxID, common.BridgeRequestProcessingStatus)
			if err != nil {
				return err
			}
		}
		if metaType == basemeta.IssuingETHResponseMeta {
			meta := tx.GetMetadata().(*metadata.IssuingETHResponse)
			reqTxID = meta.RequestedTxID
			err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, reqTxID, common.BridgeRequestAcceptedStatus)
			if err != nil {
				return err
			}
		} else if metaType == basemeta.IssuingResponseMeta {
			meta := tx.GetMetadata().(*metadata.IssuingResponse)
			reqTxID = meta.RequestedTxID
			err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, reqTxID, common.BridgeRequestAcceptedStatus)
			if err != nil {
				return err
			}
		}
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
