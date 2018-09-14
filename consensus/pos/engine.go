package pos

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/ninjadotorg/cash-prototype/transaction"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/mempool"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/wire"
)

var (
	errBlockSizeExceed     = errors.New("block size is too big")
	errNotInCommittee      = errors.New("user not in committee")
	errSigWrongOrNotExits  = errors.New("signature is wrong or not existed in block header")
	errChainNotFullySynced = errors.New("chains are not fully synced")
	errTxIsWrong           = errors.New("transaction is wrong")
	errNotEnoughChainData  = errors.New("not enough chain data")
	errCantFinalizeBlock   = errors.New("can't finalize block")
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
	upComingCommittee     []string
	knownChainsHeight     chainsHeight
	validatedChainsHeight chainsHeight
	validatorSigCh        chan blockSig
}

type ChainInfo struct {
	CurrentCommittee  []string
	UpComingCommittee []string
	ChainsHeight      []int
}
type chainsHeight struct {
	Heights []int
	sync.Mutex
}
type Config struct {
	BlockChain       *blockchain.BlockChain
	ChainParams      *blockchain.Params
	blockGen         *BlkTmplGenerator
	MemPool          *mempool.TxPool
	ValidatorKeyPair cashec.KeyPair
	Server           interface {
		// list functions callback which are assigned from Server struct
		GetPeerIdsFromPublicKey(string) []peer2.ID
		PushMessageToAll(wire.Message) error
		PushMessageToPeer(wire.Message, peer2.ID) error
		GetChainState() error
	}
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
	time.Sleep(1 * time.Second)
	Logger.log.Info("Starting Parallel Proof of Stake Consensus engine")
	self.started = true
	self.knownChainsHeight.Heights = make([]int, TOTAL_VALIDATORS)
	self.currentCommittee = make([]string, TOTAL_VALIDATORS)
	self.validatedChainsHeight.Heights = make([]int, TOTAL_VALIDATORS)

	for chainID := 0; chainID < TOTAL_VALIDATORS; chainID++ {
		self.knownChainsHeight.Heights[chainID] = int(self.config.BlockChain.BestState[chainID].Height + 1)
	}

	Logger.log.Info("Validating local blockchain...")
	for chainID := 0; chainID < TOTAL_VALIDATORS; chainID++ {
		//Don't validate genesis block
		for blockIdx := 1; blockIdx < self.knownChainsHeight.Heights[chainID]; blockIdx++ {
			block, err := self.config.BlockChain.GetBlockByBlockHeight(int32(blockIdx), byte(chainID))
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
	self.validatedChainsHeight.Heights = self.knownChainsHeight.Heights
	self.currentCommittee = self.config.BlockChain.BestState[0].BestBlock.Header.NextCommittee

	time.Sleep(2 * time.Second)
	go func() {
		for {
			self.config.Server.GetChainState()
			time.Sleep(10 * time.Second)
		}
	}()
	self.quit = make(chan struct{})
	self.wg.Add(1)

	// Test GetPeerIdsFromPublicKey
	//go func(){
	//	for {
	//		realPubKey := "vHX/aAVJsH4sHpYAHR3i1guLSE07QV3l"
	//		peerIds := self.config.Server.GetPeerIdsFromPublicKey(realPubKey)
	//		Logger.log.Info("DEBUG GetPeerIdsFromPublicKey", peerIds)
	//		time.Sleep(5 * time.Second)
	//	}
	//}()
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
	// self.wg.Wait()
	self.started = false
	Logger.log.Info("Consensus engine stopped")
	return nil
}

func New(Config *Config) *Engine {
	Config.blockGen = NewBlkTmplGenerator(Config.MemPool, Config.BlockChain)
	return &Engine{
		config: *Config,
	}
}

func (self *Engine) StopSealer() {
	if self.sealerStarted {
		Logger.log.Info("Stopping Sealer...")
		close(self.quitSealer)
		close(self.validatorSigCh)
		self.sealerStarted = false
	}
}
func (self *Engine) StartSealer(sealerPrvKey string) {
	if self.sealerStarted {
		Logger.log.Error("Sealer already started")
		return
	}

	_, err := self.config.ValidatorKeyPair.Import(sealerPrvKey)
	if err != nil {
		Logger.log.Error("Can't import sealer's key!")
		return
	}
	self.validatorSigCh = make(chan blockSig)
	self.quitSealer = make(chan struct{})
	self.sealerStarted = true
	Logger.log.Info("Starting sealer with public key: " + base64.StdEncoding.EncodeToString(self.config.ValidatorKeyPair.PublicKey))

	go func() {
		tempChainsHeight := make([]int, TOTAL_VALIDATORS)
		for {
			select {
			case <-self.quitSealer:
				return
			default:
				if !intArrayEquals(tempChainsHeight, self.knownChainsHeight.Heights) {
					chainID, validators := self.getMyChain()
					if chainID < TOTAL_VALIDATORS {
						Logger.log.Info("(๑•̀ㅂ•́)و Yay!! It's my turn")
						newBlock, err := self.createBlock()
						if err != nil {
							Logger.log.Error(err)
							continue
						}
						err = self.Finalize(newBlock, validators)
						if err != nil {
							Logger.log.Critical(err)
							continue
						}
						tempChainsHeight = self.knownChainsHeight.Heights
					}
				}
			}
		}
	}()
}

func (self *Engine) Finalize(block *blockchain.Block, chainValidators []string) error {
	finalBlock := block
	validateSigList := make(chan []string)
	timeout := time.After(5 * time.Second)

	// Collect signatures of other validators
	go func() {
		var reslist []string
		for {
			select {
			case <-timeout:
				Logger.log.Critical("oopss... Can't finalize block")
				return
			default:
				validatorSig := <-self.validatorSigCh
				if block.Hash().String() != validatorSig.BlockHash {
					Logger.log.Critical("(o_O)!")
					continue
				}
				validatorKp := cashec.KeyPair{
					PublicKey: []byte(validatorSig.Validator),
				}
				isValid, _ := validatorKp.Verify(block.Hash().CloneBytes(), []byte(validatorSig.ValidatorSig))
				if isValid {
					reslist = append(reslist, validatorSig.ValidatorSig)
				}
				if len(reslist) == 10 {
					validateSigList <- reslist
					return
				}
			}
		}
	}()

	//Request for signatures of other validators
	go func() {
		newMsg := &wire.MessageRequestSign{}
		newMsg.Block = *block
		for idx := 0; idx < CHAIN_VALIDATORS_LENGTH; idx++ {
			self.config.Server.GetPeerIdsFromPublicKey(chainValidators[idx])
		}
	}()

	// Wait for signatures of other validators
	select {
	case resList := <-validateSigList:
		fmt.Println(resList)
	case <-timeout:
		return errCantFinalizeBlock
	}

	self.UpdateChain(finalBlock)
	blockMsg, err := wire.MakeEmptyMessage(wire.CmdBlock)
	if err != nil {
		return err
	}
	self.config.Server.PushMessageToAll(blockMsg)
	return nil
}

func (self *Engine) createBlock() (*blockchain.Block, error) {
	myChainID, _ := self.getMyChain()
	newblock, err := self.config.blockGen.NewBlockTemplate(string(self.config.ValidatorKeyPair.PublicKey), self.config.BlockChain, myChainID)
	if err != nil {
		return &blockchain.Block{}, err
	}
	newblock.Block.Header.ChainID = myChainID
	return newblock.Block, nil
}

func (self *Engine) signData(data []byte) (string, error) {
	signatureByte, err := self.config.ValidatorKeyPair.Sign(data)
	if err != nil {
		return "", errors.New("Can't sign data. " + err.Error())
	}
	return string(signatureByte), nil
}

func (self *Engine) validateBlock(block *blockchain.Block) error {
	// validate steps: block size -> sealer's sig of the final block -> sealer is belong to committee -> validate each committee member's sig

	// 1. Check blocksize
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return err
	}
	if len(blockBytes) > MAX_BLOCKSIZE {
		return errBlockSizeExceed
	}

	// 2. Check whether signature of the block belongs to chain leader or not.
	k := cashec.KeyPair{
		PublicKey: []byte(block.ChainLeader),
	}
	isValidSignature, err := k.Verify(block.Hash().CloneBytes(), []byte(block.ChainLeaderSig))
	if err != nil {
		return err
	}
	if isValidSignature == false {
		return errSigWrongOrNotExits
	}
	for _, s := range block.Header.BlockCommitteeSigs {
		if strings.Compare(s, block.ChainLeaderSig) != 0 {
			return errSigWrongOrNotExits
		}
	}

	// 3. Check whether sealer of the block belongs to committee or not.
	for _, c := range self.currentCommittee {
		if strings.Compare(c, block.ChainLeader) != 0 {
			return errNotInCommittee
		}
	}

	// 4. Validate committee signatures
	for i, sig := range block.Header.BlockCommitteeSigs {
		k := cashec.KeyPair{
			PublicKey: []byte(self.currentCommittee[i]),
		}
		isValidSignature, err := k.Verify(block.Hash().CloneBytes(), []byte(sig))
		if err != nil {
			return err
		}
		if isValidSignature == false {
			return errSigWrongOrNotExits
		}
	}

	// 5. Revalidata transactions again @@
	for _, tx := range block.Transactions {
		if tx.ValidateTransaction() == false {
			return errTxIsWrong
		}
	}

	return nil
}

func (self *Engine) validatePreSignBlock(block *blockchain.Block) error {
	// validate steps: block size -> sealer is belong to committee -> validate sealer's sig -> check chainsHeight of this block -> validate each transaction

	// 1. Check whether block size is greater than MAX_BLOCKSIZE or not.
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return err
	}
	if len(blockBytes) > MAX_BLOCKSIZE {
		return errBlockSizeExceed
	}

	// 2. Check user is in current committee or not
	for _, c := range self.currentCommittee {
		if strings.Compare(c, block.ChainLeader) != 0 {
			return errNotInCommittee
		}
	}

	// 3. Check signature of the block belongs to current committee or not.
	k := cashec.KeyPair{
		PublicKey: []byte(block.ChainLeader),
	}
	isValidSignature, err := k.Verify(block.Hash().CloneBytes(), []byte(block.ChainLeaderSig))
	if err != nil {
		return err
	}
	if isValidSignature == false {
		return errSigWrongOrNotExits
	}
	for _, s := range block.Header.BlockCommitteeSigs {
		if strings.Compare(s, block.ChainLeaderSig) != 0 {
			return errSigWrongOrNotExits
		}
	}

	// 4. Check chains height of the block.
	for i := 0; i < TOTAL_VALIDATORS; i++ {
		if int(self.config.BlockChain.BestState[i].Height) < block.Header.ChainsHeight[i] {
			timer := time.NewTimer(MAX_SYNC_CHAINS_TIME * time.Second)
			<-timer.C
			break
		}
	}
	for i := 0; i < TOTAL_VALIDATORS; i++ {
		if int(self.config.BlockChain.BestState[i].Height) < block.Header.ChainsHeight[i] {
			return errChainNotFullySynced
		}
	}

	// 5. Revalidate transactions in block.
	for _, tx := range block.Transactions {
		if tx.ValidateTransaction() == false {
			return errTxIsWrong
		}
	}

	return nil
}

