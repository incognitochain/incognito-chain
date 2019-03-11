package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/big0t/constant-chain/common"
)

func CreateBeaconGenesisBlock(
	version int,
	genesisParams GenesisParams,
) *BeaconBlock {
	inst := [][]string{}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssingInstruction := []string{StakeAction}
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssingInstruction = append(beaconAssingInstruction, "beacon")

	shardAssingInstruction := []string{StakeAction}
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssingInstruction = append(shardAssingInstruction, "shard")

	inst = append(inst, beaconAssingInstruction)
	inst = append(inst, shardAssingInstruction)

	// init network param
	inst = append(inst, []string{"init", "salaryPerTx", fmt.Sprintf("%v", genesisParams.SalaryPerTx)})
	inst = append(inst, []string{"init", "basicSalary", fmt.Sprintf("%v", genesisParams.BasicSalary)})
	inst = append(inst, []string{"init", "salaryFund", strconv.Itoa(int(genesisParams.InitFundSalary))})
	inst = append(inst, []string{"init", "feePerTxKb", fmt.Sprintf("%v", genesisParams.FeePerTxKb)})

	inst = append(inst, []string{InitAction, "initialPaymentAddress", genesisParams.InitialPaymentAddress})
	inst = append(inst, []string{InitAction, "initialDCBToken", strconv.Itoa(int(genesisParams.InitialDCBToken))})
	inst = append(inst, []string{InitAction, "initialCMBToken", strconv.Itoa(int(genesisParams.InitialCMBToken))})
	inst = append(inst, []string{InitAction, "initialGOVToken", strconv.Itoa(int(genesisParams.InitialGOVToken))})
	inst = append(inst, []string{InitAction, "initialBondToken", strconv.Itoa(int(genesisParams.InitialBondToken))})

	inst = append(inst, []string{SetAction, "randomnumber", strconv.Itoa(int(0))})

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
