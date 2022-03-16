package bridgeagg

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func IsBridgeAggMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BridgeAggModifyListTokenMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta:
		return true
	case metadataCommon.IssuingUnifiedTokenRequestMeta:
		return true
	case metadataCommon.IssuingUnifiedTokenResponseMeta:
		return true
	default:
		return false
	}
}

type Vault struct {
	statedb.BridgeAggConvertedTokenState
	RewardReserve uint64 `json:"RewardReserve"`
}

func (v *Vault) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State         *statedb.BridgeAggConvertedTokenState `json:"State"`
		RewardReserve uint64                                `json:"RewardReserve"`
	}{
		State:         &v.BridgeAggConvertedTokenState,
		RewardReserve: v.RewardReserve,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		State         *statedb.BridgeAggConvertedTokenState `json:"State"`
		RewardReserve uint64                                `json:"RewardReserve"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	v.RewardReserve = temp.RewardReserve
	if temp.State != nil {
		v.BridgeAggConvertedTokenState = *temp.State
	}
	return nil
}
