package ppos

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/mempool"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/wire"
)

// PoSEngine only need to start if node runner want to be a validator

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
	BlockChain  *blockchain.BlockChain
	ConnManager *connmanager.ConnManager
	// RewardAgent
	ChainParams    *blockchain.Params
	BlockGen       *blockchain.BlkTmplGenerator
	MemPool        *mempool.TxPool
	ProducerKeySet cashec.KeySet
	Server         interface {
		// list functions callback which are assigned from Server struct
		GetPeerIDsFromPublicKey(string) []peer2.ID
		PushMessageToAll(wire.Message) error
		PushMessageToPeer(wire.Message, peer2.ID) error
		PushMessageGetChainState() error
	}
	FeeEstimator map[byte]*mempool.FeeEstimator
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

//Start start consensus engine
func (self *Engine) Start() error {
	self.Lock()
	defer self.Unlock()
	if self.started {
		return errors.New("Consensus engine is already started")
	}
	Logger.log.Info("Starting Parallel Proof of Stake Consensus engine")
	self.knownChainsHeight.Heights = make([]int, common.TotalValidators)
	self.validatedChainsHeight.Heights = make([]int, common.TotalValidators)

	self.committee.ValidatorBlkNum = make(map[string]int)
	self.committee.ValidatorReliablePts = make(map[string]int)
	self.committee.CurrentCommittee = make([]string, common.TotalValidators)

	for chainID := 0; chainID < common.TotalValidators; chainID++ {
		self.knownChainsHeight.Heights[chainID] = int(self.config.BlockChain.BestState[chainID].Height)
		self.validatedChainsHeight.Heights[chainID] = 1
	}

	copy(self.committee.CurrentCommittee, self.config.ChainParams.GenesisBlock.Header.Committee)
	Logger.log.Info("Validating local blockchain...")

	if _, ok := self.config.FeeEstimator[0]; !ok {
		// happen when FastMode = false
		validatedChainsHeight := make([]int, common.TotalValidators)
		var wg sync.WaitGroup
		errCh := make(chan error)
		for chainID := byte(0); chainID < common.TotalValidators; chainID++ {
			//Don't validate genesis block (blockHeight = 1)
			validatedChainsHeight[chainID] = 1
			self.config.FeeEstimator[chainID] = mempool.NewFeeEstimator(
				mempool.DefaultEstimateFeeMaxRollback,
				mempool.DefaultEstimateFeeMinRegisteredBlocks)
			go func(chainID byte) {
				wg.Add(1)
				var err error
				defer func() {
					wg.Done()
					if err != nil {
						errCh <- err
					}
				}()
				for blockHeight := 2; blockHeight <= self.knownChainsHeight.Heights[chainID]; blockHeight++ {
					var block *blockchain.Block
					block, err = self.config.BlockChain.GetBlockByBlockHeight(int32(blockHeight), byte(chainID))
					if err != nil {
						Logger.log.Error(err)
						return
					}
					Logger.log.Infof("block height: %d", block.Header.Height)
					//TODO Comment validateBlockSanity segment to create block with only 1 node (validator)
					/*err = self.validateBlockSanity(block)
					if err != nil {
						Logger.log.Error(err)
						return
					}*/
					// end TODO
					err = self.config.BlockChain.CreateAndSaveTxViewPointFromBlock(block)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					err = self.config.FeeEstimator[block.Header.ChainID].RegisterBlock(block)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					self.validatedChainsHeight.Lock()
					self.validatedChainsHeight.Heights[chainID] = blockHeight
					self.validatedChainsHeight.Unlock()
					self.committee.UpdateCommitteePoint(block.BlockProducer, block.Header.BlockCommitteeSigs)
				}
			}(chainID)
		}
		time.Sleep(1000 * time.Millisecond)
		wg.Wait()
		select {
		case err := <-errCh:
			return err
		default:
			break
		}
	} else {
		copy(self.validatedChainsHeight.Heights, self.knownChainsHeight.Heights)
	}

	self.started = true
	self.cQuit = make(chan struct{})

	go func() {
		for {
			self.config.Server.PushMessageGetChainState()
			time.Sleep(common.GetChainStateInterval * time.Second)
		}
	}()

	return nil
}

