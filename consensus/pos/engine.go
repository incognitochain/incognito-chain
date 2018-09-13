package pos

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/mempool"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/mining"
	"github.com/ninjadotorg/cash-prototype/wire"
)

var (
	errBlockSizeExceed     = errors.New("block size is too big")
	errNotInCommittee      = errors.New("user not in committee")
	errSigWrongOrNotExits  = errors.New("signature is wrong or not existed in block header")
	errChainNotFullySynced = errors.New("chains are not fully synced")
	errTxIsWrong           = errors.New("transaction is wrong")
	errNotEnoughChainData  = errors.New("not enough chain data")
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
	knownChainsHeight     []int
	validatedChainsHeight struct {
		Heights []int
		sync.Mutex
	}
	validatorSigCh chan blockSig
	waitForMyTurn  chan struct{}
}

type ChainInfo struct {
	CurrentCommittee  []string
	UpComingCommittee []string
	ChainsHeight      []int
}
type Config struct {
	BlockChain       *blockchain.BlockChain
	ChainParams      *blockchain.Params
	BlockGen         *mining.BlkTmplGenerator
	MemPool          *mempool.TxPool
	ValidatorKeyPair cashec.KeyPair
	Server           interface {
		// list functions callback which are assigned from Server struct
		PushMessageToAll(wire.Message) error
		PushMessageToPeer(wire.Message, peer2.ID) error
		GetChainState() error
		UpdateChain(*blockchain.Block)
	}
}

type blockSig struct {
	BlockHash    string
	Validator    string
	ValidatorSig string
}

func (self *Engine) Start() error {
	self.Lock()
	if self.started {
		self.Unlock()
		return errors.New("Consensus engine is already started")
	}
	Logger.log.Info("Starting Parallel Proof of Stake Consensus engine")
	self.started = true
	self.knownChainsHeight = make([]int, 20)
	self.currentCommittee = make([]string, 21)
	self.validatedChainsHeight.Heights = make([]int, 20)

	for chainID := 0; chainID < 20; chainID++ {
		self.knownChainsHeight[chainID] = int(self.config.BlockChain.BestState[chainID].Height)
	}

	Logger.log.Info("Validating local blockchain...")
	for chainID := byte(0); chainID < 20; chainID++ {
		for blockIdx := 0; blockIdx < self.knownChainsHeight[chainID]; blockIdx++ {
			blockHash, err := self.config.BlockChain.GetBlockHashByBlockHeight(int32(blockIdx), chainID)
			if err != nil {
				return err
			}
			err = self.validateBlock(self.config.BlockChain.GetBlockByHash(*blockHash))
			if err != nil {
				return err
			}
		}
	}

	self.quit = make(chan struct{})
	self.wg.Add(1)
	self.Unlock()
	return nil
}

func (self *Engine) Stop() error {
	self.Lock()
	defer self.Unlock()

	if !self.started {
		return errors.New("Consensus engine isn't running")
	}
	close(self.quit)
	self.StopSealer()
	// self.wg.Wait()
	self.started = false
	fmt.Print("Consensus engine stopped")
	return nil
}

func New(Config *Config) *Engine {
	return &Engine{
		config: *Config,
	}
}

func (self *Engine) StopSealer() {
	if self.sealerStarted {
		close(self.quitSealer)
		close(self.waitForMyTurn)
		close(self.validatorSigCh)
		self.sealerStarted = false
	}
}
func (self *Engine) StartSealer(sealerPrvKey []byte) {
	if self.sealerStarted {
		Logger.log.Error("Sealer already started")
		return
	}
	_, err := self.config.ValidatorKeyPair.Import(sealerPrvKey)
	if err != nil {
		Logger.log.Error("Can't import sealer's key!")
		return
	}
	self.waitForMyTurn = make(chan struct{})
	self.validatorSigCh = make(chan blockSig)
	self.quitSealer = make(chan struct{})
	self.sealerStarted = true
	Logger.log.Info("Starting sealing...")
	go func() {
		for {
			select {
			case <-self.quitSealer:
				return
			case <-self.waitForMyTurn:
				newBlock, err := self.createBlock()
				if err != nil {
					Logger.log.Critical(err)
				}
				self.Finalize(newBlock)
			}
		}
	}()

	go func() {
		// tempChainsHeight := make([]int, 20)
		for {
			select {
			case <-self.quitSealer:
				return
			default:

			}
		}
	}()
}

func (self *Engine) Finalize(block *blockchain.Block) {
	select {
	case <-self.quitSealer:
		return
	}
	// request signature from other validators
	for {

	}
	newMsg := &wire.MessageRequestSign{}
	newMsg.Block = *block

	// self.Config.Server.PushMessageToPeer()
	finalBlock := block
	validateSigList := make(chan []string)
	// Wait for signatures of other validators
	go func() {
		for {
			validatorSig := <-self.validatorSigCh
			if block.Hash().String() != validatorSig.BlockHash {

			}
		}
	}()

	select {
	case resList := <-validateSigList:
		fmt.Println(resList)
	case <-time.After(5 * time.Second):
		Logger.log.Critical()
		return
	}

	self.config.Server.UpdateChain(finalBlock)
	// todo refactor PushBlockMessage to PushMessageToAll
	//self.config.Server.PushBlockMessage(finalBlock)
}

func (self *Engine) SwitchMember() {

}

func (self *Engine) createBlock() (*blockchain.Block, error) {
	myChainID, myChainCommittee := self.getMyChain()
	fmt.Println(myChainID, myChainCommittee)
	newblock, err := self.config.BlockGen.NewBlockTemplate(string(self.config.ValidatorKeyPair.PublicKey), self.config.BlockChain, myChainID)
	if err != nil {
		return newblock.Block, err
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
	for i := 0; i < 20; i++ {
		if int(self.config.BlockChain.BestState[i].Height) < block.Header.ChainsHeight[i] {
			timer := time.NewTimer(MAX_SYNC_CHAINS_TIME * time.Second)
			<-timer.C
			break
		}
	}
	for i := 0; i < 20; i++ {
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
	for index := byte(0); index < 20; index++ {
		myChainCommittee, err = self.getChainValidators(index)
		if err != nil {
			return 20, []string{}
		}
		if string(self.config.ValidatorKeyPair.PublicKey) == myChainCommittee[0] {
			return index, myChainCommittee
		}
	}
	return 20, []string{} // the math wrong some where 😭
}

func (self *Engine) getChainValidators(chainID byte) ([]string, error) {
	var validators []string
	for index := 1; index <= 11; index++ {
		validatorID := math.Mod(float64(index+int(chainID)), 21)
		validators = append(validators, self.currentCommittee[int(validatorID)])
	}
	if len(validators) == 11 {
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
	self.config.Server.UpdateChain(block)
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

	return
}

func (self *Engine) OnChainStateReceived(msg *wire.MessageChainState) {

	return
}

func (self *Engine) OnGetChainState(msg *wire.MessageGetChainState) {
	chainInfo := ChainInfo{
		CurrentCommittee:  self.currentCommittee,
		UpComingCommittee: self.upComingCommittee,
		ChainsHeight:      self.knownChainsHeight,
	}
	newMsg, err := wire.MakeEmptyMessage(wire.CmdChainState)
	if err != nil {
		return
	}
	newMsg.(*wire.MessageChainState).ChainInfo = chainInfo
	self.config.Server.PushMessageToPeer(newMsg, msg.SenderID)
	return
}
