package portalprocess

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
)

type CurrentPortalStateV4 struct {
	UTXOs                     map[string]map[string]*statedb.UTXO                          // tokenID : hash(tokenID || walletAddress || txHash || index) : value
	ShieldingExternalTx       map[string]map[string]*statedb.ShieldingRequest              // tokenID : hash(tokenID || proofHash) : value
	WaitingUnshieldRequests   map[string]map[string]*statedb.WaitingUnshieldRequest        // tokenID : hash(tokenID || unshieldID) : value
	ProcessedUnshieldRequests map[string]map[string]*statedb.ProcessedUnshieldRequestBatch // tokenID : hash(tokenID || batchID) : value
}

func (s *CurrentPortalStateV4) Copy() *CurrentPortalStateV4 {
	v := new(CurrentPortalStateV4)
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(s)
	if err != nil {
		return nil
	}
	err = gob.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(v)
	if err != nil {
		return nil
	}
	return v
}

func InitCurrentPortalStateV4FromDB(
	stateDB *statedb.StateDB,
	lastState *CurrentPortalStateV4,
) (*CurrentPortalStateV4, error) {
	var err error
	if lastState != nil {
		return lastState, nil
	}

	// load list of UTXOs
	utxos := map[string]map[string]*statedb.UTXO{}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		utxos[tokenID], err = statedb.GetUTXOsByTokenID(stateDB, tokenID)
		if err != nil {
			return nil, err
		}
	}

	// load list of waiting unshielding requests
	waitingUnshieldRequests := map[string]map[string]*statedb.WaitingUnshieldRequest{}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		waitingUnshieldRequests[tokenID], err = statedb.GetWaitingUnshieldRequestsByTokenID(stateDB, tokenID)
		if err != nil {
			return nil, err
		}
	}

	// load list of processed unshielding requests batch
	processedUnshieldRequestsBatch := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		processedUnshieldRequestsBatch[tokenID], err = statedb.GetListProcessedBatchUnshieldRequestsByTokenID(stateDB, tokenID)
		if err != nil {
			return nil, err
		}
	}

	return &CurrentPortalStateV4{
		UTXOs:                     utxos,
		ShieldingExternalTx:       nil,
		WaitingUnshieldRequests:   waitingUnshieldRequests,
		ProcessedUnshieldRequests: processedUnshieldRequestsBatch,
	}, nil
}

func StorePortalV4StateToDB(
	stateDB *statedb.StateDB,
	currentPortalState *CurrentPortalStateV4,
) error {
	var err error
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		err = statedb.StoreUTXOs(stateDB, currentPortalState.UTXOs[tokenID])
		if err != nil {
			return err
		}
	}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		err = statedb.StoreShieldingRequests(stateDB, currentPortalState.ShieldingExternalTx[tokenID])
		if err != nil {
			return err
		}
	}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		err = statedb.StoreWaitingUnshieldRequests(stateDB, currentPortalState.WaitingUnshieldRequests[tokenID])
		if err != nil {
			return err
		}
	}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		err = statedb.StoreProcessedBatchUnshieldRequests(stateDB, currentPortalState.ProcessedUnshieldRequests[tokenID])
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdatePortalStateUTXOs(CurrentPortalStateV4 *CurrentPortalStateV4, tokenID string, listUTXO []*statedb.UTXO) {
	if CurrentPortalStateV4.UTXOs == nil {
		CurrentPortalStateV4.UTXOs = map[string]map[string]*statedb.UTXO{}
	}
	if CurrentPortalStateV4.UTXOs[tokenID] == nil {
		CurrentPortalStateV4.UTXOs[tokenID] = map[string]*statedb.UTXO{}
	}
	for _, utxo := range listUTXO {
		walletAddress := utxo.GetWalletAddress()
		txHash := utxo.GetTxHash()
		outputIdx := utxo.GetOutputIndex()
		outputAmount := utxo.GetOutputAmount()
		chainCodeSeed := utxo.GetChainCodeSeed()
		CurrentPortalStateV4.UTXOs[tokenID][statedb.GenerateUTXOObjectKey(tokenID, walletAddress, txHash, outputIdx).String()] = statedb.NewUTXOWithValue(walletAddress, txHash, outputIdx, outputAmount, chainCodeSeed)
	}
}