//Stop stop consensus engine
func (self *Engine) Stop() error {
	Logger.log.Info("Stopping Consensus engine...")
	self.Lock()
	defer self.Unlock()

	if !self.started {
		return errors.New("Consensus engine isn't running")
	}
	self.StopProducer()
	if self.cQuitSwap != nil {
		close(self.cQuitSwap)
	}
	close(self.cQuit)
	self.started = false
	Logger.log.Info("Consensus engine stopped")
	return nil
}

//StartProducer start producing block
func (self *Engine) StartProducer(producerKeySet cashec.KeySet) {
	if self.producerStarted {
		Logger.log.Error("Producer already started")
		return
	}

	self.config.ProducerKeySet = producerKeySet

	self.cQuitProducer = make(chan struct{})
	self.cQuitCommitteeWatcher = make(chan struct{})
	self.cBlockSig = make(chan blockSig)
	self.cNewBlock = make(chan blockchain.Block)

	self.producerStarted = true
	Logger.log.Info("Starting producer with public key: " + base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00)))

	go func() {
		for {
			select {
			case <-self.cQuitProducer:
				return
			default:
				if self.started {
					if common.IntArrayEquals(self.knownChainsHeight.Heights, self.validatedChainsHeight.Heights) {
						chainID := self.getMyChain()
						if chainID >= 0 && chainID < common.TotalValidators {
							go self.StartCommitteeWatcher()
							Logger.log.Info("(๑•̀ㅂ•́)و Yay!! It's my turn")
							Logger.log.Info("Current chainsHeight")
							Logger.log.Info(self.validatedChainsHeight.Heights)
							Logger.log.Info("My chainID: ", chainID)

							newBlock, err := self.createBlock()
							if err != nil {
								Logger.log.Error(err)
								continue
							}
							err = self.Finalize(newBlock)
							if err != nil {
								Logger.log.Critical(err)
								continue
							}
						}
					} else {
						self.StopCommitteeWatcher()
						for i, v := range self.knownChainsHeight.Heights {
							if v > self.validatedChainsHeight.Heights[i] {
								lastBlockHash := self.config.BlockChain.BestState[i].BestBlockHash.String()
								getBlkMsg := &wire.MessageGetBlocks{
									LastBlockHash: lastBlockHash,
								}
								self.config.Server.PushMessageToAll(getBlkMsg)
							}
						}
					}
				}
			}
		}
	}()
}

// StopProducer stop producer
func (self *Engine) StopProducer() {
	if self.producerStarted {
		Logger.log.Info("Stopping Producer...")
		close(self.cQuitProducer)
		close(self.cBlockSig)
		self.StopCommitteeWatcher()
		self.producerStarted = false
	}
}

func (self *Engine) createBlock() (*blockchain.Block, error) {
	Logger.log.Info("Start creating block...")
	myChainID := self.getMyChain()
	paymentAddress := self.config.ProducerKeySet.PaymentAddress
	privatekey := self.config.ProducerKeySet.PrivateKey
	newblock, err := self.config.BlockGen.NewBlockTemplate(&paymentAddress, &privatekey, myChainID)
	if err != nil {
		return &blockchain.Block{}, err
	}
	newblock.Header.ChainsHeight = make([]int, common.TotalValidators)
	copy(newblock.Header.ChainsHeight, self.validatedChainsHeight.Heights)
	newblock.Header.ChainID = myChainID
	newblock.BlockProducer = base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00))

	return newblock, nil
}

