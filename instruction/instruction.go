package instruction

type Inst interface {
	GetType() uint
	ToString() []string
}

const (
	I_SWAP          = 1
	I_RANDOM        = 1 << 1
	I_STAKE         = 1 << 2
	I_ASSIGN        = 1 << 3
	I_STOPAUTOSTAKE = 1 << 4
)

type InstManager struct {
	instructions []Inst
}

func ImportInstructionFromStringArray(instStr [][]string) *InstManager {
	insts := InstManager{}
	for _, inst := range instStr {
		insts.instructions = append(insts.instructions, processInstructionString(inst))
	}
	return &insts
}

func processInstructionString(str []string) Inst {
	return nil
}

func (s *InstManager) filter(instTypeMap uint, cb func(Inst) bool) {
	for _, v := range s.instructions {
		if v.GetType()&instTypeMap != 0 {
			cb(v)
		}
	}
}
