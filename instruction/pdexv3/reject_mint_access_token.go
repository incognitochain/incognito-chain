package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectMintAccessToken struct {
	otaReceiver string
	amount      uint64
	shardID     byte
	txReqID     common.Hash
}

func NewRejectMintAccessToken() *RejectMintAccessToken {
	return &RejectMintAccessToken{}
}

func NewRejectMintAccessTokenWithValue(
	otaReceiver string, amount uint64, shardID byte, txReqID common.Hash,
) *RejectMintAccessToken {
	return &RejectMintAccessToken{
		otaReceiver: otaReceiver,
		amount:      amount,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

func (r *RejectMintAccessToken) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3MintAccessTokenRequestMeta, source[0])
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

func (r *RejectMintAccessToken) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenRequestMeta))
	res = append(res, common.Pdexv3RejectStringStatus)
	data, err := json.Marshal(r)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (r *RejectMintAccessToken) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver string      `json:"OtaReceiver"`
		Amount      uint64      `json:"Amount"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{
		OtaReceiver: r.otaReceiver,
		Amount:      r.amount,
		ShardID:     r.shardID,
		TxReqID:     r.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *RejectMintAccessToken) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver string      `json:"OtaReceiver"`
		Amount      uint64      `json:"Amount"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	r.amount = temp.Amount
	r.otaReceiver = temp.OtaReceiver
	r.shardID = temp.ShardID
	r.txReqID = temp.TxReqID
	return nil
}

func (r *RejectMintAccessToken) OtaReceiver() string {
	return r.otaReceiver
}

func (r *RejectMintAccessToken) Amount() uint64 {
	return r.amount
}

func (r *RejectMintAccessToken) ShardID() byte {
	return r.shardID
}

func (r *RejectMintAccessToken) TxReqID() common.Hash {
	return r.txReqID
}
