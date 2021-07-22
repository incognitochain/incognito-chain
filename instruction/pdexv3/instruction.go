package pdexv3

import (
	"errors"
	"strconv"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type Base struct {
	metaData metadataCommon.Metadata
	txReqID  string
	shardID  byte
}

func NewBase() *Base {
	return &Base{}
}

func NewBaseWithValue(metaData metadataCommon.Metadata, txReqID string, shardID byte) *Base {
	return &Base{
		metaData: metaData,
		txReqID:  txReqID,
		shardID:  shardID,
	}
}

func (base *Base) FromStringArr(source []string) error {
	if len(source) < 3 {
		return errors.New("Invalid length of instruction")
	}
	base.metaData.FromStringArr(source[:len(source)-2])
	base.txReqID = source[len(source)-2]
	shardID, err := strconv.Atoi(source[len(source)-1])
	if err != nil {
		return err
	}
	base.shardID = byte(shardID)
	return nil
}

func (base *Base) StringArr() []string {
	res := base.metaData.StringArr()
	res = append(res, base.txReqID)
	shardID := strconv.Itoa(int(base.shardID))
	res = append(res, shardID)
	return res
}

func (base *Base) MetaData() metadataCommon.Metadata {
	return base.metaData
}

func (base *Base) TxReqID() string {
	return base.txReqID
}

func (base *Base) ShardID() byte {
	return base.shardID
}
