package blockchain

import (
	"strconv"
	"strings"
	"time"

	"github.com/ninjadotorg/constant/blockchain/btc/btcapi"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

// make sure beststate are up to date
// snapshot beststate before create new block
// beststate will be updated with new block
// if new block fail to connect to blockchain
// roll back best state
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

	tempInstruction := self.GenerateInstruction(block)
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

func (self *BlkTmplGenerator) GenerateInstruction(block *BeaconBlock) [][]string {
	// TODO:
	// - set instruction
	// - del instruction
	// - swap instruction
	//    + format
	//    + ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,...") "shard" "shardID"]
	//    + ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,...") "beacon"]
	// - random instruction
	// - assign instruction
	instructions := [][]string{}
	// TODO process unexpected swap
	if block.Header.Height%EPOCH == EPOCH-1 {
		beacanPendingValidator, beaconCommittee, swappedValidator, beaconNextCommittee, _ := SwapValidator(self.chain.BestState.Beacon.BeaconPendingValidator, self.chain.BestState.Beacon.BeaconCommittee, OFFSET)
		swapInstructions := []string{}
		swapInstructions = append(swapInstructions, "swap")
		swapInstructions = append(swapInstructions, beaconNextCommittee...)
		swapInstructions = append(swapInstructions, swappedValidator...)
		swapInstructions = append(swapInstructions, "beacon")
		instructions = append(instructions, swapInstructions)

		validatorArr := append(beaconCommittee, beacanPendingValidator...)
		var err error
		block.Header.ValidatorsRoot, err = GenerateHashFromStringArray(validatorArr)
		if err != nil {
			panic(err)
		}
	} else {
		validatorArr := append(self.chain.BestState.Beacon.BeaconCommittee, self.chain.BestState.Beacon.BeaconPendingValidator...)
		var err error
		block.Header.ValidatorsRoot, err = GenerateHashFromStringArray(validatorArr)
		if err != nil {
			panic(err)
		}
	}
	// Time to get random number and no block in this epoch get it
	if block.Header.Height%200 >= RANDOM_TIME && self.chain.BestState.Beacon.IsGetRandomNUmber == false {
		chainTimeStamp, err := btcapi.GetCurrentChainTimeStamp()
		if err != nil {
			panic(err)
		}
		if chainTimeStamp > self.chain.BestState.Beacon.CurrentRandomTimeStamp {
			randomInstruction := GenerateRandomInstruction(self.chain.BestState.Beacon.CurrentRandomTimeStamp)
			instructions = append(instructions, randomInstruction)
			Logger.log.Infof("RandomNumber %+v", randomInstruction)
			//TODO
			// process stake transaction to get staking candidate
			beaconStaker := []string{}
			shardStaker := []string{}
			beaconAssingInstruction := []string{"assign"}
			beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(beaconStaker, ","))
			beaconAssingInstruction = append(beaconAssingInstruction, "beacon")

			shardAssingInstruction := []string{"assign"}
			shardAssingInstruction = append(shardAssingInstruction, strings.Join(shardStaker, ","))
			shardAssingInstruction = append(shardAssingInstruction, "shard")
		}
	}
	return [][]string{}
}

// TODO: Get stake from staking tx (block pool)
// return param #1: beacon staker
// return param #2: shard staker
func (self *BlkTmplGenerator) GetStakerPublicKey() ([]string, []string) {
	return []string{}, []string{}
}

// ["random" "{blockheight}" "{bitcointimestamp}" "{nonce}" "{timestamp}"]
func GenerateRandomInstruction(timestamp int64) []string {
	msg := make(chan string)

	go btcapi.GenerateRandomNumber(timestamp, msg)
	res := <-msg
	reses := strings.Split(res, (","))
	strs := []string{}
	strs = append(strs, "random")
	strs = append(strs, reses...)
	strs = append(strs, strconv.Itoa(int(timestamp)))
	return strs
}
