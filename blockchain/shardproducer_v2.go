package blockchain

func (blockchain *BlockChain) NewBlockShard_V2(curView *ShardBestState, version int, proposer string, round int, shardsToBeaconLimit map[byte]uint64) (newBlock *ShardBlock, err error) {
	processState := &ShardProcessState{
		curView:          curView,
		newView:          nil,
		blockchain:       blockchain,
		version:          version,
		proposer:         proposer,
		round:            round,
		newBlock:         NewShardBlock(),
		beaconBlocks:     []*BeaconBlock{},
		crossShardBlocks: make(map[byte][]*CrossShardBlock),
	}

	if err := processState.PreProduceProcess(); err != nil {
		return nil, err
	}

	if err := processState.BuildBody(); err != nil {
		return nil, err
	}

	processState.newView, err = processState.curView.updateShardBestState(blockchain, processState.newBlock, processState.beaconBlocks, newCommitteeChange())
	if err != nil {
		return nil, err
	}

	if err := processState.BuildHeader(); err != nil {
		return nil, err
	}

	return processState.newBlock, nil
}

type ShardProcessState struct {
	//init state
	curView    *ShardBestState
	newView    *ShardBestState
	blockchain *BlockChain
	version    int
	proposer   string
	round      int

	//pre process state
	newBlock         *ShardBlock
	beaconBlocks     []*BeaconBlock
	crossShardBlocks map[byte][]*CrossShardBlock
}

func (s *ShardProcessState) PreProduceProcess() error {
	//get beacon blocks and get shard blocks
	return nil
}

func (s *ShardProcessState) BuildBody() (err error) {
	return
}

func (s *ShardProcessState) BuildHeader() (err error) {
	return
}
