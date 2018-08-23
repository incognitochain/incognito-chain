package miner

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/mining"
)

var (
	defaultNumWorkers = uint32(1) //uint32(runtime.NumCPU())
)

type Config struct {
	ChainParams *blockchain.Params

	Chain *blockchain.BlockChain

	BlockTemplateGenerator *mining.BlkTmplGenerator

	MiningAddrs []string

	Server interface {
		PushBlockMessage(*blockchain.Block) bool
	}
}

type Miner struct {
	sync.Mutex
	g                 *mining.BlkTmplGenerator
	cfg               Config
	numWorkers        uint32
	started           bool
	submitBlockLock   sync.Mutex
	wg                sync.WaitGroup
	workerWg          sync.WaitGroup
	updateNumWorkers  chan struct{}
	queryHashesPerSec chan float64
	updateHashes      chan uint64
	quit              chan struct{}
}

func (m *Miner) GenerateBlock(n uint32) ([]*common.Hash, error) {
	m.Lock()
	// Respond with an error if server is already mining.
	if m.started {
		m.Unlock()
		return nil, errors.New("Server is already CPU mining....")
	}
	m.started = true
	m.wg.Add(1)
	m.Unlock()
	fmt.Print("Generating %d blocks", n)

	i := uint32(0)
	blockHashes := make([]*common.Hash, n)

	for {
		m.submitBlockLock.Lock()
		payToAddr := m.cfg.MiningAddrs[rand.Intn(len(m.cfg.MiningAddrs))]
		template, err := m.g.NewBlockTemplate(payToAddr, m.cfg.Chain)
		m.submitBlockLock.Unlock()
		if err != nil {
			fmt.Sprintf("Failed to create new block template: %v", err)
			continue
		}
		//@todo need verify and process block before sending
		m.commitBlock(template.Block)
		blockHashes[i] = template.Block.Hash()
		i++
		if i == n {
			fmt.Sprintf("Generated %d blocks", i)
			m.Lock()
			m.wg.Wait()
			m.started = false
			m.Unlock()
			return blockHashes, nil
		}
	}

}

func (m *Miner) commitBlock(block *blockchain.Block) (bool, error) {
	m.submitBlockLock.Lock()
	defer m.submitBlockLock.Unlock()
	sended := m.cfg.Server.PushBlockMessage(block)
	if sended != true {
		fmt.Print("sending error...........")
		return false, nil
	}

	return true, nil
}

func (m *Miner) generateBlocks(quit chan struct{}) {
	fmt.Print("Starting generate blocks worker")
out:
	for {
		select {
		case <-quit:
			break out
		default:
			// Non-blocking select to fall through
		}
		m.submitBlockLock.Lock()
		payToAddr := m.cfg.MiningAddrs[rand.Intn(len(m.cfg.MiningAddrs))]

		template, err := m.g.NewBlockTemplate(payToAddr, m.cfg.Chain)
		m.submitBlockLock.Unlock()
		if err != nil || len(template.Block.Transactions) == 0 {
			fmt.Sprint("Failed to create new block template: %v", err)
			continue
		}
		m.commitBlock(template.Block)
	}
	m.workerWg.Done()
	fmt.Sprint("Generate blocks worker done")
}

func (m *Miner) workerController() {
	// launchWorkers groups common code to launch a specified number of
	// workers for generating blocks.
	var runningWorkers []chan struct{}
	launchWorkers := func(numWorkers uint32) {
		for i := uint32(0); i < numWorkers; i++ {
			quit := make(chan struct{})
			runningWorkers = append(runningWorkers, quit)

			m.workerWg.Add(1)
			go m.generateBlocks(quit)
		}
	}

	// Launch the current number of workers by default.
	runningWorkers = make([]chan struct{}, 0, m.numWorkers)
	launchWorkers(m.numWorkers)

out:
	for {
		select {
		// Update the number of running workers.
		case <-m.updateNumWorkers:
			// No change.
			numRunning := uint32(len(runningWorkers))
			if m.numWorkers == numRunning {
				continue
			}

			// Add new workers.
			if m.numWorkers > numRunning {
				launchWorkers(m.numWorkers - numRunning)
				continue
			}

			// Signal the most recently created goroutines to exit.
			for i := numRunning - 1; i >= m.numWorkers; i-- {
				close(runningWorkers[i])
				runningWorkers[i] = nil
				runningWorkers = runningWorkers[:i]
			}

		case <-m.quit:
			for _, quit := range runningWorkers {
				close(quit)
			}
			break out
		}
	}

	// Wait until all workers shut down to stop the speed monitor since
	// they rely on being able to send updates to it.
	m.workerWg.Wait()
	m.wg.Done()
}

func (m *Miner) Start() {
	m.Lock()
	defer m.Unlock()

	if m.started {
		return
	}

	m.quit = make(chan struct{})
	m.wg.Add(2)
	go m.workerController()

	m.started = true
	fmt.Print("CPU miner started")
}

func (m *Miner) Stop() {
	m.Lock()
	defer m.Unlock()

	// Nothing to do if the miner is not currently running or if running in
	// discrete mode (using GenerateNBlocks).
	if !m.started {
		return
	}

	close(m.quit)
	m.wg.Wait()
	m.started = false
	fmt.Print("CPU miner stopped")
}

// New returns a new instance of a CPU miner for the provided configuration.
// Use Start to begin the mining process.  See the documentation for CPUMiner
// type for more details.
func New(cfg *Config) *Miner {
	return &Miner{
		g:                 cfg.BlockTemplateGenerator,
		cfg:               *cfg,
		numWorkers:        defaultNumWorkers,
		updateNumWorkers:  make(chan struct{}),
		queryHashesPerSec: make(chan float64),
		updateHashes:      make(chan uint64),
	}
}
