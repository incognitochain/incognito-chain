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
	ErrCommitBeaconCommitteeState
	ErrUpdateCommitteeState
	ErrGenerateBeaconCommitteeStateHash
	ErrCommitShardCommitteeState
	ErrUpdateShardCommitteeState
	ErrGenerateShardCommitteeStateHash
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	ErrSwapInstructionSanity:          {-1000, "swap instruction sanity error"},
	ErrStakeInstructionSanity:         {-1001, "stake instruction sanity error"},
	ErrStopAutoStakeInstructionSanity: {-1002, "stop auto stake sanity error"},
	ErrAssignInstructionSanity:        {-1003, "assign sanity error"},

	ErrCommitBeaconCommitteeState:       {-2000, "commit beacon committee state error"},
	ErrUpdateCommitteeState:             {-2001, "update committee state error"},
	ErrGenerateBeaconCommitteeStateHash: {-2002, "generate beacon committee state root hash"},

	ErrCommitShardCommitteeState:       {-3000, "commit shard committee state"},
	ErrUpdateShardCommitteeState:       {-3001, " update shard committee state error"},
	ErrGenerateShardCommitteeStateHash: {-3002, " generate shard committee state root hash"},
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
