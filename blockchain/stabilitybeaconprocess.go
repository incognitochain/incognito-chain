package blockchain

import (
	"github.com/constant-money/constant-chain/database"
)

func (bsb *BestStateBeacon) processStabilityInstruction(inst []string, db database.DatabaseInterface) error {
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	default:
		return nil
	}
	return nil
}
