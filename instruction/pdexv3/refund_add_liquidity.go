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

type RefundAddLiquidity struct {
	pairHash    string
	otaReceiver string // refund pToken
	tokenID     string
	tokenAmount uint64
	txReqID     string
	shardID     byte
}

func NewRefundAddLiquidity() *RefundAddLiquidity {
	return &RefundAddLiquidity{}
}

func NewRefundAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
	txReqID string, shardID byte,
) *RefundAddLiquidity {
	return NewRefundAddLiquidityWithValue(
		metaData.PairHash(),
		metaData.RefundAddress(),
		metaData.TokenID(),
		txReqID,
		metaData.TokenAmount(),
		shardID,
	)
}

func NewRefundAddLiquidityWithValue(
	pairHash, otaReceiver,
	tokenID, txReqID string,
	tokenAmount uint64,
	shardID byte,
) *RefundAddLiquidity {
	return &RefundAddLiquidity{
		pairHash:    pairHash,
		otaReceiver: otaReceiver,
		tokenID:     tokenID,
		tokenAmount: tokenAmount,
		txReqID:     txReqID,
		shardID:     shardID,
	}
}

func (r *RefundAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PairHash    string `json:"PairHash"`
		OTAReceiver string `json:"OTAReceiver"` // refund pToken
		TokenID     string `json:"TokenID"`
		TokenAmount uint64 `json:"TokenAmount"`
		TxReqID     string `json:"TxReqID"`
		ShardID     byte   `json:"ShardID"`
	}{
		PairHash:    r.pairHash,
		OTAReceiver: r.otaReceiver,
		TokenID:     r.tokenID,
		TokenAmount: r.tokenAmount,
		TxReqID:     r.txReqID,
		ShardID:     r.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *RefundAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PairHash    string `json:"PairHash"`
		OTAReceiver string `json:"OTAReceiver"` // refund pToken
		TokenID     string `json:"TokenID"`
		TokenAmount uint64 `json:"TokenAmount"`
		TxReqID     string `json:"TxReqID"`
		ShardID     byte   `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	r.pairHash = temp.PairHash
	r.otaReceiver = temp.OTAReceiver
	r.tokenID = temp.TokenID
	r.tokenAmount = temp.TokenAmount
	r.txReqID = temp.TxReqID
	r.shardID = temp.ShardID
	return nil
}

func (r *RefundAddLiquidity) FromStringArr(source []string) error {
	if len(source) != 8 {
		return fmt.Errorf("Receive length %v but expect %v", len(source), 8)
	}
	if source[0] != strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta) {
		return fmt.Errorf("Receive metaType %v but expect %v", source[0], metadataCommon.PDexV3AddLiquidityMeta)
	}
	if source[1] != RefundStatus {
		return fmt.Errorf("Receive status %v but expect %v", source[1], RefundStatus)
	}
	if source[2] == "" {
		return errors.New("Pair hash is invalid")
	}
	r.pairHash = source[2]
	tokenID, err := common.Hash{}.NewHashFromStr(source[3])
	if err != nil {
		return err
	}
	if tokenID.IsZeroValue() {
		return errors.New("TokenID is empty")
	}
	r.tokenID = source[3]
	tokenAmount, err := strconv.ParseUint(source[4], 10, 32)
	if err != nil {
		return err
	}
	r.tokenAmount = tokenAmount
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(source[5])
	if err != nil {
		return err
	}
	if !otaReceiver.IsValid() {
		return errors.New("receiver Address is invalid")
	}
	r.otaReceiver = source[5]
	r.txReqID = source[6]
	shardID, err := strconv.Atoi(source[7])
	if err != nil {
		return err
	}
	r.shardID = byte(shardID)
	return nil
}

func (r *RefundAddLiquidity) StringArr() []string {
	metaDataType := strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta)
	res := []string{metaDataType, RefundStatus}
	res = append(res, r.pairHash)
	res = append(res, r.tokenID)
	tokenAmount := strconv.FormatUint(r.tokenAmount, 10)
	res = append(res, tokenAmount)
	res = append(res, r.otaReceiver)
	res = append(res, r.txReqID)
	shardID := strconv.Itoa(int(r.shardID))
	res = append(res, shardID)
	return res
}
