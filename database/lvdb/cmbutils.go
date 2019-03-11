package lvdb

import (
	"encoding/binary"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
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
	reserve []byte,
	members [][]byte,
	capital uint64,
	txHash []byte,
	state uint8,
	fine uint64,
) ([]byte, error) {
	// Add reserve account
	values := []byte{}
	values = append(values, reserve...)

	// Add capital
	capitalInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(capitalInBytes, capital)
	values = append(values, capitalInBytes...)

	// Add fine
	fineInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(fineInBytes, fine)
	values = append(values, fineInBytes...)

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

func parseCMBInitValue(value []byte) ([]byte, [][]byte, uint64, []byte, uint8, uint64, error) {
	// Get reserve
	if len(value) < PaymentAddressLen {
		return nil, nil, 0, nil, 0, 0, errors.Errorf("error parsing cmb value")
	}
	reserve := value[:PaymentAddressLen]

	// Get capital
	value = value[PaymentAddressLen:]
	if len(value) < 8 {
		return nil, nil, 0, nil, 0, 0, errors.Errorf("error parsing cmb value")
	}
	capital := binary.LittleEndian.Uint64(value)

	// Get fine
	value = value[8:]
	if len(value) < 8 {
		return nil, nil, 0, nil, 0, 0, errors.Errorf("error parsing cmb value")
	}
	fine := binary.LittleEndian.Uint64(value)

	// Last byte: state
	state := uint8(value[len(value)-1])

	// Last 32 bytes (not counting approvalByte): txHash
	txHash := value[len(value)-common.HashSize-1 : len(value)-1]

	// The rest: members
	value = value[8 : len(value)-common.HashSize-2]
	if len(value)%PaymentAddressLen != 0 {
		return nil, nil, 0, nil, 0, 0, errors.Errorf("error parsing cmb value")
	}
	numMembers := len(value) / PaymentAddressLen
	members := [][]byte{}

	for i := 0; i < numMembers; i += 1 {
		member := make([]byte, PaymentAddressLen)
		copy(member, value[i*PaymentAddressLen:(i+1)*PaymentAddressLen])
		members = append(members, member)
	}
	return reserve, members, capital, txHash, state, fine, nil
}

func getCMBResponseKey(mainAccount, approver []byte) []byte {
	key := append(cmbResponsePrefix, mainAccount...)
	key = append(key, Splitter...)
	key = append(key, approver...)
	return key
}

func getCMBDepositSendKey(contractID []byte) []byte {
	key := append(cmbDepositSendKeyPrefix, contractID...)
	return key
}

func getCMBDepositSendValue(txHash []byte) []byte {
	return txHash
}

func getCMBWithdrawRequestKey(contractID []byte) []byte {
	key := append(cmbWithdrawRequestPrefix, contractID...)
	return key
}

func getCMBWithdrawRequestValue(txHash []byte, state uint8) []byte {
	values := make([]byte, len(txHash), len(txHash)+1)
	copy(values, txHash)
	values = append(values, common.Uint8ToBytes(state)...)
	return values
}

func parseWithdrawRequestValue(values []byte) ([]byte, uint8, error) {
	if len(values) != 1+common.HashSize {
		return nil, 0, errors.Errorf("Error parsing withdraw request")
	}
	txHash := values[:len(values)-2]
	state := uint8(values[len(values)-1])
	return txHash, state, nil
}

func getCMBNoticeKey(blockHeight uint64, txReqHash []byte) []byte {
	// 0xjackalope: convert to uint32 before saving to db
	key := cmbNoticePrefix
	key = append(key, common.Uint32ToBytes(uint32(blockHeight))...)
	key = append(key, Splitter...)
	key = append(key, txReqHash...)
	return key
}
