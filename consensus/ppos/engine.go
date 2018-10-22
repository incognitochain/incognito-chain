package ppos

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/ninjadotorg/cash/cashec"
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/common/base58"
	"github.com/ninjadotorg/cash/mempool"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash/blockchain"
	"github.com/ninjadotorg/cash/connmanager"
	"github.com/ninjadotorg/cash/transaction"
	"github.com/ninjadotorg/cash/wire"
)

// PoSEngine only need to start if node runner want to be a validator

type Engine struct {
	sync.Mutex
	started       bool
	sealerStarted bool

	// channel
	cQuit       chan struct{}
	cQuitSealer chan struct{}
	cBlockSig   chan blockSig
	cQuitSwap   chan struct{}
	cSwapChain  chan byte

	config                EngineConfig
	knownChainsHeight     chainsHeight
	validatedChainsHeight chainsHeight

	committee committeeStruct
}

type committeeStruct struct {
	ValidatorBlkNum      map[string]int //track the number of block created by each validator
	ValidatorReliablePts map[string]int //track how reliable is the validator node
	UpcomingCommittee    []string
	CurrentCommittee     []string
	sync.Mutex
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
	ChainParams     *blockchain.Params
	BlockGen        *blockchain.BlkTmplGenerator
	MemPool         *mempool.TxPool
	ValidatorKeySet cashec.KeySetSealer
	Server interface {
		// list functions callback which are assigned from Server struct
		GetPeerIDsFromPublicKey(string) []peer2.ID
		PushMessageToAll(wire.Message) error
		PushMessageToPeer(wire.Message, peer2.ID) error
		PushMessageGetChainState() error
	}
	FeeEstimator map[byte]*mempool.FeeEstimator
}

type blockSig struct {
	BlockHash    string
	Validator    string
	ValidatorSig string
}

//Init apply configuration to consensus engine
func (self Engine) Init(cfg *EngineConfig) (*Engine, error) {
	return &Engine{
		config: *cfg,
	}, nil
}

//Start start consensus engine
func (self *Engine) Start() error {
	self.Lock()
	defer self.Unlock()
	if self.started {
		// self.Unlock()
		return errors.New("Consensus engine is already started")
	}
	Logger.log.Info("Starting Parallel Proof of Stake Consensus engine")
	self.knownChainsHeight.Heights = make([]int, common.TotalValidators)
	self.validatedChainsHeight.Heights = make([]int, common.TotalValidators)
	self.committee.ValidatorBlkNum = make(map[string]int)
	self.committee.ValidatorReliablePts = make(map[string]int)
	self.committee.UpcomingCommittee = make([]string, common.TotalValidators)
	self.committee.CurrentCommittee = make([]string, common.TotalValidators)

	for chainID := 0; chainID < common.TotalValidators; chainID++ {
		self.knownChainsHeight.Heights[chainID] = int(self.config.BlockChain.BestState[chainID].Height)
		self.validatedChainsHeight.Heights[chainID] = 1
	}

	Logger.log.Info("Validating local blockchain...")
	copy(self.committee.CurrentCommittee, self.config.ChainParams.GenesisBlock.Header.Committee)
	copy(self.committee.UpcomingCommittee, self.committee.CurrentCommittee)

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
					Logger.log.Infof("block height: %d", block.Height)
					//Comment validateBlockSanity segment to create block with only 1 node (validator)
					err = self.validateBlockSanity(block)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					err = self.config.BlockChain.CreateTxViewPoint(block)
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
					self.committee.UpdateCommittee(block.ChainLeader, block.Header.BlockCommitteeSigs)
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
	self.StopSealer()
	if self.cQuitSwap != nil {
		close(self.cQuitSwap)
	}
	close(self.cQuit)
	self.started = false
	Logger.log.Info("Consensus engine stopped")
	return nil
}

