package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/tendermint/tendermint/types"
	"sort"
	"strconv"
)

//todo: process instruction btc header relaying
func (blockchain *BlockChain) processRelayingInstructions(block *BeaconBlock, bd *[]database.BatchData) error {
	beaconHeight := block.Header.Height - 1
	db := blockchain.GetDatabase()

	relayingState, err := InitRelayingHeaderChainStateFromDB(db, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	// collect instruction RelayingBNBHeader
	// sort by block height
	// store header chain
	// update relaying state
	instsGroupByBlockHeight := make(map[uint64][][]string)
	blockHeightArr := make([]uint64, 0)
	for _, inst := range block.Body.Instructions {
		if len(inst) < 4 || inst[0] != strconv.Itoa(metadata.RelayingBNBHeaderMeta) {
			continue // Not error, just not relaying instruction
		}

		var err error
		var relayingContent metadata.RelayingBNBHeaderContent
		err = json.Unmarshal([]byte(inst[3]), &relayingContent)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling relaying header instruction: %+v", err)
			return nil
		}

		// get blockHeight in content
		blockHeight := relayingContent.BlockHeight

		// add to blockHeightArr
		if isExist, _ := common.SliceExists(blockHeightArr, blockHeight); !isExist {
			blockHeightArr = append(blockHeightArr, blockHeight)
		}

		// add to actionsGroupByBlockHeight
		if instsGroupByBlockHeight[blockHeight] != nil {
			instsGroupByBlockHeight[blockHeight] = append(instsGroupByBlockHeight[blockHeight], inst)
		} else{
			instsGroupByBlockHeight[blockHeight] = [][]string{inst}
		}
	}

	// sort blockHeightArr
	sort.Slice(blockHeightArr, func(i, j int) bool {
		return blockHeightArr[i] < blockHeightArr[j]
	})

	// process each instruction
	for _, blockHeight := range blockHeightArr {
		for _, inst := range instsGroupByBlockHeight[blockHeight] {
			err = blockchain.processRelayingHeaderInst(beaconHeight, inst, relayingState)
			if err != nil {
				Logger.log.Error(err)
				return nil
			}
		}
	}

	// store updated relayingState to leveldb with new beacon height
	err = storeRelayingHeaderStateToDB(db, beaconHeight+1, relayingState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}


func (blockchain *BlockChain) processRelayingHeaderInst(
	beaconHeight uint64, instructions []string, relayingState *RelayingHeaderChainState) error {
	if relayingState == nil {
		Logger.log.Errorf("relaying header state is nil")
		return nil
	}
	if len(instructions) !=  4 {
		return nil  // skip the instruction
	}
	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.RelayingBNBHeaderContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	var header lvdb.BNBHeader
	headerBytes, err := base64.StdEncoding.DecodeString(actionData.Header)
	if err != nil {
		return err
	}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return err
	}

	reqStatus := instructions[2]
	if reqStatus == common.RelayingHeaderUnconfirmedAcceptedChainStatus {
		//update relaying state
		relayingState.BNBHeaderChain.UnconfirmedHeaders = append(relayingState.BNBHeaderChain.UnconfirmedHeaders, header.Header)
		// track into DB ?
		//custodianDepositTrackKey := lvdb.NewCustodianDepositKey(actionData.TxReqID.String())
		//custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
		//	Status: common.PortalCustodianDepositAcceptedStatus,
		//	IncogAddressStr: actionData.IncogAddressStr,
		//	DepositedAmount : actionData.DepositedAmount,
		//}
		//
		//custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)
		//err = db.TrackCustodianDepositCollateral(
		//	[]byte(custodianDepositTrackKey),
		//	custodianDepositDataBytes,
		//)
		//if err != nil {
		//	Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
		//	return nil
		//}
	} else if reqStatus == common.RelayingHeaderConfirmedAcceptedChainStatus {
		// get new latest header
		blockIDNewLatestHeader := header.Header.LastBlockID
		//todo:
		for _, header := range relayingState.BNBHeaderChain.UnconfirmedHeaders {
			if bytes.Equal(header.Hash().Bytes(), blockIDNewLatestHeader.Hash) {
				relayingState.BNBHeaderChain.LatestHeader = header
				break
			}
		}

		//update relaying state
		relayingState.BNBHeaderChain.UnconfirmedHeaders = []*types.Header{header.Header}

		// store new confirmed header into db
		newConfirmedheader := relayingState.BNBHeaderChain.LatestHeader
		newConfirmedheaderBytes, _ := json.Marshal(newConfirmedheader)

		err := db.StoreRelayingBNBHeaderChain(actionData.BlockHeight, newConfirmedheaderBytes)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while storing new confirmed header: %+v", err)
			return nil
		}

		//// track custodian deposit into DB
		//custodianDepositTrackKey := lvdb.NewCustodianDepositKey(actionData.TxReqID.String())
		//custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
		//	Status: common.PortalCustodianDepositRefundStatus,
		//	IncogAddressStr: actionData.IncogAddressStr,
		//	DepositedAmount : actionData.DepositedAmount,
		//}
		//
		//custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)
		//err = db.TrackCustodianDepositCollateral(
		//	[]byte(custodianDepositTrackKey),
		//	custodianDepositDataBytes,
		//)
		//if err != nil {
		//	Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
		//	return nil
		//}
	} else if reqStatus == common.RelayingHeaderRejectedChainStatus {
		// track into DB ?
	}

	return nil
}