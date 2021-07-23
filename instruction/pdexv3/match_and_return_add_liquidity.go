package pdexv3

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
)

type MatchAndReturnAddLiquidity struct {
	Base
	returnAmount              uint64
	existedTokenActualAmount  uint64
	existedTokenID            string
	existedTokenReturnAmount  uint64
	existedTokenRefundAddress string
	nfctID                    string
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
	existedTokenID string, existedTokenReturnAmount uint64,
	existedTokenRefundAddress, nfctID string,
) *MatchAndReturnAddLiquidity {
	return NewMatchAndReturnAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
		returnAmount, existedTokenActualAmount,
		existedTokenID, existedTokenReturnAmount,
		existedTokenRefundAddress, nfctID,
	)
}

func NewMatchAndReturnAddLiquidityWithValue(
	base Base,
	returnAmount, existedTokenActualAmount uint64,
	existedTokenID string, existedTokenReturnAmount uint64,
	existedTokenRefundAddress, nfctID string,
) *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		Base:                      base,
		returnAmount:              returnAmount,
		existedTokenActualAmount:  existedTokenActualAmount,
		existedTokenID:            existedTokenID,
		existedTokenReturnAmount:  existedTokenReturnAmount,
		existedTokenRefundAddress: existedTokenRefundAddress,
		nfctID:                    nfctID,
	}
}

func (m *MatchAndReturnAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 8 {
		return errors.New("Length of source can not be smaller than 8")
	}
	err := m.Base.FromStringSlice(temp[:len(temp)-7])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-7:]
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
		return errors.New("ExistedTokenID is empty")
	}
	m.existedTokenID = temp[2]
	existedTokenReturnAmount, err := strconv.ParseUint(temp[3], 10, 32)
	if err != nil {
		return err
	}
	m.existedTokenReturnAmount = existedTokenReturnAmount
	existedTokenRefundAddress := privacy.OTAReceiver{}
	err = existedTokenRefundAddress.FromString(temp[4])
	if err != nil {
		return err
	}
	if !existedTokenRefundAddress.IsValid() {
		return errors.New("ExistedTokenRefundAddress is not valid")
	}
	m.existedTokenRefundAddress = temp[4]
	nfctID, err := common.Hash{}.NewHashFromStr(temp[5])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.nfctID = temp[5]
	if temp[6] != MatchAndReturnStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[6], MatchAndReturnStatus)
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
	existedTokenReturnAmount := strconv.FormatUint(m.existedTokenReturnAmount, 10)
	res = append(res, existedTokenReturnAmount)
	res = append(res, m.existedTokenRefundAddress)
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

func (m *MatchAndReturnAddLiquidity) ExistedTokenReturnAmount() uint64 {
	return m.existedTokenReturnAmount
}

func (m *MatchAndReturnAddLiquidity) ExistedTokenRefundAddress() string {
	return m.existedTokenRefundAddress
}
