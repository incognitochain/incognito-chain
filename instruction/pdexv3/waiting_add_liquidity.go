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

type WaitingAddLiquidity struct {
	poolPairID      string // only "" for the first contribution of pool
	pairHash        string
	receiverAddress string // receive nfct
	refundAddress   string // refund pToken
	tokenID         string
	tokenAmount     uint64
	amplifier       uint // only set for the first contribution
	txReqID         string
	shardID         byte
}

func NewWaitingAddLiquidity() *WaitingAddLiquidity {
	return &WaitingAddLiquidity{}
}

func NewWaitingAddLiquidityFromMetadata(
	metaData metadataPdexV3.AddLiquidity,
	txReqID string, shardID byte,
) *WaitingAddLiquidity {
	return NewWaitingAddLiquidityWithValue(
		metaData.PoolPairID(),
		metaData.PairHash(),
		metaData.ReceiverAddress(),
		metaData.RefundAddress(),
		metaData.TokenID(),
		txReqID,
		metaData.TokenAmount(),
		metaData.Amplifier(),
		shardID,
	)
}

func NewWaitingAddLiquidityWithValue(
	poolPairID, pairHash,
	receiverAddress, refundAddress,
	tokenID, txReqID string,
	tokenAmount uint64, amplifier uint,
	shardID byte,
) *WaitingAddLiquidity {
	return &WaitingAddLiquidity{
		poolPairID:      poolPairID,
		pairHash:        pairHash,
		receiverAddress: receiverAddress,
		refundAddress:   refundAddress,
		tokenID:         tokenID,
		tokenAmount:     tokenAmount,
		amplifier:       amplifier,
		txReqID:         txReqID,
		shardID:         shardID,
	}
}

func (w *WaitingAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID      string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash        string `json:"PairHash"`
		ReceiverAddress string `json:"ReceiverAddress"` // receive nfct
		RefundAddress   string `json:"RefundAddress"`   // refund pToken
		TokenID         string `json:"TokenID"`
		TokenAmount     uint64 `json:"TokenAmount"`
		Amplifier       uint   `json:"Amplifier"` // only set for the first contribution
		TxReqID         string `json:"TxReqID"`
		ShardID         byte   `json:"ShardID"`
	}{
		PoolPairID:      w.poolPairID,
		PairHash:        w.pairHash,
		ReceiverAddress: w.receiverAddress,
		RefundAddress:   w.refundAddress,
		TokenID:         w.tokenID,
		TokenAmount:     w.tokenAmount,
		Amplifier:       w.amplifier,
		TxReqID:         w.txReqID,
		ShardID:         w.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (w *WaitingAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID      string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash        string `json:"PairHash"`
		ReceiverAddress string `json:"ReceiverAddress"` // receive nfct
		RefundAddress   string `json:"RefundAddress"`   // refund pToken
		TokenID         string `json:"TokenID"`
		TokenAmount     uint64 `json:"TokenAmount"`
		Amplifier       uint   `json:"Amplifier"` // only set for the first contribution
		TxReqID         string `json:"TxReqID"`
		ShardID         byte   `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	w.poolPairID = temp.PoolPairID
	w.pairHash = temp.PairHash
	w.receiverAddress = temp.ReceiverAddress
	w.refundAddress = temp.RefundAddress
	w.tokenID = temp.TokenID
	w.tokenAmount = temp.TokenAmount
	w.amplifier = temp.Amplifier
	w.txReqID = temp.TxReqID
	w.shardID = temp.ShardID
	return nil
}

func (w *WaitingAddLiquidity) FromStringArr(source []string) error {
	if len(source) != 10 {
		return fmt.Errorf("Receive length %v but expect %v", len(source), 9)
	}
	if source[0] != strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta) {
		return fmt.Errorf("Receive metaType %v but expect %v", source[0], metadataCommon.PDexV3AddLiquidityMeta)
	}
	w.poolPairID = source[1]
	if source[2] == "" {
		return errors.New("Pair hash is invalid")
	}
	w.pairHash = source[2]
	tokenID, err := common.Hash{}.NewHashFromStr(source[3])
	if err != nil {
		return err
	}
	if tokenID.IsZeroValue() {
		return errors.New("TokenID is empty")
	}
	w.tokenID = source[3]
	tokenAmount, err := strconv.ParseUint(source[4], 10, 32)
	if err != nil {
		return err
	}
	w.tokenAmount = tokenAmount
	amplifier, err := strconv.ParseUint(source[5], 10, 32)
	if err != nil {
		return err
	}
	w.amplifier = uint(amplifier)
	receiverAddress := privacy.OTAReceiver{}
	err = receiverAddress.FromString(source[6])
	if err != nil {
		return err
	}
	w.receiverAddress = source[6]

	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(source[6])
	if err != nil {
		return err
	}
	w.receiverAddress = source[7]
	w.txReqID = source[8]
	shardID, err := strconv.Atoi(source[9])
	if err != nil {
		return err
	}
	w.shardID = byte(shardID)

	return nil
}

func (w *WaitingAddLiquidity) StringArr() ([]string, error) {
	metaDataType := strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta)
	res := []string{metaDataType}
	res = append(res, w.poolPairID)
	res = append(res, w.pairHash)
	res = append(res, w.tokenID)
	tokenAmount := strconv.FormatUint(w.tokenAmount, 10)
	res = append(res, tokenAmount)
	amplifier := strconv.FormatUint(uint64(w.amplifier), 10)
	res = append(res, amplifier)
	res = append(res, w.receiverAddress)
	res = append(res, w.refundAddress)
	res = append(res, w.txReqID)
	shardID := strconv.Itoa(int(w.shardID))
	res = append(res, shardID)
	return res, nil
}
