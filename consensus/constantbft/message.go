package constantbft

import (
	"encoding/json"
	"time"

	"github.com/ninjadotorg/constant/common"

	"github.com/ninjadotorg/constant/blockchain"

	"github.com/ninjadotorg/constant/wire"
)

func (self *Engine) OnBFTMsg(msg wire.Message) {
	self.cBFTMsg <- msg
	return
}

// func (self *Engine) OnBFTPropose(msg *wire.MessageBFTPropose) {
// 	self.cBFTMsg <- msg
// 	return
// }

// func (self *Engine) OnBFTPrepare(msg *wire.MessageBFTPrepare) {
// 	self.cBFTMsg <- msg
// 	return

// }

// func (self *Engine) OnBFTCommit(msg *wire.MessageBFTCommit) {
// 	self.cBFTMsg <- msg
// 	return

// }

// func (self *Engine) OnBFTReady(msg *wire.MessageBFTReady) {
// 	self.cBFTMsg <- msg
// 	return

// }

func (self *Engine) OnInvalidBlockReceived(blockHash string, shardID byte, reason string) {
	// leave empty for now
	Logger.log.Error(blockHash, shardID, reason)
	return
}

func MakeMsgBFTPropose(block json.RawMessage) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTPropose)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBFTPropose).Block = block
	return msg, nil
}

func MakeMsgBFTPrepare(Ri []byte, pubkey string, blkHash string) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTPrepare)
	if err != nil {
		Logger.log.Error(err)

		return msg, err
	}
	msg.(*wire.MessageBFTPrepare).Ri = Ri
	msg.(*wire.MessageBFTPrepare).Pubkey = pubkey
	msg.(*wire.MessageBFTPrepare).BlkHash = blkHash
	return msg, nil
}

func MakeMsgBFTCommit(commitSig string, R string, validatorsIdx []int, pubkey string) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTCommit)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBFTCommit).CommitSig = commitSig
	msg.(*wire.MessageBFTCommit).R = R
	msg.(*wire.MessageBFTCommit).ValidatorsIdx = validatorsIdx
	msg.(*wire.MessageBFTCommit).Pubkey = pubkey
	return msg, nil
}

func MakeMsgBFTReady(bestStateHash common.Hash) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTReady)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBFTReady).BestStateHash = bestStateHash
	msg.(*wire.MessageBFTReady).Timestamp = time.Now().Unix()
	return msg, nil
}

func MakeMsgBeaconBlock(block *blockchain.BeaconBlock) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBlockBeacon).Block = *block
	return msg, nil
}

func MakeMsgShardBlock(block *blockchain.ShardBlock) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBlockShard).Block = *block
	return msg, nil
}

func MakeMsgShardToBeaconBlock(block *blockchain.ShardToBeaconBlock) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageShardToBeacon).Block = *block
	return msg, nil
}

func MakeMsgCrossShardBlock(block *blockchain.CrossShardBlock) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdCrossShard)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageCrossShard).Block = *block
	return msg, nil
}
