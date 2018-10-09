package ppos

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/common/base58"
	"github.com/ninjadotorg/cash-prototype/mempool"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/wire"
)

// PoSEngine only need to start if node runner want to be a validator

type Engine struct {
	sync.Mutex
	started bool
	wg      sync.WaitGroup
	quit    chan struct{}

	sealerStarted bool
	quitSealer    chan struct{}

	config                Config
	currentCommittee      []string
	candidates            []string
	knownChainsHeight     chainsHeight
	validatedChainsHeight chainsHeight
	blockSigCh            chan blockSig
}

type ChainInfo struct {
	CurrentCommittee  []string
	CandidateListHash string
	ChainsHeight      []int
}
type chainsHeight struct {
	Heights []int
	sync.Mutex
}
type Config struct {
	BlockChain      *blockchain.BlockChain
	ChainParams     *blockchain.Params
	blockGen        *BlkTmplGenerator
	MemPool         *mempool.TxPool
	ValidatorKeySet cashec.KeySetSealer
	Server interface {
		// list functions callback which are assigned from Server struct
		GetPeerIdsFromPublicKey(string) []peer2.ID
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

func (self *Engine) Start() error {
	self.Lock()
	defer self.Unlock()
	if self.started {
		self.Unlock()
		return errors.New("Consensus engine is already started")
	}
	Logger.log.Info("Starting Parallel Proof of Stake Consensus engine")
	self.knownChainsHeight.Heights = make([]int, TOTAL_VALIDATORS)
	self.validatedChainsHeight.Heights = make([]int, TOTAL_VALIDATORS)
	self.currentCommittee = []string{}

	for chainID := 0; chainID < TOTAL_VALIDATORS; chainID++ {
		self.knownChainsHeight.Heights[chainID] = int(self.config.BlockChain.BestState[chainID].Height)
	}

	Logger.log.Info("Validating local blockchain...")
	for chainID := 0; chainID < TOTAL_VALIDATORS; chainID++ {
		//Don't validate genesis block (blockHeight = 1)
		for blockHeight := 2; blockHeight < self.knownChainsHeight.Heights[chainID]; blockHeight++ {
			block, err := self.config.BlockChain.GetBlockByBlockHeight(int32(blockHeight), byte(chainID))
			if err != nil {
				Logger.log.Error(err)
				return err
			}
			err = self.validateBlock(block)
			if err != nil {
				Logger.log.Error(err)
				return err
			}
		}
	}

	copy(self.validatedChainsHeight.Heights, self.knownChainsHeight.Heights)
	Logger.log.Info("-------------------------------------------")
	Logger.log.Info(len(self.validatedChainsHeight.Heights))
	Logger.log.Info(len(self.knownChainsHeight.Heights))
	Logger.log.Info("-------------------------------------------")
	for key := range self.config.BlockChain.BestState[0].BestBlock.Header.CommitteeSigs {
		self.currentCommittee = append(self.currentCommittee, key)
	}

	go func() {
		for {
			self.config.Server.PushMessageGetChainState()
			time.Sleep(GETCHAINSTATE_INTERVAL * time.Second)
			Logger.log.Info("Heights ", self.knownChainsHeight.Heights)
			Logger.log.Info("Heights ", self.validatedChainsHeight.Heights)
		}
	}()

	self.started = true
	self.quit = make(chan struct{})
	self.wg.Add(1)

	return nil
}

func (self *Engine) Stop() error {
	Logger.log.Info("Stopping Consensus engine...")
	self.Lock()
	defer self.Unlock()

	if !self.started {
		return errors.New("Consensus engine isn't running")
	}
	self.StopSealer()
	close(self.quit)
	self.started = false
	Logger.log.Info("Consensus engine stopped")
	return nil
}

func New(cfg *Config) *Engine {
	cfg.blockGen = NewBlkTmplGenerator(cfg.MemPool, cfg.BlockChain)
	return &Engine{
		config: *cfg,
	}
}

func (self *Engine) StartSealer(sealerKeySet cashec.KeySetSealer) {
	if self.sealerStarted {
		Logger.log.Error("Sealer already started")
		return
	}
	self.config.ValidatorKeySet = sealerKeySet

	self.quitSealer = make(chan struct{})
	self.blockSigCh = make(chan blockSig)
	self.sealerStarted = true
	Logger.log.Info("Starting sealer with public key: " + base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)))

	go func() {
		for {
			select {
			case <-self.quitSealer:
				return
			default:
				if self.started {
					if common.IntArrayEquals(self.knownChainsHeight.Heights, self.validatedChainsHeight.Heights) {
						chainID := self.getMyChain()
						fmt.Println(base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)), len(self.currentCommittee))
						if chainID < TOTAL_VALIDATORS {
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

func (self *Engine) StopSealer() {
	if self.sealerStarted {
		Logger.log.Info("Stopping Sealer...")
		close(self.quitSealer)
		close(self.blockSigCh)
		self.sealerStarted = false
	}
}

func (self *Engine) createBlock() (*blockchain.Block, error) {
	Logger.log.Info("Start creating block...")
	myChainID := self.getMyChain()
	paymentAddress, err := self.config.ValidatorKeySet.GetPaymentAddress()
	newblock, err := self.config.blockGen.NewBlockTemplate(paymentAddress, self.config.BlockChain, myChainID)
	if err != nil {
		return &blockchain.Block{}, err
	}
	newblock.Block.Header.ChainsHeight = make([]int, TOTAL_VALIDATORS)
	copy(newblock.Block.Header.ChainsHeight, self.validatedChainsHeight.Heights)
	newblock.Block.Header.ChainID = myChainID
	newblock.Block.ChainLeader = base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))
	// newblock.Block.Header.Committee = self.GetNextCommittee()

	for _, validator := range self.GetNextCommittee() {
		newblock.Block.Header.CommitteeSigs[validator] = ""
	}

	sig, err := self.signData([]byte(newblock.Block.Hash().String()))
	if err != nil {
		return &blockchain.Block{}, err
	}
	newblock.Block.Header.CommitteeSigs[newblock.Block.ChainLeader] = sig
	return newblock.Block, nil
}

func (self *Engine) Finalize(block *blockchain.Block) error {
	Logger.log.Info("Start finalizing block...")
	finalBlock := block
	allSigReceived := make(chan struct{})
	cancel := make(chan struct{})
	committee := []string{}
	for validator := range block.Header.CommitteeSigs {
		committee = append(committee, validator)
	}
	defer func() {
		close(cancel)
		close(allSigReceived)
	}()

	// Collect signatures of other validators
	go func(blockHash string) {
		var sigsReceived int
		for {
			select {
			case <-cancel:
				return
			case blocksig := <-self.blockSigCh:
				Logger.log.Info("Validator's signature received", sigsReceived)

				if blockHash != blocksig.BlockHash {
					Logger.log.Critical("(o_O)!", blocksig, "this block", blockHash)
					continue
				}

				if sig, ok := block.Header.CommitteeSigs[blocksig.Validator]; ok {
					if sig != "" {
						if common.IndexOfStr(blocksig.Validator, committee) < len(committee) {
							err := cashec.ValidateDataB58(blocksig.Validator, blocksig.ValidatorSig, []byte(block.Hash().String()))

							if err != nil {
								Logger.log.Error("Validate sig error:", err)
								continue
							} else {
								sigsReceived++
								finalBlock.Header.CommitteeSigs[blocksig.Validator] = blocksig.ValidatorSig
							}
						}
					} else {
						Logger.log.Error("Already received this validator blocksig")
					}
				}

				if sigsReceived == (MINIMUM_BLOCKSIGS - 1) {
					allSigReceived <- struct{}{}
					return
				}
			}
		}
	}(block.Hash().String())
	//Request for signatures of other validators
	go func() {
		reqSigMsg, _ := wire.MakeEmptyMessage(wire.CmdRequestSign)
		reqSigMsg.(*wire.MessageRequestSign).Block = *block
		for idx := 0; idx < TOTAL_VALIDATORS; idx++ {
			//@TODO: retry on failed validators
			if committee[idx] != block.ChainLeader {
				go func(validator string) {
					peerIDs := self.config.Server.GetPeerIdsFromPublicKey(validator)
					if len(peerIDs) != 0 {
						Logger.log.Info("Request signaure from "+peerIDs[0], validator)
						self.config.Server.PushMessageToPeer(reqSigMsg, peerIDs[0])
					} else {
						fmt.Println("Validator's peer not found!", validator)
					}
				}(committee[idx])
			}
		}
	}()

	// Wait for signatures of other validators
	select {
	case <-allSigReceived:
		Logger.log.Info("Validator sigs: ", finalBlock.Header.CommitteeSigs)
	case <-time.After(MAX_BLOCKSIGN_WAIT_TIME * time.Second):
		return errCantFinalizeBlock
	}

	cmsBytes, _ := json.Marshal(finalBlock.Header.CommitteeSigs)
	sig, err := self.signData(cmsBytes)
	if err != nil {
		return err
	}
	finalBlock.ChainLeaderSig = sig
	self.UpdateChain(finalBlock)
	blockMsg, err := wire.MakeEmptyMessage(wire.CmdBlock)
	if err != nil {
		return err
	}
	blockMsg.(*wire.MessageBlock).Block = *finalBlock
	self.config.Server.PushMessageToAll(blockMsg)
	return nil
}

func (self *Engine) validateBlock(block *blockchain.Block) error {
	// validate steps: block size -> sealer's sig of the final block -> sealer is belong to committee -> validate each committee member's sig
	if block.Header.PrevBlockHash.String() != self.config.BlockChain.BestState[block.Header.ChainID].BestBlockHash.String() {
		return errChainNotFullySynced
	}
	// 1. Check blocksize
	err := self.CheckBlockSize(block)
	if err != nil {
		return err
	}

	// 2. Check whether signature of the block belongs to chain leader or not.
	decPubkey, _, err := base58.Base58Check{}.Decode(block.ChainLeader)
	if err != nil {
		return err
	}
	k := cashec.KeySetSealer{
		SpublicKey: decPubkey,
	}
	decSig, _, err := base58.Base58Check{}.Decode(block.ChainLeaderSig)
	if err != nil {
		return err
	}
	cmsBytes, _ := json.Marshal(block.Header.CommitteeSigs)
	isValidSignature, err := k.Verify(cmsBytes, decSig)
	if err != nil {
		return err
	}
	if isValidSignature == false {
		return errSigWrongOrNotExits
	}

	// 3. Check whether we acquire enough data to validate this block
	if self.validatedChainsHeight.Heights[block.Header.ChainID] == (int(block.Height) - 1) {
		notFullySync := false
		for i := 0; i < TOTAL_VALIDATORS; i++ {
			if self.validatedChainsHeight.Heights[i] < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
				notFullySync = true
				getBlkMsg := &wire.MessageGetBlocks{
					LastBlockHash: self.config.BlockChain.BestState[i].BestBlockHash.String(),
				}
				peerIDs := self.config.Server.GetPeerIdsFromPublicKey(block.ChainLeader)
				if len(peerIDs) != 0 {
					Logger.log.Info("Send getblock to "+peerIDs[0], block.ChainLeader)
					self.config.Server.PushMessageToPeer(getBlkMsg, peerIDs[0])
				} else {
					fmt.Println("Validator's peer not found!", block.ChainLeader)
				}
			}
		}
		if notFullySync {
			timer := time.NewTimer(MAX_SYNC_CHAINS_TIME * time.Second)
			<-timer.C
			for i := 0; i < TOTAL_VALIDATORS; i++ {
				if int(self.config.BlockChain.BestState[i].Height) < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
					return errChainNotFullySynced
				}
			}
		}
	} else {
		return errChainNotFullySynced
	}

	// 4. Validate MerkleRootCommitments
	err = self.ValidateMerkleRootCommitments(block)
	if err != nil {
		return err
	}

	// 5. Validate committee member signatures
	err = self.ValidateCommitteeSigs([]byte(block.Hash().String()), block.Header.CommitteeSigs)
	if err != nil {
		return err
	}
	// 6. Validate transactions
	return self.ValidateTxList(block.Transactions)

}

func (self *Engine) validatePreSignBlock(block *blockchain.Block) error {
	// validate steps: block size -> sealer is belong to committee -> validate sealer's sig -> check chainsHeight of this block -> validate each transaction

	if block.Header.PrevBlockHash.String() != self.config.BlockChain.BestState[block.Header.ChainID].BestBlockHash.String() {
		return errChainNotFullySynced
	}
	// 1. Check whether block size is greater than MAX_BLOCKSIZE or not.
	err := self.CheckBlockSize(block)
	if err != nil {
		return err
	}

	// 2. Check signature of the block leader
	decPubkey, _, err := base58.Base58Check{}.Decode(block.ChainLeader)
	if err != nil {
		return err
	}
	k := cashec.KeySetSealer{
		SpublicKey: decPubkey,
	}
	decSig, _, err := base58.Base58Check{}.Decode(block.Header.CommitteeSigs[block.ChainLeader])
	if err != nil {
		return err
	}
	isValidSignature, err := k.Verify([]byte(block.Hash().String()), decSig)
	if err != nil {
		return err
	}
	if isValidSignature == false {
		return errSigWrongOrNotExits
	}

	// 4. Check chains height of the block.
	if self.validatedChainsHeight.Heights[block.Header.ChainID] == (int(block.Height) - 1) {
		notFullySync := false
		for i := 0; i < TOTAL_VALIDATORS; i++ {
			Logger.log.Info("--------------------------------------------------------")
			Logger.log.Info(len(self.validatedChainsHeight.Heights))
			Logger.log.Info(len(block.Header.ChainsHeight))
			if self.validatedChainsHeight.Heights[i] < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
				notFullySync = true
				getBlkMsg := &wire.MessageGetBlocks{
					LastBlockHash: self.config.BlockChain.BestState[i].BestBlockHash.String(),
				}
				peerIDs := self.config.Server.GetPeerIdsFromPublicKey(block.ChainLeader)
				if len(peerIDs) != 0 {
					Logger.log.Info("Send getblock to "+peerIDs[0], block.ChainLeader)
					self.config.Server.PushMessageToPeer(getBlkMsg, peerIDs[0])
				} else {
					fmt.Println("Validator's peer not found!", block.ChainLeader)
				}
			}
		}
		if notFullySync {
			timer := time.NewTimer(MAX_SYNC_CHAINS_TIME * time.Second)
			<-timer.C
			for i := 0; i < TOTAL_VALIDATORS; i++ {
				if int(self.config.BlockChain.BestState[i].Height) < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
					return errChainNotFullySynced
				}
			}
		}
	} else {
		return errChainNotFullySynced
	}

	// 5. Validate MerkleRootCommitments
	err = self.ValidateMerkleRootCommitments(block)
	if err != nil {
		return err
	}

	// 6. Validate transactions
	return self.ValidateTxList(block.Transactions)
}

