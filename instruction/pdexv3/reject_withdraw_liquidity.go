package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectWithdrawLiquidity struct {
	txReqID    common.Hash
	shardID    byte
	poolPairID string
	accessID   *common.Hash
	accessOTA  []byte
}

func NewRejectWithdrawLiquidity() *RejectWithdrawLiquidity {
	return &RejectWithdrawLiquidity{}
}

func NewRejectWithdrawLiquidityWithValue(
	txReqID common.Hash, shardID byte, poolPairID string, accessID *common.Hash, accessOTA []byte,
) *RejectWithdrawLiquidity {
	return &RejectWithdrawLiquidity{
		shardID:    shardID,
		txReqID:    txReqID,
		poolPairID: poolPairID,
		accessID:   accessID,
		accessOTA:  accessOTA,
	}
}

func (r *RejectWithdrawLiquidity) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3WithdrawLiquidityRequestMeta, source[0])
	}
	if source[1] != common.PDEWithdrawalRejectedChainStatus {
		return fmt.Errorf("Expect status %s but get %v", common.PDEWithdrawalRejectedChainStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), r)
	return err
}

func (r *RejectWithdrawLiquidity) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
	res = append(res, common.PDEWithdrawalRejectedChainStatus)
	data, err := json.Marshal(r)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (r *RejectWithdrawLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxReqID    common.Hash  `json:"TxReqID"`
		ShardID    byte         `json:"ShardID"`
		PoolPairID string       `json:"PoolPairID,omitempty"`
		AccessID   *common.Hash `json:"AccessID,omitempty"`
		AccessOTA  []byte       `json:"AccessOTA,omitempty"`
	}{
		TxReqID:    r.txReqID,
		ShardID:    r.shardID,
		PoolPairID: r.poolPairID,
		AccessID:   r.accessID,
		AccessOTA:  r.accessOTA,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *RejectWithdrawLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxReqID    common.Hash  `json:"TxReqID"`
		ShardID    byte         `json:"ShardID"`
		PoolPairID string       `json:"PoolPairID,omitempty"`
		AccessID   *common.Hash `json:"AccessID,omitempty"`
		AccessOTA  []byte       `json:"AccessOTA,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	r.txReqID = temp.TxReqID
	r.shardID = temp.ShardID
	r.accessID = temp.AccessID
	r.poolPairID = temp.PoolPairID
	r.accessOTA = temp.AccessOTA
	return nil
}

func (r *RejectWithdrawLiquidity) TxReqID() common.Hash {
	return r.txReqID
}

func (r *RejectWithdrawLiquidity) ShardID() byte {
	return r.shardID
}

func (r *RejectWithdrawLiquidity) PoolPairID() string {
	return r.poolPairID
}

func (r *RejectWithdrawLiquidity) AccessID() *common.Hash {
	return r.accessID
}

func (r *RejectWithdrawLiquidity) AccessOTA() []byte {
	return r.accessOTA
}
