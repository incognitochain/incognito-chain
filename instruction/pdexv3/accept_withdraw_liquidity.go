package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type AcceptWithdrawLiquidity struct {
	poolPairID string
	metadataPdexv3.AccessOption
	tokenID     common.Hash
	tokenAmount uint64
	shareAmount uint64
	otaReceiver string
	accessOTA   []byte
	txReqID     common.Hash
	shardID     byte
}

func NewAcceptWithdrawLiquidity() *AcceptWithdrawLiquidity {
	return &AcceptWithdrawLiquidity{}
}

func NewAcceptWithdrawLiquidityWithValue(
	poolPairID string,
	tokenID common.Hash,
	tokenAmount, shareAmount uint64,
	otaReceiver string,
	txReqID common.Hash, shardID byte,
	accessOption metadataPdexv3.AccessOption,
	accessOTA []byte,
) *AcceptWithdrawLiquidity {
	return &AcceptWithdrawLiquidity{
		poolPairID:   poolPairID,
		AccessOption: accessOption,
		txReqID:      txReqID,
		shardID:      shardID,
		tokenID:      tokenID,
		tokenAmount:  tokenAmount,
		shareAmount:  shareAmount,
		otaReceiver:  otaReceiver,
		accessOTA:    accessOTA,
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
		PoolPairID string `json:"PoolPairID"`
		metadataPdexv3.AccessOption
		TokenID     common.Hash `json:"TokenID"`
		TokenAmount uint64      `json:"TokenAmount"`
		ShareAmount uint64      `json:"ShareAmount"`
		OtaReceiver string      `json:"OtaReceiver"`
		TxReqID     common.Hash `json:"TxReqID"`
		ShardID     byte        `json:"ShardID"`
		AccessOTA   []byte      `json:"AccessOTA,omitempty"`
	}{
		PoolPairID:   a.poolPairID,
		AccessOption: a.AccessOption,
		TokenID:      a.tokenID,
		TokenAmount:  a.tokenAmount,
		ShareAmount:  a.shareAmount,
		OtaReceiver:  a.otaReceiver,
		TxReqID:      a.txReqID,
		ShardID:      a.shardID,
		AccessOTA:    a.accessOTA,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptWithdrawLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID string `json:"PoolPairID"`
		metadataPdexv3.AccessOption
		TokenID     common.Hash `json:"TokenID"`
		TokenAmount uint64      `json:"TokenAmount"`
		OtaReceiver string      `json:"OtaReceiver"`
		ShareAmount uint64      `json:"ShareAmount"`
		TxReqID     common.Hash `json:"TxReqID"`
		ShardID     byte        `json:"ShardID"`
		AccessOTA   []byte      `json:"AccessOTA,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.tokenID = temp.TokenID
	a.tokenAmount = temp.TokenAmount
	a.poolPairID = temp.PoolPairID
	a.AccessOption = temp.AccessOption
	a.shareAmount = temp.ShareAmount
	a.otaReceiver = temp.OtaReceiver
	a.txReqID = temp.TxReqID
	a.shardID = temp.ShardID
	a.accessOTA = temp.AccessOTA
	return nil
}

func (a *AcceptWithdrawLiquidity) TxReqID() common.Hash {
	return a.txReqID
}

func (a *AcceptWithdrawLiquidity) ShardID() byte {
	return a.shardID
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

func (a *AcceptWithdrawLiquidity) ShareAmount() uint64 {
	return a.shareAmount
}

func (a *AcceptWithdrawLiquidity) OtaReceiver() string {
	return a.otaReceiver
}

func (a *AcceptWithdrawLiquidity) AccessOTA() []byte {
	return a.accessOTA
}
