package portalprocess

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/portal/portalv4"

	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type CurrentPortalStateV4 struct {
	UTXOs                     map[string]map[string]*statedb.UTXO                          // tokenID : hash(tokenID || walletAddress || txHash || index) : value
	ShieldingExternalTx       map[string]map[string]*statedb.ShieldingRequest              // tokenID : hash(tokenID || proofHash) : value
	WaitingUnshieldRequests   map[string]map[string]*statedb.WaitingUnshieldRequest        // tokenID : hash(tokenID || unshieldID) : value
	ProcessedUnshieldRequests map[string]map[string]*statedb.ProcessedUnshieldRequestBatch // tokenID : hash(tokenID || batchID) : value

	DeletedUTXOKeyHashes                 []common.Hash
	DeletedWaitingUnshieldReqKeyHashes   []common.Hash
	DeletedProcessedUnshieldReqKeyHashes []common.Hash
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
	portalParamV4 portalv4.PortalParams,
) (*CurrentPortalStateV4, error) {
	var err error
	if lastState != nil {
		// reset temporary states
		lastState.DeletedUTXOKeyHashes = []common.Hash{}
		lastState.DeletedWaitingUnshieldReqKeyHashes = []common.Hash{}
		lastState.DeletedProcessedUnshieldReqKeyHashes = []common.Hash{}
		return lastState, nil
	}

	// load list of UTXOs
	utxos := map[string]map[string]*statedb.UTXO{}
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
		utxos[tokenID], err = statedb.GetUTXOsByTokenID(stateDB, tokenID)
		if err != nil {
			return nil, err
		}
	}

	// load list of waiting unshielding requests
	waitingUnshieldRequests := map[string]map[string]*statedb.WaitingUnshieldRequest{}
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
		waitingUnshieldRequests[tokenID], err = statedb.GetWaitingUnshieldRequestsByTokenID(stateDB, tokenID)
		if err != nil {
			return nil, err
		}
	}

	// load list of processed unshielding requests batch
	processedUnshieldRequestsBatch := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
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

		DeletedUTXOKeyHashes:                 []common.Hash{},
		DeletedWaitingUnshieldReqKeyHashes:   []common.Hash{},
		DeletedProcessedUnshieldReqKeyHashes: []common.Hash{},
	}, nil
}

func StorePortalV4StateToDB(
	stateDB *statedb.StateDB,
	currentPortalState *CurrentPortalStateV4,
	portalParamV4 portalv4.PortalParams,
) error {
	var err error
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
		err = statedb.StoreUTXOs(stateDB, currentPortalState.UTXOs[tokenID])
		if err != nil {
			return err
		}
	}
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
		err = statedb.StoreShieldingRequests(stateDB, currentPortalState.ShieldingExternalTx[tokenID])
		if err != nil {
			return err
		}
	}
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
		err = statedb.StoreWaitingUnshieldRequests(stateDB, currentPortalState.WaitingUnshieldRequests[tokenID])
		if err != nil {
			return err
		}
	}
	for _, tokenID := range portalParamV4.PortalV4TokenIDs {
		err = statedb.StoreProcessedBatchUnshieldRequests(stateDB, currentPortalState.ProcessedUnshieldRequests[tokenID])
		if err != nil {
			return err
		}
	}

	err = statedb.DeleteUTXOs(stateDB, currentPortalState.DeletedUTXOKeyHashes)
	if err != nil {
		return err
	}
	err = statedb.DeleteWaitingUnshieldRequests(stateDB, currentPortalState.DeletedWaitingUnshieldReqKeyHashes)
	if err != nil {
		return err
	}
	err = statedb.DeletePortalBatchUnshieldRequests(stateDB, currentPortalState.DeletedProcessedUnshieldReqKeyHashes)
	if err != nil {
		return err
	}
	return nil
}

func (s *CurrentPortalStateV4) AddUTXOs(utxos []*statedb.UTXO, tokenID string) {
	if s.UTXOs == nil {
		s.UTXOs = map[string]map[string]*statedb.UTXO{}
	}
	if s.UTXOs[tokenID] == nil {
		s.UTXOs[tokenID] = map[string]*statedb.UTXO{}
	}

	var walletAddress string
	var txHash string
	var outputIdx uint32
	var outputAmount uint64
	var chainCodeSeed string
	var utxoKeyStr string
	for _, utxo := range utxos {
		walletAddress = utxo.GetWalletAddress()
		txHash = utxo.GetTxHash()
		outputIdx = utxo.GetOutputIndex()
		outputAmount = utxo.GetOutputAmount()
		chainCodeSeed = utxo.GetChainCodeSeed()
		utxoKeyStr = statedb.GenerateUTXOObjectKey(tokenID, walletAddress, txHash, outputIdx).String()
		s.UTXOs[tokenID][utxoKeyStr] = statedb.NewUTXOWithValue(walletAddress, txHash, outputIdx, outputAmount, chainCodeSeed)
	}
}

