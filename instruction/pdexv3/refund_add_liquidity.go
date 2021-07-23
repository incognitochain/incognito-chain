package pdexv3

import (
	"errors"
	"fmt"

	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type RefundAddLiquidity struct {
	Base
}

func NewRefundAddLiquidity() *RefundAddLiquidity {
	return &RefundAddLiquidity{
		Base: Base{
			metaData: metadataPdexv3.NewAddLiquidity(),
		},
	}
}

func NewRefundAddLiquidityFromMetadata(
	metaData metadataPdexv3.AddLiquidity,
	txReqID string, shardID byte,
) *RefundAddLiquidity {
	return NewRefundAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
	)
}

func NewRefundAddLiquidityWithValue(
	base Base,
) *RefundAddLiquidity {
	return &RefundAddLiquidity{
		Base: base,
	}
}

func (r *RefundAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 2 {
		return errors.New("Length of source can not be smaller than 2")
	}
	err := r.Base.FromStringSlice(temp[:len(temp)-1])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-1:]
	if temp[0] != RefundStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[0], RefundStatus)
	}
	return nil
}

func (r *RefundAddLiquidity) StringSlice() []string {
	res := r.Base.StringSlice()
	res = append(res, RefundStatus)
	return res
}
