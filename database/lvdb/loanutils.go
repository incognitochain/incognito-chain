package lvdb

import (
	"encoding/binary"

	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"
)

func getLoanPaymentValue(
	principle, interest uint64,
	deadline uint32,
) []byte {
	// Add principle
	values := []byte{}
	principleInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(principleInBytes, principle)
	values = append(values, principleInBytes...)

	// Add interest
	interestInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(interestInBytes, interest)
	values = append(values, interestInBytes...)

	// Add deadline
	values = append(values, common.Uint32ToBytes(deadline)...)

	return values
}

func parseLoanPaymentValue(value []byte) (uint64, uint64, uint32, error) {
	// Get principle
	if len(value) < 8 {
		return 0, 0, 0, errors.Errorf("Error parsing loan payment value")
	}
	principle := binary.LittleEndian.Uint64(value)

	// Get interest
	value = value[8:]
	if len(value) < 8 {
		return 0, 0, 0, errors.Errorf("Error parsing loan payment value")
	}
	interest := binary.LittleEndian.Uint64(value)

	// The rest: deadline
	value = value[8:]
	if len(value) < 4 {
		return 0, 0, 0, errors.Errorf("Error parsing loan payment value")
	}
	deadline := common.BytesToUint32(value)
	return principle, interest, deadline, nil
}
