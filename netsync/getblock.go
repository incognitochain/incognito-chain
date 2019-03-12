package netsync

import (
	"sort"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

func (netSync *NetSync) GetBlkShardByHashAndSend(peerID libp2p.ID, blkType byte, blkHashes []common.Hash, crossShardID byte) {
	for _, blkHash := range blkHashes {
		blk, err, _ := netSync.config.BlockChain.GetShardBlockByHash(&blkHash)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		blkMsg, err := netSync.CreateBlkShardMsgByType(blk, blkType, crossShardID)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		err = netSync.config.Server.PushMessageToPeer(blkMsg, peerID)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

func (netSync *NetSync) GetBlkBeaconByHashAndSend(peerID libp2p.ID, blkHashes []common.Hash) {
	for _, blkHash := range blkHashes {
		blk, err, _ := netSync.config.BlockChain.GetBeaconBlockByHash(&blkHash)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg.(*wire.MessageBlockBeacon).Block = *blk
		netSync.config.Server.PushMessageToPeer(newMsg, peerID)
	}
}

func (netSync *NetSync) GetBlkShardByHeightAndSend(peerID libp2p.ID, fromPool bool, blkType byte, specificHeight bool, shardID byte, blkHeights []uint64, crossShardID byte) {
	if !specificHeight {
		if len(blkHeights) != 2 || blkHeights[1] < blkHeights[0] {
			return
		}
	}
	sort.Slice(blkHeights, func(i, j int) bool { return blkHeights[i] < blkHeights[j] })
	var (
		blkHeight uint64
		idx       int
		err       error
	)
	if !specificHeight {
		blkHeight = blkHeights[0] - 1
	}
	for blkHeight < blkHeights[len(blkHeights)-1] {
		if specificHeight {
			blkHeight = blkHeights[idx]
			idx++
		} else {
			blkHeight++
		}
		if blkHeight <= 1 {
			continue
		}
		var blkMsg wire.Message
		if fromPool {
			switch blkType {
			case 1:
				blkToSend := netSync.config.CrossShardPool[shardID].GetBlockByHeight(crossShardID, blkHeight)
				if blkToSend == nil {
					Logger.log.Error(err)
					continue
				}
				blkMsg, err = wire.MakeEmptyMessage(wire.CmdCrossShard)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				blkMsg.(*wire.MessageCrossShard).Block = *blkToSend
			case 2:
				blkToSend := netSync.config.ShardToBeaconPool.GetBlockByHeight(shardID, blkHeight)
				if blkToSend == nil {
					Logger.log.Error(err)
					continue
				}
				blkMsg, err = wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				blkMsg.(*wire.MessageShardToBeacon).Block = *blkToSend
			}
		} else {
			blk, err := netSync.config.BlockChain.GetShardBlockByHeight(blkHeight, shardID)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			blkMsg, err = netSync.CreateBlkShardMsgByType(blk, blkType, crossShardID)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
		}
		err = netSync.config.Server.PushMessageToPeer(blkMsg, peerID)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

func (netSync *NetSync) GetBlkBeaconByHeightAndSend(peerID libp2p.ID, fromPool bool, specificHeight bool, blkHeights []uint64) {
	if !specificHeight {
		if len(blkHeights) != 2 || blkHeights[1] < blkHeights[0] {
			return
		}
	}
	sort.Slice(blkHeights, func(i, j int) bool { return blkHeights[i] < blkHeights[j] })
	var (
		blkHeight uint64
		idx       int
	)
	if !specificHeight {
		blkHeight = blkHeights[0] - 1
	}
	for blkHeight < blkHeights[len(blkHeights)-1] {
		if specificHeight {
			blkHeight = blkHeights[idx]
			idx++
		} else {
			blkHeight++
		}
		if blkHeight <= 1 {
			continue
		}
		blk, err := netSync.config.BlockChain.GetBeaconBlockByHeight(blkHeight)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		msgBeaconBlk, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		msgBeaconBlk.(*wire.MessageBlockBeacon).Block = *blk
		err = netSync.config.Server.PushMessageToPeer(msgBeaconBlk, peerID)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

// blkType:
// 0: normal
// 1: crossShard
// 2: shardToBeacon
func (netSync *NetSync) CreateBlkShardMsgByType(block *blockchain.ShardBlock, blkType byte, crossShardID byte) (wire.Message, error) {
	var (
		blkMsg wire.Message
		err    error
	)
	switch blkType {
	case 0:
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageBlockShard).Block = *block
	case 1:
		blkToSend, err := block.CreateCrossShardBlock(crossShardID)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdCrossShard)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageCrossShard).Block = *blkToSend
	case 2:
		blkToSend := block.CreateShardToBeaconBlock(netSync.config.BlockChain)
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageShardToBeacon).Block = *blkToSend
	}
	return blkMsg, nil
}
