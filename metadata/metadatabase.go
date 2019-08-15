package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"strconv"
)

type MetadataBase struct {
	Type int
}

func NewMetadataBase(thisType int) *MetadataBase {
	return &MetadataBase{Type: thisType}
}

func (mb MetadataBase) IsMinerCreatedMetaType() bool {
	metaType := mb.GetType()
	for _, mType := range minerCreatedMetaTypes {
		if metaType == mType {
			return true
		}
	}
	return false
}

func (mb *MetadataBase) CalculateSize() uint64 {
	return 0
}

func (mb *MetadataBase) Validate() error {
	return nil
}

func (mb *MetadataBase) Process() error {
	return nil
}

func (mb MetadataBase) GetType() int {
	return mb.Type
}

func (mb MetadataBase) Hash() *common.Hash {
	record := strconv.Itoa(mb.Type)
	hash := common.HashH([]byte(record))
	return &hash
}

func (mb MetadataBase) CheckTransactionFee(tx Transaction, minFeePerKbTx uint64) bool {
	txFee := tx.GetTxFee()
	fullFee := minFeePerKbTx * tx.GetTxActualSize()
	return !(txFee < fullFee)
}

func (mb *MetadataBase) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	return [][]string{}, nil
}

func (mb MetadataBase) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	txr Transaction,
	bcr BlockchainRetriever,
	accumulatedValues *AccumulatedValues,
) (bool, error) {
	return true, nil
}
