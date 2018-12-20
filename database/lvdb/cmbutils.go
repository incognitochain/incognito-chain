package lvdb

import (
	"encoding/binary"

	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

const PubkeyLen = 33

func errUnexpected(err error, content string) *database.DatabaseError {
	if err == nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Errorf(content))
	}
	return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, content))
}

func getCMBValue(capital uint64, members [][]byte) ([]byte, error) {
	// Add capital
	values := []byte{}
	capitalInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(capitalInBytes, capital)
	values = append(values, capitalInBytes...)

	// Add members
	for _, member := range members {
		if len(member) != PubkeyLen {
			return nil, errors.Errorf("provided bytes are not pubkey")
		}
		values = append(values, member...)
	}
	return values, nil
}
