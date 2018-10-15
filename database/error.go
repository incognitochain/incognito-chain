package database

import "fmt"

const (
	DriverExistErr       = "DriverExistErr"
	DriverNotRegisterErr = "DriverNotRegisterErr"

	// LevelDB
	OpenDbErr     = "OpenDbErr"
	NotExistValue = "NotExistValue"

	// BlockChain err

)

var ErrCodeMessage = map[string]struct {
	code    int
	message string
}{
	// -1xxx
	DriverExistErr:       {-1000, "Driver is already registered"},
	DriverNotRegisterErr: {-1001, "Driver is not registered"},

	// -2xxx
	OpenDbErr:     {-2000, "Open database error"},
	NotExistValue: {-2001, "Value is not existed"},
}

type DatabaseError struct {
	err     error
	code    int
	message string
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.message)
}

func NewDatabaseError(key string, err error) *DatabaseError {
	return &DatabaseError{
		err:     err,
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
