package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
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

func (or *OracleReward) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
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
	record += string(or.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
