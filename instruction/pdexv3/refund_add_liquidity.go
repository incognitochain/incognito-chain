package pdexv3

import (
	"errors"
	"fmt"

	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type RefundAddLiquidity struct {
	Base
}

func NewRefundAddLiquidity() *RefundAddLiquidity {
	return &RefundAddLiquidity{}
}

func NewRefundAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
	txReqID string, shardID byte,
) *RefundAddLiquidity {
	return NewRefundAddLiquidityWithValue(*NewBaseWithValue(&metaData, txReqID, shardID))
}

func NewRefundAddLiquidityWithValue(base Base) *RefundAddLiquidity {
	return &RefundAddLiquidity{
		Base: base,
	}
}

func (r *RefundAddLiquidity) FromStringArr(source []string) error {
	temp := source
	if len(temp) < 2 {
		return errors.New("Length of source can not be smaller than 2")
	}
	r.Base.FromStringArr(temp[:len(temp)-1])
	temp = temp[len(temp)-1:]
	if temp[0] != RefundStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[0], RefundStatus)
	}
	return nil
}

func (r *RefundAddLiquidity) StringArr() []string {
	res := r.Base.StringArr()
	res = append(res, RefundStatus)
	return res
}
