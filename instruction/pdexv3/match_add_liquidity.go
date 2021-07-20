package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
)

type MatchAddLiquidity struct {
	pairHash    string
	otaReceiver string // receive nfct
	tokenID     string
	tokenAmount uint64
	txReqID     string
	shardID     byte
}

func NewMatchAddLiquidity() *MatchAddLiquidity {
	return &MatchAddLiquidity{}
}

func NewMatchAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
	txReqID string, shardID byte,
) *MatchAddLiquidity {
	return NewMatchAddLiquidityWithValue(
		metaData.PairHash(),
		metaData.ReceiveAddress(),
		metaData.TokenID(),
		txReqID,
		metaData.TokenAmount(),
		shardID,
	)
}

func NewMatchAddLiquidityWithValue(
	pairHash, otaReceiver,
	tokenID, txReqID string,
	tokenAmount uint64,
	shardID byte,
) *MatchAddLiquidity {
	return &MatchAddLiquidity{
		pairHash:    pairHash,
		otaReceiver: otaReceiver,
		tokenID:     tokenID,
		tokenAmount: tokenAmount,
		txReqID:     txReqID,
		shardID:     shardID,
	}
}

func (m *MatchAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PairHash    string `json:"PairHash"`
		OTAReceiver string `json:"OTAReceiver"` // receive nfct
		TokenID     string `json:"TokenID"`
		TokenAmount uint64 `json:"TokenAmount"`
		TxReqID     string `json:"TxReqID"`
		ShardID     byte   `json:"ShardID"`
	}{
		PairHash:    m.pairHash,
		OTAReceiver: m.otaReceiver,
		TokenID:     m.tokenID,
		TokenAmount: m.tokenAmount,
		TxReqID:     m.txReqID,
		ShardID:     m.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (m *MatchAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PairHash    string `json:"PairHash"`
		OTAReceiver string `json:"OTAReceiver"` // Receive nfct
		TokenID     string `json:"TokenID"`
		TokenAmount uint64 `json:"TokenAmount"`
		TxReqID     string `json:"TxReqID"`
		ShardID     byte   `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	m.pairHash = temp.PairHash
	m.otaReceiver = temp.OTAReceiver
	m.tokenID = temp.TokenID
	m.tokenAmount = temp.TokenAmount
	m.txReqID = temp.TxReqID
	m.shardID = temp.ShardID
	return nil
}

func (m *MatchAddLiquidity) FromStringArr(source []string) error {
	if len(source) != 8 {
		return fmt.Errorf("Receive length %v but expect %v", len(source), 8)
	}
	if source[0] != strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta) {
		return fmt.Errorf("Receive metaType %v but expect %v", source[0], metadataCommon.PDexV3AddLiquidityMeta)
	}
	if source[1] != MatchStatus {
		return fmt.Errorf("Receive status %v but expect %v", source[1], MatchStatus)
	}
	if source[2] == "" {
		return errors.New("Pair hash is invalid")
	}
	m.pairHash = source[2]
	tokenID, err := common.Hash{}.NewHashFromStr(source[3])
	if err != nil {
		return err
	}
	if tokenID.IsZeroValue() {
		return errors.New("TokenID is empty")
	}
	m.tokenID = source[3]
	tokenAmount, err := strconv.ParseUint(source[4], 10, 32)
	if err != nil {
		return err
	}
	m.tokenAmount = tokenAmount
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(source[5])
	if err != nil {
		return err
	}
	if !otaReceiver.IsValid() {
		return errors.New("receiver Address is invalid")
	}
	m.otaReceiver = source[5]
	m.txReqID = source[6]
	shardID, err := strconv.Atoi(source[7])
	if err != nil {
		return err
	}
	m.shardID = byte(shardID)
	return nil
}

func (m *MatchAddLiquidity) StringArr() []string {
	metaDataType := strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta)
	res := []string{metaDataType, MatchStatus}
	res = append(res, m.pairHash)
	res = append(res, m.tokenID)
	tokenAmount := strconv.FormatUint(m.tokenAmount, 10)
	res = append(res, tokenAmount)
	res = append(res, m.otaReceiver)
	res = append(res, m.txReqID)
	shardID := strconv.Itoa(int(m.shardID))
	res = append(res, shardID)
	return res
}
