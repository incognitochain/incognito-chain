package database

import "fmt"

const ()

var ErrCodeMessage = map[string]struct {
	code    int
	message string
}{
	// -1xxx
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
