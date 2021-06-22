package peerv2

import (
	"context"
	"sort"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/peerv2/wrapper"
	"github.com/incognitochain/incognito-chain/wire"
)

func NewBlockProvider(p *p2pgrpc.GRPCProtocol, bg BlockGetter) *BlockProvider {
	bp := &BlockProvider{BlockGetter: bg}
	proto.RegisterHighwayServiceServer(p.GetGRPCServer(), bp)
	go p.Serve() // NOTE: must serve after registering all services
	return bp
}

func (bp *BlockProvider) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	Logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardByHash(ctx context.Context, req *proto.GetBlockShardByHashRequest) (*proto.GetBlockShardByHashResponse, error) {
	uuid := req.GetUUID()
	hashes := []common.Hash{}
	for _, blkHashBytes := range req.Hashes {
		blkHash := common.Hash{}
		err := blkHash.SetBytes(blkHashBytes)
		if err != nil {
			continue
		}
		hashes = append(hashes, blkHash)
	}
	Logger.Infof("[blkbyhash] Receive GetBlockShardByHash shard %v request hash %v, uuid = %s", req.Shard, hashes, uuid)
	blkMsgs := bp.getBlockShardByHash(
		hashes,
	)
	Logger.Infof("[blkbyhash] Blockshard received from netsync: %d, uuid = %s", len(blkMsgs), uuid)
	resp := &proto.GetBlockShardByHashResponse{}
	for _, msg := range blkMsgs {
		encoded, err := encodeMessage(msg)
		if err != nil {
			Logger.Warnf("ERROR Failed encoding message %v", msg.MessageType())
			continue
		}
		resp.Data = append(resp.Data, []byte(encoded))
	}
	return resp, nil
}

func (bp *BlockProvider) GetBlockBeaconByHash(ctx context.Context, req *proto.GetBlockBeaconByHashRequest) (*proto.GetBlockBeaconByHashResponse, error) {
	uuid := req.GetUUID()
	hashes := []common.Hash{}
	for _, blkHashBytes := range req.Hashes {
		blkHash := common.Hash{}
		err := blkHash.SetBytes(blkHashBytes)
		if err != nil {
			continue
		}
		hashes = append(hashes, blkHash)
	}
	Logger.Infof("[blkbyhash] Receive GetBlockBeaconByHash request hash %v, uuid = %s", hashes, uuid)
	blkMsgs := bp.getBlockBeaconByHash(
		hashes,
	)
	Logger.Infof("[blkbyhash] Block beacon received from netsync: %d, uuid = %s", len(blkMsgs), uuid)
	resp := &proto.GetBlockBeaconByHashResponse{}
	for _, msg := range blkMsgs {
		encoded, err := encodeMessage(msg)
		if err != nil {
			Logger.Warnf("ERROR Failed encoding message %v", msg.MessageType())
			continue
		}
		resp.Data = append(resp.Data, []byte(encoded))
	}
	return resp, nil
}

func (bp *BlockProvider) GetBlockCrossShardByHash(ctx context.Context, req *proto.GetBlockCrossShardByHashRequest) (*proto.GetBlockCrossShardByHashResponse, error) {
	Logger.Info("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

func (bp *BlockProvider) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	stream proto.HighwayService_StreamBlockByHeightServer,
) error {
	uuid := req.GetUUID()
	cnt := 0
	Logger.Infof("[stream] Block provider received request stream block type %v, spec %v, height [%v..%v] len %v, from %v to %v, uuid = %s ", req.Type, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights), req.From, req.To, uuid)
	blkRecv := bp.BlockGetter.StreamBlockByHeight(false, req)
	for blk := range blkRecv {
		cnt++
		rdata, err := wrapper.EnCom(blk)
		blkData := append([]byte{byte(req.Type)}, rdata...)
		if err != nil {
			Logger.Infof("[stream] block channel return error when marshal %v, uuid = %s", err, uuid)
			Logger.Infof("[stream] Successfully sent %v blocks to client, uuid %v", cnt, uuid)
			return err
		}
		if err := stream.Send(&proto.BlockData{Data: blkData}); err != nil {
			Logger.Infof("[stream] Server send block to client return err %v, uuid = %s", err, uuid)
			Logger.Infof("[stream] Successfully sent %v blocks to client, uuid %v", cnt, uuid)
			return err
		}
	}
	Logger.Infof("[stream] Successfully sent %v blocks to client, uuid %v", cnt, uuid)
	return nil
}

func (bp *BlockProvider) StreamBlockByHash(
	req *proto.BlockByHashRequest,
	stream proto.HighwayService_StreamBlockByHashServer,
) error {
	uuid := req.GetUUID()
	Logger.Infof("[stream] Block provider received request stream block type %v, hashes [%v..%v] len %v, from %v to %v, uuid = %s ", req.Type, req.Hashes[0], req.Hashes[len(req.Hashes)-1], len(req.Hashes), req.From, req.To, uuid)
	cnt := 0
	blkRecv := bp.BlockGetter.StreamBlockByHash(false, req)
	for blk := range blkRecv {
		cnt++
		rdata, err := wrapper.EnCom(blk)
		blkData := append([]byte{byte(req.Type)}, rdata...)
		if err != nil {
			Logger.Infof("[stream] blkbyhash block channel return error when marshal %v, uuid = %s", err, uuid)
			Logger.Infof("[stream] Successfully sent %v blocks to client, uuid %v", cnt, uuid)
			return err
		}
		if err := stream.Send(&proto.BlockData{Data: blkData}); err != nil {
			Logger.Infof("[stream] blkbyhash Server send block to client return err %v, uuid = %s", err, uuid)
			Logger.Infof("[stream] Successfully sent %v blocks to client, uuid %v", cnt, uuid)
			return err
		}
	}
	Logger.Infof("[stream] Successfully sent %v blocks to client, uuid %v", cnt, uuid)
	return nil
}

