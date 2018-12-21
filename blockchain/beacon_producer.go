package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

// make sure beststate are up to date
func (self *BlkTmplGenerator) NewBlockBeacon(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey) (*BeaconBlock, error) {
	block := &BeaconBlock{}
	block.Header.Version = VERSION
	block.Header.ParentHash = self.chain.BestState.Beacon.BestBlockHash
	block.Header.Height = self.chain.BestState.Beacon.BeaconHeight + 1
	block.Header.Epoch = self.chain.BestState.Beacon.BeaconEpoch
	if block.Header.Height%200 == 0 {
		block.Header.Epoch++
	}
	block.Header.Timestamp = time.Now().Unix()
	// tempShardState := self.GetShardState()

	return block, nil
}

// Get from blockpool
func (self *BlkTmplGenerator) GetShardState() [][]common.Hash {
	return [][]common.Hash{}
}

func (self *BlkTmplGenerator) GenerateInstruction() ([][]string, error) {
	// TODO:
	// - set instruction
	// - del instruction
	// - swap instruction
	// - random instruction
	return [][]string{}, nil
}