// get validator chainID and committee of that chainID
func (self *Engine) getMyChain() (byte, []string) {
	var myChainCommittee []string
	var err error
	for idx := byte(0); idx < byte(TOTAL_VALIDATORS); idx++ {
		myChainCommittee, err = self.getChainValidators(idx)
		if err != nil {
			return TOTAL_VALIDATORS, []string{}
		}
		if base64.StdEncoding.EncodeToString(self.config.ValidatorKeyPair.PublicKey) == myChainCommittee[0] {
			return idx, myChainCommittee
		}
	}
	return TOTAL_VALIDATORS, []string{} // nope, you're not in the committee
}

func (self *Engine) getChainValidators(chainID byte) ([]string, error) {
	var validators []string
	for index := 1; index <= CHAIN_VALIDATORS_LENGTH; index++ {
		validatorID := math.Mod(float64(index+int(chainID)), 20)
		validators = append(validators, self.currentCommittee[int(validatorID)])
	}
	if len(validators) == CHAIN_VALIDATORS_LENGTH {
		return validators, nil
	}
	return nil, errors.New("can't get chain's validators")
}

func (self *Engine) GetTxSenderChain(senderLastByte byte) (byte, error) {
	modResult := senderLastByte % 100
	for index := byte(0); index < 5; index++ {
		if (modResult-index)%5 == 0 {
			return byte((modResult - index) / 5), nil
		}
	}
	return 0, errors.New("can't get sender's chainID")
}

