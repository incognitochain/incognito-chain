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
	existedTokenReturnAmount uint64
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
	returnAmount, existedTokenActualAmount, existedTokenReturnAmount uint64,
	existedTokenID, nfctID string,
) *MatchAndReturnAddLiquidity {
	return NewMatchAndReturnAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
		returnAmount, existedTokenActualAmount, existedTokenReturnAmount,
		existedTokenID, nfctID,
	)
}

func NewMatchAndReturnAddLiquidityWithValue(
	base Base,
	returnAmount, existedTokenActualAmount, existedTokenReturnAmount uint64,
	existedTokenID, nfctID string,
) *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		Base:                     base,
		returnAmount:             returnAmount,
		existedTokenActualAmount: existedTokenActualAmount,
		existedTokenID:           existedTokenID,
		existedTokenReturnAmount: existedTokenReturnAmount,
		nfctID:                   nfctID,
	}
}

func (m *MatchAndReturnAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 7 {
		return errors.New("Length of source can not be smaller than 7")
	}
	err := m.Base.FromStringSlice(temp[:len(temp)-6])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-6:]
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
	existedTokenReturnAmount, err := strconv.ParseUint(temp[2], 10, 32)
	if err != nil {
		return err
	}
	m.existedTokenReturnAmount = existedTokenReturnAmount
	existedTokenID, err := common.Hash{}.NewHashFromStr(temp[3])
	if err != nil {
		return err
	}
	if existedTokenID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.existedTokenID = temp[3]
	nfctID, err := common.Hash{}.NewHashFromStr(temp[4])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.nfctID = temp[4]
	if temp[5] != strconv.Itoa(common.PDEContributionMatchedNReturnedStatus) {
		return fmt.Errorf("Receive status %s expect %s", temp[5], strconv.Itoa(common.PDEContributionMatchedNReturnedStatus))
	}
	return nil
}

func (m *MatchAndReturnAddLiquidity) StringSlice() []string {
	res := m.Base.StringSlice()
	returnAmount := strconv.FormatUint(m.returnAmount, 10)
	res = append(res, returnAmount)
	existedTokenActualAmount := strconv.FormatUint(m.existedTokenActualAmount, 10)
	res = append(res, existedTokenActualAmount)
	existedTokenReturnAmount := strconv.FormatUint(m.existedTokenReturnAmount, 10)
	res = append(res, existedTokenReturnAmount)
	res = append(res, m.existedTokenID)
	res = append(res, m.nfctID)
	res = append(res, strconv.Itoa(common.PDEContributionMatchedNReturnedStatus))
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

func (m *MatchAndReturnAddLiquidity) ExistedTokenReturnAmount() uint64 {
	return m.existedTokenReturnAmount
}

func (m *MatchAndReturnAddLiquidity) ExistedTokenID() string {
	return m.existedTokenID
}
