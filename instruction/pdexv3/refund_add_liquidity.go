package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RefundAddLiquidity struct {
	contribution statedb.Pdexv3ContributionState
}

func NewRefundAddLiquidity() *RefundAddLiquidity {
	return &RefundAddLiquidity{}
}

func NewRefundAddLiquidityWithValue(contribution statedb.Pdexv3ContributionState) *RefundAddLiquidity {
	return &RefundAddLiquidity{
		contribution: contribution,
	}
}

func (r *RefundAddLiquidity) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3AddLiquidityRequestMeta, source[0])
	}
	if source[1] != common.PDEContributionRefundChainStatus {
		return fmt.Errorf("Expect status %s but get %v", common.PDEContributionRefundChainStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), r)
	if err != nil {
		return err
	}
	return nil
}

func (r *RefundAddLiquidity) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
	res = append(res, common.PDEContributionRefundChainStatus)
	data, err := json.Marshal(r)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (r *RefundAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Contribution statedb.Pdexv3ContributionState `json:"Contribution"`
	}{
		Contribution: r.contribution,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *RefundAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		Contribution statedb.Pdexv3ContributionState `json:"Contribution"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	r.contribution = temp.Contribution
	return nil
}

func (r *RefundAddLiquidity) Contribution() statedb.Pdexv3ContributionState {
	return r.contribution
}
