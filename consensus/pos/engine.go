package pos

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/mempool"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/mining"
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
	upComingCommittee     []string
	knownChainsHeight     []int
	validatedChainsHeight struct {
		Heights []int
		sync.Mutex
	}
	validatorSigCh chan blockSig
	waitForMyTurn  chan struct{}
}

type Config struct {
	BlockChain       *blockchain.BlockChain
	ChainParams      *blockchain.Params
	BlockGen         *mining.BlkTmplGenerator
	MemPool          *mempool.TxPool
	ValidatorKeyPair cashec.KeyPair
	Server           interface {
		// list functions callback which are assigned from Server struct
		PushBlockMessage(*blockchain.Block) error
		PushMessageToAll(wire.Message) error
		PushMessageToPeerID(wire.Message, peer2.ID) error
		GetChainState() error
		UpdateChain(*blockchain.Block)
	}
}

type blockSig struct {
	BlockHash string
	ChainID   byte
	BlockSig  string
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
	// for chainID := 0; chainID < 20; chainID++ {

	// }
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
	if err := self.checkIsLatest(); err != nil {

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

	// self.Config.Server.PushMessageToPeerID()
	var finalBlock *blockchain.Block

	// Wait for signatures of other validators
	for {
		// validatorSig := <-self.validatorSigCh

	}

	self.config.Server.UpdateChain(finalBlock)
	self.config.Server.PushBlockMessage(finalBlock)
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
	return nil
}

func (self *Engine) validatePreSignBlock(block *blockchain.Block) error {
	// validate steps: block size -> sealer is belong to committee -> validate sealer's sig -> check chainsHeight of this block -> validate each transaction
	return nil
}
func (self *Engine) checkIsLatest() error {

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
	signedBlockMsg := &wire.MessageSignedBlock{
		BlockHash: block.Hash().String(),
		ChainID:   block.Header.ChainID,
		Validator: string(self.config.ValidatorKeyPair.PublicKey),
		BlockSig:  string(sig),
	}
	dataByte, _ := signedBlockMsg.JsonSerialize()
	signedBlockMsg.ValidatorSig, err = self.signData(dataByte)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	err = self.config.Server.PushMessageToPeerID(signedBlockMsg, msgBlock.SenderID)
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

func (self *Engine) OnBlockSigReceived(blockHash string, chainID byte, sig string) {
	self.validatorSigCh <- blockSig{
		BlockHash: blockHash,
		ChainID:   chainID,
		BlockSig:  sig,
	}
	return
}

func (self *Engine) OnInvalidBlockReceived(blockHash string, chainID byte, reason string) {
	return
}

func (self *Engine) OnChainStateReceived(*wire.MessageChainState) {
	return
}

func (self *Engine) OnGetChainState(*wire.MessageGetChainState) {
	return
}

func (self *Engine) RequestChainState() []int {

	return nil
}
