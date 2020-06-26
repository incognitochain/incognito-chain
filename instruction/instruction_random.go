package instruction

import (
	"strconv"
)

//Random Instruction which get nonce from bitcoin block
type RandomInst struct {
	btcNonce       uint64
	btcBlockHeight uint64
	checkpointTime uint64
	btcBlockTime   uint64
}

func NewRandomInst() *RandomInst {
	s := &RandomInst{}
	return s
}

func (s *RandomInst) GetType() string {
	return RANDOM_ACTION
}

func (s *RandomInst) SetNonce(n uint64) *RandomInst {
	s.btcNonce = n
	return s
}

func (s *RandomInst) SetBtcBlockHeight(n uint64) *RandomInst {
	s.btcNonce = n
	return s
}

func (s *RandomInst) SetBtcBlockTime(n uint64) *RandomInst {
	s.btcBlockTime = n
	return s
}

func (s *RandomInst) SetCheckpointTime(n uint64) *RandomInst {
	s.checkpointTime = n
	return s
}

func (s *RandomInst) ToString() []string {
	strs := []string{}
	strs = append(strs, "random")
	strs = append(strs, strconv.Itoa(int(s.btcNonce)))
	strs = append(strs, strconv.Itoa(int(s.btcBlockHeight)))
	strs = append(strs, strconv.Itoa(int(s.checkpointTime)))
	strs = append(strs, strconv.Itoa(int(s.btcBlockTime)))
	return strs
}
