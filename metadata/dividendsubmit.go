package metadata

import (
	"encoding/json"
	"strconv"

	"github.com/big0t/constant-chain/common"
)

//func (ds *DividendSubmit) Hash() *common.Hash {
//	record := ds.TokenID.String()
//	record += strconv.FormatUint(ds.DividendID, 10)
//	record += strconv.FormatUint(ds.TotalTokenAmount, 10)
//	record += string(ds.ShardID)
//
//	// final hash
//	record += ds.MetadataBase.Hash().String()
//	hash := common.DoubleHashH([]byte(record))
//	return &hash
//}
//
//func (ds *DividendSubmit) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
//	// Check shardID
//	if ds.ShardID != shardID {
//		return false, errors.Errorf("Expect shard id %d, got %d", shardID, ds.ShardID)
//	}
//
//	// Check latest dividend proposal
//	forDCB := ds.TokenID.IsEqual(&common.DCBTokenID)
//	id, _ := bcr.GetLatestDividendProposal(forDCB)
//	if ds.DividendID != id {
//		return false, errors.Errorf("Expect dividend id %d, got %d", id, ds.DividendID)
//	}
//
//	totalTokenAmount, _, _, err := bcr.GetAmountPerAccount(ds.TokenID)
//	if err != nil {
//		return false, err
//	}
//	if ds.TotalTokenAmount != totalTokenAmount {
//		return false, errors.Errorf("Expect total token amount %d, got %d", totalTokenAmount, ds.TotalTokenAmount)
//	}
//
//	return false, nil
//}
//
//func (ds *DividendSubmit) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
//	return false, true, nil // No need to check for fee
//}
//
//func (ds *DividendSubmit) ValidateMetadataByItself() bool {
//	if !ds.TokenID.IsEqual(&common.DCBTokenID) && ds.TokenID.IsEqual(&common.GOVTokenID) {
//		return false
//	}
//	return true
//}
//
//func (ds *DividendSubmit) CheckTransactionFee(tr Transaction, minFee uint64) bool {
//	return true
//}

// Each shard submits this action to aggregate total outstanding tokens across the whole network
type DividendSubmit struct {
	TokenID          *common.Hash
	DividendID       uint64
	TotalTokenAmount uint64
	ShardID          byte
}

func BuildDividendSubmitInst(
	tokenID *common.Hash,
	dividendID uint64,
	totalTokenAmount uint64,
	shardID byte,
) ([][]string, error) {
	ds := &DividendSubmit{
		TokenID:          tokenID,
		DividendID:       dividendID,
		TotalTokenAmount: totalTokenAmount,
		ShardID:          shardID,
	}
	lrActionValue, err := getDividendSubmitActionValue(ds)
	if err != nil {
		return nil, err
	}
	lrAction := []string{strconv.Itoa(DividendSubmitMeta), lrActionValue}
	return [][]string{lrAction}, nil
}

func getDividendSubmitActionValue(ds *DividendSubmit) (string, error) {
	value, err := json.Marshal(ds)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func ParseDividendSubmitActionValue(value string) (*DividendSubmit, error) {
	ds := &DividendSubmit{}
	err := json.Unmarshal([]byte(value), ds)
	if err != nil {
		return nil, err
	}
	return ds, nil
}
