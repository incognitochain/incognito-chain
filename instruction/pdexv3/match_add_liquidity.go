package pdexv3

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type MatchAddLiquidity struct {
	Base
	newPoolPairID string
	nfctID        string
}

func NewMatchAddLiquidity() *MatchAddLiquidity {
	return &MatchAddLiquidity{
		Base: Base{
			metaData: metadataPdexv3.NewAddLiquidity(),
		},
	}
}

func NewMatchAddLiquidityFromMetadata(
	metaData metadataPdexv3.AddLiquidity,
	txReqID string, shardID byte,
	newPoolPairID, nfctID string,
) *MatchAddLiquidity {
	return NewMatchAddLiquidityWithValue(
		*NewBaseWithValue(&metaData, txReqID, shardID),
		newPoolPairID, nfctID,
	)
}

func NewMatchAddLiquidityWithValue(
	base Base, newPoolPairID, nfctID string,
) *MatchAddLiquidity {
	return &MatchAddLiquidity{
		Base:          base,
		newPoolPairID: newPoolPairID,
		nfctID:        nfctID,
	}
}

func (m *MatchAddLiquidity) FromStringSlice(source []string) error {
	temp := source
	if len(temp) < 4 {
		return errors.New("Length of source can not be smaller than 4")
	}
	err := m.Base.FromStringSlice(temp[:len(temp)-3])
	if err != nil {
		return err
	}
	temp = temp[len(temp)-3:]
	if temp[0] == utils.EmptyString {
		return errors.New("PoolPairID can not be empty")
	}
	m.newPoolPairID = temp[0]
	nfctID, err := common.Hash{}.NewHashFromStr(temp[1])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.nfctID = temp[1]
	if temp[2] != strconv.Itoa(common.PDEContributionAcceptedStatus) {
		return fmt.Errorf("Receive status %s expect %s", temp[2], strconv.Itoa(common.PDEContributionAcceptedStatus))
	}
	return nil
}

func (m *MatchAddLiquidity) StringSlice() []string {
	res := m.Base.StringSlice()
	res = append(res, m.newPoolPairID)
	res = append(res, m.nfctID)
	res = append(res, strconv.Itoa(common.PDEContributionAcceptedStatus))
	return res
}

func (m *MatchAddLiquidity) NewPoolPairID() string {
	return m.newPoolPairID
}

func (m *MatchAddLiquidity) NfctID() string {
	return m.nfctID
}
