package common

const (
	UnexpectedErr = iota
	Base58CheckDecodeErr
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedErr: {-10001, "Unexpected error"},
	Base58CheckDecodeErr : {-10002, "Unexpected error"},
}

//type CashecError struct {
//	code    int
//	message string
//	err     error
//}
//
//func (e CashecError) Error() string {
//	return fmt.Sprintf("%+v: %+v", e.code, e.message)
//}
//
//func (e CashecError) GetCode() int {
//	return e.code
//}
//
//func NewCashecError(key int, err error) *CashecError {
//	return &CashecError{
//		err:     errors.Wrap(err, ErrCodeMessage[key].message),
//		code:    ErrCodeMessage[key].code,
//		message: ErrCodeMessage[key].message,
//	}
//}