//StartSealer start sealing block
func (self *Engine) StartSealer(sealerKeySet cashec.KeySetSealer) {
	if self.sealerStarted {
		Logger.log.Error("Sealer already started")
		return
	}
	self.config.ValidatorKeySet = sealerKeySet

	self.cQuitSealer = make(chan struct{})
	self.cBlockSig = make(chan blockSig)
	self.sealerStarted = true
	Logger.log.Info("Starting sealer with public key: " + base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)))

	go func() {
		for {
			select {
			case <-self.cQuitSealer:
				return
			default:
				if self.started {
					if common.IntArrayEquals(self.knownChainsHeight.Heights, self.validatedChainsHeight.Heights) {
						chainID := self.getMyChain()
						if chainID < common.TotalValidators {
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

// StopSealer stop sealer
func (self *Engine) StopSealer() {
	if self.sealerStarted {
		Logger.log.Info("Stopping Sealer...")
		close(self.cQuitSealer)
		close(self.cBlockSig)
		self.sealerStarted = false
	}
}

func (self *Engine) createBlock() (*blockchain.Block, error) {
	Logger.log.Info("Start creating block...")
	myChainID := self.getMyChain()
	paymentAddress, err := self.config.ValidatorKeySet.GetPaymentAddress()
	newblock, err := self.config.BlockGen.NewBlockTemplate(paymentAddress, myChainID)
	if err != nil {
		return &blockchain.Block{}, err
	}
	newblock.Block.Header.ChainsHeight = make([]int, common.TotalValidators)
	copy(newblock.Block.Header.ChainsHeight, self.validatedChainsHeight.Heights)
	newblock.Block.Header.ChainID = myChainID
	newblock.Block.ChainLeader = base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))

	// hash candidate list and set to block header
	candidates := self.GetCndList(newblock.Block)
	candidateBytes, _ := json.Marshal(candidates)
	newblock.Block.Header.CandidateHash = common.HashH(candidateBytes)

	return newblock.Block, nil
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
		close(cancel)
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

				if blockHash != blocksig.BlockHash {
					Logger.log.Critical("Block hash not match!", blocksig, "this block", blockHash)
					continue
				}

				if idx := common.IndexOfStr(blocksig.Validator, committee); idx != -1 {
					if finalBlock.Header.BlockCommitteeSigs[idx] == "" {
						err := cashec.ValidateDataB58(blocksig.Validator, blocksig.ValidatorSig, []byte(blockHash))

						if err != nil {
							Logger.log.Error("Validate sig error:", err)
							continue
						} else {
							sigsReceived++
							finalBlock.Header.BlockCommitteeSigs[idx] = blocksig.ValidatorSig
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
		//Uncomment this segment to create block with only 1 node (validator)
		// allSigReceived <- struct{}{}

		reqSigMsg, _ := wire.MakeEmptyMessage(wire.CmdRequestBlockSign)
		reqSigMsg.(*wire.MessageRequestBlockSign).Block = block
		for idx := 0; idx < common.TotalValidators; idx++ {
			//@TODO: retry on failed validators
			if committee[idx] != finalBlock.ChainLeader {
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
		//blocksig wait time exceeded -> get a new commitee list and retry
		Logger.log.Error(errExceedSigWaitTime)
		if retryTime == 5 {
			cancel <- struct{}{}
			return errExceedBlockRetry
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
	finalBlock.ChainLeaderSig = sig

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
	self.config.BlockChain.BestState[block.Header.ChainID].Candidates = self.GetCndList(block)
	self.config.BlockChain.BestState[block.Header.ChainID].Update(block)
	self.config.BlockChain.StoreBestState(block.Header.ChainID)

	self.knownChainsHeight.Lock()
	if self.knownChainsHeight.Heights[block.Header.ChainID] < int(block.Height) {
		self.knownChainsHeight.Heights[block.Header.ChainID] = int(block.Height)
		self.sendBlockMsg(block)
	}
	self.knownChainsHeight.Unlock()
	self.validatedChainsHeight.Lock()
	self.validatedChainsHeight.Heights[block.Header.ChainID] = int(block.Height)
	self.validatedChainsHeight.Unlock()

	self.committee.UpdateCommittee(block.ChainLeader, block.Header.BlockCommitteeSigs)
}

func (self *Engine) GetCndList(block *blockchain.Block) map[string]blockchain.CndInfo {
	bestState := self.config.BlockChain.BestState[block.Header.ChainID]
	candidates := bestState.Candidates
	if candidates == nil {
		candidates = make(map[string]blockchain.CndInfo)
	}
	for _, tx := range block.Transactions {
		if tx.GetType() == common.TxVotingType {
			txV, ok := tx.(*transaction.TxVoting)
			nodeAddr := txV.NodeAddr
			cndVal, ok := candidates[nodeAddr]
			_ = cndVal
			if !ok {
				candidates[nodeAddr] = blockchain.CndInfo{
					Value:     txV.GetValue(),
					Timestamp: block.Header.Timestamp,
					ChainID:   block.Header.ChainID,
				}
			} else {
				candidates[nodeAddr] = blockchain.CndInfo{
					Value:     cndVal.Value + txV.GetValue(),
					Timestamp: block.Header.Timestamp,
					ChainID:   block.Header.ChainID,
				}
			}
		}
	}
	return candidates
}

func (self *Engine) IsExistedNodeAddr(nodeAddr string) bool {
	for _, bestState := range self.config.BlockChain.BestState {
		_, ok := bestState.Candidates[nodeAddr]
		if ok {
			return true
		}
	}
	return false
}

func (self *Engine) StartSwap() error {
	self.cQuitSwap = make(chan struct{})
	for {
		select {
		case <-self.cQuitSwap:
			{
				Logger.log.Info("Consensus engine stopped swap")
				return nil
			}
		case chainId := <-self.cSwapChain:
			{
				Logger.log.Infof("Consensus engine swap period %s START", chainId)

				//cndList := self.config.BlockChain.GetCndList()
				//for _, cndNode := range cndList {
				//	self.config.ConnManager.Connect()
				//}

				Logger.log.Infof("Consensus engine swap period %s END", chainId)
				continue
			}
		}
	}

	return nil
}