func (s *CurrentPortalStateV4) AddShieldingExternalTx(tokenID string, shieldingProofTxHash string,
	shieldingExternalTxHash string, incAddress string, amount uint64) {
	if s.ShieldingExternalTx == nil {
		s.ShieldingExternalTx = map[string]map[string]*statedb.ShieldingRequest{}
	}
	if s.ShieldingExternalTx[tokenID] == nil {
		s.ShieldingExternalTx[tokenID] = map[string]*statedb.ShieldingRequest{}
	}
	shieldKeyStr := statedb.GenerateShieldingRequestObjectKey(tokenID, shieldingProofTxHash).String()
	s.ShieldingExternalTx[tokenID][shieldKeyStr] = statedb.NewShieldingRequestWithValue(shieldingExternalTxHash, incAddress, amount)
}

func (s *CurrentPortalStateV4) IsExistedShieldingExternalTx(tokenID string, shieldingProofTxHash string) bool {
	if s.ShieldingExternalTx == nil || s.ShieldingExternalTx[tokenID] == nil {
		return false
	}
	shieldKeyStr := statedb.GenerateShieldingRequestObjectKey(tokenID, shieldingProofTxHash).String()
	_, isExisted := s.ShieldingExternalTx[tokenID][shieldKeyStr]
	return isExisted
}

func (s *CurrentPortalStateV4) AddWaitingUnshieldRequest(
	unshieldID string, tokenID string, remoteAddress string, unshieldAmt uint64, beaconHeight uint64) {
	if s.WaitingUnshieldRequests == nil {
		s.WaitingUnshieldRequests = map[string]map[string]*statedb.WaitingUnshieldRequest{}
	}
	if s.WaitingUnshieldRequests[tokenID] == nil {
		s.WaitingUnshieldRequests[tokenID] = map[string]*statedb.WaitingUnshieldRequest{}
	}

	keyWaitingUnshieldRequest := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, unshieldID).String()
	waitingUnshieldRequest := statedb.NewWaitingUnshieldRequestStateWithValue(remoteAddress, unshieldAmt, unshieldID, beaconHeight)
	s.WaitingUnshieldRequests[tokenID][keyWaitingUnshieldRequest] = waitingUnshieldRequest
}

func (s *CurrentPortalStateV4) AddBatchProcessedUnshieldRequest(
	batchID string, utxos []*statedb.UTXO, beaconHeight uint64, externalFee uint, unshieldIDs []string,
	tokenID string) {
	if s.ProcessedUnshieldRequests == nil {
		s.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}
	if s.ProcessedUnshieldRequests[tokenID] == nil {
		s.ProcessedUnshieldRequests[tokenID] = map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}

	externalFees := map[uint64]statedb.ExternalFeeInfo{
		beaconHeight: {
			NetworkFee:    externalFee,
			RBFReqIncTxID: "",
		},
	}

	keyProcessedUnshieldRequest := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenID, batchID).String()
	s.ProcessedUnshieldRequests[tokenID][keyProcessedUnshieldRequest] = statedb.NewProcessedUnshieldRequestBatchWithValue(
		batchID, unshieldIDs, utxos, externalFees)
}

func (s *CurrentPortalStateV4) UpdatePortalStateAfterProcessBatchUnshieldRequest(
	batchID string, utxos []*statedb.UTXO, beaconHeight uint64, externalFee uint, unshieldIDs []string,
	tokenID string) {
	// remove unshieldIDs from WaitingUnshieldRequests
	s.RemoveWaitingUnshieldReqs(unshieldIDs, tokenID)

	// remove list utxos from state
	s.RemoveUTXOs(utxos, tokenID)

	// add batch process to ProcessedUnshieldRequests
	s.AddBatchProcessedUnshieldRequest(batchID, utxos, beaconHeight, externalFee, unshieldIDs, tokenID)
}

