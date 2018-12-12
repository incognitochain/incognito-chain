package blockchain

import "time"

type BeaconBlockGenerator struct{}

// @Hung: genesis should be build as configuration file like JSON
func (self *BeaconBlockGenerator) CreateBeaconGenesisBlock(
	version int,
	beaconNodes []string,
	shardNodes []string,
	icoParams IcoParams,
	salaryPerTx uint64,
	basicSalary uint64,
) *BlockV2 {

	loc, _ := time.LoadLocation("America/New_York")
	time := time.Date(2018, 8, 1, 0, 0, 0, 0, loc)

	//TODO: build param
	inst := [][]string{}
	// build validator beacon
	inst = append(inst, []string{"assign", "...", "beacon"})
	// build validator shard
	inst = append(inst, []string{"assign", "...", "shard"})
	// build network param

	body := &BeaconBlockBody{ShardState: nil, Instructions: nil}
	header := &BeaconBlockHeader{
		BlockHeaderGeneric: BlockHeaderGeneric{
			PrevBlockHash: nil,
			Timestamp:     time.Unix(),
			Height:        1,
			Version:       1,
		},
		DataHash: body.Hash(),
	}

	block := &BlockV2{
		AggregatedSig: nil,
		ProducerSig:   nil,
		ValidatorsIdx: nil,
		Type:          "beacon",
		Body:          body,
		Header:        header,
	}

	return block
}
