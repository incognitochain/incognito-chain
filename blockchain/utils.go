package blockchain

import "github.com/ninjadotorg/constant/common"

const Decimals = uint64(10000) // Each float number is multiplied by this value to store as uint64

func GetInterestAmount(principle, interestRate uint64) uint64 {
	return principle * interestRate / Decimals
}

// blockExists determines whether a block with the given hash exists either in
// the main chain or any side chains.
//
// This function is safe for concurrent access.
func (self *BlockChain) BlockExists(hash *common.Hash) (bool, error) {
	result, err := self.config.DataBase.HasBlock(hash)
	if err != nil {
		return false, NewBlockChainError(UnExpectedError, err)
	} else {
		return result, nil
	}
}
