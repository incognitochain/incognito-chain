package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type MintAccessToken struct {
	otaReceiver string
	shardID     byte
	txReqID     common.Hash
}

func NewMintAccessToken() *MintAccessToken {
	return &MintAccessToken{}
}

func NewMintAccessTokenWithValue(
	otaReceiver string,
	shardID byte,
	txReqID common.Hash,
) *MintAccessToken {
	return &MintAccessToken{
		otaReceiver: otaReceiver,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

// FromStringSlice verify format [{mint-access-token-metaType}, {action}, {data}]
// won't verify source[1] will be verify in other place
func (a *MintAccessToken) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3MintAccessTokenMeta, source[0])
	}
	err := json.Unmarshal([]byte(source[2]), a)
	if err != nil {
		return err
	}
	return nil
}

// StringSlice format [{mint-access-token-metaType}, {action}, {data}]
func (a *MintAccessToken) StringSlice(action string) ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenMeta))
	res = append(res, action)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *MintAccessToken) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver string      `json:"OtaReceiver"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{
		OtaReceiver: a.otaReceiver,
		ShardID:     a.shardID,
		TxReqID:     a.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *MintAccessToken) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver string      `json:"OtaReceiver"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.otaReceiver = temp.OtaReceiver
	a.shardID = temp.ShardID
	a.txReqID = temp.TxReqID
	return nil
}

func (a *MintAccessToken) OtaReceiver() string {
	return a.otaReceiver
}

func (a *MintAccessToken) ShardID() byte {
	return a.shardID
}

func (a *MintAccessToken) TxReqID() common.Hash {
	return a.txReqID
}