// get validator chainID and committee of that chainID
func (self *Engine) getMyChain() byte {
	pkey := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))
	for idx := byte(0); idx < byte(TOTAL_VALIDATORS); idx++ {
		validator := self.currentCommittee[int((1+int(idx))%TOTAL_VALIDATORS)]
		if pkey == validator {
			return idx
		}
	}
	return TOTAL_VALIDATORS // nope, you're not in the committee
}

func (self *Engine) OnRequestSign(msgBlock *wire.MessageRequestSign) {
	block := &msgBlock.Block
	err := self.validatePreSignBlock(block)
	if err != nil {
		invalidBlockMsg := &wire.MessageInvalidBlock{
			Reason:    err.Error(),
			BlockHash: block.Hash().String(),
			ChainID:   block.Header.ChainID,
			Validator: base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)),
		}
		dataByte, _ := invalidBlockMsg.JsonSerialize()
		invalidBlockMsg.ValidatorSig, err = self.signData(dataByte)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		Logger.log.Critical("Invalid block msg", invalidBlockMsg)
		err = self.config.Server.PushMessageToAll(invalidBlockMsg)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		return
	}

	sig, err := self.signData([]byte(block.Hash().String()))
	if err != nil {
		Logger.log.Critical("OHSHITT", err)
		// ??? something went terribly wrong
		return
	}
	blockSigMsg := wire.MessageBlockSig{
		BlockHash:    block.Hash().String(),
		Validator:    base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)),
		ValidatorSig: sig,
	}
	peerID, err := peer2.IDB58Decode(msgBlock.SenderID)
	if err != nil {
		Logger.log.Error("ERROR", msgBlock.SenderID, peerID, err)
	}
	Logger.log.Info(block.Hash().String(), blockSigMsg)
	err = self.config.Server.PushMessageToPeer(&blockSigMsg, peerID)
	if err != nil {
		Logger.log.Error(err)
	}
	return
}

