package bnb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	bnbdb "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/types"
	"sort"
)

var dbPath = ""
func setDBPath(path string) {
	if dbPath == "" {
		dbPath = path
	}
}

type BNBChainState struct {
	FinalBlocks         []*types.Block           // there are two blocks behind
	LatestBlock         *types.Block             // there is one block behind (in candidate next blocks)
	CandidateNextBlocks []*types.Block           // candidates for next latest block
	OrphanBlocks        map[int64][]*types.Block // orphan blocks, waiting to be appended to candidates

	ChainDB bnbdb.DB
}

func (b *BNBChainState) GetBNBBlockByHeight(h int64) (*types.Block, error) {
	blockStore := NewBlockStore(b.ChainDB)
	block := blockStore.LoadBlock(h)
	if block == nil {
		return nil, errors.New("[GetBNBBlockByHeight] Block is nil")
	}
	return block, nil
}

func (b *BNBChainState) GetDataHashBNBBlockByHeight(h int64) ([]byte, error) {
	block, err := b.GetBNBBlockByHeight(h)
	if err != nil {
		return nil, err
	}
	return block.DataHash, nil
}

func newLatestBlockKey() []byte {
	return []byte("latestblock")
}

func newCandidateBlocksKey() []byte {
	return []byte("candidateblock")
}

func newOrphanBlockKey() []byte {
	return []byte("orphanblock")
}

func (b *BNBChainState) StoreBNBChainState() error {
	var err error
	// store FinalBlocks into db (final blocks, there are two blocks behind)
	blockStore := NewBlockStore(b.ChainDB)
	for i := 0; i < len(b.FinalBlocks); i++ {
		block := b.FinalBlocks[i]
		commit := new(types.Commit)
		if i == len(b.FinalBlocks)-1 {
			commit = b.LatestBlock.LastCommit
		} else {
			commit = b.FinalBlocks[i+1].LastCommit
		}
		err := blockStore.SaveBlock(block, block.MakePartSet(types.BlockPartSizeBytes), commit)
		if err != nil {
			Logger.log.Errorf("Error when store final blocks %v\n", err)
			return NewBNBRelayingError(StoreBNBChainErr, err)
		}
	}
	b.FinalBlocks = []*types.Block{}

	// store LatestBlock
	err = storeLatestBlock(b.ChainDB, b.LatestBlock)
	if err != nil {
		Logger.log.Errorf("Error when store bnb latest block %v\n", err)
		return NewBNBRelayingError(StoreBNBChainErr, err)
	}

	// store CandidateNextBlocks
	err = storeCandidateBlocks(b.ChainDB, b.CandidateNextBlocks)
	if err != nil {
		Logger.log.Errorf("Error when store bnb latest block %v\n", err)
		return NewBNBRelayingError(StoreBNBChainErr, err)
	}

	// store OrphanBlocks
	err = storeOrphanBlocks(b.ChainDB, b.OrphanBlocks)
	if err != nil {
		Logger.log.Errorf("Error when store bnb latest block %v\n", err)
		return NewBNBRelayingError(StoreBNBChainErr, err)
	}
	return nil
}

func storeLatestBlock(db bnbdb.DB, b *types.Block) error {
	key := newLatestBlockKey()
	value, err := b.Marshal()
	if err != nil {
		Logger.log.Errorf("Error when marshaling block %v\n", err)
		return NewBNBRelayingError(StoreBNBChainErr, err)
	}

	db.Set(key, value)
	return nil
}

func storeCandidateBlocks(db bnbdb.DB, blocks []*types.Block) error {
	key := newCandidateBlocksKey()
	value, err := json.Marshal(blocks)

	if err != nil {
		Logger.log.Errorf("Error when marshaling block %v\n", err)
		return NewBNBRelayingError(StoreBNBChainErr, err)
	}

	db.Set(key, value)
	return nil
}

func storeOrphanBlocks(db bnbdb.DB, orphanBlocks map[int64][]*types.Block) error {
	key := newOrphanBlockKey()
	value, err := json.Marshal(orphanBlocks)

	if err != nil {
		Logger.log.Errorf("Error when marshaling block %v\n", err)
		return NewBNBRelayingError(StoreBNBChainErr, err)
	}

	db.Set(key, value)
	return nil
}

func getLatestBlock(db bnbdb.DB) (*types.Block, error) {
	key := newLatestBlockKey()
	value := db.Get(key)
	block := new(types.Block)
	if len(value) > 0 {
		err := block.Unmarshal(value)
		if err != nil {
			Logger.log.Errorf("Error when unmarshaling block %v\n", err)
			return nil, NewBNBRelayingError(GetBNBChainErr, err)
		}
	}
	return block, nil
}

func getCandidateBlocks(db bnbdb.DB) ([]*types.Block, error) {
	key := newCandidateBlocksKey()
	value := db.Get(key)
	blocks := []*types.Block{}
	if len(value) > 0 {
		err := json.Unmarshal(value, &blocks)
		if err != nil {
			Logger.log.Errorf("Error when unmarshaling block %v\n", err)
			return nil, NewBNBRelayingError(GetBNBChainErr, err)
		}
	}
	return blocks, nil
}

