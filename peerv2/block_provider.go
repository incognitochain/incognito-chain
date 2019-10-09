package peerv2

import (
	"context"

	p2pgrpc "github.com/paralin/go-libp2p-grpc"
)

func NewBlockProvider(p *p2pgrpc.GRPCProtocol) *BlockProvider {
	bp := &BlockProvider{}
	RegisterHighwayServiceServer(p.GetGRPCServer(), bp)
	return bp
}

func (bp *BlockProvider) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardByHeight(context.Context, *GetBlockShardByHeightRequest) (*GetBlockShardByHeightResponse, error) {
	logger.Println("Receive GetBlockShardByHeight request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardByHash(context.Context, *GetBlockShardByHashRequest) (*GetBlockShardByHashResponse, error) {
	logger.Println("Receive GetBlockShardByHash request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockBeaconByHeight(context.Context, *GetBlockBeaconByHeightRequest) (*GetBlockBeaconByHeightResponse, error) {
	logger.Println("Receive GetBlockBeaconByHeight request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockBeaconByHash(context.Context, *GetBlockBeaconByHashRequest) (*GetBlockBeaconByHashResponse, error) {
	logger.Println("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockCrossShardByHeight(context.Context, *GetBlockCrossShardByHeightRequest) (*GetBlockCrossShardByHeightResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHeight request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockCrossShardByHash(context.Context, *GetBlockCrossShardByHashRequest) (*GetBlockCrossShardByHashResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type BlockProvider struct{}
