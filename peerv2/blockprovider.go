package peerv2

import (
	"context"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
)

func NewBlockProvider(p *p2pgrpc.GRPCProtocol, ns NetSync) *BlockProvider {
	bp := &BlockProvider{NetSync: ns}
	RegisterHighwayServiceServer(p.GetGRPCServer(), bp)
	go p.Serve() // NOTE: must serve after registering all services
	return bp
}

func (bp *BlockProvider) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	Logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardByHeight(ctx context.Context, req *GetBlockShardByHeightRequest) (*GetBlockShardByHeightResponse, error) {
	Logger.Info("Receive GetBlockShardByHeight request")
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		blockShard,
		req.Specific,
		byte(req.Shard),
		[]uint64{req.FromHeight, req.ToHeight},
		0,
	)
	Logger.Info("Blockshard received from netsync:", blkMsgs)
	resp := &GetBlockShardByHeightResponse{}
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

func (bp *BlockProvider) GetBlockShardByHash(ctx context.Context, req *GetBlockShardByHashRequest) (*GetBlockShardByHashResponse, error) {
	Logger.Info("Receive GetBlockShardByHash request")
	hashes := []common.Hash{}
	for _, blkHashBytes := range req.Hashes {
		blkHash := common.Hash{}
		err := blkHash.SetBytes(blkHashBytes)
		if err != nil {
			continue
		}
		hashes = append(hashes, blkHash)
	}
	blkMsgs := bp.NetSync.GetBlockShardByHash(
		hashes,
	)
	Logger.Info("Blockshard received from netsync:", blkMsgs)
	resp := &GetBlockShardByHashResponse{}
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

func (bp *BlockProvider) GetBlockBeaconByHeight(ctx context.Context, req *GetBlockBeaconByHeightRequest) (*GetBlockBeaconByHeightResponse, error) {
	Logger.Info("Receive GetBlockBeaconByHeight request")
	blkMsgs := bp.NetSync.GetBlockBeaconByHeight(
		req.FromPool,
		false,
		[]uint64{req.FromHeight, req.ToHeight},
	)
	resp := &GetBlockBeaconByHeightResponse{}
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

func (bp *BlockProvider) GetBlockBeaconByHash(ctx context.Context, req *GetBlockBeaconByHashRequest) (*GetBlockBeaconByHashResponse, error) {
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
	blkMsgs := bp.NetSync.GetBlockBeaconByHash(
		hashes,
	)
	Logger.Info("Block beacon received from netsync:", blkMsgs)
	resp := &GetBlockBeaconByHashResponse{}
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

func (bp *BlockProvider) GetBlockCrossShardByHeight(ctx context.Context, req *GetBlockCrossShardByHeightRequest) (*GetBlockCrossShardByHeightResponse, error) {
	Logger.Info("Receive GetBlockCrossShardByHeight request:", req.Heights)
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		crossShard,
		req.Specific,
		byte(req.FromShard),
		req.Heights,
		byte(req.ToShard),
	)
	Logger.Info("BlockCS received from netsync:", blkMsgs)
	resp := &GetBlockCrossShardByHeightResponse{}
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

func (bp *BlockProvider) GetBlockCrossShardByHash(ctx context.Context, req *GetBlockCrossShardByHashRequest) (*GetBlockCrossShardByHashResponse, error) {
	Logger.Info("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardToBeaconByHeight(ctx context.Context, req *GetBlockShardToBeaconByHeightRequest) (*GetBlockShardToBeaconByHeightResponse, error) {
	Logger.Info("[sync] Receive GetBlockShardToBeaconByHeight request:", req.FromHeight, req.ToHeight)
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
	Logger.Info("BlockS2B received from netsync:", blkMsgs)
	resp := &GetBlockShardToBeaconByHeightResponse{}
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
	GetBlockBeaconByHeight(bool, bool, []uint64) []wire.Message
	GetBlockBeaconByHash(blkHashes []common.Hash) []wire.Message
}
