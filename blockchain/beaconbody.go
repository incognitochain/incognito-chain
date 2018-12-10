package blockchain

import "github.com/ninjadotorg/constant/common"

type BlockBodyBeacon struct {
	ParamsInstruction []string
}

func (f BlockBodyBeacon) Hash() common.Hash {
	return common.Hash{}
}

func (f BlockBodyBeacon) UnmarshalJSON(data []byte) error {
	return nil
}
