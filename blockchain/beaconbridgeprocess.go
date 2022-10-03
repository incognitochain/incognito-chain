package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/pkg/errors"
)

// NOTE: for whole bridge's deposit process, anytime an error occurs it will be logged for debugging and the request will be skipped for retry later. No error will be returned so that the network can still continue to process others.

type BurningReqAction struct {
	Meta          metadataBridge.BurningRequest `json:"meta"`
	RequestedTxID *common.Hash                  `json:"RequestedTxID"`
}

func (blockchain *BlockChain) processBridgeInstructions(curView *BeaconBestState, block *types.BeaconBlock, bridgeAggUnshieldTxIDs map[string]bool) error {
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.IssuingETHRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertETHTxHashIssued, false)

		case strconv.Itoa(metadata.IssuingBSCRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertBSCTxHashIssued, false)

		case strconv.Itoa(metadata.IssuingPRVERC20RequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertPRVEVMTxHashIssued, true)

		case strconv.Itoa(metadata.IssuingPRVBEP20RequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertPRVEVMTxHashIssued, true)

		case strconv.Itoa(metadata.IssuingPLGRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertPLGTxHashIssued, false)

		case strconv.Itoa(metadata.IssuingFantomRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertFTMTxHashIssued, false)

		case strconv.Itoa(metadata.IssuingNearRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertNEARTxHashIssued, false)

		case strconv.Itoa(metadata.IssuingAuroraRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertAURORATxHashIssued, false)

		case strconv.Itoa(metadata.IssuingAvaxRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertAVAXTxHashIssued, false)

		case strconv.Itoa(metadata.IssuingRequestMeta):
			updatingInfoByTokenID, err = blockchain.processIssuingReq(curView.featureStateDB, inst, updatingInfoByTokenID)

		case strconv.Itoa(metadata.ContractingRequestMeta):
			updatingInfoByTokenID, err = blockchain.processContractingReq(curView.featureStateDB, inst, updatingInfoByTokenID)

		case strconv.Itoa(metadata.BurningConfirmMeta), strconv.Itoa(metadata.BurningConfirmForDepositToSCMeta), strconv.Itoa(metadata.BurningConfirmMetaV2), strconv.Itoa(metadata.BurningConfirmForDepositToSCMetaV2):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, "", bridgeAggUnshieldTxIDs)

		case strconv.Itoa(metadata.BurningBSCConfirmMeta), strconv.Itoa(metadata.BurningPBSCConfirmForDepositToSCMeta):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, common.BSCPrefix, bridgeAggUnshieldTxIDs)

		case strconv.Itoa(metadata.BurningPLGConfirmMeta), strconv.Itoa(metadata.BurningPLGConfirmForDepositToSCMeta):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, common.PLGPrefix, bridgeAggUnshieldTxIDs)

		case strconv.Itoa(metadata.BurningFantomConfirmMeta), strconv.Itoa(metadata.BurningFantomConfirmForDepositToSCMeta):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, common.FTMPrefix, bridgeAggUnshieldTxIDs)

		case strconv.Itoa(metadata.BurningAuroraConfirmMeta), strconv.Itoa(metadata.BurningAuroraConfirmForDepositToSCMeta):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, common.AURORAPrefix, bridgeAggUnshieldTxIDs)

		case strconv.Itoa(metadata.BurningAvaxConfirmMeta), strconv.Itoa(metadata.BurningAvaxConfirmForDepositToSCMeta):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, common.AVAXPrefix, bridgeAggUnshieldTxIDs)

		case strconv.Itoa(metadata.BurningNearConfirmMeta):
			updatingInfoByTokenID, err = blockchain.processBurningReq(curView, inst, updatingInfoByTokenID, common.NEARPrefix, bridgeAggUnshieldTxIDs)
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
			curView.featureStateDB,
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

