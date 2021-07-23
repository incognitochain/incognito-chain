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
	returnAmount             uint64
	existedTokenActualAmount uint64
	existedTokenID           string
	nfctID                   string
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
	returnAmount, existedTokenActualAmount uint64,
	existedTokenID, nfctID string,
) *MatchAndReturnAddLiquidity {
	return NewMatchAndReturnAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
		returnAmount, existedTokenActualAmount,
		existedTokenID, nfctID,
	)
}

func NewMatchAndReturnAddLiquidityWithValue(
	base Base,
	returnAmount, existedTokenActualAmount uint64,
	existedTokenID, nfctID string,
) *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		Base:                     base,
		returnAmount:             returnAmount,
		existedTokenActualAmount: existedTokenActualAmount,
		existedTokenID:           existedTokenID,
		nfctID:                   nfctID,
	}
}

func (m *MatchAndReturnAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 6 {
		return errors.New("Length of source can not be smaller than 6")
	}
	err := m.Base.FromStringSlice(temp[:len(temp)-5])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-5:]
	returnAmount, err := strconv.ParseUint(temp[0], 10, 32)
	if err != nil {
		return err
	}
	m.returnAmount = returnAmount
	existedTokenActualAmount, err := strconv.ParseUint(temp[1], 10, 32)
	if err != nil {
		return err
	}
	m.existedTokenActualAmount = existedTokenActualAmount
	existedTokenID, err := common.Hash{}.NewHashFromStr(temp[2])
	if err != nil {
		return err
	}
	if existedTokenID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.existedTokenID = temp[2]
	nfctID, err := common.Hash{}.NewHashFromStr(temp[3])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.nfctID = temp[3]
	if temp[4] != MatchAndReturnStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[4], MatchAndReturnStatus)
	}
	return nil
}

func (m *MatchAndReturnAddLiquidity) StringSlice() []string {
	res := m.Base.StringSlice()
	returnAmount := strconv.FormatUint(m.returnAmount, 10)
	res = append(res, returnAmount)
	existedTokenActualAmount := strconv.FormatUint(m.existedTokenActualAmount, 10)
	res = append(res, existedTokenActualAmount)
	res = append(res, m.existedTokenID)
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

func (m *MatchAndReturnAddLiquidity) ExistedTokenActualAmount() uint64 {
	return m.existedTokenActualAmount
}

func (m *MatchAndReturnAddLiquidity) ExistedTokenID() string {
	return m.existedTokenID
}
