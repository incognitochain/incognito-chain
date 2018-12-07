package database

import "fmt"

const (
	DriverExistErr = iota
	DriverNotRegisterErr

	// LevelDB
	OpenDbErr
	NotExistValue

	// BlockChain err
	NotImplHashMethod
	BlockExisted
	UnexpectedError
	KeyExisted

	//voting err
	NotEnoughCandidateDCB
	NotEnoughCandidateGOV
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	// -1xxx driver
	DriverExistErr:       {-1000, "Driver is already registered"},
	DriverNotRegisterErr: {-1001, "Driver is not registered"},

	// -2xxx levelDb
	OpenDbErr:     {-2000, "Open database error"},
	NotExistValue: {-2001, "H is not existed"},

	// -3xxx blockchain
	NotImplHashMethod: {-3000, "Data does not implement Hash() method"},
	BlockExisted:      {-3001, "Block already existed"},
	UnexpectedError:   {-3002, "Unexpected error"},
	KeyExisted:        {-3003, "PubKey already existed in database"},

	// -4xxx voting
	NotEnoughCandidateDCB: {-4000, "Note enough candidate for DCB Board"},
	NotEnoughCandidateGOV: {-4001, "Note enough candidate for GOV Board"},
}

type DatabaseError struct {
	err     error
	code    int
	message string
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("%+v: %+v", e.code, e.message)
}

func NewDatabaseError(key int, err error) *DatabaseError {
	return &DatabaseError{
		err:     err,
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
