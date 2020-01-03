package peerv2

import (
	"context"
	"sync"
	"time"

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
	prtc    GRPCDialer
	stop    chan int
	sync.RWMutex
}

type GRPCDialer interface {
	Dial(ctx context.Context, peerID peer.ID, dialOpts ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewRequester(prtc GRPCDialer) *BlockRequester {
	req := &BlockRequester{
		prtc:    prtc,
		peerIDs: make(chan peer.ID, 100),
		conn:    nil,
		stop:    make(chan int, 1),
		RWMutex: sync.RWMutex{},
	}
	go req.keepConnection()
	return req
}

// keepConnection dials highway to establish gRPC connection if it isn't available
func (c *BlockRequester) keepConnection() {
	currentHWID := peer.ID("")
	watchTimestep := time.Tick(RequesterDialTimestep)

	closeConnection := func() {
		c.Lock()
		defer c.Unlock()
		if c.conn == nil {
			return
		}

		Logger.Info("Closing old requester connection")
		err := c.conn.Close()
		if err != nil {
			Logger.Errorf("Failed closing old requester connection: %+v", err)
		}
		c.conn = nil
	}

	for {
		select {
		case <-watchTimestep:
			c.RLock()
			if c.ready() {
				continue
			}
			c.RUnlock()

			Logger.Warn("BlockRequester is not ready, dialing")
			closeConnection()
			ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
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
				c.Lock()
				c.conn = conn
				c.Unlock()
			}
			cancel()

		case hwID := <-c.peerIDs:
			Logger.Infof("Received new highway peerID, old = %s, new = %s", currentHWID.String(), hwID.String())
			if hwID != currentHWID && c.conn != nil {
				closeConnection()
			}
			currentHWID = hwID

		case <-c.stop:
			Logger.Info("Stop keeping blockrequester connection to highway")
			closeConnection()
			return
		}
	}
}

func (c *BlockRequester) ready() bool {
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	Logger.Infof("[blkbyheight] Requesting block shard %v (by specific %v): from = %v to = %v; height: %v", shardID, bySpecific, from, to, heights)

	client := proto.NewHighwayServiceClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardByHeight(
		ctx,
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	blkHashBytes := [][]byte{}
	for _, hash := range hashes {
		blkHashBytes = append(blkHashBytes, hash.GetBytes())
	}
	Logger.Infof("[blkbyhash] Requesting shard block by hash: %v", hashes)
	client := proto.NewHighwayServiceClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardByHash(
		ctx,
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	Logger.Infof("[blkbyheight] Requesting beaconblock (by specific %v): from = %v to = %v; height: %v", bySpecific, from, to, heights)
	client := proto.NewHighwayServiceClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockBeaconByHeight(
		ctx,
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	blkHashBytes := [][]byte{}
	for _, hash := range hashes {
		blkHashBytes = append(blkHashBytes, hash.GetBytes())
	}
	Logger.Infof("Requesting beacon block by hash: %v", hashes)
	client := proto.NewHighwayServiceClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockBeaconByHash(
		ctx,
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}

	Logger.Infof("[sync] Requesting blkshdtobcn by specific height %v: from = %v to = %v; Heights: %v", bySpecific, from, to, heights)
	client := proto.NewHighwayServiceClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardToBeaconByHeight(
		ctx,
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
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}

	Logger.Infof("Requesting block crossshard by height: shard %v to %v, height %v", fromShard, toShard, heights)
	client := proto.NewHighwayServiceClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockCrossShardByHeight(
		ctx,
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
