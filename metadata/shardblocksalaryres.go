package metadata

import (
	"errors"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
)

type ShardBlockSalaryRes struct {
	MetadataBase
	ShardBlockHeight uint64
	ProducerAddress  privacy.PaymentAddress
}

func NewShardBlockSalaryRes(
	shardBlockHeight uint64,
	producerAddress privacy.PaymentAddress,
	metaType int,
) *ShardBlockSalaryRes {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &ShardBlockSalaryRes{
		ShardBlockHeight: shardBlockHeight,
		ProducerAddress:  producerAddress,
		MetadataBase:     metadataBase,
	}
}

func (sbsRes *ShardBlockSalaryRes) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes *ShardBlockSalaryRes) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return false, nil
}

func (sbsRes *ShardBlockSalaryRes) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(sbsRes.ProducerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	if len(sbsRes.ProducerAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	if sbsRes.ShardBlockHeight == 0 {
		return false, false, errors.New("Wrong request info's shard block height")
	}
	return false, true, nil
}

func (sbsRes *ShardBlockSalaryRes) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes *ShardBlockSalaryRes) Hash() *common.Hash {
	record := sbsRes.ProducerAddress.String()
	record += string(sbsRes.ShardBlockHeight)
	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