func (self *Engine) OnBlockReceived(block *blockchain.Block) {
	if self.config.BlockChain.BestState[block.Header.ChainID].Height < block.Height {
		if _, _, err := self.config.BlockChain.GetBlockHeightByBlockHash(block.Hash()); err != nil {
			err := self.validateBlock(block)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			isMainChain, ok, err := self.config.BlockChain.ProcessBlock(block)
			_ = isMainChain
			_ = ok
			if err != nil {
				Logger.log.Error(err)
				return
			}
			self.UpdateChain(block)
			err = self.config.FeeEstimator[block.Header.ChainID].RegisterBlock(block)
			if err != nil {
				Logger.log.Error(err)
			}
			self.knownChainsHeight.Lock()
			if self.knownChainsHeight.Heights[block.Header.ChainID] < int(block.Height) {
				self.knownChainsHeight.Heights[block.Header.ChainID] = int(block.Height)
			}
			self.knownChainsHeight.Unlock()
		}
	} else {
		//save block to cache
	}
	return
}

func (self *Engine) OnBlockSigReceived(blockHash string, validator string, sig string) {
	Logger.log.Info("Received a block signature")
	self.blockSigCh <- blockSig{
		BlockHash:    blockHash,
		Validator:    validator,
		ValidatorSig: sig,
	}
	return
}

