package peerv2

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p-core/peer"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
	"google.golang.org/grpc"
)

type GRPCService_Client struct {
	conn *grpc.ClientConn
}

func NewClient(pr *p2pgrpc.GRPCProtocol, peerID peer.ID) (*GRPCService_Client, error) {
	conn, err := pr.Dial(
		context.Background(),
		peerID,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &GRPCService_Client{conn: conn}, nil
}

func (c *GRPCService_Client) Register(
	ctx context.Context,
	pubkey string,
	messages []string,
) ([]*MessageTopicPair, error) {
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.Register(
		ctx,
		&RegisterRequest{
			CommitteePublicKey: pubkey,
			WantedMessages:     messages,
		},
	)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return reply.Pair, nil
}

func (c *GRPCService_Client) GetBlockShardByHeight(
	shardID int32,
	from uint64,
	to uint64,
) ([]byte, error) {
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   false,
			FromHeight: from,
			ToHeight:   to,
			Heights:    nil,
			FromPool:   false,
		},
	)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return reply.Data, nil
}
