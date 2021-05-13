package statedb

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

// ================= Shielding Request =================
func StoreShieldingRequestStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalShieldingRequestStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalV4Status(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalV4ShieldingRequestStatusError, err)
	}

	return nil
}

func GetShieldingRequestStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalShieldingRequestStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalV4Status(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalV4ShieldingRequestStatusError, err)
	}

	return data, nil
}

func IsExistsShieldingRequest(stateDB *StateDB, tokenID string, proofHash string) (bool, error) {
	keyStr := GenerateShieldingRequestObjectKey(tokenID, proofHash).String()
	key, err := common.Hash{}.NewHashFromStr(keyStr)
	if err != nil {
		return false, NewStatedbError(GetPortalShieldingRequestsError, err)
	}
	return stateDB.Exist(PortalV4ShieldRequestObjectType, *key)
}

func StoreShieldingRequests(stateDB *StateDB, shieldingRequests map[string]*ShieldingRequest) error {
	for keyStr, shieldingReq := range shieldingRequests {
		key, err := common.Hash{}.NewHashFromStr(keyStr)
		if err != nil {
			return NewStatedbError(StorePortalShieldingRequestsError, err)
		}
		err = stateDB.SetStateObject(PortalV4ShieldRequestObjectType, *key, shieldingReq)
		if err != nil {
			return NewStatedbError(StorePortalShieldingRequestsError, err)
		}
	}

	return nil
}

// ================= List UTXOs =================
func GetUTXOsByTokenID(stateDB *StateDB, tokenID string) (map[string]*UTXO, error) {
	return stateDB.getUTXOsByTokenID(tokenID), nil
}

func StoreUTXOs(stateDB *StateDB, utxos map[string]*UTXO) error {
	for keyStr, utxo := range utxos {
		key, err := common.Hash{}.NewHashFromStr(keyStr)
		if err != nil {
			return NewStatedbError(StorePortalV4UTXOsError, err)
		}
		err = stateDB.SetStateObject(PortalV4UTXOObjectType, *key, utxo)
		if err != nil {
			return NewStatedbError(StorePortalV4UTXOsError, err)
		}
	}

	return nil
}

func DeleteUTXOs(stateDB *StateDB, utxoKeys []common.Hash) error {
	for _, keyHash := range utxoKeys {
		stateDB.MarkDeleteStateObject(PortalV4UTXOObjectType, keyHash)
	}

	return nil
}

// ================= List Waiting Unshielding Requests =================
func GetWaitingUnshieldRequestsByTokenID(stateDB *StateDB, tokenID string) (map[string]*WaitingUnshieldRequest, error) {
	return stateDB.getListWaitingUnshieldRequestsByTokenID(tokenID), nil
}

func StoreWaitingUnshieldRequests(
	stateDB *StateDB,
	waitingUnshieldReqs map[string]*WaitingUnshieldRequest) error {
	for keyStr, waitingReq := range waitingUnshieldReqs {
		key, err := common.Hash{}.NewHashFromStr(keyStr)
		if err != nil {
			return NewStatedbError(StorePortalListWaitingUnshieldRequestError, err)
		}
		err = stateDB.SetStateObject(PortalWaitingUnshieldObjectType, *key, waitingReq)
		if err != nil {
			return NewStatedbError(StorePortalListWaitingUnshieldRequestError, err)
		}
	}

	return nil
}

func DeleteWaitingUnshieldRequests(stateDB *StateDB, wUnshieldReqKeys []common.Hash) error {
	for _, keyHash := range wUnshieldReqKeys {
		stateDB.MarkDeleteStateObject(PortalWaitingUnshieldObjectType, keyHash)
	}

	return nil
}

// ================= Unshielding Request Status =================
// Store and get the status of the Unshield Request by unshieldID
func StorePortalUnshieldRequestStatus(stateDB *StateDB, unshieldID string, statusContent []byte) error {
	statusType := PortalUnshieldRequestStatusPrefix()
	statusSuffix := []byte(unshieldID)
	err := StorePortalV4Status(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalUnshieldRequestStatusError, err)
	}

	return nil
}

func GetPortalUnshieldRequestStatus(stateDB *StateDB, unshieldID string) ([]byte, error) {
	statusType := PortalUnshieldRequestStatusPrefix()
	statusSuffix := []byte(unshieldID)
	data, err := GetPortalV4Status(stateDB, statusType, statusSuffix)
	if err != nil && err.(*StatedbError).GetErrorCode() != ErrCodeMessage[GetPortalStatusNotFoundError].Code {
		return []byte{}, NewStatedbError(GetPortalUnshieldRequestStatusError, err)
	}

	return data, nil
}

