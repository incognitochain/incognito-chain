package peerv2

import (
	"context"
	"log"

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
	log.Printf("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	return nil, nil
}

func (bp *BlockProvider) GetBlockShardByHeight(ctx context.Context, req *GetBlockShardByHeightRequest) (*GetBlockShardByHeightResponse, error) {
	log.Println("[db] Receive GetBlockShardByHeight request")
	blkType := byte(0) // TODO(@0xbunyip): define in common file
	blkMsgs := bp.NetSync.GetBlockShardByHeight(
		req.FromPool,
		blkType,
		false,
		byte(req.Shard),
		[]uint64{req.FromHeight, req.ToHeight},
		0,
	)
	log.Println("Blockshard received from netsync:", blkMsgs)
	resp := &GetBlockShardByHeightResponse{}
	for _, msg := range blkMsgs {
		encoded, err := encodeMessage(msg)
		if err != nil {
			log.Printf("ERROR Failed encoding message %v", msg.MessageType())
			continue
		}
		resp.Data = append(resp.Data, []byte(encoded))
	}
	return resp, nil
}

func (bp *BlockProvider) GetBlockShardByHash(ctx context.Context, req *GetBlockShardByHashRequest) (*GetBlockShardByHashResponse, error) {
	log.Println("Receive GetBlockShardByHash request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockBeaconByHeight(ctx context.Context, req *GetBlockBeaconByHeightRequest) (*GetBlockBeaconByHeightResponse, error) {
	log.Println("[db] Receive GetBlockBeaconByHeight request")
	blkMsgs := bp.NetSync.GetBlockBeaconByHeight(
		req.FromPool,
		false,
		[]uint64{req.FromHeight, req.ToHeight},
	)
	resp := &GetBlockBeaconByHeightResponse{}
	for _, msg := range blkMsgs {
		encoded, err := encodeMessage(msg)
		if err != nil {
			log.Printf("ERROR Failed encoding message %v", msg.MessageType())
			continue
		}
		resp.Data = append(resp.Data, []byte(encoded))
	}
	return resp, nil
}

func (bp *BlockProvider) GetBlockBeaconByHash(ctx context.Context, req *GetBlockBeaconByHashRequest) (*GetBlockBeaconByHashResponse, error) {
	log.Println("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockCrossShardByHeight(ctx context.Context, req *GetBlockCrossShardByHeightRequest) (*GetBlockCrossShardByHeightResponse, error) {
	log.Println("Receive GetBlockCrossShardByHeight request")
	return nil, nil
}

func (bp *BlockProvider) GetBlockCrossShardByHash(ctx context.Context, req *GetBlockCrossShardByHashRequest) (*GetBlockCrossShardByHashResponse, error) {
	log.Println("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type BlockProvider struct {
	NetSync NetSync
}

type NetSync interface {
	GetBlockShardByHeight(bool, byte, bool, byte, []uint64, byte) []wire.Message
	GetBlockBeaconByHeight(bool, bool, []uint64) []wire.Message
}
