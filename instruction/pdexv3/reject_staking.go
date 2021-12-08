package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectStaking struct {
	tokenID     common.Hash
	amount      uint64
	otaReceiver string
	shardID     byte
	txReqID     common.Hash
}

func NewRejectStaking() *RejectStaking { return &RejectStaking{} }

func NewRejectStakingWithValue(
	otaReceiver string, tokenID, txReqID common.Hash, shardID byte, amount uint64,
) *RejectStaking {
	return &RejectStaking{
		otaReceiver: otaReceiver,
		tokenID:     tokenID,
		txReqID:     txReqID,
		shardID:     shardID,
		amount:      amount,
	}
}

func (r *RejectStaking) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3StakingRequestMeta, source[0])
	}
	if source[1] != common.Pdexv3RejectStringStatus {
		return fmt.Errorf("Expect status %s but get %v", common.Pdexv3RejectStringStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), r)
	if err != nil {
		return err
	}
	return nil
}

func (r *RejectStaking) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta))
	res = append(res, common.Pdexv3RejectStringStatus)
	data, err := json.Marshal(r)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (r *RejectStaking) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount      uint64      `json:"Amount"`
		TokenID     common.Hash `json:"TokenID"`
		OtaReceiver string      `json:"OtaReceiver"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{
		TokenID:     r.tokenID,
		OtaReceiver: r.otaReceiver,
		ShardID:     r.shardID,
		TxReqID:     r.txReqID,
		Amount:      r.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *RejectStaking) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount      uint64      `json:"Amount"`
		TokenID     common.Hash `json:"TokenID"`
		OtaReceiver string      `json:"OtaReceiver"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	r.tokenID = temp.TokenID
	r.otaReceiver = temp.OtaReceiver
	r.shardID = temp.ShardID
	r.txReqID = temp.TxReqID
	r.amount = temp.Amount
	return nil
}

func (r *RejectStaking) TokenID() common.Hash {
	return r.tokenID
}

func (r *RejectStaking) OtaReceiver() string {
	return r.otaReceiver
}

func (r *RejectStaking) ShardID() byte {
	return r.shardID
}

func (r *RejectStaking) TxReqID() common.Hash {
	return r.txReqID
}

func (r *RejectStaking) Amount() uint64 {
	return r.amount
}
