package constantpos

import (
	"sync"
)

type ChainInfo struct {
	CurrentCommittee        []string
	CandidateListMerkleHash string
	ChainsHeight            []int
}

type chainsHeight struct {
	Heights []int
	sync.Mutex
}

type blockSig struct {
	Validator string
	BlockSig  string
}

type swapSig struct {
	Validator string
	SwapSig   string
}
