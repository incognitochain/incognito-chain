package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type MatchAddLiquidity struct {
	contribution  statedb.Pdexv3ContributionState
	newPoolPairID string
	nfctID        common.Hash
}

func NewMatchAddLiquidity() *MatchAddLiquidity {
	return &MatchAddLiquidity{}
}

func NewMatchAddLiquidityWithValue(
	contribution statedb.Pdexv3ContributionState,
	newPoolPairID string,
	nfctID common.Hash,
) *MatchAddLiquidity {
	return &MatchAddLiquidity{
		contribution:  contribution,
		newPoolPairID: newPoolPairID,
		nfctID:        nfctID,
	}
}

func (m *MatchAddLiquidity) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3AddLiquidityRequestMeta, source[0])
	}
	if source[1] != common.PDEContributionMatchedChainStatus {
		return fmt.Errorf("Expect status %s but get %v", common.PDEContributionMatchedChainStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), m)
	if err != nil {
		return err
	}
	return nil
}

func (m *MatchAddLiquidity) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
	res = append(res, common.PDEContributionMatchedChainStatus)
	data, err := json.Marshal(m)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (m *MatchAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Contribution  statedb.Pdexv3ContributionState `json:"Contribution"`
		NewPoolPairID string                          `json:"NewPoolPairID"`
		NfctID        common.Hash                     `json:"NfctID"`
	}{
		Contribution:  m.contribution,
		NewPoolPairID: m.newPoolPairID,
		NfctID:        m.nfctID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (m *MatchAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		Contribution  statedb.Pdexv3ContributionState `json:"Contribution"`
		NewPoolPairID string                          `json:"NewPoolPairID"`
		NfctID        common.Hash                     `json:"NfctID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	m.contribution = temp.Contribution
	m.newPoolPairID = temp.NewPoolPairID
	m.nfctID = temp.NfctID
	return nil
}

func (m *MatchAddLiquidity) NewPoolPairID() string {
	return m.newPoolPairID
}

func (m *MatchAddLiquidity) NfctID() common.Hash {
	return m.nfctID
}

func (m *MatchAddLiquidity) Contribution() statedb.Pdexv3ContributionState {
	return m.contribution
}