// ================= List Batching Unshielding Request =================
func GetListProcessedBatchUnshieldRequestsByTokenID(stateDB *StateDB, tokenID string) (map[string]*ProcessedUnshieldRequestBatch, error) {
	return stateDB.getListProcessedBatchUnshieldRequestsByTokenID(tokenID), nil
}

func StoreProcessedBatchUnshieldRequests(
	stateDB *StateDB,
	processedBatchUnshieldReqs map[string]*ProcessedUnshieldRequestBatch) error {
	for keyStr, batchReq := range processedBatchUnshieldReqs {
		key, err := common.Hash{}.NewHashFromStr(keyStr)
		if err != nil {
			return NewStatedbError(StorePortalListProcessedBatchUnshieldRequestError, err)
		}
		err = stateDB.SetStateObject(PortalProcessedUnshieldRequestBatchObjectType, *key, batchReq)
		if err != nil {
			return NewStatedbError(StorePortalListProcessedBatchUnshieldRequestError, err)
		}
	}

	return nil
}

// Store and get the status of the Unshield Request by unshieldID
func StorePortalBatchUnshieldRequestStatus(stateDB *StateDB, batchID string, statusContent []byte) error {
	statusType := PortalBatchUnshieldRequestStatusPrefix()
	statusSuffix := []byte(batchID)
	err := StorePortalV4Status(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalBatchUnshieldRequestStatusError, err)
	}

	return nil
}

func DeletePortalBatchUnshieldRequests(stateDB *StateDB, batchProcessedUnshieldReqKeys []common.Hash) error {
	for _, keyHash := range batchProcessedUnshieldReqKeys {
		stateDB.MarkDeleteStateObject(PortalProcessedUnshieldRequestBatchObjectType, keyHash)
	}

	return nil
}

// ================= Batching Unshielding Request Status =================
func GetPortalBatchUnshieldRequestStatus(stateDB *StateDB, batchID string) ([]byte, error) {
	statusType := PortalBatchUnshieldRequestStatusPrefix()
	statusSuffix := []byte(batchID)
	data, err := GetPortalV4Status(stateDB, statusType, statusSuffix)
	if err != nil && err.(*StatedbError).GetErrorCode() != ErrCodeMessage[GetPortalStatusNotFoundError].Code {
		return []byte{}, NewStatedbError(GetPortalBatchUnshieldRequestStatusError, err)
	}

	return data, nil
}

// ================= Unshielding Batch Fee Replacement Status =================
// Store and get the status of the Unshield Batch Replacement Request by batchID
func StorePortalUnshieldBatchReplacementRequestStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalUnshielFeeReplacementBatchStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalV4Status(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalUnshieldBatchFeeReplacementRequestStatusError, err)
	}

	return nil
}

func GetPortalUnshieldBatchReplacementRequestStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalUnshielFeeReplacementBatchStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalV4Status(stateDB, statusType, statusSuffix)
	if err != nil && err.(*StatedbError).GetErrorCode() != ErrCodeMessage[GetPortalStatusNotFoundError].Code {
		return []byte{}, NewStatedbError(GetPortalUnshieldBatchFeeReplacementRequestStatusError, err)
	}

	return data, nil
}

// ================= Submit unshield batch confirmed Status =================
// Store and get the status of the Unshield Batch Replacement Request by batchID
func StorePortalSubmitConfirmedTxRequestStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalSubmitConfirmedTxStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalV4Status(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalSubmitConfirmedTxRequestStatusError, err)
	}

	return nil
}

func GetPortalSubmitConfirmedTxRequestStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalSubmitConfirmedTxStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalV4Status(stateDB, statusType, statusSuffix)
	if err != nil && err.(*StatedbError).GetErrorCode() != ErrCodeMessage[GetPortalStatusNotFoundError].Code {
		return []byte{}, NewStatedbError(GetPortalSubmitConfirmedTxRequestStatusError, err)
	}

	return data, nil
}

// ================= Portal v4 Status =================
func StorePortalV4Status(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePortalV4StatusObjectKey(statusType, statusSuffix)
	value := NewPortalV4StatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(PortalV4StatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePortalV4StatusError, err)
	}
	return nil
}

func GetPortalV4Status(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePortalV4StatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPortalV4StatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalV4StatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPortalStatusNotFoundError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent, nil
}
