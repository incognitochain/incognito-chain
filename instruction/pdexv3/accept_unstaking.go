package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AcceptUnstaking struct {
	stakingPoolID common.Hash
	nftID         common.Hash
	amount        uint64
	otaReceiver   string
	txReqID       common.Hash
	shardID       byte
}

func NewAcceptUnstaking() *AcceptUnstaking {
	return &AcceptUnstaking{}
}

func NewAcceptUnstakingWithValue(
	stakingPoolID, nftID common.Hash,
	amount uint64,
	otaReceiver string,
	txReqID common.Hash, shardID byte,
) *AcceptUnstaking {
	return &AcceptUnstaking{
		stakingPoolID: stakingPoolID,
		nftID:         nftID,
		txReqID:       txReqID,
		shardID:       shardID,
		amount:        amount,
		otaReceiver:   otaReceiver,
	}
}

func (a *AcceptUnstaking) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3UnstakingRequestMeta, source[0])
	}
	if source[1] != common.Pdexv3AcceptUnstakingStatus {
		return fmt.Errorf("Expect status %s but get %v", common.Pdexv3AcceptUnstakingStatus, source[1])
	}
	err := json.Unmarshal([]byte(source[2]), a)
	return err
}

func (a *AcceptUnstaking) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
	res = append(res, common.Pdexv3AcceptUnstakingStatus)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *AcceptUnstaking) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		StakingPoolID common.Hash `json:"StakingPoolID"`
		NftID         common.Hash `json:"NftID"`
		Amount        uint64      `json:"Amount"`
		OtaReceiver   string      `json:"OtaReceiver"`
		TxReqID       common.Hash `json:"TxReqID"`
		ShardID       byte        `json:"ShardID"`
	}{
		StakingPoolID: a.stakingPoolID,
		NftID:         a.nftID,
		Amount:        a.amount,
		OtaReceiver:   a.otaReceiver,
		TxReqID:       a.txReqID,
		ShardID:       a.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptUnstaking) UnmarshalJSON(data []byte) error {
	temp := struct {
		StakingPoolID common.Hash `json:"StakingPoolID"`
		NftID         common.Hash `json:"NftID"`
		Amount        uint64      `json:"Amount"`
		OtaReceiver   string      `json:"OtaReceiver"`
		TxReqID       common.Hash `json:"TxReqID"`
		ShardID       byte        `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.nftID = temp.NftID
	a.amount = temp.Amount
	a.stakingPoolID = temp.StakingPoolID
	a.otaReceiver = temp.OtaReceiver
	a.txReqID = temp.TxReqID
	a.shardID = temp.ShardID
	return nil
}

func (a *AcceptUnstaking) TxReqID() common.Hash {
	return a.txReqID
}

func (a *AcceptUnstaking) ShardID() byte {
	return a.shardID
}

func (a *AcceptUnstaking) NftID() common.Hash {
	return a.nftID
}

func (a *AcceptUnstaking) StakingPoolID() common.Hash {
	return a.stakingPoolID
}

func (a *AcceptUnstaking) Amount() uint64 {
	return a.amount
}

func (a *AcceptUnstaking) OtaReceiver() string {
	return a.otaReceiver
}
