package peerv2

import (
	"context"
	"log"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc"
)

type BlockRequester struct {
	conn *grpc.ClientConn
}

func NewRequester(pr *p2pgrpc.GRPCProtocol, peerID peer.ID) (*BlockRequester, error) {
	conn, err := pr.Dial(
		context.Background(),
		peerID,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &BlockRequester{conn: conn}, nil
}

func (c *BlockRequester) Register(
	ctx context.Context,
	pubkey string,
	messages []string,
	selfID peer.ID,
) ([]*MessageTopicPair, error) {
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.Register(
		ctx,
		&RegisterRequest{
			CommitteePublicKey: pubkey,
			WantedMessages:     messages,
			PeerID:             peer.IDB58Encode(selfID),
		},
	)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return reply.Pair, nil
}

func (c *BlockRequester) GetBlockShardByHeight(
	shardID int32,
	from uint64,
	to uint64,
) ([][]byte, error) {
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
		return nil, err
	}
	return reply.Data, nil
}