func (self *Engine) OnRequestSign(msgBlock *wire.MessageRequestSign) {
	block := &msgBlock.Block
	err := self.validatePreSignBlock(block)
	if err != nil {
		invalidBlockMsg := &wire.MessageInvalidBlock{
			Reason:    err.Error(),
			BlockHash: block.Hash().String(),
			ChainID:   block.Header.ChainID,
			Validator: string(self.config.ValidatorKeyPair.PublicKey),
		}
		dataByte, _ := invalidBlockMsg.JsonSerialize()
		invalidBlockMsg.ValidatorSig, err = self.signData(dataByte)
		if err != nil {
			Logger.log.Error(err)
			return
		}

		err = self.config.Server.PushMessageToAll(invalidBlockMsg)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		return
	}

	sig, err := self.config.ValidatorKeyPair.Sign([]byte(block.Hash().String()))
	if err != nil {
		// ??? something went terribly wrong
		return
	}
	blockSigMsg := &wire.MessageBlockSig{
		BlockHash:    block.Hash().String(),
		Validator:    string(self.config.ValidatorKeyPair.PublicKey),
		ValidatorSig: string(sig),
	}

	err = self.config.Server.PushMessageToPeer(blockSigMsg, msgBlock.SenderID)
	if err != nil {
		Logger.log.Error(err)
	}
	return
}