func (blockchain *BlockChain) processIssuingBridgeReq(curView *BeaconBestState, instruction []string, updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo, insertEVMTxHashIssued func(*statedb.StateDB, []byte) error, isPRV bool) (map[common.Hash]metadata.UpdatingInfo, error) {
	if len(instruction) != 4 {
		return updatingInfoByTokenID, nil // skip the instruction
	}
	if instruction[2] == "rejected" {
		txReqID, err := common.Hash{}.NewHashFromStr(instruction[3])
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while building tx request id in bytes from string: ", err)
			return updatingInfoByTokenID, nil
		}
		err = statedb.TrackBridgeReqWithStatus(curView.featureStateDB, *txReqID, common.BridgeRequestRejectedStatus)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while tracking bridge request with rejected status to leveldb: ", err)
		}
		return updatingInfoByTokenID, nil
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while decoding content string of accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	var issuingEVMAcceptedInst metadataBridge.IssuingEVMAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingEVMAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}

	err = insertEVMTxHashIssued(curView.featureStateDB, issuingEVMAcceptedInst.UniqTx)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while inserting EVM tx hash issued to leveldb: ", err)
		return updatingInfoByTokenID, nil
	}

	if !isPRV {
		updatingInfo, found := updatingInfoByTokenID[issuingEVMAcceptedInst.IncTokenID]
		if found {
			updatingInfo.CountUpAmt += issuingEVMAcceptedInst.IssuingAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:      issuingEVMAcceptedInst.IssuingAmount,
				DeductAmt:       0,
				TokenID:         issuingEVMAcceptedInst.IncTokenID,
				ExternalTokenID: issuingEVMAcceptedInst.ExternalTokenID,
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[issuingEVMAcceptedInst.IncTokenID] = updatingInfo
	}

	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) processIssuingReq(bridgeStateDB *statedb.StateDB, instruction []string, updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo) (map[common.Hash]metadata.UpdatingInfo, error) {
	if len(instruction) != 4 {
		return updatingInfoByTokenID, nil // skip the instruction
	}

	if instruction[2] == "rejected" {
		txReqID, err := common.Hash{}.NewHashFromStr(instruction[3])
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while building tx request id in bytes from string: ", err)
			return updatingInfoByTokenID, nil
		}
		err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, *txReqID, common.BridgeRequestRejectedStatus)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while tracking bridge request with rejected status to leveldb: ", err)
		}
		return updatingInfoByTokenID, nil
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while decoding content string of accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while unmarshaling accepted issuance instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	updatingInfo, found := updatingInfoByTokenID[issuingAcceptedInst.IncTokenID]
	if found {
		updatingInfo.CountUpAmt += issuingAcceptedInst.DepositedAmount
	} else {
		updatingInfo = metadata.UpdatingInfo{
			CountUpAmt:    issuingAcceptedInst.DepositedAmount,
			DeductAmt:     0,
			TokenID:       issuingAcceptedInst.IncTokenID,
			IsCentralized: true,
		}
	}
	updatingInfoByTokenID[issuingAcceptedInst.IncTokenID] = updatingInfo
	tmpBytes, _ := json.Marshal(updatingInfo)
	Logger.log.Infof("updatingIssuedInfo[%v]: %v\n", issuingAcceptedInst.IncTokenID.String(), string(tmpBytes))
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) processContractingReq(
	bridgeStateDB *statedb.StateDB,
	instruction []string,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) (map[common.Hash]metadata.UpdatingInfo, error) {
	if len(instruction) != 4 {
		return updatingInfoByTokenID, nil // skip the instruction
	}
	if instruction[2] == "rejected" {
		return updatingInfoByTokenID, nil // skip the instruction
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while decoding content string of accepted contracting instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	var contractingReqAction metadata.ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while unmarshaling accepted contracting instruction: ", err)
		return updatingInfoByTokenID, nil
	}
	md := contractingReqAction.Meta

	bridgeTokenExisted, err := statedb.IsBridgeTokenExistedByType(bridgeStateDB, md.TokenID, true)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while checking whether token (%s) existed in centralized bridge token list: %+v", md.TokenID.String(), err)
		return updatingInfoByTokenID, nil
	}
	if !bridgeTokenExisted {
		Logger.log.Warnf("WARNING: token (%s) did not exist in centralized bridge token list (from tx: %s)", md.TokenID.String(), contractingReqAction.TxReqID.String())
		return updatingInfoByTokenID, nil
	}

	updatingInfo, found := updatingInfoByTokenID[md.TokenID]
	if found {
		updatingInfo.DeductAmt += md.BurnedAmount
	} else {
		updatingInfo = metadata.UpdatingInfo{
			CountUpAmt:    0,
			DeductAmt:     md.BurnedAmount,
			TokenID:       md.TokenID,
			IsCentralized: true,
		}
	}
	updatingInfoByTokenID[md.TokenID] = updatingInfo
	tmpBytes, _ := json.Marshal(updatingInfo)
	Logger.log.Infof("updatingContractInfo[%v]: %v\n", md.TokenID.String(), string(tmpBytes))
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) processBurningReq(
	curView *BeaconBestState,
	instruction []string,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
	prefix string,
	bridgeAggUnshieldTxIDs map[string]bool,
) (map[common.Hash]metadata.UpdatingInfo, error) {
	if len(instruction) < 8 {
		return updatingInfoByTokenID, nil // skip the instruction
	}
	txIDStr := instruction[5]
	if bridgeAggUnshieldTxIDs[txIDStr] {
		// confirm instruction from bridge agg => do nothing
		Logger.log.Debugf("Skip process bridge agg confirm instruction in the bridge process flow %v", txIDStr)
		return updatingInfoByTokenID, nil
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

	incTokenID := &common.Hash{}
	incTokenID, _ = (*incTokenID).NewHash(incTokenIDBytes)

	if bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), externalTokenID) {
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else {
		amount = amt.Uint64()
	}

	bridgeTokenExisted, err := statedb.IsBridgeTokenExistedByType(curView.featureStateDB, *incTokenID, false)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while checking whether token (%s) existed in decentralized bridge token list: %+v", incTokenID.String(), err)
		return updatingInfoByTokenID, nil
	}
	if !bridgeTokenExisted {
		Logger.log.Warnf("WARNING: token (%s) did not exist in decentralized bridge token list", incTokenID.String())
		return updatingInfoByTokenID, nil
	}

	updatingInfo, found := updatingInfoByTokenID[*incTokenID]
	if found {
		updatingInfo.DeductAmt += amount
	} else {
		updatingInfo = metadata.UpdatingInfo{
			CountUpAmt:      0,
			DeductAmt:       amount,
			TokenID:         *incTokenID,
			ExternalTokenID: externalTokenID,
			IsCentralized:   false,
		}
	}
	updatingInfoByTokenID[*incTokenID] = updatingInfo
	tmpBytes, _ := json.Marshal(updatingInfo)
	Logger.log.Infof("updatingBurnedInfo[%v]: %v\n", incTokenID.String(), string(tmpBytes))
	return updatingInfoByTokenID, nil
}

