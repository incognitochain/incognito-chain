package instruction

import (
	"strconv"
)

//Random Instruction which get nonce from bitcoin block
type RandomInstruction struct {
	BtcNonce       uint64
	BtcBlockHeight uint64
	CheckpointTime uint64
	BtcBlockTime   uint64
}

func NewRandomInst() *RandomInstruction {
	s := &RandomInstruction{}
	return s
}

func (s *RandomInstruction) GetType() string {
	return RANDOM_ACTION
}

func (s *RandomInstruction) SetNonce(n uint64) *RandomInstruction {
	s.BtcNonce = n
	return s
}

func (s *RandomInstruction) SetBtcBlockHeight(n uint64) *RandomInstruction {
	s.BtcNonce = n
	return s
}

func (s *RandomInstruction) SetBtcBlockTime(n uint64) *RandomInstruction {
	s.BtcBlockTime = n
	return s
}

func (s *RandomInstruction) SetCheckpointTime(n uint64) *RandomInstruction {
	s.CheckpointTime = n
	return s
}

func (s *RandomInstruction) ToString() []string {
	strs := []string{}
	strs = append(strs, RANDOM_ACTION)
	strs = append(strs, strconv.Itoa(int(s.BtcNonce)))
	strs = append(strs, strconv.Itoa(int(s.BtcBlockHeight)))
	strs = append(strs, strconv.Itoa(int(s.CheckpointTime)))
	strs = append(strs, strconv.Itoa(int(s.BtcBlockTime)))
	return strs
}
