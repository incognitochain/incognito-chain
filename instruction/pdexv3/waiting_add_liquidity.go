package pdexv3

import (
	"errors"
	"fmt"

	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type WaitingAddLiquidity struct {
	Base
}

func NewWaitingAddLiquidity() *WaitingAddLiquidity {
	return &WaitingAddLiquidity{
		Base: Base{
			metaData: metadataPdexv3.NewAddLiquidity(),
		},
	}
}

func NewWaitingAddLiquidityFromMetadata(
	metaData metadataPdexv3.AddLiquidity,
	txReqID string, shardID byte,
) *WaitingAddLiquidity {
	return NewWaitingAddLiquidityWithValue(*NewBaseWithValue(&metaData, txReqID, shardID))
}

func NewWaitingAddLiquidityWithValue(
	base Base,
) *WaitingAddLiquidity {
	return &WaitingAddLiquidity{
		Base: base,
	}
}

func (w *WaitingAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 2 {
		return errors.New("Length of source can not be smaller than 2")
	}
	err := w.Base.FromStringSlice(temp[:len(temp)-1])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-1:]
	if temp[0] != WaitingStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[0], WaitingStatus)
	}
	return nil
}

func (w *WaitingAddLiquidity) StringSlice() []string {
	return append(w.Base.StringSlice(), WaitingStatus)
}
