package blockchain

import "time"

type BeaconBlockGenerator struct{}

// @Hung: genesis should be build as configuration file like JSON
func (self *BeaconBlockGenerator) createBeaconGenesisBlock() *BlockV2 {
	loc, _ := time.LoadLocation("America/New_York")
	time := time.Date(2018, 8, 1, 0, 0, 0, 0, loc)

	//TODO: build param

	body := &BeaconBlockBody{ShardState: nil, StateInstruction: nil}
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
