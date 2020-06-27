package instruction

import (
	"strconv"
)

//Random Instruction which get nonce from bitcoin block
type RandomInstruction struct {
	BtcNonce       int64
	BtcBlockHeight int
	CheckPointTime int64
	BtcBlockTime   int64
}

func NewRandomInstruction() *RandomInstruction {
	s := &RandomInstruction{}
	return s
}

func (s *RandomInstruction) GetType() string {
	return RANDOM_ACTION
}

func (s *RandomInstruction) SetNonce(n int64) *RandomInstruction {
	s.BtcNonce = n
	return s
}

func (s *RandomInstruction) SetBtcBlockHeight(n int) *RandomInstruction {
	s.BtcBlockHeight = n
	return s
}

func (s *RandomInstruction) SetBtcBlockTime(n int64) *RandomInstruction {
	s.BtcBlockTime = n
	return s
}

func (s *RandomInstruction) SetCheckPointTime(n int64) *RandomInstruction {
	s.CheckPointTime = n
	return s
}

func (s *RandomInstruction) ToString() []string {
	strs := []string{}
	strs = append(strs, RANDOM_ACTION)
	strs = append(strs, strconv.Itoa(int(s.BtcNonce)))
	strs = append(strs, strconv.Itoa(int(s.BtcBlockHeight)))
	strs = append(strs, strconv.Itoa(int(s.CheckPointTime)))
	strs = append(strs, strconv.Itoa(int(s.BtcBlockTime)))
	return strs
}
