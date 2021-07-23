package pdexv3

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
)

type RefundAddLiquidity struct {
	Base
	existedTokenID     string
	existedTokenAmount uint64
	refundAddress      string
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
	existedTokenID, refundAddress string,
	existedTokenAmount uint64,
) *RefundAddLiquidity {
	return NewRefundAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
		existedTokenID, refundAddress, existedTokenAmount,
	)
}

func NewRefundAddLiquidityWithValue(
	base Base,
	existedTokenID, refundAddress string,
	existedTokenAmount uint64,
) *RefundAddLiquidity {
	return &RefundAddLiquidity{
		Base:               base,
		existedTokenID:     existedTokenID,
		existedTokenAmount: existedTokenAmount,
		refundAddress:      refundAddress,
	}
}

func (r *RefundAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 5 {
		return errors.New("Length of source can not be smaller than 5")
	}
	err := r.Base.FromStringSlice(temp[:len(temp)-4])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-4:]
	existedTokenID, err := common.Hash{}.NewHashFromStr(temp[0])
	if err != nil {
		return err
	}
	if existedTokenID.IsZeroValue() {
		return errors.New("ExistedTokenID is empty")
	}
	r.existedTokenID = temp[0]
	existedTokenAmount, err := strconv.ParseUint(temp[1], 10, 32)
	if err != nil {
		return err
	}
	r.existedTokenAmount = existedTokenAmount
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(temp[2])
	if err != nil {
		return err
	}
	if !refundAddress.IsValid() {
		return errors.New("ReceiveAddress is not valid")
	}
	if temp[3] != RefundStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[0], RefundStatus)
	}
	return nil
}

func (r *RefundAddLiquidity) StringSlice() []string {
	res := r.Base.StringSlice()
	res = append(res, r.existedTokenID)
	existedTokenAmount := strconv.FormatUint(r.existedTokenAmount, 10)
	res = append(res, existedTokenAmount)
	res = append(res, r.refundAddress)
	res = append(res, RefundStatus)
	return res
}

func (r *RefundAddLiquidity) ExistedTokenID() string {
	return r.existedTokenID
}

func (r *RefundAddLiquidity) ExistedTokenAmount() uint64 {
	return r.existedTokenAmount
}
