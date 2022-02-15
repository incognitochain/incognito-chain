package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type AcceptStaking struct {
	metadataPdexv3.AccessOption
	stakingPoolID common.Hash
	liquidity     uint64
	accessOTA     []byte
	shardID       byte
	txReqID       common.Hash
}

func NewAcceptStaking() *AcceptStaking { return &AcceptStaking{} }

func NewAcceptStakingWithAccessID(
	stakingPoolID, txReqID common.Hash, shardID byte, liquidity uint64, nextAccessOTA []byte,
	accessOption metadataPdexv3.AccessOption, accessID common.Hash,
) *AcceptStaking {
	if !accessOption.UseNft() && !accessID.IsZeroValue() {
		accessOption.AccessID = &accessID
	}
	return NewAcceptStakingWithValue(stakingPoolID, txReqID, shardID, liquidity, nextAccessOTA, accessOption)
}

func NewAcceptStakingWithValue(
	stakingPoolID, txReqID common.Hash, shardID byte, liquidity uint64, nextAccessOTA []byte,
	accessOption metadataPdexv3.AccessOption,
) *AcceptStaking {
	return &AcceptStaking{
		accessOTA:     nextAccessOTA,
		stakingPoolID: stakingPoolID,
		txReqID:       txReqID,
		shardID:       shardID,
		liquidity:     liquidity,
		AccessOption:  accessOption,
	}
}

func (a *AcceptStaking) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3StakingRequestMeta, source[0])
	}
	if source[1] != common.Pdexv3AcceptStringStatus {
		return fmt.Errorf("Expect status %s but get %v", common.Pdexv3AcceptStringStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), a)
	if err != nil {
		return err
	}
	return nil
}

func (a *AcceptStaking) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta))
	res = append(res, common.Pdexv3AcceptStringStatus)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *AcceptStaking) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		metadataPdexv3.AccessOption
		StakingPoolID common.Hash `json:"StakingPoolID"`
		Liquidity     uint64      `json:"Liquidity"`
		AccessOTA     []byte      `json:"AccessOTA,omitempty"`
		ShardID       byte        `json:"ShardID"`
		TxReqID       common.Hash `json:"TxReqID"`
	}{
		AccessOption:  a.AccessOption,
		StakingPoolID: a.stakingPoolID,
		ShardID:       a.shardID,
		Liquidity:     a.liquidity,
		AccessOTA:     a.accessOTA,
		TxReqID:       a.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptStaking) UnmarshalJSON(data []byte) error {
	temp := struct {
		metadataPdexv3.AccessOption
		StakingPoolID common.Hash `json:"StakingPoolID"`
		Liquidity     uint64      `json:"Liquidity"`
		AccessOTA     []byte      `json:"AccessOTA,omitempty"`
		ShardID       byte        `json:"ShardID"`
		TxReqID       common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.stakingPoolID = temp.StakingPoolID
	a.liquidity = temp.Liquidity
	a.AccessOption = temp.AccessOption
	a.shardID = temp.ShardID
	a.txReqID = temp.TxReqID
	a.accessOTA = temp.AccessOTA
	return nil
}

func (a *AcceptStaking) AccessOTA() []byte {
	return a.accessOTA
}

func (a *AcceptStaking) Liquidity() uint64 {
	return a.liquidity
}

func (a *AcceptStaking) StakingPoolID() common.Hash {
	return a.stakingPoolID
}

func (a *AcceptStaking) ShardID() byte {
	return a.shardID
}

func (a *AcceptStaking) TxReqID() common.Hash {
	return a.txReqID
}

func (a AcceptStaking) NftID() *common.Hash {
	return a.AccessOption.NftID
}
