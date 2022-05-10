package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectUserMintNft struct {
	otaReceiver string
	amount      uint64
	shardID     byte
	txReqID     common.Hash
}

func NewRejectUserMintNft() *RejectUserMintNft {
	return &RejectUserMintNft{}
}

func NewRejectUserMintNftWithValue(otaReceiver string, amount uint64, shardID byte, txReqID common.Hash) *RejectUserMintNft {
	return &RejectUserMintNft{
		otaReceiver: otaReceiver,
		amount:      amount,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

func (r *RejectUserMintNft) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3UserMintNftRequestMeta, source[0])
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

func (r *RejectUserMintNft) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta))
	res = append(res, common.Pdexv3RejectStringStatus)
	data, err := json.Marshal(r)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (r *RejectUserMintNft) MarshalJSON() ([]byte, error) {
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

func (r *RejectUserMintNft) UnmarshalJSON(data []byte) error {
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

func (r *RejectUserMintNft) OtaReceiver() string {
	return r.otaReceiver
}

func (r *RejectUserMintNft) Amount() uint64 {
	return r.amount
}

func (r *RejectUserMintNft) ShardID() byte {
	return r.shardID
}

func (r *RejectUserMintNft) TxReqID() common.Hash {
	return r.txReqID
}