func (blockchain *BlockChain) storeBurningConfirm(stateDB *statedb.StateDB, instructions [][]string, blockHeight uint64, metas []string) error {
	for _, inst := range instructions {
		found := false
		for _, meta := range metas {
			if inst[0] == meta {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		BLogger.log.Infof("storeBurningConfirm for block %d, inst %v, meta type %v", blockHeight, inst, inst[0])

		txID, err := bridgeagg.GetTxIDFromBurningConfirmInst(inst)
		if err != nil {
			return errors.Wrap(err, "txid invalid")
		}

		if err := statedb.StoreBurningConfirm(stateDB, *txID, blockHeight); err != nil {
			return errors.Wrapf(err, "store failed, txID: %x", txID)
		}
	}
	return nil
}

func (blockchain *BlockChain) updateBridgeIssuanceStatus(bridgeStateDB *statedb.StateDB, block *types.ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		metaType := tx.GetMetadataType()
		var err error
		var reqTxID common.Hash
		if metaType == metadata.IssuingETHRequestMeta || metaType == metadata.IssuingRequestMeta ||
			metaType == metadata.IssuingBSCRequestMeta || metaType == metadata.IssuingPRVERC20RequestMeta ||
			metaType == metadata.IssuingPRVBEP20RequestMeta || metaType == metadata.IssuingPLGRequestMeta ||
			metaType == metadata.IssuingFantomRequestMeta || metaType == metadataCommon.IssuingUnifiedTokenRequestMeta ||
			metaType == metadata.IssuingAuroraRequestMeta || metaType == metadata.IssuingAvaxRequestMeta ||
			metaType == metadata.IssuingNearRequestMeta {
			reqTxID = *tx.Hash()
			err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, reqTxID, common.BridgeRequestProcessingStatus)
			if err != nil {
				return err
			}
		}
		if metaType == metadata.IssuingETHResponseMeta || metaType == metadata.IssuingBSCResponseMeta ||
			metaType == metadata.IssuingPRVERC20ResponseMeta || metaType == metadata.IssuingPRVBEP20ResponseMeta ||
			metaType == metadata.IssuingPLGResponseMeta || metaType == metadata.IssuingFantomResponseMeta ||
			metaType == metadataCommon.IssuingUnifiedTokenResponseMeta || metaType == metadata.IssuingAuroraResponseMeta ||
			metaType == metadata.IssuingAvaxResponseMeta || metaType == metadata.IssuingNearResponseMeta {
			if metaType == metadataCommon.IssuingUnifiedTokenResponseMeta {
				meta := tx.GetMetadata().(*metadataBridge.ShieldResponse)
				reqTxID = meta.RequestedTxID
			} else {
				meta := tx.GetMetadata().(*metadataBridge.IssuingEVMResponse)
				reqTxID = meta.RequestedTxID
			}
			err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, reqTxID, common.BridgeRequestAcceptedStatus)
			if err != nil {
				return err
			}
		} else if metaType == metadata.IssuingResponseMeta {
			meta := tx.GetMetadata().(*metadata.IssuingResponse)
			reqTxID = meta.RequestedTxID
			err = statedb.TrackBridgeReqWithStatus(bridgeStateDB, reqTxID, common.BridgeRequestAcceptedStatus)
			if err != nil {
				return err
			}
		} else if metaType == metadataCommon.IssuingReshieldResponseMeta {
			meta := tx.GetMetadata().(*metadataBridge.IssuingReshieldResponse)
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
