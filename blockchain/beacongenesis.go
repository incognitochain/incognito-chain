package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

func CreateBeaconGenesisBlock(
	version int,
	genesisParams GenesisParams,
) *BeaconBlock {
	inst := [][]string{}
	shardAutoStaking := []string{}
	beaconAutoStaking := []string{}
	for i := 0; i < TestNetMinShardCommitteeSize; i++ {
		shardAutoStaking = append(shardAutoStaking, "false")
	}
	for i := 0; i < TestNetMinBeaconCommitteeSize; i++ {
		beaconAutoStaking = append(beaconAutoStaking, "false")
	}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssingInstruction := []string{StakeAction}
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssingInstruction = append(beaconAssingInstruction, "beacon")
	beaconAssingInstruction = append(beaconAssingInstruction, []string{""}...)
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPaymentAddress[:], ","))
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(beaconAutoStaking[:], ","))

	shardAssingInstruction := []string{StakeAction}
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssingInstruction = append(shardAssingInstruction, "shard")
	shardAssingInstruction = append(shardAssingInstruction, []string{""}...)
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPaymentAddress[:], ","))
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(shardAutoStaking[:], ","))

	inst = append(inst, beaconAssingInstruction)
	inst = append(inst, shardAssingInstruction)

	// init network param
	inst = append(inst, []string{SetAction, "randomnumber", strconv.Itoa(int(0))})

	layout := "2006-01-02T15:04:05.000Z"
	str := "2018-08-01T00:00:00.000Z"
	genesisTime, err := time.Parse(layout, str)

	if err != nil {
		fmt.Println(err)
	}
	body := BeaconBody{ShardState: nil, Instructions: inst}
	header := BeaconHeader{
		Timestamp:                       genesisTime.Unix(),
		Height:                          1,
		Version:                         1,
		Round:                           1,
		Epoch:                           1,
		PreviousBlockHash:               common.Hash{},
		BeaconCommitteeAndValidatorRoot: common.Hash{},
		BeaconCandidateRoot:             common.Hash{},
		ShardCandidateRoot:              common.Hash{},
		ShardCommitteeAndValidatorRoot:  common.Hash{},
		ShardStateHash:                  common.Hash{},
		InstructionHash:                 common.Hash{},
	}

	block := &BeaconBlock{
		Body:   body,
		Header: header,
	}

	return block
}
