package blockchain

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/common"

	"github.com/ninjadotorg/constant/metadata"
)

const Decimals = uint64(10000) // Each float number is multiplied by this value to store as uint64

func GetShardIDFromLastByte(b byte) byte {
	return byte(int(b) % common.SHARD_NUMBER)
}

func GetInterestAmount(principle, interestRate uint64) uint64 {
	return principle * interestRate / Decimals
}

func parseMetadata(meta interface{}) (metadata.Metadata, error) {
	if meta == nil {
		return nil, nil
	}

	mtTemp := map[string]interface{}{}
	metaInBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaInBytes, &mtTemp)
	if err != nil {
		return nil, err
	}
	var md metadata.Metadata
	switch int(mtTemp["Type"].(float64)) {
	case metadata.BuyFromGOVRequestMeta:
		md = &metadata.BuySellRequest{}

	case metadata.BuyBackRequestMeta:
		md = &metadata.BuyBackRequest{}

	case metadata.BuyFromGOVResponseMeta:
		md = &metadata.BuySellResponse{}

	case metadata.BuyBackResponseMeta:
		md = &metadata.BuyBackResponse{}

	case metadata.LoanRequestMeta:
		md = &metadata.LoanRequest{}

	case metadata.LoanResponseMeta:
		md = &metadata.LoanResponse{}

	default:
		return nil, errors.New("Could not parse metadata with known types.")
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

// blockExists determines whether a block with the given hash exists either in
// the main chain or any side chains.
//
// This function is safe for concurrent access.
// func (self *BlockChain) BlockExists(hash *common.Hash) (bool, error) {
// 	result, err := self.config.DataBase.HasBlock(hash)
// 	if err != nil {
// 		return false, NewBlockChainError(UnExpectedError, err)
// 	} else {
// 		return result, nil
// 	}
// }
