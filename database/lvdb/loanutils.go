package lvdb

import (
	"encoding/binary"

	"github.com/constant-money/constant-chain/common"
	"github.com/pkg/errors"
)

func getLoanPaymentValue(
	principle, interest, deadline uint64,
) []byte {
	// Add principle
	values := []byte{}
	values = append(values, common.Uint64ToBytes(principle)...)

	// Add interest
	values = append(values, common.Uint64ToBytes(interest)...)

	// Add deadline
	values = append(values, common.Uint64ToBytes(deadline)...)

	return values
}

func parseLoanPaymentValue(value []byte) (uint64, uint64, uint64, error) {
	if len(value) < 8*3 { // priciple, interest and deadline
		return 0, 0, 0, errors.Errorf("Error parsing loan payment value")
	}

	// Get principle
	principle := common.BytesToUint64(value)

	// Get interest
	value = value[8:]
	interest := common.BytesToUint64(value)

	// The rest: deadline
	value = value[8:]
	deadline := binary.LittleEndian.Uint64(value)
	return principle, interest, deadline, nil
}