// Finalize after successfully create a block we will send this block to other validators to get their signatures
func (self *Engine) Finalize(finalBlock *blockchain.Block) error {
	Logger.log.Info("Start finalizing block...")
	allSigReceived := make(chan struct{})
	retryTime := 0
	cancel := make(chan struct{})
	defer func() {
		close(cancel)
		close(allSigReceived)
	}()
finalizing:
	finalBlock.Header.BlockCommitteeSigs = make([]string, common.TotalValidators)
	finalBlock.Header.Committee = make([]string, common.TotalValidators)

	copy(finalBlock.Header.Committee, self.GetCommittee())
	sig, err := self.signData([]byte(finalBlock.Hash().String()))
	if err != nil {
		return err
	}
	finalBlock.Header.BlockCommitteeSigs[finalBlock.Header.ChainID] = sig

	committee := finalBlock.Header.Committee

	// Collect signatures of other validators
	go func(blockHash string) {
		sigsReceived := 0
		for {
			select {
			case <-self.cQuit:
				return
			case <-cancel:
				return
			case blocksig := <-self.cBlockSig:

				if idx := common.IndexOfStr(blocksig.Validator, committee); idx != -1 {
					if finalBlock.Header.BlockCommitteeSigs[idx] == common.EmptyString {
						err := cashec.ValidateDataB58(blocksig.Validator, blocksig.BlockSig, []byte(blockHash))

						if err != nil {
							Logger.log.Error("ValidateTransaction sig error:", err)
							continue
						} else {
							sigsReceived++
							finalBlock.Header.BlockCommitteeSigs[idx] = blocksig.BlockSig
							Logger.log.Info("Validator's signature received", sigsReceived)
						}
					} else {
						Logger.log.Error("Already received this validator blocksig")
					}
				}

				if sigsReceived == (common.MinBlockSigs - 1) {
					allSigReceived <- struct{}{}
					return
				}
			}
		}
	}(finalBlock.Hash().String())

	//Request for signatures of other validators
	go func(block blockchain.Block) {
		//TODO Uncomment this segment to create block with only 1 node (validator)
		allSigReceived <- struct{}{}
		// end TODO

		reqSigMsg, _ := wire.MakeEmptyMessage(wire.CmdBlockSigReq)
		reqSigMsg.(*wire.MessageBlockSigReq).Block = block
		for idx := 0; idx < common.TotalValidators; idx++ {
			//@TODO: retry on failed validators
			if committee[idx] != finalBlock.BlockProducer {
				go func(validator string) {
					peerIDs := self.config.Server.GetPeerIDsFromPublicKey(validator)
					if len(peerIDs) != 0 {
						Logger.log.Info("Request signature from "+peerIDs[0], validator)
						self.config.Server.PushMessageToPeer(reqSigMsg, peerIDs[0])
					} else {
						Logger.log.Error("Validator's peer not found!", validator)
					}
				}(committee[idx])
			}
		}
	}(*finalBlock)
	// Wait for signatures of other validators
	select {
	case <-self.cQuit:
		cancel <- struct{}{}
		return nil
	case <-allSigReceived:
		Logger.log.Info("Validator sigs: ", finalBlock.Header.BlockCommitteeSigs)
	case <-time.After(common.MaxBlockSigWaitTime * time.Second):
		//blocksig wait time exceeded -> get a new committee list and retry
		Logger.log.Error(ErrExceedSigWaitTime)
		if retryTime == 5 {
			cancel <- struct{}{}
			return NewConsensusError(ErrExceedBlockRetry, nil)
		}
		retryTime++
		Logger.log.Infof("Start finalizing block... %d time", retryTime)
		cancel <- struct{}{}
		goto finalizing
	}

	headerBytes, _ := json.Marshal(finalBlock.Header)
	sig, err = self.signData(headerBytes)
	if err != nil {
		return err
	}
	finalBlock.BlockProducerSig = sig

	self.UpdateChain(finalBlock)
	self.sendBlockMsg(finalBlock)
	return nil
}

func (self *Engine) UpdateChain(block *blockchain.Block) {
	err := self.config.BlockChain.ConnectBlock(block)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	// save block into fee estimator
	err = self.config.FeeEstimator[block.Header.ChainID].RegisterBlock(block)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	// update tx pool
	for _, tx := range block.Transactions {
		self.config.MemPool.RemoveTx(tx)
	}

	// update candidate list
	err = self.config.BlockChain.BestState[block.Header.ChainID].Update(block)
	if err != nil {
		Logger.log.Errorf("Can not update merkle tree for block: %+v", err)
		return
	}
	self.config.BlockChain.StoreBestState(block.Header.ChainID)

	self.knownChainsHeight.Lock()
	if self.knownChainsHeight.Heights[block.Header.ChainID] < int(block.Header.Height) {
		self.knownChainsHeight.Heights[block.Header.ChainID] = int(block.Header.Height)
		self.sendBlockMsg(block)
	}
	self.knownChainsHeight.Unlock()
	self.validatedChainsHeight.Lock()
	self.validatedChainsHeight.Heights[block.Header.ChainID] = int(block.Header.Height)
	self.validatedChainsHeight.Unlock()

	self.committee.UpdateCommitteePoint(block.BlockProducer, block.Header.BlockCommitteeSigs)
}
