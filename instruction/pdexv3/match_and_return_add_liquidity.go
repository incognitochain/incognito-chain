package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type MatchAndReturnAddLiquidity struct {
	shareAmount              uint64
	contribution             statedb.Pdexv3ContributionState
	returnAmount             uint64
	existedTokenActualAmount uint64
	existedTokenReturnAmount uint64
	existedTokenID           common.Hash
	accessOTA                []byte
}

func NewMatchAndReturnAddLiquidity() *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{}
}

func NewMatchAndReturnAddLiquidityWithValue(
	contribution statedb.Pdexv3ContributionState,
	shareAmount, returnAmount, existedTokenActualAmount, existedTokenReturnAmount uint64,
	existedTokenID common.Hash, accessOTA []byte,
) *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		shareAmount:              shareAmount,
		contribution:             contribution,
		returnAmount:             returnAmount,
		existedTokenActualAmount: existedTokenActualAmount,
		existedTokenID:           existedTokenID,
		existedTokenReturnAmount: existedTokenReturnAmount,
		accessOTA:                accessOTA,
	}
}

func (m *MatchAndReturnAddLiquidity) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3AddLiquidityRequestMeta, source[0])
	}
	if source[1] != common.PDEContributionMatchedNReturnedChainStatus {
		return fmt.Errorf("Expect status %s but get %v", common.PDEContributionWaitingChainStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), m)
	if err != nil {
		return err
	}
	return nil
}

func (m *MatchAndReturnAddLiquidity) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
	res = append(res, common.PDEContributionMatchedNReturnedChainStatus)
	data, err := json.Marshal(m)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (m *MatchAndReturnAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ShareAmount              uint64                           `json:"ShareAmount"`
		Contribution             *statedb.Pdexv3ContributionState `json:"Contribution"`
		ReturnAmount             uint64                           `json:"ReturnAmount"`
		ExistedTokenActualAmount uint64                           `json:"ExistedTokenActualAmount"`
		ExistedTokenReturnAmount uint64                           `json:"ExistedTokenReturnAmount"`
		ExistedTokenID           common.Hash                      `json:"ExistedTokenID"`
		AccessOTA                []byte                           `json:"AccessOTA,omitempty"`
	}{
		ShareAmount:              m.shareAmount,
		Contribution:             &m.contribution,
		ReturnAmount:             m.returnAmount,
		ExistedTokenActualAmount: m.existedTokenActualAmount,
		ExistedTokenReturnAmount: m.existedTokenReturnAmount,
		ExistedTokenID:           m.existedTokenID,
		AccessOTA:                m.accessOTA,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (m *MatchAndReturnAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		ShareAmount              uint64                           `json:"ShareAmount"`
		Contribution             *statedb.Pdexv3ContributionState `json:"Contribution"`
		ReturnAmount             uint64                           `json:"ReturnAmount"`
		ExistedTokenActualAmount uint64                           `json:"ExistedTokenActualAmount"`
		ExistedTokenReturnAmount uint64                           `json:"ExistedTokenReturnAmount"`
		ExistedTokenID           common.Hash                      `json:"ExistedTokenID"`
		AccessOTA                []byte                           `json:"AccessOTA,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	m.shareAmount = temp.ShareAmount
	if temp.Contribution != nil {
		m.contribution = *temp.Contribution
	}
	m.returnAmount = temp.ReturnAmount
	m.existedTokenActualAmount = temp.ExistedTokenActualAmount
	m.existedTokenReturnAmount = temp.ExistedTokenReturnAmount
	m.existedTokenID = temp.ExistedTokenID
	m.accessOTA = temp.AccessOTA
	return nil
}

func (m *MatchAndReturnAddLiquidity) ReturnAmount() uint64 {
	return m.returnAmount
}

func (m *MatchAndReturnAddLiquidity) ExistedTokenActualAmount() uint64 {
	return m.existedTokenActualAmount
}

func (m *MatchAndReturnAddLiquidity) ExistedTokenReturnAmount() uint64 {
	return m.existedTokenReturnAmount
}

func (m *MatchAndReturnAddLiquidity) ExistedTokenID() common.Hash {
	return m.existedTokenID
}

func (m *MatchAndReturnAddLiquidity) Contribution() statedb.Pdexv3ContributionState {
	return m.contribution
}

func (m *MatchAndReturnAddLiquidity) ShareAmount() uint64 {
	return m.shareAmount
}

func (m *MatchAndReturnAddLiquidity) AccessOTA() []byte {
	return m.accessOTA
}
