package peerv2

import (
	"context"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
)

func NewBlockProvider(p *p2pgrpc.GRPCProtocol, ns NetSync) *BlockProvider {
	bp := &BlockProvider{NetSync: ns}
	proto.RegisterHighwayServiceServer(p.GetGRPCServer(), bp)
	go p.Serve() // NOTE: must serve after registering all services
	return bp
}

func (bp *BlockProvider) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	Logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardByHeight(
	ctx context.Context,
	req *proto.GetBlockShardByHeightRequest,
) (
	*proto.GetBlockShardByHeightResponse,
	error,
) {
	Logger.Infof("[blkbyheight] Receive GetBlockShardByHeight request (by Specific %v): from %v, to %v, height: %v ", req.Specific, req.FromHeight, req.ToHeight, req.GetHeights())
	reqHeights := []uint64{req.FromHeight, req.ToHeight}
	if req.Specific {
		reqHeights = req.Heights
	}
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		blockShard,
		req.Specific,
		byte(req.Shard),
		reqHeights,
		0,
	)
	Logger.Info("[blkbyheight] Blockshard received from netsync:", len(blkMsgs))
	resp := &proto.GetBlockShardByHeightResponse{}
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

func (bp *BlockProvider) GetBlockShardByHash(ctx context.Context, req *proto.GetBlockShardByHashRequest) (*proto.GetBlockShardByHashResponse, error) {
	hashes := []common.Hash{}
	for _, blkHashBytes := range req.Hashes {
		blkHash := common.Hash{}
		err := blkHash.SetBytes(blkHashBytes)
		if err != nil {
			continue
		}
		hashes = append(hashes, blkHash)
	}
	Logger.Infof("[blkbyhash] Receive GetBlockShardByHash shard %v request hash %v", req.Shard, hashes)
	blkMsgs := bp.NetSync.GetBlockShardByHash(
		hashes,
	)
	Logger.Info("[blkbyhash] Blockshard received from netsync:", len(blkMsgs))
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

func (bp *BlockProvider) GetBlockBeaconByHeight(
	ctx context.Context,
	req *proto.GetBlockBeaconByHeightRequest,
) (
	*proto.GetBlockBeaconByHeightResponse,
	error,
) {
	Logger.Infof("[blkbyheight] Receive GetBlockBeaconByHeight request (by Specific %v): from %v, to %v, height: %v ", req.Specific, req.FromHeight, req.ToHeight, req.GetHeights())
	reqHeights := []uint64{req.FromHeight, req.ToHeight}
	if req.Specific {
		reqHeights = req.Heights
	}
	blkMsgs := bp.NetSync.GetBlockBeaconByHeight(
		req.FromPool,
		req.Specific,
		reqHeights,
	)
	resp := &proto.GetBlockBeaconByHeightResponse{}
	for _, msg := range blkMsgs {
		encoded, err := encodeMessage(msg)
		if err != nil {
			Logger.Warnf("ERROR Failed encoding message %v", msg.MessageType())
			continue
		}
		resp.Data = append(resp.Data, []byte(encoded))
	}
	Logger.Info("[blkbyheight] Blockbeacon received from netsync: ", len(blkMsgs))
	return resp, nil
}

func (bp *BlockProvider) GetBlockBeaconByHash(ctx context.Context, req *proto.GetBlockBeaconByHashRequest) (*proto.GetBlockBeaconByHashResponse, error) {
	Logger.Info("Receive GetBlockBeaconByHash request")
	hashes := []common.Hash{}
	for _, blkHashBytes := range req.Hashes {
		blkHash := common.Hash{}
		err := blkHash.SetBytes(blkHashBytes)
		if err != nil {
			continue
		}
		hashes = append(hashes, blkHash)
	}
	Logger.Infof("[blkbyhash] Receive GetBlockBeaconByHash request hash %v", hashes)
	blkMsgs := bp.NetSync.GetBlockBeaconByHash(
		hashes,
	)
	Logger.Info("[blkbyhash] Block beacon received from netsync:", len(blkMsgs))
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

func (bp *BlockProvider) GetBlockCrossShardByHeight(ctx context.Context, req *proto.GetBlockCrossShardByHeightRequest) (*proto.GetBlockCrossShardByHeightResponse, error) {
	Logger.Info("[blkbyheight] Receive GetBlockCrossShardByHeight request:", req.Heights)
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		crossShard,
		req.Specific,
		byte(req.FromShard),
		req.Heights,
		byte(req.ToShard),
	)
	Logger.Info("[blkbyheight] BlockCS received from netsync:", len(blkMsgs))
	resp := &proto.GetBlockCrossShardByHeightResponse{}
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

func (bp *BlockProvider) GetBlockShardToBeaconByHeight(ctx context.Context, req *proto.GetBlockShardToBeaconByHeightRequest) (*proto.GetBlockShardToBeaconByHeightResponse, error) {
	Logger.Info("[blkbyheight] Receive GetBlockShardToBeaconByHeight request:", req.FromHeight, req.ToHeight)
	reqHeights := []uint64{req.FromHeight, req.ToHeight}
	if req.Specific {
		reqHeights = req.Heights
	}
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		shardToBeacon,
		req.Specific,
		byte(req.FromShard),
		reqHeights,
		0,
	)
	Logger.Info("[blkbyheight] BlockS2B received from netsync:", len(blkMsgs))
	resp := &proto.GetBlockShardToBeaconByHeightResponse{}
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

type BlockProvider struct {
	NetSync NetSync
}

type NetSync interface {
	//GetBlockShardByHeight fromPool bool, blkType byte, specificHeight bool, shardID byte, blkHeights []uint64, crossShardID byte
	GetBlockShardByHeight(bool, byte, bool, byte, []uint64, byte) []wire.Message
	GetBlockShardByHash(blkHashes []common.Hash) []wire.Message
	//GetBlockBeaconByHeight fromPool bool, specificHeight bool, blkHeights []uint64
	GetBlockBeaconByHeight(bool, bool, []uint64) []wire.Message
	GetBlockBeaconByHash(blkHashes []common.Hash) []wire.Message
}
