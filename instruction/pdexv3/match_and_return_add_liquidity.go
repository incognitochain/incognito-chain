package pdexv3

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type MatchAndReturnAddLiquidity struct {
	Base
	returnAmount uint64
	nfctID       string
}

func NewMatchAndReturnAddLiquidity() *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		Base: Base{
			metaData: metadataPdexv3.NewAddLiquidity(),
		},
	}
}

func NewMatchAndReturnAddLiquidityFromMetadata(
	metaData metadataPdexv3.AddLiquidity,
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

func (m *MatchAndReturnAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 4 {
		return errors.New("Length of source can not be smaller than 4")
	}
	err := m.Base.FromStringSlice(temp[:len(temp)-3])
	if err != nil {
		return err
	}
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

func (m *MatchAndReturnAddLiquidity) StringSlice() []string {
	res := m.Base.StringSlice()
	returnAmount := strconv.FormatUint(m.returnAmount, 10)
	res = append(res, returnAmount)
	res = append(res, m.nfctID)
	res = append(res, MatchAndReturnStatus)
	return res
}

func (m *MatchAndReturnAddLiquidity) ReturnAmount() uint64 {
	return m.returnAmount
}

func (m *MatchAndReturnAddLiquidity) NfctID() string {
	return m.nfctID
}
