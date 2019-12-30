package peerv2

import (
	"context"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

type BlockRequester struct {
	conn *grpc.ClientConn

	peerIDs chan peer.ID
	prtc    *p2pgrpc.GRPCProtocol
}

func NewRequester(prtc *p2pgrpc.GRPCProtocol) *BlockRequester {
	req := &BlockRequester{
		prtc:    prtc,
		peerIDs: make(chan peer.ID, 100),
		conn:    nil,
	}
	go req.keepConnection()
	return req
}

// keepConnection dials highway to establish gRPC connection if it isn't available
func (c *BlockRequester) keepConnection() {
	currentHWID := peer.ID("")
	watchTimestep := time.Tick(RequesterDialTimestep)
	for {
		select {
		case <-watchTimestep:
			if c.Ready() {
				continue
			}

			Logger.Warn("BlockRequester is not ready, dialing")
			if c.conn != nil {
				Logger.Info("Closing old requester connection")
				err := c.conn.Close()
				if err != nil {
					Logger.Errorf("Failed closing old requester connection: %+v", err)
				}
				c.conn = nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if conn, err := c.prtc.Dial(
				ctx,
				currentHWID,
				grpc.WithInsecure(),
				grpc.WithBlock(),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:    RequesterKeepaliveTime,
					Timeout: RequesterKeepaliveTimeout,
				}),
			); err != nil {
				Logger.Error("Could not dial to highway grpc server:", err, currentHWID)
			} else {
				c.conn = conn
			}
			cancel()

		case hwID := <-c.peerIDs:
			Logger.Infof("Received new highway peerID, old = %s, new = %s", currentHWID.String(), hwID.String())
			if hwID != currentHWID && c.conn != nil {
				if err := c.conn.Close(); err != nil { // Close gRPC connection
					Logger.Errorf("Failed closing connection to highway: %v %v %+v", hwID, currentHWID, err)
				}
			}
			currentHWID = hwID
		}
	}
}

func (c *BlockRequester) Ready() bool {
	return c.conn != nil && c.conn.GetState() == connectivity.Ready
}

func (c *BlockRequester) UpdateTarget(p peer.ID) {
	c.peerIDs <- p
}

func (c *BlockRequester) Target() string {
	if c.conn == nil {
		return ""
	}
	return c.conn.Target()
}

func (c *BlockRequester) Register(
	ctx context.Context,
	pubkey string,
	messages []string,
	committeeIDs []byte,
	selfID peer.ID,
	role string,
) ([]*proto.MessageTopicPair, *proto.UserRole, error) {
	if !c.Ready() {
		return nil, nil, errors.New("requester not ready")
	}

	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.Register(
		ctx,
		&proto.RegisterRequest{
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

	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&proto.GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   bySpecific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
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
	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardByHash(
		context.Background(),
		&proto.GetBlockShardByHashRequest{
			Shard:  shardID,
			Hashes: blkHashBytes,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
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
	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&proto.GetBlockBeaconByHeightRequest{
			Specific:   bySpecific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
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
	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockBeaconByHash(
		context.Background(),
		&proto.GetBlockBeaconByHashRequest{
			Hashes: blkHashBytes,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
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
	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockShardToBeaconByHeight(
		context.Background(),
		&proto.GetBlockShardToBeaconByHeightRequest{
			FromShard:  shardID,
			Specific:   bySpecific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
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
	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.GetBlockCrossShardByHeight(
		context.Background(),
		&proto.GetBlockCrossShardByHeightRequest{
			FromShard:  fromShard,
			ToShard:    toShard,
			Specific:   true,
			FromHeight: 0,
			ToHeight:   0,
			Heights:    heights,
			FromPool:   getFromPool,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, err
	} else if reply != nil {
		Logger.Infof("Received block s2b data len: %v", len(reply.Data))
	}
	return reply.Data, nil
}
