package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/blockchain/btc/btcapi"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

// make sure beststate are up to date
func (self *BlkTmplGenerator) NewBlockBeacon(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey) (*BeaconBlock, error) {
	block := &BeaconBlock{}
	// Create Header
	block.Header.Producer = base58.Base58Check{}.Encode(self.chain.config.Wallet.MasterAccount.Key.KeySet.PaymentAddress.Pk, byte(0x00))
	block.Header.Version = VERSION
	block.Header.ParentHash = self.chain.BestState.Beacon.BestBlockHash
	block.Header.Height = self.chain.BestState.Beacon.BeaconHeight + 1
	block.Header.Epoch = self.chain.BestState.Beacon.BeaconEpoch
	if block.Header.Height%200 == 0 {
		block.Header.Epoch++
	}
	block.Header.Timestamp = time.Now().Unix()

	tempShardState := self.GetShardState()
	tempShardStateArr := []common.Hash{}
	for _, hashes := range tempShardState {
		tempShardStateArr = append(tempShardStateArr, hashes...)
	}
	tempShardStateHash, err := GenerateHashFromHashArray(tempShardStateArr)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	block.Header.ShardStateHash = tempShardStateHash

	tempInstruction := self.GenerateInstruction()
	tempInstructionArr := []string{}
	for _, strs := range tempInstruction {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := GenerateHashFromStringArray(tempInstructionArr)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	block.Header.InstructionHash = tempInstructionHash

	//Update Validator root and candidate root

	//Create Body
	block.Body.Instructions = tempInstruction
	block.Body.ShardState = tempShardState
	return block, nil
}

// TODO: Get from blockpool
func (self *BlkTmplGenerator) GetShardState() [][]common.Hash {
	return [][]common.Hash{}
}

func (self *BlkTmplGenerator) GenerateInstruction() [][]string {
	// TODO:
	// - set instruction
	// - del instruction
	// - swap instruction
	// - random instruction
	random := GenerateRandomNumber(self.chain.BestState.Beacon.CurrentRandomTimeStamp)
	Logger.log.Info("RandomNumber", random)
	return [][]string{}
}

func GenerateRandomNumber(timestamp int64) int64 {
	msg := make(chan int64)

	go btcapi.GenerateRandomNumber(timestamp, msg)
	res := <-msg
	return res
}
