package netsync

import (
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

func (netSync *NetSync) GetBlockByHeight(
	blkType proto.BlkType,
	height uint64,
	fromcID byte,
	tocID byte,
) (
	interface{},
	error,
) {
	bc := netSync.config.BlockChain
	switch blkType {
	case proto.BlkType_BlkBc:
		return bc.GetBeaconBlockByHeight(height)
	case proto.BlkType_BlkShard:
		return bc.GetShardBlockByHeight(height, fromcID)
	case proto.BlkType_BlkXShard:
		blk, err := bc.GetShardBlockByHeight(height, fromcID)
		if err != nil {
			return nil, err
		}
		return blk.CreateCrossShardBlock(tocID)
	case proto.BlkType_BlkS2B:
		blk, err := bc.GetShardBlockByHeight(height, fromcID)
		if err != nil {
			return nil, err
		}
		return blk.CreateShardToBeaconBlock(bc), nil
	default:
		return nil, errors.Errorf("Invalid block type")
	}
}

func (netSync *NetSync) GetBlockByHash(
	blkType byte,
	hash common.Hash,
) (
	interface{},
	error,
) {
	return nil, nil
}

func (netSync *NetSync) GetBlockShardByHash(blkHashes []common.Hash) []wire.Message {
	blkMsgs := []wire.Message{}
	for _, blkHash := range blkHashes {
		blk, _, err := netSync.config.BlockChain.GetShardBlockByHash(blkHash)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg.(*wire.MessageBlockShard).Block = blk
		blkMsgs = append(blkMsgs, newMsg)
	}
	return blkMsgs
}

func (netSync *NetSync) getBlockShardByHashAndSend(peerID libp2p.ID, blkType byte, blkHashes []common.Hash, crossShardID byte) {
	for _, blkHash := range blkHashes {
		blk, _, err := netSync.config.BlockChain.GetShardBlockByHash(blkHash)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		blkMsg, err := netSync.createBlockShardMsgByType(blk, blkType, crossShardID)
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

func (netSync *NetSync) GetBlockBeaconByHash(
	blkHashes []common.Hash,
) []wire.Message {
	blkMsgs := []wire.Message{}
	for _, blkHash := range blkHashes {
		blk, _, err := netSync.config.BlockChain.GetBeaconBlockByHash(blkHash)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg.(*wire.MessageBlockBeacon).Block = blk
		blkMsgs = append(blkMsgs, newMsg)
	}
	return blkMsgs
}

func (netSync *NetSync) getBlockBeaconByHashAndSend(peerID libp2p.ID, blkHashes []common.Hash) {
	for _, blkHash := range blkHashes {
		blk, _, err := netSync.config.BlockChain.GetBeaconBlockByHash(blkHash)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		newMsg.(*wire.MessageBlockBeacon).Block = blk
		err = netSync.config.Server.PushMessageToPeer(newMsg, peerID)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
	}
}

func (netSync *NetSync) GetBlockShardByHeight(fromPool bool, blkType byte, specificHeight bool, shardID byte, blkHeights []uint64, crossShardID byte) []wire.Message {
	if !specificHeight {
		if len(blkHeights) != 2 || blkHeights[1] < blkHeights[0] {
			return nil
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
	blkMsgs := []wire.Message{}
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
			case crossShard:
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
				blkMsg.(*wire.MessageCrossShard).Block = blkToSend
			case shardToBeacon:
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
				blkMsg.(*wire.MessageShardToBeacon).Block = blkToSend
			}
		} else {
			blk, err := netSync.config.BlockChain.GetShardBlockByHeight(blkHeight, shardID)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			blkMsg, err = netSync.createBlockShardMsgByType(blk, blkType, crossShardID)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
		}
		blkMsgs = append(blkMsgs, blkMsg)
	}

	return blkMsgs
}

func (netSync *NetSync) getBlockShardByHeightAndSend(peerID libp2p.ID, fromPool bool, blkType byte, specificHeight bool, shardID byte, blkHeights []uint64, crossShardID byte) {
	//fmt.Println("GETCROSS: ", fromPool, blkType, specificHeight, shardID, crossShardID, blkHeights)

	blkMsgs := netSync.GetBlockShardByHeight(fromPool, blkType, specificHeight, shardID, blkHeights, crossShardID)

	for _, blkMsg := range blkMsgs {
		err := netSync.config.Server.PushMessageToPeer(blkMsg, peerID)
		// fmt.Println("CROSS:", blkHeights, err)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

func (netSync *NetSync) GetBlockBeaconByHeight(fromPool bool, specificHeight bool, blkHeights []uint64) []wire.Message {
	if !specificHeight {
		if len(blkHeights) != 2 || blkHeights[1] < blkHeights[0] {
			return nil
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
	blkMsgs := []wire.Message{}
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
			continue
		}
		msgBeaconBlk, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		msgBeaconBlk.(*wire.MessageBlockBeacon).Block = blk
		blkMsgs = append(blkMsgs, msgBeaconBlk)
	}
	return blkMsgs
}

func (netSync *NetSync) getBlockBeaconByHeightAndSend(peerID libp2p.ID, fromPool bool, specificHeight bool, blkHeights []uint64) {
	blkMsgs := netSync.GetBlockBeaconByHeight(fromPool, specificHeight, blkHeights)
	for _, blkMsg := range blkMsgs {
		err := netSync.config.Server.PushMessageToPeer(blkMsg, peerID)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

// blkType:
// 0: normal
// 1: crossShard
// 2: shardToBeacon
func (netSync *NetSync) createBlockShardMsgByType(block *blockchain.ShardBlock, blkType byte, crossShardID byte) (wire.Message, error) {
	var (
		blkMsg wire.Message
		err    error
	)
	switch blkType {
	case blockShard:
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageBlockShard).Block = block
	case crossShard:
		blkToSend, err := block.CreateCrossShardBlock(crossShardID)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}

		// fmt.Println("CROSS: ", block.Header.Height, blkToSend, crossShardID)
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdCrossShard)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageCrossShard).Block = blkToSend
	case shardToBeacon:
		blkToSend := block.CreateShardToBeaconBlock(netSync.config.BlockChain)
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageShardToBeacon).Block = blkToSend
	}
	return blkMsg, nil
}

func (netSync *NetSync) StreamBlockBeaconByHeight(fromPool bool, specificHeight bool, blkHeights []uint64) chan interface{} {
	Logger.log.Infof("[stream] Netsync received request get block %v %v [%v...%v] len %v", fromPool, specificHeight, blkHeights[0], blkHeights[len(blkHeights)-1], len(blkHeights))
	blkCh := make(chan interface{})
	if !specificHeight {
		if len(blkHeights) != 2 || blkHeights[1] < blkHeights[0] {
			return nil
		}
	}
	sort.Slice(blkHeights, func(i, j int) bool { return blkHeights[i] < blkHeights[j] })

	go func(
		height []uint64,
		specific bool,
		blkCh chan interface{},
		BlkByHeightGetter func(uint64) (*blockchain.BeaconBlock, error),
	) {
		blkHeight := blkHeights[0] - 1
		idx := 0
		for blkHeight < blkHeights[len(blkHeights)-1] {
			if specific {
				blkHeight = blkHeights[idx]
				idx++
			} else {
				blkHeight++
			}
			if blkHeight <= 1 {
				continue
			}
			blk, err := BlkByHeightGetter(blkHeight)
			if err != nil {
				Logger.log.Infof("[stream] Netsync cannot get block, return error %v", err)
				break
			}
			blkCh <- blk
			Logger.log.Infof("[stream] Netsync push block to channel")
		}
		close(blkCh)
		return
	}(blkHeights, specificHeight, blkCh, netSync.config.BlockChain.GetBeaconBlockByHeight)
	return blkCh
}

// func (netSync *NetSync) StreamBlockByHeight(
// 	fromPool bool,
// 	blktype byte,
// 	specificHeight bool,
// 	blkHeights []uint64,
// 	fromcID byte,
// 	tocID byte,
// ) chan interface{} {
// 	Logger.log.Infof("[stream] Netsync received request get block %v %v [%v...%v] len %v", fromPool, specificHeight, blkHeights[0], blkHeights[len(blkHeights)-1], len(blkHeights))
// 	blkCh := make(chan interface{})
// 	if !specificHeight {
// 		if len(blkHeights) != 2 || blkHeights[1] < blkHeights[0] {
// 			return nil
// 		}
// 	}
// 	sort.Slice(blkHeights, func(i, j int) bool { return blkHeights[i] < blkHeights[j] })
// 	go netSync.streamBlkByHeight(blktype, specificHeight, blkHeights, blkCh, fromcID, tocID)
// 	return blkCh
// }

func (netSync *NetSync) StreamBlockByHeight(
	fromPool bool,
	req *proto.BlockByHeightRequest,
) chan interface{} {
	// Logger.log.Infof("[stream] Netsync received request get block %v %v [%v...%v] len %v", fromPool, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights))
	Logger.log.Infof("[stream] Netsync received request stream block type %v, spec %v, height [%v..%v] len %v, from %v to %v", req.Type, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights), req.From, req.To)
	blkCh := make(chan interface{})
	if !req.Specific {
		if len(req.Heights) != 2 || req.Heights[1] < req.Heights[0] {
			return nil
		}
	}
	sort.Slice(req.Heights, func(i, j int) bool { return req.Heights[i] < req.Heights[j] })
	go netSync.streamBlkByHeight(req, blkCh)
	return blkCh
}

func (netSync *NetSync) streamBlkByHeight(
	req *proto.BlockByHeightRequest,
	blkCh chan interface{},
) {

	blkHeight := req.Heights[0] - 1
	idx := 0
	for blkHeight < req.Heights[len(req.Heights)-1] {
		if req.Specific {
			blkHeight = req.Heights[idx]
			idx++
		} else {
			blkHeight++
		}
		if blkHeight <= 1 {
			continue
		}
		blk, err := netSync.GetBlockByHeight(req.Type, blkHeight, byte(req.From), byte(req.To))
		if err != nil {
			Logger.log.Infof("[stream] Netsync cannot get block, return error %v", err)
			break
		}
		blkCh <- blk
		Logger.log.Infof("[stream] Netsync push block to channel")
	}
	close(blkCh)
	return
}