func getOrphanBlocks(db bnbdb.DB) (map[int64][]*types.Block, error) {
	key := newOrphanBlockKey()
	value := db.Get(key)
	blocks := map[int64][]*types.Block{}
	if len(value) > 0 {
		err := json.Unmarshal(value, &blocks)
		if err != nil {
			Logger.log.Errorf("Error when unmarshaling block %v\n", err)
			return nil, NewBNBRelayingError(GetBNBChainErr, err)
		}
	}
	return blocks, nil
}

func (b *BNBChainState) LoadBNBChainState(path string, chainID string) error {
	setDBPath(path)
	// if there is no block in db, add genesis block to final blocks
	var err error
	if b.ChainDB == nil {
		b.ChainDB = bnbdb.NewDB("bnbchain", bnbdb.GoLevelDBBackend, path)
	}

	blockStore := NewBlockStore(b.ChainDB)
	if blockStore.Height() == 0 {
		genesisBlock, _ := getGenesisBlock(chainID)
		b.LatestBlock = genesisBlock
	} else {
		// load LatestBlock
		b.LatestBlock, err = getLatestBlock(b.ChainDB)
		if err != nil {
			Logger.log.Errorf("Error when get bnb latest block %v\n", err)
			return NewBNBRelayingError(GetBNBChainErr, err)
		}
	}

	// load CandidateNextBlocks
	b.CandidateNextBlocks, err = getCandidateBlocks(b.ChainDB)
	if err != nil {
		Logger.log.Errorf("Error when get candidate blocks %v\n", err)
		return NewBNBRelayingError(GetBNBChainErr, err)
	}

	// load OrphanBlocks
	b.OrphanBlocks, err = getOrphanBlocks(b.ChainDB)
	if err != nil {
		Logger.log.Errorf("Error when get bnb orphan blocks %v\n", err)
		return NewBNBRelayingError(GetBNBChainErr, err)
	}

	return nil
}

func appendBlockToBlocksArray(b *types.Block, blocks []*types.Block) ([]*types.Block, error) {
	if blocks == nil {
		return []*types.Block{b}, nil
	}
	newBlockHash := b.Header.Hash().Bytes()
	for _, block := range blocks {
		if bytes.Equal(block.Hash().Bytes(), newBlockHash) {
			Logger.log.Errorf("Block is existed %v\n", newBlockHash)
			return blocks, fmt.Errorf("Block is existed %v\n", newBlockHash)
		}
	}
	blocks = append(blocks, b)
	return blocks, nil
}

func removeBlock(blocks []*types.Block, index int) []*types.Block {
	return append(blocks[:index], blocks[index+1:]...)
}

func (b *BNBChainState) addBlockToOrphanBlocks(block *types.Block) error {
	var err error
	if len(b.OrphanBlocks) >= MaxOrphanBlocks {
		Logger.log.Errorf("[addBlockToOrphanBlocks] Orphan blocks is full")
		return NewBNBRelayingError(FullOrphanBlockErr, errors.New("orphan blocks is full"))
	}

	b.OrphanBlocks[block.Height], err = appendBlockToBlocksArray(block, b.OrphanBlocks[block.Height])
	if err != nil {
		Logger.log.Errorf("[addBlockToOrphanBlocks] Error when append block to orphan blocks %v\n", err)
		return NewBNBRelayingError(AddBlockToOrphanBlockErr, err)
	}
	return nil
}

func (b *BNBChainState) ProcessNewBlock(block *types.Block, chainID string) error {
	// verify lastCommit
	if block.LastCommit == nil {
		Logger.log.Errorf("[ProcessNewBlock] last commit is nil")
		return NewBNBRelayingError(InvalidNewHeaderErr, errors.New("last commit is nil"))
	}
	if !bytes.Equal(block.LastCommitHash, block.LastCommit.Hash()) {
		Logger.log.Errorf("[ProcessNewBlock] invalid last commit hash")
		return NewBNBRelayingError(InvalidNewHeaderErr, errors.New("invalid last commit hash"))
	}

	// validate block height
	if block.Height <= b.LatestBlock.Height {
		Logger.log.Errorf("[ProcessNewBlock] block.Height must be greater than latestBlock.Height")
		return NewBNBRelayingError(InvalidNewHeaderErr, errors.New("block.Height must be greater than latestBlock.Height"))
	}

	isCase1, isCase2, err := b.handleBlock(block, chainID)
	if isCase2 && err != nil {
		// add to orphan blocks, pre-processing
		// maybe the block is the confirmation block of other block that hasn't pushed to candidate blocks yet
		err := b.addBlockToOrphanBlocks(block)
		if err != nil {
			Logger.log.Errorf("[ProcessNewBlock] Error when append block to orphan blocks %v\n", err)
			return NewBNBRelayingError(InvalidNewHeaderErr, err)
		}
	}

	// else
	// add the block to orphan blocks
	if !isCase1 && !isCase2 {
		err := b.addBlockToOrphanBlocks(block)
		if err != nil {
			Logger.log.Errorf("[ProcessNewBlock] Error when append block to orphan blocks %v\n", err)
			return NewBNBRelayingError(InvalidNewHeaderErr, err)
		}
	}

	// check if there is any block in orphan blocks that confirms one of blocks in candidate blocks
	// or confirms latest block
	// remove orphan blocks that have block height < height of the current latest block + 1
	err = b.checkOrphanBlocks(chainID)
	if err != nil {
		Logger.log.Errorf("[ProcessNewBlock] Error when append block to orphan blocks %v\n", err)
		return NewBNBRelayingError(CheckOrphanBlockErr, err)
	}

	return nil
}