func (self *Engine) OnInvalidBlockReceived(blockHash string, chainID byte, reason string) {
	// leave empty for now
	Logger.log.Error(blockHash, chainID, reason)
	return
}

func (self *Engine) OnChainStateReceived(msg *wire.MessageChainState) {
	// fmt.Println(msg)
	chainInfo := msg.ChainInfo.(map[string]interface{})
	for i, v := range self.validatedChainsHeight.Heights {
		if chainInfo["ChainsHeight"] != nil {
			if v < int(chainInfo["ChainsHeight"].([]interface{})[i].(float64)) {
				self.knownChainsHeight.Heights[i] = int(chainInfo["ChainsHeight"].([]interface{})[i].(float64))
				lastBlockHash := self.config.BlockChain.BestState[i].BestBlockHash.String()
				Logger.log.Info("############################")
				Logger.log.Infof("ChainId: %d", i)
				Logger.log.Infof("best state with block has: %d", lastBlockHash)
				Logger.log.Info("############################")
				getBlkMsg := &wire.MessageGetBlocks{
					LastBlockHash: lastBlockHash,
				}
				Logger.log.Info("Send getblock to " + msg.SenderID)
				peerID, err := peer2.IDB58Decode(msg.SenderID)
				if err != nil {
					continue
				}
				self.config.Server.PushMessageToPeer(getBlkMsg, peerID)
			}
		} else {
			Logger.log.Error("what the ...")
		}
	}
	return
}

