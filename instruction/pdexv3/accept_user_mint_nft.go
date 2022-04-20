package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AcceptUserMintNft struct {
	nftID       common.Hash
	burntAmount uint64
	otaReceiver string
	shardID     byte
	txReqID     common.Hash
}

func NewAcceptUserMintNft() *AcceptUserMintNft {
	return &AcceptUserMintNft{}
}

func NewAcceptUserMintNftWithValue(
	otaReceiver string, burntAmount uint64, shardID byte, nftID, txReqID common.Hash,
) *AcceptUserMintNft {
	return &AcceptUserMintNft{
		otaReceiver: otaReceiver,
		burntAmount: burntAmount,
		nftID:       nftID,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

func (a *AcceptUserMintNft) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3UserMintNftRequestMeta, source[0])
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

func (a *AcceptUserMintNft) StringSlice() ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta))
	res = append(res, common.Pdexv3AcceptStringStatus)
	data, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (a *AcceptUserMintNft) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver string      `json:"OtaReceiver"`
		BurntAmount uint64      `json:"BurntAmount"`
		NftID       common.Hash `json:"NftID"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{
		OtaReceiver: a.otaReceiver,
		BurntAmount: a.burntAmount,
		ShardID:     a.shardID,
		NftID:       a.nftID,
		TxReqID:     a.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (a *AcceptUserMintNft) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver string      `json:"OtaReceiver"`
		BurntAmount uint64      `json:"BurntAmount"`
		NftID       common.Hash `json:"NftID"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	a.otaReceiver = temp.OtaReceiver
	a.burntAmount = temp.BurntAmount
	a.nftID = temp.NftID
	a.shardID = temp.ShardID
	a.txReqID = temp.TxReqID
	return nil
}

func (a *AcceptUserMintNft) OtaReceiver() string {
	return a.otaReceiver
}

func (a *AcceptUserMintNft) NftID() common.Hash {
	return a.nftID
}

func (a *AcceptUserMintNft) ShardID() byte {
	return a.shardID
}

func (a *AcceptUserMintNft) TxReqID() common.Hash {
	return a.txReqID
}

func (a *AcceptUserMintNft) BurntAmount() uint64 {
	return a.burntAmount
}