func (s *CurrentPortalStateV4) RemoveUTXOs(utxos []*statedb.UTXO, tokenID string) {
	// remove list utxos that spent
	for _, u := range utxos {
		utxoKeyHash := statedb.GenerateUTXOObjectKey(tokenID, u.GetWalletAddress(), u.GetTxHash(), u.GetOutputIndex())
		delete(s.UTXOs[tokenID], utxoKeyHash.String())
		s.DeletedUTXOKeyHashes = append(s.DeletedUTXOKeyHashes, utxoKeyHash)
	}
}

func (s *CurrentPortalStateV4) RemoveWaitingUnshieldReqs(unshieldIDs []string, tokenID string) {
	// remove unshieldIDs from WaitingUnshieldRequests
	for _, unshieldID := range unshieldIDs {
		keyWaitingUnshieldRequest := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, unshieldID)
		delete(s.WaitingUnshieldRequests[tokenID], keyWaitingUnshieldRequest.String())
		s.DeletedWaitingUnshieldReqKeyHashes = append(s.DeletedWaitingUnshieldReqKeyHashes, keyWaitingUnshieldRequest)
	}
}

func (s *CurrentPortalStateV4) AddExternalFeeForBatchProcessedUnshieldRequest(
	batchID string, tokenID string, externalFee uint, beaconHeight uint64, rbfReqIncTxID string) {
	if s.ProcessedUnshieldRequests == nil {
		s.ProcessedUnshieldRequests = map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}
	if s.ProcessedUnshieldRequests[tokenID] == nil {
		s.ProcessedUnshieldRequests[tokenID] = map[string]*statedb.ProcessedUnshieldRequestBatch{}
	}

	keyProcessedUnshieldRequest := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenID, batchID).String()
	externalFees := s.ProcessedUnshieldRequests[tokenID][keyProcessedUnshieldRequest].GetExternalFees()
	if externalFees == nil {
		externalFees = map[uint64]statedb.ExternalFeeInfo{}
	}
	externalFees[beaconHeight] = statedb.ExternalFeeInfo{
		NetworkFee:    externalFee,
		RBFReqIncTxID: rbfReqIncTxID,
	}
	s.ProcessedUnshieldRequests[tokenID][keyProcessedUnshieldRequest].SetExternalFees(externalFees)
}

func (s *CurrentPortalStateV4) RemoveBatchProcessedUnshieldRequest(tokenIDStr string, batchKey common.Hash) {
	delete(s.ProcessedUnshieldRequests[tokenIDStr], batchKey.String())
	s.DeletedProcessedUnshieldReqKeyHashes = append(s.DeletedProcessedUnshieldReqKeyHashes, batchKey)
}

// get latest beaconheight
func GetMaxKeyValue(input map[uint64]statedb.ExternalFeeInfo) (max uint64) {
	max = 0
	for k := range input {
		if k > max {
			max = k
		}
	}
	return max
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
		OTAPubKeyStr:   unshieldRequest.OTAPubKeyStr,
		TxRandomStr:    unshieldRequest.TxRandomStr,
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

func UpdateNewStatusBatchUnshield(
	batchID string, newStatus byte,
	newExternalFee map[uint64]metadata.ExternalFeeInfo, stateDB *statedb.StateDB) error {
	// get unshield request by unshield ID
	batchUnshieldStatusBytes, err := statedb.GetPortalBatchUnshieldRequestStatus(stateDB, batchID)
	if err != nil {
		return err
	}
	var batchUnshield metadata.PortalUnshieldRequestBatchStatus
	err = json.Unmarshal(batchUnshieldStatusBytes, &batchUnshield)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", batchUnshieldStatusBytes, err)
		return err
	}

	updateExternalFees := batchUnshield.NetworkFees
	for beaconHeight, externalFeeInfo := range newExternalFee {
		updateExternalFees[beaconHeight] = externalFeeInfo
	}

	// update new status and store to db
	unshieldRequestNewStatus := metadata.PortalUnshieldRequestBatchStatus{
		BatchID:       batchID,
		TokenID:       batchUnshield.TokenID,
		UnshieldIDs:   batchUnshield.UnshieldIDs,
		UTXOs:         batchUnshield.UTXOs,
		RawExternalTx: batchUnshield.RawExternalTx,
		NetworkFees:   updateExternalFees,
		BeaconHeight:  batchUnshield.BeaconHeight,
		Status:        newStatus,
	}
	unshieldRequestNewStatusBytes, _ := json.Marshal(unshieldRequestNewStatus)
	err = statedb.StorePortalBatchUnshieldRequestStatus(
		stateDB,
		batchID,
		unshieldRequestNewStatusBytes)
	if err != nil {
		return err
	}
	return nil
}
