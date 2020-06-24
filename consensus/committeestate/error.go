package committeestate

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	ErrSwapInstructionSanity = iota
	ErrStakeInstructionSanity
	ErrStopAutoStakeInstructionSanity
	ErrAssignInstructionSanity
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	ErrSwapInstructionSanity:          {-1000, "swap instruction sanity error"},
	ErrStakeInstructionSanity:         {-1001, "stake instruction sanity error"},
	ErrStopAutoStakeInstructionSanity: {-1002, "stop auto stake sanity error"},
	ErrAssignInstructionSanity:        {-1003, "assign sanity error"},
}

type CommitteeStateError struct {
	Code    int
	Message string
	err     error
}

func (e CommitteeStateError) Error() string {
	return fmt.Sprintf("%d: %s \n %+v", e.Code, e.Message, e.err)
}

func NewCommitteeStateError(key int, err error) *CommitteeStateError {
	return &CommitteeStateError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
