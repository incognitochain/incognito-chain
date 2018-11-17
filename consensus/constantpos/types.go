package constantpos

import (
	"sync"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/mempool"
	"github.com/ninjadotorg/constant/wire"
)

type Engine struct {
	sync.Mutex
	started         bool
	producerStarted bool
	committeeMutex  sync.Mutex

	// channel
	cQuit                 chan struct{}
	cQuitProducer         chan struct{}
	cBlockSig             chan blockSig
	cQuitSwap             chan struct{}
	cSwapChain            chan byte
	cSwapSig              chan swapSig
	cQuitCommitteeWatcher chan struct{}
	cNewBlock             chan blockchain.Block

	config                EngineConfig
	knownChainsHeight     chainsHeight
	validatedChainsHeight chainsHeight

	committee committeeStruct
}

type committeeStruct struct {
	ValidatorBlkNum      map[string]int //track the number of block created by each validator
	ValidatorReliablePts map[string]int //track how reliable is the validator node
	CurrentCommittee     []string
	cmWatcherStarted     bool
	sync.Mutex
	LastUpdate int64
}

type ChainInfo struct {
	CurrentCommittee        []string
	CandidateListMerkleHash string
	ChainsHeight            []int
}

type chainsHeight struct {
	Heights []int
	sync.Mutex
}

type EngineConfig struct {
	BlockChain     *blockchain.BlockChain
	ConnManager    *connmanager.ConnManager
	ChainParams    *blockchain.Params
	BlockGen       *blockchain.BlkTmplGenerator
	MemPool        *mempool.TxPool
	ProducerKeySet cashec.KeySetProducer
	Server         interface {
		// list functions callback which are assigned from Server struct
		GetPeerIDsFromPublicKey(string) []libp2p.ID
		PushMessageToAll(wire.Message) error
		PushMessageToPeer(wire.Message, libp2p.ID) error
		PushMessageGetChainState() error
	}
}

type blockSig struct {
	Validator string
	BlockSig  string
}

type swapSig struct {
	Validator string
	SwapSig   string
}

//Init apply configuration to consensus engine
func (self Engine) Init(cfg *EngineConfig) (*Engine, error) {
	return &Engine{
		committeeMutex: sync.Mutex{},
		config:         *cfg,
	}, nil
}
