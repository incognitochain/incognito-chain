package pos

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/mempool"

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

	Config             Config
	CurrentCommittee   []string
	NextBlockCandidate []string
	CurrentLeader      string
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
		PushBlockSignature(*wire.MessageSignedBlock) error
		PushRequestSignBlock(*blockchain.Block, string) error
		PushInvalidBlockMessage(*wire.MessageInvalidBlock) error
		UpdateChain(*blockchain.Block)
	}
}

func (self *Engine) Start(sealerPrvKey []byte) error {
	self.Lock()
	// Respond with an error if server is already mining.
	if self.started {
		self.Unlock()
		return errors.New("Consensus engine is already started")
	}
	self.quit = make(chan struct{})
	Logger.log.Info("Starting Parallel Proof of Stake Consensus engine")
	_, err := self.Config.ValidatorKeyPair.Import(sealerPrvKey)
	if err != nil {
		return errors.New("Can't import sealer's key!")
	}
	self.started = true
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
	// self.wg.Wait()
	self.started = false
	fmt.Print("Consensus engine stopped")
	return nil
}

func (self *Engine) createBlock(chainID byte) (*blockchain.Block, error) {
	newblock, err := self.Config.BlockGen.NewBlockTemplate(self.CurrentLeader, self.Config.BlockChain, chainID)
	if err != nil {
		return newblock.Block, err
	}
	return newblock.Block, nil
}

func (self *Engine) signData(data []byte) (string, error) {
	signatureByte, err := self.Config.ValidatorKeyPair.Sign(data)
	if err != nil {
		return "", errors.New("Can't sign data. " + err.Error())
	}
	return string(signatureByte), nil
}

func (self *Engine) validateBlock(block *blockchain.Block) error {
	return nil
}

func (self *Engine) GetChainValidators(chainID byte) ([]string, error) {
	var validators []string
	for index := 1; index <= 11; index++ {
		validatorID := math.Mod(float64(index+int(chainID)), 21)
		validators = append(validators, self.CurrentCommittee[int(validatorID)])
	}
	if len(validators) == 11 {
		return validators, nil
	}
	return nil, errors.New("can't get chain's validators")
}

func (self *Engine) GetSenderChain(senderLastByte byte) (byte, error) {
	// addrBig := new(big.Int)
	// addrBig.SetBytes([]byte{senderLastByte})

	// b := make([]byte, 4)
	// binary.BigEndian.PutUint32(b, uint32(100))
	// modNum := new(big.Int)
	// modNum.SetBytes(b)

	// modResult := new(big.Int)
	// modResult = modResult.Mod(addrBig, modNum)

	// for index := uint64(0); index < 5; index++ {
	// 	if (modResult.Uint64()-index)%5 == 0 {
	// 		return byte((modResult.Uint64() - index) / 5), nil
	// 	}
	// }

	modResult := senderLastByte % 100
	for index := byte(0); index < 5; index++ {
		if (modResult-index)%5 == 0 {
			return byte((modResult - index) / 5), nil
		}
	}
	return 0, errors.New("can't get sender's chainID")
}

func (self *Engine) OnRequestSign(block *blockchain.Block) {
	err := self.validateBlock(block)
	if err != nil {
		invalidBlockMsg := &wire.MessageInvalidBlock{
			Reason:    err.Error(),
			BlockHash: block.Hash().String(),
			ChainID:   block.Header.ChainID,
			Validator: string(self.Config.ValidatorKeyPair.PublicKey),
		}
		dataByte, _ := invalidBlockMsg.JsonSerialize()
		invalidBlockMsg.ValidatorSig, err = self.signData(dataByte)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		err = self.Config.Server.PushInvalidBlockMessage(invalidBlockMsg)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		return
	}

	return
}

func (self *Engine) OnBlockReceived(block *blockchain.Block) {
	err := self.validateBlock(block)
	if err != nil {
		return
	}
	self.Config.Server.UpdateChain(block)
	return
}

// func (self *Engine) filterTx() []*transaction.Transaction {

// }

func New(Config *Config) *Engine {
	return &Engine{
		Config: *Config,
	}
}
