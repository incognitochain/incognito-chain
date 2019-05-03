package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

type ReturnStakingMeta struct {
	MetadataBase
	TxID            string
	ProducerAddress privacy.PaymentAddress
}

func NewReturnStaking(
	txID string,
	producerAddress privacy.PaymentAddress,
	metaType int,
) *ReturnStakingMeta {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &ReturnStakingMeta{
		TxID:            txID,
		ProducerAddress: producerAddress,
		MetadataBase:    metadataBase,
	}
}

func (sbsRes *ReturnStakingMeta) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes *ReturnStakingMeta) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with request tx (via RequestedTxID) in current block
	return false, nil
}

func (sbsRes *ReturnStakingMeta) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(sbsRes.ProducerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	if len(sbsRes.ProducerAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	// if sbsRes.ShardBlockHeight == 0 {
	// 	return false, false, errors.New("Wrong request info's shard block height")
	// }
	return false, true, nil
}

func (sbsRes *ReturnStakingMeta) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes *ReturnStakingMeta) Hash() *common.Hash {
	record := sbsRes.ProducerAddress.String()
	record += sbsRes.TxID

	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

//validate in shard block
func (sbsRes *ReturnStakingMeta) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	//TODO: check if tx staking is existed
	// check if producer is swap
	// check if producer is in trhis shard

	return true, nil
}