func (self *Engine) OnGetChainState(msg *wire.MessageGetChainState) {
	chainInfo := ChainInfo{
		CurrentCommittee:  self.currentCommittee,
		CandidateListHash: "",
		ChainsHeight:      self.validatedChainsHeight.Heights,
	}
	newMsg, err := wire.MakeEmptyMessage(wire.CmdChainState)
	if err != nil {
		return
	}
	newMsg.(*wire.MessageChainState).ChainInfo = chainInfo
	peerID, _ := peer2.IDB58Decode(msg.SenderID)
	self.config.Server.PushMessageToPeer(newMsg, peerID)
	return
}

func (self *Engine) UpdateChain(block *blockchain.Block) {
	// save block into fee estimator
	self.config.FeeEstimator[block.Header.ChainID].RegisterBlock(block)

	// save best state
	newBestState := &blockchain.BestState{}
	// numTxns := uint64(len(block.Transactions))
	for _, tx := range block.Transactions {
		self.config.MemPool.RemoveTx(tx)
	}
	tree := self.config.BlockChain.BestState[block.Header.ChainID].CmTree
	self.config.BlockChain.UpdateMerkleTreeForBlock(tree, block)
	newBestState.Init(block, tree)
	self.config.BlockChain.BestState[block.Header.ChainID] = newBestState
	self.config.BlockChain.StoreBestState(block.Header.ChainID)

	// save index of block
	self.config.BlockChain.StoreBlockIndex(block)
	self.validatedChainsHeight.Lock()
	self.validatedChainsHeight.Heights[block.Header.ChainID] = int(block.Height)
	self.validatedChainsHeight.Unlock()
}
