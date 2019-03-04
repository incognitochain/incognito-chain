package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ninjadotorg/constant/common"
)

func CreateBeaconGenesisBlock(
	version int,
	genesisParams GenesisParams,
) *BeaconBlock {
	inst := [][]string{}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssingInstruction := []string{"stake"}
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssingInstruction = append(beaconAssingInstruction, "beacon")

	shardAssingInstruction := []string{"stake"}
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssingInstruction = append(shardAssingInstruction, "shard")

	inst = append(inst, beaconAssingInstruction)
	inst = append(inst, shardAssingInstruction)
	// build network param
	inst = append(inst, []string{"set", "salaryPerTx", fmt.Sprintf("%v", genesisParams.SalaryPerTx)})
	inst = append(inst, []string{"set", "basicSalary", fmt.Sprintf("%v", genesisParams.BasicSalary)})
	inst = append(inst, []string{"set", "initialPaymentAddress", genesisParams.InitialPaymentAddress})
	inst = append(inst, []string{"set", "initFundSalary", strconv.Itoa(int(genesisParams.InitFundSalary))})
	inst = append(inst, []string{"set", "initialDCBToken", strconv.Itoa(int(genesisParams.InitialDCBToken))})
	inst = append(inst, []string{"set", "initialCMBToken", strconv.Itoa(int(genesisParams.InitialCMBToken))})
	inst = append(inst, []string{"set", "initialGOVToken", strconv.Itoa(int(genesisParams.InitialGOVToken))})
	inst = append(inst, []string{"set", "initialBondToken", strconv.Itoa(int(genesisParams.InitialBondToken))})
	inst = append(inst, []string{"set", "randomnumber", strconv.Itoa(int(0))})

	body := BeaconBody{ShardState: nil, Instructions: inst}
	header := BeaconHeader{
		Timestamp:           time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:              1,
		Version:             1,
		Round:               1,
		Epoch:               1,
		PrevBlockHash:       common.Hash{},
		ValidatorsRoot:      common.Hash{},
		BeaconCandidateRoot: common.Hash{},
		ShardCandidateRoot:  common.Hash{},
		ShardValidatorsRoot: common.Hash{},
		ShardStateHash:      common.Hash{},
		InstructionHash:     common.Hash{},
	}

	block := &BeaconBlock{
		Body:   body,
		Header: header,
	}

	return block
}
