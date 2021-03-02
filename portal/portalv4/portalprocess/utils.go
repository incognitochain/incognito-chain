package portalprocess

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
)

type CurrentPortalStateV4 struct {
	UTXOs                     map[string]map[string]*statedb.UTXO                          // tokenID : hash(tokenID || walletAddress || txHash || index) : value
	ShieldingExternalTx       map[string]map[string]*statedb.ShieldingRequest              // tokenID : hash(tokenID || proofHash) : value
	//WaitingUnshieldRequests   map[string]map[string]*statedb.WaitingUnshieldRequest        // tokenID : hash(tokenID || unshieldID) : value
	//ProcessedUnshieldRequests map[string]map[string]*statedb.ProcessedUnshieldRequestBatch // tokenID : hash(tokenID || batchID) : value
}

func InitCurrentPortalStateV4FromDB(
	stateDB *statedb.StateDB,
) (*CurrentPortalStateV4, error) {
	var err error

	// load list of UTXOs
	utxos := map[string]map[string]*statedb.UTXO{}
	for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
		utxos[tokenID], err = statedb.GetUTXOsByTokenID(stateDB, tokenID)
		if err != nil {
			return nil, err
		}
	}

	//// load list of waiting unshielding requests
	//waitingUnshieldRequests := map[string]map[string]*statedb.WaitingUnshieldRequest{}
	//for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
	//	waitingUnshieldRequests[tokenID], err = statedb.GetWaitingUnshieldRequestsByTokenID(stateDB, tokenID)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	// load list of processed unshielding requests batch
	//processedUnshieldRequestsBatch := map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{}
	//for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
	//	waitingUnshieldRequests[tokenID], err = statedb.GetWaitingUnshieldRequestsByTokenID(stateDB, tokenID)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	return &CurrentPortalStateV4{
		UTXOs:                     utxos,
		ShieldingExternalTx:       nil,
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
	//for _, tokenID := range portalcommonv4.PortalV4SupportedIncTokenIDs {
	//	err = statedb.StoreWaitingUnshieldRequests(stateDB, currentPortalState.WaitingUnshieldRequests[tokenID])
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}

func UpdatePortalStateUTXOs(currentPortalV4State *CurrentPortalStateV4, tokenID string, listUTXO []*statedb.UTXO) {
	if currentPortalV4State.UTXOs == nil {
		currentPortalV4State.UTXOs = map[string]map[string]*statedb.UTXO{}
	}
	if currentPortalV4State.UTXOs[tokenID] == nil {
		currentPortalV4State.UTXOs[tokenID] = map[string]*statedb.UTXO{}
	}
	for _, utxo := range listUTXO {
		walletAddress := utxo.GetWalletAddress()
		txHash := utxo.GetTxHash()
		outputIdx := utxo.GetOutputIndex()
		outputAmount := utxo.GetOutputAmount()
		currentPortalV4State.UTXOs[tokenID][statedb.GenerateUTXOObjectKey(tokenID, walletAddress, txHash, outputIdx).String()] = statedb.NewUTXOWithValue(walletAddress, txHash, outputIdx, outputAmount)
	}
}

func UpdatePortalStateShieldingExternalTx(currentPortalV4State *CurrentPortalStateV4, tokenID string, shieldingProofTxHash string, shieldingExternalTxHash string, incAddress string, amount uint64) {
	if currentPortalV4State.ShieldingExternalTx == nil {
		currentPortalV4State.ShieldingExternalTx = map[string]map[string]*statedb.ShieldingRequest{}
	}
	if currentPortalV4State.ShieldingExternalTx[tokenID] == nil {
		currentPortalV4State.ShieldingExternalTx[tokenID] = map[string]*statedb.ShieldingRequest{}
	}
	currentPortalV4State.ShieldingExternalTx[tokenID][statedb.GenerateShieldingRequestObjectKey(tokenID, shieldingProofTxHash).String()] = statedb.NewShieldingRequestWithValue(shieldingExternalTxHash, incAddress, amount)
}

func IsExistsProofInPortalState(currentPortalV4State *CurrentPortalStateV4, tokenID string, shieldingProofTxHash string) bool {
	if currentPortalV4State.ShieldingExternalTx == nil {
		return false
	}
	if currentPortalV4State.ShieldingExternalTx[tokenID] == nil {
		return false
	}
	_, exists := currentPortalV4State.ShieldingExternalTx[tokenID][statedb.GenerateShieldingRequestObjectKey(tokenID, shieldingProofTxHash).String()]
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