func (self *Engine) OnBlockReceived(block *blockchain.Block) {
	err := self.validateBlock(block)
	if err != nil {
		return
	}
	self.UpdateChain(block)
	return
}

func (self *Engine) OnBlockSigReceived(blockHash string, validator string, sig string) {
	self.validatorSigCh <- blockSig{
		BlockHash:    blockHash,
		Validator:    validator,
		ValidatorSig: sig,
	}
	return
}

func (self *Engine) OnInvalidBlockReceived(blockHash string, chainID byte, reason string) {
	// leave empty for now
	return
}

func (self *Engine) OnChainStateReceived(msg *wire.MessageChainState) {
	chainInfo := msg.ChainInfo.(ChainInfo)
	self.knownChainsHeight.Heights = chainInfo.ChainsHeight
	return
}

func (self *Engine) OnGetChainState(msg *wire.MessageGetChainState) {
	chainInfo := ChainInfo{
		CurrentCommittee:  self.currentCommittee,
		UpComingCommittee: self.upComingCommittee,
		ChainsHeight:      self.knownChainsHeight.Heights,
	}
	newMsg, err := wire.MakeEmptyMessage(wire.CmdChainState)
	if err != nil {
		return
	}
	newMsg.(*wire.MessageChainState).ChainInfo = chainInfo
	self.config.Server.PushMessageToPeer(newMsg, msg.SenderID)
	return
}

func (self *Engine) UpdateChain(block *blockchain.Block) {
	// save block
	self.config.BlockChain.StoreBlock(block)

	// save best state
	newBestState := &blockchain.BestState{}
	numTxns := uint64(len(block.Transactions))
	for _, tx := range block.Transactions {
		self.config.MemPool.RemoveTx(*tx.(*transaction.Tx))
	}
	newBestState.Init(block, 0, 0, numTxns, numTxns, time.Unix(block.Header.Timestamp.Unix(), 0))
	self.config.BlockChain.BestState[block.Header.ChainID] = newBestState
	self.config.BlockChain.StoreBestState(block.Header.ChainID)

	// save index of block
	self.config.BlockChain.StoreBlockIndex(block)
	self.validatedChainsHeight.Lock()
	self.knownChainsHeight.Lock()
	self.validatedChainsHeight.Heights[block.Header.ChainID]++
	self.knownChainsHeight.Heights[block.Header.ChainID]++
	self.knownChainsHeight.Unlock()
	self.validatedChainsHeight.Unlock()
}

func intArrayEquals(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