type BlockProvider struct {
	proto.UnimplementedHighwayServiceServer
	BlockGetter BlockGetter
}

type BlockGetter interface {
	StreamBlockByHeight(fromPool bool, req *proto.BlockByHeightRequest) chan interface{}
	StreamBlockByHash(fromPool bool, req *proto.BlockByHashRequest) chan interface{}
	GetShardBlockByHeight(height uint64, shardID byte) (map[common.Hash]*types.ShardBlock, error)
	GetShardBlockByHash(hash common.Hash) (*types.ShardBlock, uint64, error)
	GetBeaconBlockByHeight(height uint64) ([]*types.BeaconBlock, error)
	GetBeaconBlockByHash(beaconBlockHash common.Hash) (*types.BeaconBlock, uint64, error)
}

func (bp *BlockProvider) getBlockShardByHash(blkHashes []common.Hash) []wire.Message {
	blkMsgs := []wire.Message{}
	for _, blkHash := range blkHashes {
		blk, _, err := bp.BlockGetter.GetShardBlockByHash(blkHash)
		if err != nil {
			Logger.Error(err)
			continue
		}
		newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			Logger.Error(err)
			continue
		}
		newMsg.(*wire.MessageBlockShard).Block = blk
		blkMsgs = append(blkMsgs, newMsg)
	}
	return blkMsgs
}

func (bp *BlockProvider) getBlockBeaconByHash(
	blkHashes []common.Hash,
) []wire.Message {
	blkMsgs := []wire.Message{}
	for _, blkHash := range blkHashes {
		blk, _, err := bp.BlockGetter.GetBeaconBlockByHash(blkHash)
		if err != nil {
			Logger.Error(err)
			continue
		}
		newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.Error(err)
			continue
		}
		newMsg.(*wire.MessageBlockBeacon).Block = blk
		blkMsgs = append(blkMsgs, newMsg)
	}
	return blkMsgs
}

// GetBlockShardByHeight return list message contains all of blocks of given heights
func (bp *BlockProvider) GetBlockShardByHeight(
	fromPool bool,
	blkType byte,
	specificHeight bool,
	shardID byte,
	blkHeights []uint64,
	crossShardID byte,
) []wire.Message {
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
				blkMsg, err = wire.MakeEmptyMessage(wire.CmdCrossShard)
				if err != nil {
					Logger.Error(err)
					continue
				}
			}
			blkMsgs = append(blkMsgs, blkMsg)
		} else {
			blks, err := bp.BlockGetter.GetShardBlockByHeight(blkHeight, shardID)
			if err != nil {
				Logger.Error(err)
				continue
			}
			for _, blk := range blks {
				blkMsg, err = bp.createBlockShardMsgByType(blk, blkType, crossShardID)
				if err != nil {
					Logger.Error(err)
					continue
				}
				blkMsgs = append(blkMsgs, blkMsg)
			}
		}
	}
	return blkMsgs
}

// GetBlockBeaconByHeight return list message contains all of blocks of given heights
func (bp *BlockProvider) GetBlockBeaconByHeight(
	fromPool bool,
	specificHeight bool,
	blkHeights []uint64,
) []wire.Message {
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
		blks, err := bp.BlockGetter.GetBeaconBlockByHeight(blkHeight)
		if err != nil {
			continue
		}
		for _, blk := range blks {
			msgBeaconBlk, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
			if err != nil {
				Logger.Error(err)
				continue
			}
			msgBeaconBlk.(*wire.MessageBlockBeacon).Block = blk
			blkMsgs = append(blkMsgs, msgBeaconBlk)
		}
	}
	return blkMsgs
}

// blkType:
// 0: normal
// 1: crossShard
// 2: shardToBeacon
func (bp *BlockProvider) createBlockShardMsgByType(
	block *types.ShardBlock,
	blkType byte,
	crossShardID byte,
) (
	wire.Message,
	error,
) {
	var (
		blkMsg wire.Message
		err    error
	)
	switch blkType {
	case blockShard:
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			Logger.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageBlockShard).Block = block
	case crossShard:
		blkToSend, err := types.CreateCrossShardBlock(block, crossShardID)
		if err != nil {
			Logger.Error(err)
			return nil, err
		}

		// fmt.Println("CROSS: ", block.Header.Height, blkToSend, crossShardID)
		blkMsg, err = wire.MakeEmptyMessage(wire.CmdCrossShard)
		if err != nil {
			Logger.Error(err)
			return nil, err
		}
		blkMsg.(*wire.MessageCrossShard).Block = blkToSend
	}
	return blkMsg, nil
}
