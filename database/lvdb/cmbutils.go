package lvdb

import (
	"encoding/binary"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

const PaymentAddressLen = 66

func errUnexpected(err error, content string) *database.DatabaseError {
	if err == nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Errorf(content))
	}
	return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, content))
}

func getCMBInitKey(mainAccount []byte) []byte {
	// Add main account
	key := append(cmbPrefix, mainAccount...)
	return key
}

func getCMBInitValue(
	members [][]byte,
	capital uint64,
	txHash []byte,
	state uint8,
) ([]byte, error) {
	// Add capital
	values := []byte{}
	capitalInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(capitalInBytes, capital)
	values = append(values, capitalInBytes...)

	// Add members
	for _, member := range members {
		if len(member) != PaymentAddressLen {
			return nil, errors.Errorf("provided bytes are not payment address")
		}
		values = append(values, member...)
	}

	// Add txHash
	values = append(values, txHash...)

	// Add state
	values = append(values, common.Uint8ToBytes(state)...)
	return values, nil
}

func parseCMBInitValue(value []byte) ([][]byte, uint64, []byte, uint8, error) {
	// Get capital (first 8 bytes)
	if len(value) < 8 {
		return nil, 0, nil, 0, errors.Errorf("error parsing cmb value")
	}
	capital := binary.LittleEndian.Uint64(value)

	// Last byte: state
	state := uint8(value[len(value)-1])

	// Last 32 bytes (not counting approvalByte): txHash
	txHash := value[len(value)-common.HashSize-1 : len(value)-1]

	// The rest: members
	value = value[8 : len(value)-common.HashSize-2]
	if len(value)%PaymentAddressLen != 0 {
		return nil, 0, nil, 0, errors.Errorf("error parsing cmb value")
	}
	numMembers := len(value) / PaymentAddressLen
	members := [][]byte{}

	for i := 0; i < numMembers; i += 1 {
		member := make([]byte, PaymentAddressLen)
		copy(member, value[i*PaymentAddressLen:(i+1)*PaymentAddressLen])
		members = append(members, member)
	}
	return members, capital, txHash, state, nil
}

func getCMBResponseKey(mainAccount, approver []byte) []byte {
	key := append(cmbResponsePrefix, mainAccount...)
	key = append(key, Splitter...)
	key = append(key, approver...)
	return key
}
