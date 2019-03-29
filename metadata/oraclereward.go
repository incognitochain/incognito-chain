package metadata

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
)

type OracleReward struct {
	MetadataBase
	OracleFeedTxID common.Hash
}

func NewOracleReward(oracleFeedTxID common.Hash, metaType int) *OracleReward {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &OracleReward{
		OracleFeedTxID: oracleFeedTxID,
		MetadataBase:   metadataBase,
	}
}

func (or *OracleReward) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (or *OracleReward) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via OracleFeedTxID) in current block
	return false, nil
}

func (or *OracleReward) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (or *OracleReward) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (or *OracleReward) Hash() *common.Hash {
	record := or.OracleFeedTxID.String()
	record += or.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (or *OracleReward) CalculateSize() uint64 {
	return calculateSize(or)
}
