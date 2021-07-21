package pdexv3

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type MatchAddLiquidity struct {
	Base
	newPoolPairID string
	nfctID        string
}

func NewMatchAddLiquidity() *MatchAddLiquidity {
	return &MatchAddLiquidity{}
}

func NewMatchAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
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

func (m *MatchAddLiquidity) FromStringArr(source []string) error {
	temp := source
	if len(temp) < 4 {
		return errors.New("Length of source can not be smaller than 4")
	}
	m.Base.FromStringArr(temp[:len(temp)-3])
	temp = temp[len(temp)-3:]
	if temp[0] == utils.EmptyString {
		return errors.New("PoolPairID can not be empty")
	}
	nfctID, err := common.Hash{}.NewHashFromStr(temp[1])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.nfctID = temp[1]
	if temp[2] != MatchStatus {
		return fmt.Errorf("Receive status %s expect %s", temp[2], MatchStatus)
	}
	return nil
}

func (m *MatchAddLiquidity) StringArr() []string {
	res := m.Base.StringArr()
	res = append(res, m.newPoolPairID)
	res = append(res, m.nfctID)
	res = append(res, MatchStatus)
	return res
}

func (m *MatchAddLiquidity) NewPoolPairID() string {
	return m.newPoolPairID
}

func (m *MatchAddLiquidity) NfctID() string {
	return m.nfctID
}
