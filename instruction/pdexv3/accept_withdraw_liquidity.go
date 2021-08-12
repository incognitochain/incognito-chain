package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AcceptWithdrawLiquidity struct {
	poolPairID  string
	nftID       common.Hash
	index       common.Hash
	tokenID     common.Hash
	tokenAmount uint64
	shareAmount uint64
	txReqID     common.Hash
	shardID     byte
}

func NewAcceptWithdrawLiquidity() *AcceptWithdrawLiquidity {
	return &AcceptWithdrawLiquidity{}
}

func NewAcceptWithdrawLiquidityWithValue(
	poolPairID string,
	nftID, index, tokenID common.Hash,
	tokenAmount, shareAmount uint64,
	txReqID common.Hash, shardID byte,
) *AcceptWithdrawLiquidity {
	return &AcceptWithdrawLiquidity{
		poolPairID:  poolPairID,
		nftID:       nftID,
		index:       index,
		txReqID:     txReqID,
		shardID:     shardID,
		tokenID:     tokenID,
		tokenAmount: tokenAmount,
		shareAmount: shareAmount,
	}
}

func (a *AcceptWithdrawLiquidity) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3WithdrawLiquidityRequestMeta, source[0])
	}
	if source[1] != common.PDEWithdrawalAcceptedChainStatus {
		return fmt.Errorf("Expect status %s but get %v", common.PDEWithdrawalAcceptedChainStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), a)
	return err
}

func (a *AcceptWithdrawLiquidity) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
	res = append(res, common.PDEWithdrawalAcceptedChainStatus)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *AcceptWithdrawLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID  string      `json:"PoolPairID"`
		NftID       common.Hash `json:"NftID"`
		Index       common.Hash `json:"Index"`
		TokenID     common.Hash `json:"TokenID"`
		TokenAmount uint64      `json:"TokenAmount"`
		ShareAmount uint64      `json:"ShareAmount"`
		TxReqID     common.Hash `json:"TxReqID"`
		ShardID     byte        `jdon:"ShardID"`
	}{
		PoolPairID:  a.poolPairID,
		NftID:       a.nftID,
		Index:       a.index,
		TokenID:     a.tokenID,
		TokenAmount: a.tokenAmount,
		ShareAmount: a.shareAmount,
		TxReqID:     a.txReqID,
		ShardID:     a.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptWithdrawLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID  string      `json:"PoolPairID"`
		NftID       common.Hash `json:"NftID"`
		Index       common.Hash `json:"Index"`
		TokenID     common.Hash `json:"TokenID"`
		TokenAmount uint64      `json:"TokenAmount"`
		ShareAmount uint64      `json:"ShareAmount"`
		TxReqID     common.Hash `json:"TxReqID"`
		ShardID     byte        `jdon:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.tokenID = temp.TokenID
	a.tokenAmount = temp.TokenAmount
	a.poolPairID = temp.PoolPairID
	a.nftID = temp.NftID
	a.index = temp.Index
	a.shareAmount = temp.ShareAmount
	a.txReqID = temp.TxReqID
	a.shardID = temp.ShardID
	return nil
}

func (a *AcceptWithdrawLiquidity) TxReqID() common.Hash {
	return a.txReqID
}

func (a *AcceptWithdrawLiquidity) ShardID() byte {
	return a.shardID
}

func (a *AcceptWithdrawLiquidity) NftID() common.Hash {
	return a.nftID
}

func (a *AcceptWithdrawLiquidity) PoolPairID() string {
	return a.poolPairID
}

func (a *AcceptWithdrawLiquidity) TokenID() common.Hash {
	return a.tokenID
}

func (a *AcceptWithdrawLiquidity) TokenAmount() uint64 {
	return a.tokenAmount
}

func (a *AcceptWithdrawLiquidity) Index() common.Hash {
	return a.index
}

func (a *AcceptWithdrawLiquidity) ShareAmount() uint64 {
	return a.shareAmount
}
