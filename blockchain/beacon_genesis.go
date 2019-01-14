package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// @Hung: genesis should be build as configuration file like JSON
func CreateBeaconGenesisBlock(
	version int,
	beaconNodes []string,
	icoParams IcoParams,
	salaryPerTx uint64,
	basicSalary uint64,
	randomnumber int,
) *BeaconBlock {
	//TODO: build param
	inst := [][]string{}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssingInstruction := []string{"stake"}
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(preSelectBeaconNodeTestnetSerializedPubkey[:], ","))
	beaconAssingInstruction = append(beaconAssingInstruction, "beacon")

	shardAssingInstruction := []string{"stake"}
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(preSelectShardNodeTestnetSerializedPubkey[:], ","))
	shardAssingInstruction = append(shardAssingInstruction, "shard")

	inst = append(inst, beaconAssingInstruction)
	inst = append(inst, shardAssingInstruction)
	// build network param
	inst = append(inst, []string{"set", "salaryPerTx", fmt.Sprintf("%v", salaryPerTx)})
	inst = append(inst, []string{"set", "basicSalary", fmt.Sprintf("%v", basicSalary)})
	inst = append(inst, []string{"set", "initialPaymentAddress", icoParams.InitialPaymentAddress})
	inst = append(inst, []string{"set", "initFundSalary", strconv.Itoa(int(icoParams.InitFundSalary))})
	inst = append(inst, []string{"set", "initialDCBToken", strconv.Itoa(int(icoParams.InitialDCBToken))})
	inst = append(inst, []string{"set", "initialCMBToken", strconv.Itoa(int(icoParams.InitialCMBToken))})
	inst = append(inst, []string{"set", "initialGOVToken", strconv.Itoa(int(icoParams.InitialGOVToken))})
	inst = append(inst, []string{"set", "initialBondToken", strconv.Itoa(int(icoParams.InitialBondToken))})
	inst = append(inst, []string{"set", "randomnumber", strconv.Itoa(int(0))})

	body := BeaconBody{ShardState: nil, Instructions: inst}
	header := BeaconHeader{
		Timestamp: time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:    1,
		Version:   1,
		//TODO:

	}

	block := &BeaconBlock{
		Body:   body,
		Header: header,
	}

	return block
}
