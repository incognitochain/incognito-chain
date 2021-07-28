package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type WaitingAddLiquidity struct {
	contribution statedb.Pdexv3ContributionState
}

func NewWaitingAddLiquidity() *WaitingAddLiquidity {
	return &WaitingAddLiquidity{}
}

func NewWaitingAddLiquidityWithValue(contribution statedb.Pdexv3ContributionState) *WaitingAddLiquidity {
	return &WaitingAddLiquidity{
		contribution: contribution,
	}
}

func (w *WaitingAddLiquidity) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3AddLiquidityRequestMeta, source[0])
	}
	if source[1] != common.PDEContributionWaitingChainStatus {
		return fmt.Errorf("Expect status %s but get %v", common.PDEContributionWaitingChainStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), w)
	if err != nil {
		return err
	}
	return nil
}

func (w *WaitingAddLiquidity) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
	res = append(res, common.PDEContributionWaitingChainStatus)
	data, err := json.Marshal(w)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (w *WaitingAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Contribution statedb.Pdexv3ContributionState `json:"Contribution"`
	}{
		Contribution: w.contribution,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (w *WaitingAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		Contribution statedb.Pdexv3ContributionState `json:"Contribution"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	w.contribution = temp.Contribution
	return nil
}

func (w *WaitingAddLiquidity) Contribution() statedb.Pdexv3ContributionState {
	return w.contribution
}
