package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type MintNft struct {
	nftID       common.Hash
	otaReceiver string
	shardID     byte
	txReqID     common.Hash
}

func NewMintNft() *MintNft {
	return &MintNft{}
}

func NewMintNftWithValue(nftID common.Hash, otaReceiver string, shardID byte, txReqID common.Hash) *MintNft {
	return &MintNft{
		nftID:       nftID,
		otaReceiver: otaReceiver,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

// FromStringSlice verify format [{mintnft-metaType}, {action}, {data}]
// won't verify source[1] will be verify in other place
func (m *MintNft) FromStringSlice(source []string) error {
	if len(source) != 3 {
		return fmt.Errorf("Expect length %v but get %v", 3, len(source))
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3MintNftRequestMeta) {
		return fmt.Errorf("Expect metaType %v but get %s", metadataCommon.Pdexv3MintNftRequestMeta, source[0])
	}
	err := json.Unmarshal([]byte(source[2]), m)
	if err != nil {
		return err
	}
	return nil
}

// StringSlice format [{mintnft-metaType}, {action}, {data}]
func (m *MintNft) StringSlice(action string) ([]string, error) {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.Pdexv3MintNftRequestMeta))
	res = append(res, action)
	data, err := json.Marshal(m)
	if err != nil {
		return res, err
	}
	res = append(res, string(data))
	return res, nil
}

func (m *MintNft) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID       common.Hash `json:"NftID"`
		OtaReceiver string      `json:"OtaReceiver"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{
		NftID:       m.nftID,
		OtaReceiver: m.otaReceiver,
		ShardID:     m.shardID,
		TxReqID:     m.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (m *MintNft) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID       common.Hash `json:"NftID"`
		OtaReceiver string      `json:"OtaReceiver"`
		ShardID     byte        `json:"ShardID"`
		TxReqID     common.Hash `json:"TxReqID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	m.nftID = temp.NftID
	m.otaReceiver = temp.OtaReceiver
	m.shardID = temp.ShardID
	m.txReqID = temp.TxReqID
	return nil
}

func (m *MintNft) NftID() common.Hash {
	return m.nftID
}

func (m *MintNft) OtaReceiver() string {
	return m.otaReceiver
}

func (m *MintNft) ShardID() byte {
	return m.shardID
}

func (m *MintNft) TxReqID() common.Hash {
	return m.txReqID
}