func UpdatePortalStateShieldingExternalTx(CurrentPortalStateV4 *CurrentPortalStateV4, tokenID string, shieldingProofTxHash string, shieldingExternalTxHash string, incAddress string, amount uint64) {
	if CurrentPortalStateV4.ShieldingExternalTx == nil {
		CurrentPortalStateV4.ShieldingExternalTx = map[string]map[string]*statedb.ShieldingRequest{}
	}
	if CurrentPortalStateV4.ShieldingExternalTx[tokenID] == nil {
		CurrentPortalStateV4.ShieldingExternalTx[tokenID] = map[string]*statedb.ShieldingRequest{}
	}
	CurrentPortalStateV4.ShieldingExternalTx[tokenID][statedb.GenerateShieldingRequestObjectKey(tokenID, shieldingProofTxHash).String()] = statedb.NewShieldingRequestWithValue(shieldingExternalTxHash, incAddress, amount)
}

func IsExistsProofInPortalState(CurrentPortalStateV4 *CurrentPortalStateV4, tokenID string, shieldingProofTxHash string) bool {
	if CurrentPortalStateV4.ShieldingExternalTx == nil {
		return false
	}
	if CurrentPortalStateV4.ShieldingExternalTx[tokenID] == nil {
		return false
	}
	_, exists := CurrentPortalStateV4.ShieldingExternalTx[tokenID][statedb.GenerateShieldingRequestObjectKey(tokenID, shieldingProofTxHash).String()]
	return exists
}

// get latest beaconheight
func GetMaxKeyValue(input map[uint64]uint) (max uint64) {
	max = 0
	for k := range input {
		if k > max {
			max = k
		}
	}
	return max
}

func UpdatePortalStateAfterUnshieldRequest(
	currentPortalStateV4 *CurrentPortalStateV4,
	unshieldID string, tokenID string, remoteAddress string, unshieldAmt uint64, beaconHeight uint64) {

	if currentPortalStateV4.WaitingUnshieldRequests == nil {
		currentPortalStateV4.WaitingUnshieldRequests = map[string]map[string]*statedb.WaitingUnshieldRequest{}
	}
	if currentPortalStateV4.WaitingUnshieldRequests[tokenID] == nil {
		currentPortalStateV4.WaitingUnshieldRequests[tokenID] = map[string]*statedb.WaitingUnshieldRequest{}
	}

	keyWaitingUnshieldRequest := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, unshieldID).String()
	waitingUnshieldRequest := statedb.NewWaitingUnshieldRequestStateWithValue(remoteAddress, unshieldAmt, unshieldID, beaconHeight)
	currentPortalStateV4.WaitingUnshieldRequests[tokenID][keyWaitingUnshieldRequest] = waitingUnshieldRequest
}

func UpdatePortalStateAfterProcessBatchUnshieldRequest(
	CurrentPortalStateV4 *CurrentPortalStateV4,
	batchID string, utxos []*statedb.UTXO, externalFees map[uint64]uint, unshieldIDs []string, tokenID string) {
	// remove unshieldIDs from WaitingUnshieldRequests
	RemoveListWaitingUnshieldFromState(CurrentPortalStateV4, unshieldIDs, tokenID)

	// remove list utxos from state
	RemoveListUtxoFromState(CurrentPortalStateV4, utxos, tokenID)

	// add batch process to ProcessedUnshieldRequests
	if CurrentPortalStateV4.ProcessedUnshieldRequests == nil {
		CurrentPortalStateV4.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}
	if CurrentPortalStateV4.ProcessedUnshieldRequests[tokenID] == nil {
		CurrentPortalStateV4.ProcessedUnshieldRequests[tokenID] = map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}

	keyProcessedUnshieldRequest := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenID, batchID).String()
	CurrentPortalStateV4.ProcessedUnshieldRequests[tokenID][keyProcessedUnshieldRequest] = statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, unshieldIDs, utxos, externalFees)
}

