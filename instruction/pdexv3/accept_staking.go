package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AcceptStaking struct {
	nftID         common.Hash
	stakingPoolID common.Hash
	liquidity     uint64
	shardID       byte
	txReqID       common.Hash
}

func NewAcceptStaking() *AcceptStaking { return &AcceptStaking{} }

func NewAcceptStakingWtihValue(
	nftID, stakingPoolID, txReqID common.Hash, shardID byte, liquidity uint64,
) *AcceptStaking {
	return &AcceptStaking{
		nftID:         nftID,
		stakingPoolID: stakingPoolID,
		txReqID:       txReqID,
		shardID:       shardID,
		liquidity:     liquidity,
	}
}

func (a *AcceptStaking) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3StakingRequestMeta, source[0])
	}
	if source[1] != common.Pdexv3AcceptStakingStatus {
		return fmt.Errorf("Expect status %s but get %v", common.Pdexv3AcceptStakingStatus, source[1])
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
	res = append(res, common.Pdexv3AcceptStakingStatus)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *AcceptStaking) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID         common.Hash `json:"NftID"`
		StakingPoolID common.Hash `json:"StakingPoolID"`
		Liquidity     uint64      `json:"Liquidity"`
		ShardID       byte        `json:"ShardID"`
		TxReqID       common.Hash `json:"TxReqID"`
	}{
		NftID:         a.nftID,
		StakingPoolID: a.stakingPoolID,
		ShardID:       a.shardID,
		Liquidity:     a.liquidity,
		TxReqID:       a.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptStaking) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID         common.Hash `json:"NftID"`
		StakingPoolID common.Hash `json:"StakingPoolID"`
		Liquidity     uint64      `json:"Liquidity"`
		ShardID       byte        `json:"ShardID"`
		TxReqID       common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.stakingPoolID = temp.StakingPoolID
	a.liquidity = temp.Liquidity
	a.nftID = temp.NftID
	a.shardID = temp.ShardID
	a.txReqID = temp.TxReqID
	return nil
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

func (a *AcceptStaking) NftID() common.Hash {
	return a.nftID
}
