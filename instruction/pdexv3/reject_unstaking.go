package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectUnstaking struct {
	txReqID common.Hash
	shardID byte
}

func NewRejectUnstaking() *RejectUnstaking {
	return &RejectUnstaking{}
}

func NewRejectUnstakingWithValue(txReqID common.Hash, shardID byte) *RejectUnstaking {
	return &RejectUnstaking{
		shardID: shardID,
		txReqID: txReqID,
	}
}

func (r *RejectUnstaking) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3UnstakingRequestMeta, source[0])
	}
	if source[1] != common.Pdexv3RejectUnstakingStatus {
		return fmt.Errorf("Expect status %s but get %v", common.Pdexv3RejectUnstakingStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), r)
	return err
}

func (r *RejectUnstaking) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
	res = append(res, common.Pdexv3RejectUnstakingStatus)
	data, err := json.Marshal(r)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (r *RejectUnstaking) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxReqID common.Hash `json:"TxReqID"`
		ShardID byte        `json:"ShardID"`
	}{
		TxReqID: r.txReqID,
		ShardID: r.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *RejectUnstaking) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxReqID common.Hash `json:"TxReqID"`
		ShardID byte        `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	r.txReqID = temp.TxReqID
	r.shardID = temp.ShardID
	return nil
}

func (r *RejectUnstaking) TxReqID() common.Hash {
	return r.txReqID
}

func (r *RejectUnstaking) ShardID() byte {
	return r.shardID
}
