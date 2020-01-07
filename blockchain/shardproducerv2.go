package blockchain

import (
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (blockchain *BlockChain) processInstructionFromBeaconV2(beaconBlocks []*BeaconBlock, shardID byte, committeeChange *committeeChange) ([]string, []string, map[string]string) {
	newShardPendingValidator := []string{}
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardPendingValidator)
	if err != nil {
		panic(err)
	}
	assignInstructions := [][]string{}
	stakingTx := make(map[string]string)
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// Process Assign Instruction
			if l[0] == AssignAction && l[2] == "shard" {
				if strings.Compare(l[3], strconv.Itoa(int(shardID))) == 0 {
					tempNewShardPendingValidator := strings.Split(l[1], ",")
					shardPendingValidator = append(shardPendingValidator, tempNewShardPendingValidator...)
					newShardPendingValidator = append(newShardPendingValidator, tempNewShardPendingValidator...)
					assignInstructions = append(assignInstructions, l)
				}
			}
			// Get Staking Tx
			// assume that stake instruction already been validated by beacon committee
			if l[0] == StakeAction && l[2] == "beacon" {
				beacon := strings.Split(l[1], ",")
				newBeaconCandidates := []string{}
				newBeaconCandidates = append(newBeaconCandidates, beacon...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						txShardID, _, _, _, err := blockchain.GetTransactionByHash(*txHash)
						if err != nil {
							continue
						}
						if txShardID != shardID {
							continue
						}
						// if transaction belong to this shard then add to shard beststate
						stakingTx[newBeaconCandidates[i]] = v
					}
				}
			}
			if l[0] == StakeAction && l[2] == "shard" {
				shard := strings.Split(l[1], ",")
				newShardCandidates := []string{}
				newShardCandidates = append(newShardCandidates, shard...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						txShardID, _, _, _, err := blockchain.GetTransactionByHash(*txHash)
						if err != nil {
							continue
						}
						if txShardID != shardID {
							continue
						}
						// if transaction belong to this shard then add to shard beststate
						stakingTx[newShardCandidates[i]] = v
					}
				}
			}
		}
	}
	return shardPendingValidator, newShardPendingValidator, stakingTx
}
