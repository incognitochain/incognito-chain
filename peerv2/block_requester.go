package peerv2

import (
	"context"
	"log"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// TODO(@0xbunyip): cache all requests to prevent querying the same height multiple times

type BlockRequester struct {
	conn       *grpc.ClientConn
	highwayPID peer.ID
	prtc       *p2pgrpc.GRPCProtocol
}

func NewRequester(prtc *p2pgrpc.GRPCProtocol, peerID peer.ID) (*BlockRequester, error) {
	req := &BlockRequester{
		prtc:       prtc,
		conn:       nil,
		highwayPID: peerID,
	}
	go req.keepConnection()
	return req, nil
}

// keepConnection dials highway to establish gRPC connection if it isn't available
func (c *BlockRequester) keepConnection() {
	for ; true; <-time.Tick(10 * time.Second) {
		if c.Ready() {
			continue
		}

		log.Println("BlockRequester is not ready, dialing")
		if conn, err := c.prtc.Dial(
			context.Background(),
			c.highwayPID,
			grpc.WithInsecure(),
		); err != nil {
			log.Println("Could not dial to highway grpc server:", err, c.highwayPID)
		} else {
			c.conn = conn
		}
	}
}

func (c *BlockRequester) Ready() bool {
	return c.conn != nil && c.conn.GetState() == connectivity.Ready
}

func (c *BlockRequester) Register(
	ctx context.Context,
	pubkey string,
	messages []string,
	selfID peer.ID,
) ([]*MessageTopicPair, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}

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
		log.Println(err)
		return nil, err
	}
	return reply.Pair, nil
}

func (c *BlockRequester) GetBlockShardByHeight(
	shardID int32,
	from uint64,
	to uint64,
	dstCandidatePublicKey string,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}

	log.Printf("Requesting shard block by height: shard = %v from = %v to = %v", shardID, from, to)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&GetBlockShardByHeightRequest{
			Shard:         shardID,
			Specific:      false,
			FromHeight:    from,
			ToHeight:      to,
			Heights:       nil,
			FromPool:      false,
			FromCommittee: dstCandidatePublicKey,
		},
	)
	log.Printf("Received block shard data %v", reply)
	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (c *BlockRequester) GetBlockBeaconByHeight(
	from uint64,
	to uint64,
	dstCandidatePublicKey string,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}

	log.Printf("Requesting beaconblock by height: from = %v to = %v", from, to)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&GetBlockBeaconByHeightRequest{
			Specific:      false,
			FromHeight:    from,
			ToHeight:      to,
			Heights:       nil,
			FromPool:      false,
			FromCommittee: dstCandidatePublicKey,
		},
	)
	if err != nil {
		return nil, err
	} else if reply != nil {
		log.Printf("Received block beacon data len: %v", len(reply.Data))
	}
	return reply.Data, nil
}
