package pdexv3

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type MatchAndReturnAddLiquidity struct {
	Base
	returnAmount uint64
	nfctID       string
}

func NewMatchAndReturnAddLiquidity() *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{}
}

func NewMatchAndReturnAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
	txReqID string, shardID byte,
	returnAmount uint64, nfctID string,
) *MatchAndReturnAddLiquidity {
	return NewMatchAndReturnAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
		returnAmount,
		nfctID,
	)
}

func NewMatchAndReturnAddLiquidityWithValue(
	base Base, returnAmount uint64, nfctID string,
) *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		Base:         base,
		returnAmount: returnAmount,
		nfctID:       nfctID,
	}
}

func (m *MatchAndReturnAddLiquidity) FromStringArr(source []string) error {
	temp := source
	if len(temp) < 4 {
		return errors.New("Length of source can not be smaller than 4")
	}
	m.Base.FromStringArr(temp[:len(temp)-3])
	temp = temp[len(temp)-3:]
	returnAmount, err := strconv.ParseUint(temp[0], 10, 32)
	if err != nil {
		return err
	}
	m.returnAmount = returnAmount
	nfctID, err := common.Hash{}.NewHashFromStr(temp[1])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.nfctID = temp[1]
	if temp[2] != MatchAndReturnStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[2], MatchAndReturnStatus)
	}
	return nil
}

func (m *MatchAndReturnAddLiquidity) StringArr() []string {
	res := m.Base.StringArr()
	res = append(res, m.nfctID)
	res = append(res, MatchAndReturnStatus)
	return res
}
