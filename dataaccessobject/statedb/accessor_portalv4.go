package statedb

import "github.com/incognitochain/incognito-chain/common"

// ================= Shielding Request =================
func StoreShieldingRequestStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalShieldingRequestStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalV4ShieldingRequestStatusError, err)
	}

	return nil
}

func GetShieldingRequestStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalShieldingRequestStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalV4ShieldingRequestStatusError, err)
	}

	return data, nil
}

func GetShieldingRequestsByTokenID(stateDB *StateDB, tokenID string) (map[string]*ShieldingRequest, error) {
	return stateDB.getShieldingRequestsByTokenID(tokenID), nil
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