package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AcceptMintAccessToken struct {
	burntAmount uint64
	otaReceiver string
	shardID     byte
	txReqID     common.Hash
}

func NewAcceptMintAccessToken() *AcceptMintAccessToken {
	return &AcceptMintAccessToken{}
}

func NewAcceptMintAccessTokenWithValue(
	burntAmount uint64,
	otaReceiver string,
	shardID byte,
	txReqID common.Hash,
) *AcceptMintAccessToken {
	return &AcceptMintAccessToken{
		burntAmount: burntAmount,
		otaReceiver: otaReceiver,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

func (a *AcceptMintAccessToken) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3MintAccessTokenRequestMeta, source[0])
	}
	if source[1] != common.Pdexv3AcceptStringStatus {
		return fmt.Errorf("Expect status %s but get %v", common.Pdexv3AcceptStringStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), a)
	if err != nil {
		return err
	}
	return nil
}

func (a *AcceptMintAccessToken) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenRequestMeta))
	res = append(res, common.Pdexv3AcceptStringStatus)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *AcceptMintAccessToken) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver string      `json:"OtaReceiver"`
		BurntAmount uint64      `json:"BurntAmount"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{
		OtaReceiver: a.otaReceiver,
		BurntAmount: a.burntAmount,
		ShardID:     a.shardID,
		TxReqID:     a.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptMintAccessToken) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver string      `json:"OtaReceiver"`
		BurntAmount uint64      `json:"BurntAmount"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.otaReceiver = temp.OtaReceiver
	a.burntAmount = temp.BurntAmount
	a.shardID = temp.ShardID
	a.txReqID = temp.TxReqID
	return nil
}

func (a *AcceptMintAccessToken) OtaReceiver() string {
	return a.otaReceiver
}

func (a *AcceptMintAccessToken) ShardID() byte {
	return a.shardID
}

func (a *AcceptMintAccessToken) TxReqID() common.Hash {
	return a.txReqID
}

func (a *AcceptMintAccessToken) BurntAmount() uint64 {
	return a.burntAmount
}
