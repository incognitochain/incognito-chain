package portalprocess

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	"strconv"
)

/* =======
Portal Unshield Request Batching Processor
======= */
type PortalUnshieldBatchingProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalUnshieldBatchingProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalUnshieldBatchingProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalUnshieldBatchingProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildUnshieldBatchingInst(
	batchID string,
	rawExtTx string,
	tokenID string,
	unshieldIDs []string,
	utxos map[string][]*statedb.UTXO,
	networkFee map[uint64]uint,
	metaType int,
	status string,
) []string {
	unshieldBatchContent := metadata.PortalUnshieldRequestBatchContent{
		BatchID:       batchID,
		RawExternalTx: rawExtTx,
		TokenID:       tokenID,
		UnshieldIDs:   unshieldIDs,
		UTXOs:         utxos,
		NetworkFee:    networkFee,
	}
	unshieldBatchContentBytes, _ := json.Marshal(unshieldBatchContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(-1),
		status,
		string(unshieldBatchContentBytes),
	}
}

// batchID is hash of current beacon height and unshieldIDs that processed
func GetBatchID(beaconHeight uint64, unshieldIDs []string) string {
	dataBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	for _, id := range unshieldIDs {
		dataBytes = append(dataBytes, []byte(id)...)
	}
	dataHash := common.HashH(dataBytes)
	return dataHash.String()
}

func (p *PortalUnshieldBatchingProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	CurrentPortalStateV4 *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	if CurrentPortalStateV4 == nil {
		Logger.log.Warn("WARN - [Batch Unshield Request]: Current Portal state V4 is null.")
		return [][]string{}, nil
	}

	newInsts := [][]string{}
	wUnshieldRequests := CurrentPortalStateV4.WaitingUnshieldRequests
	for tokenID, wReqs := range wUnshieldRequests {
		portalTokenProcessor := portalParams.PortalTokens[tokenID]
		if portalTokenProcessor == nil {
			Logger.log.Errorf("[Batch Unshield Request]: Portal token ID %v is null.", tokenID)
			continue
		}

		// use default unshield fee in nano ptoken
		feeUnshield := portalParams.DefaultFeeUnshields[tokenID]

		// only process for waiting unshield request has enough minimum confirmation incognito blocks (avoid fork beacon chain)
		wReqForProcess := map[string]*statedb.WaitingUnshieldRequest{}
		for key, wR := range wReqs {
			if wR.GetBeaconHeight() + uint64(portalParams.MinConfirmationIncBlockNum) > beaconHeight {
				continue
			}
			wReqForProcess[key] = statedb.NewWaitingUnshieldRequestStateWithValue(
				wR.GetRemoteAddress(),
				wR.GetAmount(),
				wR.GetUnshieldID(),
				wR.GetBeaconHeight())
		}

		// choose waiting unshield IDs to process with current UTXOs
		utxos := CurrentPortalStateV4.UTXOs[tokenID]
		broadCastTxs := portalTokenProcessor.ChooseUnshieldIDsFromCandidates(utxos, wReqForProcess)

		// create raw external txs
		for _, bcTx := range broadCastTxs {
			// prepare outputs for tx in pbtc amount (haven't paid network fee)
			outputTxs := []*portaltokens.OutputTx{}
			for _, chosenUnshieldID := range bcTx.UnshieldIDs {
				keyWaitingUnshieldRequest := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, chosenUnshieldID).String()
				wUnshieldReq := wUnshieldRequests[tokenID][keyWaitingUnshieldRequest]
				outputTxs = append(outputTxs, &portaltokens.OutputTx{
					ReceiverAddress: wUnshieldReq.GetRemoteAddress(),
					Amount:          wUnshieldReq.GetAmount(),
				})
			}

			// memo in tx: batchId: combine beacon height and list of unshieldIDs
			batchID := GetBatchID(beaconHeight+1, bcTx.UnshieldIDs)
			memo := batchID

			// create raw tx
			hexRawExtTxStr, _, err := portalTokenProcessor.CreateRawExternalTx(
				bcTx.UTXOs, outputTxs, feeUnshield, memo, bc)
			if err != nil {
				Logger.log.Errorf("[Batch Unshield Request]: Error when creating raw external tx %v", err)
				continue
			}

			externalFees := map[uint64]uint{
				beaconHeight + 1: uint(feeUnshield),
			}
			chosenUTXOs := map[string][]*statedb.UTXO{
				portalParams.MultiSigAddresses[tokenID]: bcTx.UTXOs,
			}
			// update current portal state
			// remove chosen waiting unshield requests from waiting list
			// remove utxos
			UpdatePortalStateAfterProcessBatchUnshieldRequest(
				CurrentPortalStateV4, batchID, chosenUTXOs, externalFees, bcTx.UnshieldIDs, tokenID)

			// build new instruction with new raw external tx
			newInst := buildUnshieldBatchingInst(
				batchID, hexRawExtTxStr, tokenID, bcTx.UnshieldIDs, chosenUTXOs, externalFees,
				metadata.PortalV4UnshieldBatchingMeta, portalcommonv4.PortalV4RequestAcceptedChainStatus)
			newInsts = append(newInsts, newInst)
		}
	}
	return newInsts, nil
}

func (p *PortalUnshieldBatchingProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	CurrentPortalStateV4 *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if CurrentPortalStateV4 == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalUnshieldRequestBatchContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		// add new processed batch unshield request to batch unshield list
		// remove waiting unshield request from waiting list
		UpdatePortalStateAfterProcessBatchUnshieldRequest(
			CurrentPortalStateV4, actionData.BatchID, actionData.UTXOs, actionData.NetworkFee, actionData.UnshieldIDs, actionData.TokenID)
		RemoveListUtxoFromDB(stateDB, actionData.UTXOs, actionData.TokenID)
		RemoveListWaitingUnshieldFromDB(stateDB, actionData.UnshieldIDs, actionData.TokenID)

		for _, unshieldID := range actionData.UnshieldIDs {
			// update status of unshield request that processed
			err := UpdateNewStatusUnshieldRequest(unshieldID, portalcommonv4.PortalUnshieldReqProcessedStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("[processPortalBatchUnshieldRequest] Error when updating status of unshielding request with unshieldID %v: %v\n", unshieldID, err)
				return nil
			}
		}

		// store status of batch unshield by batchID
		batchUnshieldRequestStatus := metadata.PortalUnshieldRequestBatchStatus{
			BatchID:       actionData.BatchID,
			RawExternalTx: actionData.RawExternalTx,
			BeaconHeight:  beaconHeight + 1,
			TokenID:       actionData.TokenID,
			UnshieldIDs:   actionData.UnshieldIDs,
			UTXOs:         actionData.UTXOs,
			NetworkFee:    actionData.NetworkFee,
			Status:        portalcommonv4.PortalBatchUnshieldReqProcessedStatus,
		}
		batchUnshieldRequestStatusBytes, _ := json.Marshal(batchUnshieldRequestStatus)
		err := statedb.StorePortalBatchUnshieldRequestStatus(
			stateDB,
			actionData.BatchID,
			batchUnshieldRequestStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalBatchUnshieldRequest] Error when storing status of redeem request by redeemID: %v\n", err)
			return nil
		}
	}

	return nil
}