func RemoveListUtxoFromState(
	CurrentPortalStateV4 *CurrentPortalStateV4,
	utxos []*statedb.UTXO, tokenID string) {
	// remove list utxos that spent
	for _, u := range utxos {
		keyUtxo := statedb.GenerateUTXOObjectKey(tokenID, u.GetWalletAddress(), u.GetTxHash(), u.GetOutputIndex()).String()
		delete(CurrentPortalStateV4.UTXOs[tokenID], keyUtxo)
	}
}

func RemoveListUtxoFromDB(
	stateDB *statedb.StateDB,
	utxos []*statedb.UTXO, tokenID string) {
	// remove list utxos that spent
	for _, u := range utxos {
		statedb.DeleteUTXO(stateDB, tokenID, u.GetWalletAddress(), u.GetTxHash(), u.GetOutputIndex())
	}
}

func RemoveListWaitingUnshieldFromState(
	CurrentPortalStateV4 *CurrentPortalStateV4,
	unshieldIDs []string, tokenID string) {
	// remove unshieldIDs from WaitingUnshieldRequests
	for _, unshieldID := range unshieldIDs {
		keyWaitingUnshieldRequest := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, unshieldID).String()
		delete(CurrentPortalStateV4.WaitingUnshieldRequests[tokenID], keyWaitingUnshieldRequest)
	}
}

func RemoveListWaitingUnshieldFromDB(
	stateDB *statedb.StateDB,
	unshieldIDs []string, tokenID string) {
	// delete waiting unshield request from db
	for _, unshieldID := range unshieldIDs {
		statedb.DeleteWaitingUnshieldRequest(stateDB, tokenID, unshieldID)
	}
}

func UpdateNewStatusUnshieldRequest(unshieldID string, newStatus int, externalTxID string, externalFee uint64, stateDB *statedb.StateDB) error {
	// get unshield request by unshield ID
	unshieldRequestBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, unshieldID)
	if err != nil {
		return err
	}
	var unshieldRequest metadata.PortalUnshieldRequestStatus
	err = json.Unmarshal(unshieldRequestBytes, &unshieldRequest)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", unshieldRequestBytes, err)
		return err
	}
	newExternalTxID := unshieldRequest.ExternalTxID
	if externalTxID != "" {
		newExternalTxID = externalTxID
	}

	newExternalFee := unshieldRequest.ExternalFee
	if externalFee > 0 {
		newExternalFee = externalFee
	}

	// update new status and store to db
	unshieldRequestNewStatus := metadata.PortalUnshieldRequestStatus{
		IncAddressStr:  unshieldRequest.IncAddressStr,
		RemoteAddress:  unshieldRequest.RemoteAddress,
		TokenID:        unshieldRequest.TokenID,
		UnshieldAmount: unshieldRequest.UnshieldAmount,
		UnshieldID:     unshieldRequest.UnshieldID,
		ExternalTxID:   newExternalTxID,
		ExternalFee:    newExternalFee,
		Status:         newStatus,
	}
	unshieldRequestNewStatusBytes, _ := json.Marshal(unshieldRequestNewStatus)
	err = statedb.StorePortalUnshieldRequestStatus(
		stateDB,
		unshieldID,
		unshieldRequestNewStatusBytes)
	if err != nil {
		return err
	}
	return nil
}

func UpdatePortalStateAfterReplaceFeeRequest(
	currentPortalV4State *CurrentPortalStateV4, unshieldBatch *statedb.ProcessedUnshieldRequestBatch, beaconHeight uint64, fee uint, tokenIDStr, batchIDStr string) {
	if currentPortalV4State.ProcessedUnshieldRequests == nil {
		currentPortalV4State.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}
	if currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr] == nil {
		currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr] = map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}
	keyWaitingReplacementRequest := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenIDStr, batchIDStr).String()
	fees := unshieldBatch.GetExternalFees()
	fees[beaconHeight] = fee
	waitingReplacementRequest := statedb.NewProcessedUnshieldRequestBatchWithValue(unshieldBatch.GetBatchID(), unshieldBatch.GetUnshieldRequests(), unshieldBatch.GetUTXOs(), fees)
	currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr][keyWaitingReplacementRequest] = waitingReplacementRequest
}

func UpdatePortalStateAfterSubmitConfirmedTx(currentPortalV4State *CurrentPortalStateV4, tokenIDStr, batchKey string) {
	delete(currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr], batchKey)
}