func (b *BNBChainState) handleBlock(block *types.Block, chainID string) (bool, bool, error) {
	// Case 1:
	// if block is the next block of the latest block, add the block to candidate blocks
	if block.Height == b.LatestBlock.Height+1 {
		// check last blockID
		if bytes.Equal(block.LastBlockID.Hash.Bytes(), b.LatestBlock.Hash()) {
			newSignedHeader := NewSignedHeader(&b.LatestBlock.Header, block.LastCommit)
			isValid, err := VerifySignedHeader(newSignedHeader, chainID)
			if isValid && err == nil {
				b.CandidateNextBlocks, err = appendBlockToBlocksArray(block, b.CandidateNextBlocks)
				if err != nil {
					Logger.log.Errorf("[ProcessNewBlock] Error when append block to candidate blocks %v\n", err)
					return true, false, NewBNBRelayingError(InvalidNewHeaderErr, err)
				}
				Logger.log.Infof("[ProcessNewBlock] Case 1: Receive new candidate block %v\n", block.Height)
				return true, false, nil
			} else {
				Logger.log.Errorf("[ProcessNewBlock] invalid new signed header %v", err)
				return true, false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
		}
	}

	// Case 2:
	// if block A is the next block of one of blocks (block B) in candidate blocks,
	// add current latest block to final blocks
	// block B becomes latest block
	// clean candidate blocks and add block A to candidate blocks
	if block.Height == b.LatestBlock.Height+2 {
		for _, cb := range b.CandidateNextBlocks {
			if block.Height == cb.Height+1 && bytes.Equal(block.LastBlockID.Hash.Bytes(), cb.Hash()) {
				newSignedHeader := NewSignedHeader(&cb.Header, block.LastCommit)
				isValid, err := VerifySignedHeader(newSignedHeader, chainID)
				if isValid && err == nil {
					b.FinalBlocks, err = appendBlockToBlocksArray(b.LatestBlock, b.FinalBlocks)
					if err != nil {
						Logger.log.Errorf("[ProcessNewBlock] Error when append current latest block to final blocks %v\n", err)
						return false, true, NewBNBRelayingError(InvalidNewHeaderErr, err)
					}
					b.LatestBlock = cb
					b.CandidateNextBlocks = []*types.Block{block}
					Logger.log.Infof("[ProcessNewBlock] Case 2 new confirmation block for one of candidate blocks %v\n", block.Height)
					return false, true, nil
				} else {
					Logger.log.Errorf("[ProcessNewBlock] invalid new signed header %v", err)
					return false, true, NewBNBRelayingError(InvalidNewHeaderErr, err)
				}
			}
		}
	}

	return false, false, nil
}

func (b *BNBChainState) checkOrphanBlocks(chainID string) error {
	// sort orphan blocks before checking
	blkHeightKeys := []int64{}
	for key := range b.OrphanBlocks {
		blkHeightKeys = append(blkHeightKeys, key)
	}
	sort.Slice(blkHeightKeys, func(i, j int) bool {
		return blkHeightKeys[i] < blkHeightKeys[j]
	})

	// if block height < height of the current latest block
	// remove it from orphan blocks

	// if block height = latest block height + 1
	// add to candidate next blocks (if valid)
	// remove it (if invalid)

	// if block height = latest block height + 2
	// check whether the block is the confirmation block of one of candidate blocks

	// else, do nothing
	for blkHeight, blks := range b.OrphanBlocks {
		if blkHeight <= b.LatestBlock.Height {
			delete(b.OrphanBlocks, blkHeight)
		} else {
			for i := 0; i < len(blks); i++ {
				isCase1, isCase2, err := b.handleBlock(blks[i], chainID)
				if isCase1 || (isCase2 && err == nil) {
					b.OrphanBlocks[blkHeight] = removeBlock(b.OrphanBlocks[blkHeight], i)
					if len(b.OrphanBlocks[blkHeight]) == 0 {
						delete(b.OrphanBlocks, blkHeight)
					}
				}
			}
		}
	}
	return nil
}
