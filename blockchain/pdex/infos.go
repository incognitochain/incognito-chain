package pdex

import (
	"math"
	"reflect"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Infos struct {
	LiquidityMintedEpochs uint64
}

func NewInfos() *Infos {
	return &Infos{
		LiquidityMintedEpochs: 0,
	}
}

func NewInfosWithValue(infosState *statedb.Pdexv3Infos) *Infos {
	return &Infos{
		LiquidityMintedEpochs: infosState.LiquidityMintedEpochs(),
	}
}

func (p *Infos) Clone() *Infos {
	result := &Infos{}
	*result = *p

	return result
}

func EmptyInfos() *Infos {
	return &Infos{
		LiquidityMintedEpochs: math.MaxUint64,
	}
}

func (infos *Infos) IsZeroValue() bool {
	return reflect.DeepEqual(infos, EmptyInfos()) || infos == nil
}
