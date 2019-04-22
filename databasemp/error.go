package databasemp

import (
	"fmt"
	"github.com/pkg/errors"
)


const (
	DriverExistErr = iota
	DriverNotRegisterErr
	
	// LevelDB
	OpenDbErr
	NotExistValue
	LvDbNotFound
	
	// BlockChain err
	NotImplHashMethod
	UnexpectedError
	KeyExisted

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
	LvDbNotFound:  {-2002, "lvdb not found"},
	
	// -3xxx blockchain
	NotImplHashMethod: {-3000, "Data does not implement this method"},
	UnexpectedError:   {-3002, "Unexpected error"},
	KeyExisted:        {-3003, "PubKey already existed in database"},
	
}

type DatabaseMempoolError struct {
	err     error
	code    int
	message string
}

func (e DatabaseMempoolError) GetErrorCode() int {
	return e.code
}

func (e DatabaseMempoolError) Error() string {
	return fmt.Sprintf("%d: %+v", e.code, e.err)
}

func NewDatabaseMempoolError(key int, err error) *DatabaseMempoolError {
	return &DatabaseMempoolError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
