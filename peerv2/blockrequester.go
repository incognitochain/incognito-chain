package peerv2

import (
	"context"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/incognitochain/incognito-chain/common"
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

		Logger.Warn("BlockRequester is not ready, dialing")
		if conn, err := c.prtc.Dial(
			context.Background(),
			c.highwayPID,
			grpc.WithInsecure(),
		); err != nil {
			Logger.Error("Could not dial to highway grpc server:", err, c.highwayPID)
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
	committeeIDs []byte,
	selfID peer.ID,
	role string,
) ([]*MessageTopicPair, *UserRole, error) {
	if !c.Ready() {
		return nil, nil, errors.New("requester not ready")
	}

	client := NewHighwayServiceClient(c.conn)
	reply, err := client.Register(
		ctx,
		&RegisterRequest{
			CommitteePublicKey: pubkey,
			WantedMessages:     messages,
			CommitteeID:        committeeIDs,
			PeerID:             peer.IDB58Encode(selfID),
			Role:               role,
		},
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return reply.Pair, reply.Role, nil
}

func (c *BlockRequester) GetBlockShardByHeight(
	shardID int32,
	bySpecific bool,
	from uint64,
	heights []uint64,
	to uint64,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}
	Logger.Infof("[blkbyheight] Requesting block shard %v (by specific %v): from = %v to = %v; height: %v", shardID, bySpecific, from, to, heights)

	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   bySpecific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	if err != nil {
		return nil, err
	}
	Logger.Infof("[blkbyheight] Received block shard data %v", len(reply.Data))
	return reply.Data, nil
}

func (c *BlockRequester) GetBlockShardByHash(
	shardID int32,
	hashes []common.Hash,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}
	blkHashBytes := [][]byte{}
	for _, hash := range hashes {
		blkHashBytes = append(blkHashBytes, hash.GetBytes())
	}
	Logger.Infof("[blkbyhash] Requesting shard block by hash: %v", hashes)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardByHash(
		context.Background(),
		&GetBlockShardByHashRequest{
			Shard:  shardID,
			Hashes: blkHashBytes,
		},
	)
	if err != nil {
		return nil, err
	}
	Logger.Infof("[blkbyhash] Received block shard data %v", len(reply.Data))
	return reply.Data, nil
}

func (c *BlockRequester) GetBlockBeaconByHeight(
	bySpecific bool,
	from uint64,
	heights []uint64,
	to uint64,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}
	Logger.Infof("[blkbyheight] Requesting beaconblock (by specific %v): from = %v to = %v; height: %v", bySpecific, from, to, heights)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&GetBlockBeaconByHeightRequest{
			Specific:   bySpecific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	if err != nil {
		return nil, err
	} else if reply != nil {
		Logger.Infof("[blkbyheight] Received block beacon data len: %v", len(reply.Data))
	}
	return reply.Data, nil
}

func (c *BlockRequester) GetBlockBeaconByHash(
	hashes []common.Hash,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}
	blkHashBytes := [][]byte{}
	for _, hash := range hashes {
		blkHashBytes = append(blkHashBytes, hash.GetBytes())
	}
	Logger.Infof("Requesting beacon block by hash: %v", hashes)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockBeaconByHash(
		context.Background(),
		&GetBlockBeaconByHashRequest{
			Hashes: blkHashBytes,
		},
	)
	if err != nil {
		return nil, err
	}
	Logger.Infof("Received block beacon data %v", len(reply.Data))
	return reply.Data, nil
}

func (c *BlockRequester) GetBlockShardToBeaconByHeight(
	shardID int32,
	bySpecific bool,
	from uint64,
	heights []uint64,
	to uint64,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}

	Logger.Infof("[sync] Requesting blkshdtobcn by specific height %v: from = %v to = %v; Heights: %v", bySpecific, from, to, heights)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardToBeaconByHeight(
		context.Background(),
		&GetBlockShardToBeaconByHeightRequest{
			FromShard:  shardID,
			Specific:   bySpecific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	if err != nil {
		Logger.Infof("[sync] Received err: %v from = %v to = %v; Heights: %v", err, from, to, heights)
		return nil, err
	} else if reply != nil {
		Logger.Infof("[sync] Received block s2b data len: %v from = %v to = %v; Heights: %v ", len(reply.Data), from, to, heights)
	}
	return reply.Data, nil
}

func (c *BlockRequester) GetBlockCrossShardByHeight(
	fromShard int32,
	toShard int32,
	heights []uint64,
	getFromPool bool,
) ([][]byte, error) {
	if !c.Ready() {
		return nil, errors.New("requester not ready")
	}

	Logger.Infof("Requesting block crossshard by height: shard %v to %v, height %v", fromShard, toShard, heights)
	client := NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockCrossShardByHeight(
		context.Background(),
		&GetBlockCrossShardByHeightRequest{
			FromShard:  fromShard,
			ToShard:    toShard,
			Specific:   true,
			FromHeight: 0,
			ToHeight:   0,
			Heights:    heights,
			FromPool:   getFromPool,
		},
	)
	if err != nil {
		return nil, err
	} else if reply != nil {
		Logger.Infof("Received block s2b data len: %v", len(reply.Data))
	}
	return reply.Data, nil
}
