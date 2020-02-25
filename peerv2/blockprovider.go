package peerv2

import (
	"context"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/peerv2/wrapper"
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
	uuid := req.GetUUID()
	Logger.Infof("[blkbyheight] Receive GetBlockShardByHeight request (by Specific %v): from %v, to %v, height: %v, uuid = %s", req.Specific, req.FromHeight, req.ToHeight, req.GetHeights(), uuid)
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
	Logger.Infof("[blkbyheight] Blockshard received from netsync: %d, uuid = %s", len(blkMsgs), uuid)
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
	blkMsgs := bp.NetSync.GetBlockShardByHash(
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

func (bp *BlockProvider) GetBlockBeaconByHeight(
	ctx context.Context,
	req *proto.GetBlockBeaconByHeightRequest,
) (
	*proto.GetBlockBeaconByHeightResponse,
	error,
) {
	uuid := req.GetUUID()
	Logger.Infof("[blkbyheight] Receive GetBlockBeaconByHeight request (by Specific %v): from %v, to %v, height: %v, uuid = %s", req.Specific, req.FromHeight, req.ToHeight, req.GetHeights(), uuid)
	reqHeights := []uint64{req.FromHeight, req.ToHeight}
	if req.Specific {
		reqHeights = req.Heights
	}
	blkMsgs := bp.NetSync.GetBlockBeaconByHeight(
		req.FromPool,
		req.Specific,
		reqHeights,
	)
	Logger.Infof("[blkbyheight] Blockbeacon received from netsync: %d, uuid = %s", len(blkMsgs), uuid)
	resp := &proto.GetBlockBeaconByHeightResponse{}
	for _, msg := range blkMsgs {
		encoded, err := encodeMessage(msg)
		if err != nil {
			Logger.Warnf("ERROR Failed encoding message %v, uuid = %s", msg.MessageType(), uuid)
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
	blkMsgs := bp.NetSync.GetBlockBeaconByHash(
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

func (bp *BlockProvider) GetBlockCrossShardByHeight(ctx context.Context, req *proto.GetBlockCrossShardByHeightRequest) (*proto.GetBlockCrossShardByHeightResponse, error) {
	uuid := req.GetUUID()
	Logger.Infof("[blkbyheight] Receive GetBlockCrossShardByHeight request: %v, uuid = %s", req.Heights, uuid)
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		crossShard,
		req.Specific,
		byte(req.FromShard),
		req.Heights,
		byte(req.ToShard),
	)
	Logger.Infof("[blkbyheight] BlockCS received from netsync: %d, uuid = %s", len(blkMsgs), uuid)
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
	uuid := req.GetUUID()
	Logger.Infof("[blkbyheight] Receive GetBlockShardToBeaconByHeight request: %d -> %d, uuid = %s", req.FromHeight, req.ToHeight, uuid)
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
	Logger.Infof("[blkbyheight] BlockS2B received from netsync: %d, uuid = %s", len(blkMsgs), uuid)
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

func (bp *BlockProvider) StreamBlockBeaconByHeight(req *proto.GetBlockBeaconByHeightRequest, stream proto.HighwayService_StreamBlockBeaconByHeightServer) error {
	var heights []uint64
	if req.Specific {
		Logger.Infof("[stream] Block provider received request [%v..%v], len %v", req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights))
		heights = req.GetHeights()
	} else {
		Logger.Infof("[stream] Block provider received request %v %v", req.GetFromHeight(), req.GetToHeight())
		heights = []uint64{req.GetFromHeight(), req.GetToHeight()}
	}
	blkRecv := bp.NetSync.StreamBlockBeaconByHeight(false, req.GetSpecific(), heights)
	for blk := range blkRecv {
		rdata, err := wrapper.EnCom(blk)
		blkData := append([]byte{blockbeacon}, rdata...)
		if err != nil {
			Logger.Infof("[stream] block channel return error when marshal %v", err)
			return err
		}
		Logger.Infof("[stream] block channel return block ok")
		if err := stream.Send(&proto.BlockData{Data: blkData}); err != nil {
			Logger.Infof("[stream] Server send block to client return err %v", err)
			return err
		}
		Logger.Infof("[stream] Server send block to client ok")
	}
	Logger.Infof("[stream] Provider return StreamBlockBeaconByHeight")
	return nil
}

func (bp *BlockProvider) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	stream proto.HighwayService_StreamBlockByHeightServer,
) error {
	// Logger.Infof("[stream] Block provider received request block type %v, blk heights specific %v [%v..%v], len %v", req.GetType(), req.GetSpecific(), req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights))
	Logger.Infof("[stream] Block provider received request stream block type %v, spec %v, height [%v..%v] len %v, from %v to %v", req.Type, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights), req.From, req.To)
	blkRecv := bp.NetSync.StreamBlockByHeight(false, req)
	for blk := range blkRecv {
		rdata, err := wrapper.EnCom(blk)
		blkData := append([]byte{byte(req.Type)}, rdata...)
		if err != nil {
			Logger.Infof("[stream] block channel return error when marshal %v", err)
			return err
		}
		Logger.Infof("[stream] block channel return block ok")
		if err := stream.Send(&proto.BlockData{Data: blkData}); err != nil {
			Logger.Infof("[stream] Server send block to client return err %v", err)
			return err
		}
		Logger.Infof("[stream] Server send block to client ok")
	}
	Logger.Infof("[stream] Provider return StreamBlockBeaconByHeight")
	return nil
}

type BlockProvider struct {
	proto.UnimplementedHighwayServiceServer
	NetSync NetSync
}

type NetSync interface {
	//GetBlockShardByHeight fromPool bool, blkType byte, specificHeight bool, shardID byte, blkHeights []uint64, crossShardID byte
	GetBlockShardByHeight(bool, byte, bool, byte, []uint64, byte) []wire.Message
	GetBlockShardByHash(blkHashes []common.Hash) []wire.Message
	//GetBlockBeaconByHeight fromPool bool, specificHeight bool, blkHeights []uint64
	GetBlockBeaconByHeight(bool, bool, []uint64) []wire.Message
	GetBlockBeaconByHash(blkHashes []common.Hash) []wire.Message
	StreamBlockBeaconByHeight(fromPool bool, specificHeight bool, blkHeights []uint64) chan interface{}
	StreamBlockByHeight(fromPool bool, req *proto.BlockByHeightRequest) chan interface{}
}
