package blockchain

import "github.com/ninjadotorg/constant/common"

type BlockHeaderBeacon struct {
	BlockHeaderGeneric
	TestParam string
}

func (f BlockHeaderBeacon) Hash() common.Hash {
	return common.Hash{}
}

func (f BlockHeaderBeacon) UnmarshalJSON([]byte) error {
	return nil
}
