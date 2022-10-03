package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"testing"
)

// test repropose block
// -> previous valid block
// -> lock block hash
// test prevote
// -> only prevote if not lock other hash
func TestActorV3_ReproposeBlock(t *testing.T) {
	backendLog := common.NewBackend(logWriter{})
	logger := backendLog.Logger("Consensus log", false)
	chain := blockchain.NewBeaconChain(nil, nil, nil, "beacon")
	bftactor := NewActorV3WithValue(chain, chain, "beacon", 9, -1, nil, logger)
	bftactor.Stop()
	bftactor.currentBestViewHeight = 1
	bftactor.currentTimeSlot = common.CalculateTimeSlot(100)

	x, _ := InitReceiveBlockByHash(-1)
	if len(x) != 0 {
		t.Error("First init should be 0")
	}

	if bftactor.getBlockForPropose(2) != nil {
		t.Error("Should not get repropose block if there is no record!")
	}

	block := &types.BeaconBlock{Header: types.BeaconHeader{Version: 9, Height: 2, Timestamp: 100, ProposeTime: 100}}

	proposeBlockInfo := &ProposeBlockInfo{
		block:    block,
		Votes:    make(map[string]*BFTVote),
		PreVotes: make(map[string]*BFTVote),
	}
	bftactor.AddReceiveBlockByHash(block.ProposeHash().String(), proposeBlockInfo)
	if bftactor.getBlockForPropose(2) != nil {
		t.Error("Should not get repropose block if block is not valid!")
	}

	proposeBlockInfo.IsValid = true
	bftactor.AddReceiveBlockByHash(block.ProposeHash().String(), proposeBlockInfo)

	if bftactor.getBlockForPropose(2) == nil {
		t.Error("Should get repropose block if block is valid!")
	}

	bftactor.currentTimeSlot = common.CalculateTimeSlot(110)
	block = &types.BeaconBlock{Header: types.BeaconHeader{Version: 9, Height: 2, Timestamp: 110, ProposeTime: 110}}
	proposeLockBlockInfo := &ProposeBlockInfo{
		block:    block,
		Votes:    make(map[string]*BFTVote),
		PreVotes: make(map[string]*BFTVote),
	}
	proposeLockBlockInfo.IsValid = true
	proposeLockBlockInfo.IsVoted = true
	bftactor.AddReceiveBlockByHash(block.ProposeHash().String(), proposeLockBlockInfo)

	if len(bftactor.GetSortedReceiveBlockByHeight(2)) != 2 {
		t.Error("Should update 2 proposed block, get", len(bftactor.GetSortedReceiveBlockByHeight(2)))
	}

	lockBlockHash := bftactor.getBlockForPropose(2)
	if lockBlockHash.GetProposeTime() != 110 {
		t.Error("Locked blockhash should be block that is voted")
	}

	bftactor.currentTimeSlot = common.CalculateTimeSlot(120)
	block.Header.ProposeTime = 120
	proposeLockBlockInfo1 := &ProposeBlockInfo{
		block:    block,
		Votes:    make(map[string]*BFTVote),
		PreVotes: make(map[string]*BFTVote),
	}
	proposeLockBlockInfo1.IsValid = true
	if !bftactor.shouldPrevote(proposeLockBlockInfo1) {
		t.Error("must repropose lock block hash")
	}

	if bftactor.shouldPrevote(proposeBlockInfo) {
		t.Error("must not repropose block if different than locked blockhash")
	}

	x, _ = InitReceiveBlockByHash(-1)
	if len(x) != 2 {
		t.Error("restore from db should return all stored records")
	}
	bftactor.receiveBlockByHash = x
	lockBlockHash = bftactor.getBlockForPropose(2)
	if lockBlockHash.GetProposeTime() != 110 {
		t.Error("Locked blockhash should be voted block")
	}

}
