package pdexv3

import (
	"errors"
	"fmt"

	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type WaitingAddLiquidity struct {
	Base
}

func NewWaitingAddLiquidity() *WaitingAddLiquidity {
	return &WaitingAddLiquidity{}
}

func NewWaitingAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
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

func (w *WaitingAddLiquidity) FromStringArr(source []string) error {
	temp := source
	if len(temp) < 2 {
		return errors.New("Length of source can not be smaller than 2")
	}
	w.Base.FromStringArr(temp[:len(temp)-1])
	temp = temp[len(temp)-1:]
	if temp[0] != WaitingStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[0], WaitingStatus)
	}
	return nil
}

func (w *WaitingAddLiquidity) StringArr() []string {
	return append(w.Base.StringArr(), WaitingStatus)
}
