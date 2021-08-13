package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectWithdrawLiquidity struct {
	txReqID common.Hash
	shardID byte
}

func NewRejectWithdrawLiquidity() *RejectWithdrawLiquidity {
	return &RejectWithdrawLiquidity{}
}

func NewRejectWithdrawLiquidityWithValue(txReqID common.Hash, shardID byte) *RejectWithdrawLiquidity {
	return &RejectWithdrawLiquidity{
		shardID: shardID,
		txReqID: txReqID,
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

func (r *RejectWithdrawLiquidity) UnmarshalJSON(data []byte) error {
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

func (r *RejectWithdrawLiquidity) TxReqID() common.Hash {
	return r.txReqID
}

func (r *RejectWithdrawLiquidity) ShardID() byte {
	return r.shardID
}
