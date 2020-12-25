package committeestate

import "github.com/incognitochain/incognito-chain/instruction"

type AssignRule interface {
	GenInstruction([]string, int64, int, map[string]uint64, uint64, byte) (*instruction.AssignInstruction, error)
}
