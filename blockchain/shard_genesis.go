package blockchain

import "time"

//TODO Write function to create Shard block of shard chain here
func CreateShardGenesisBlock(
	version int,
	shardNodes []string,
	shardsNum int,
	icoParams IcoParams,
) *ShardBlock {

	body := ShardBody{}
	header := ShardHeader{
		Timestamp: time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:    1,
		Version:   1,
		//TODO:

	}

	block := &ShardBlock{
		Body:   body,
		Header: header,
	}

	return block
}
