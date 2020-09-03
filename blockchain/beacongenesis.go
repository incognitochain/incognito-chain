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
	net uint16,
	genesisBlockTime string,
	genesisParams *GenesisParams,
) *BeaconBlock {
	inst := [][]string{}
	shardAutoStaking := []string{}
	beaconAutoStaking := []string{}
	for i := 0; i < len(genesisParams.PreSelectShardNodeSerializedPubkey); i++ {
		shardAutoStaking = append(shardAutoStaking, "false")
	}
	for i := 0; i < len(genesisParams.PreSelectBeaconNodeSerializedPubkey); i++ {
		beaconAutoStaking = append(beaconAutoStaking, "false")
	}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssignInstruction := []string{StakeAction}
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, "beacon")
	beaconAssignInstruction = append(beaconAssignInstruction, []string{""}...)
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPaymentAddress[:], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(beaconAutoStaking[:], ","))

	shardAssignInstruction := []string{StakeAction}
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssignInstruction = append(shardAssignInstruction, "shard")
	shardAssignInstruction = append(shardAssignInstruction, []string{""}...)
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPaymentAddress[:], ","))
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(shardAutoStaking[:], ","))

	inst = append(inst, beaconAssignInstruction)
	inst = append(inst, shardAssignInstruction)

	// init network param
	inst = append(inst, []string{SetAction, "randomnumber", strconv.Itoa(int(0))})

	layout := "2006-01-02T15:04:05.000Z"
	str := genesisBlockTime
	genesisTime, err := time.Parse(layout, str)

	if err != nil {
		fmt.Println(err)
	}
	body := BeaconBody{ShardState: nil, Instructions: inst}
	header := BeaconHeader{
		Timestamp:                       genesisTime.Unix(),
		Version:                         version,
		Epoch:                           1,
		Height:                          1,
		Round:                           1,
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

func GetBeaconSwapInstructionKeyListV2(genesisParams *GenesisParams, epoch uint64) ([]string, []string) {
	newCommittees := genesisParams.SelectBeaconNodeSerializedPubkeyV2[epoch]
	newRewardReceivers := genesisParams.SelectBeaconNodeSerializedPaymentAddressV2[epoch]

	// TODO - in next replacement of committee validator key -> need read oldCommittees from prev-committee instead of from genesis block
	oldCommittees := genesisParams.PreSelectBeaconNodeSerializedPubkey
	// TODO

	beaconSwapInstructionKeyListV2 := []string{SwapAction, strings.Join(newCommittees, ","), strings.Join(oldCommittees, ","), "beacon", "", "", strings.Join(newRewardReceivers, ",")}
	return beaconSwapInstructionKeyListV2, newCommittees
}
