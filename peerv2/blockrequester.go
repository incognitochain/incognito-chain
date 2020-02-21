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
			ready := c.ready()
			c.RUnlock()
			if ready {
				continue
			}

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

	uuid := genUUID()
	client := proto.NewHighwayServiceClient(c.conn)
	reply, err := client.Register(
		ctx,
		&proto.RegisterRequest{
			CommitteePublicKey: pubkey,
			WantedMessages:     messages,
			CommitteeID:        committeeIDs,
			PeerID:             peer.IDB58Encode(selfID),
			Role:               role,
			UUID:               uuid,
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
	rangeBlks := batchingBlkForSync(defaultMaxBlkReqPerPeer, syncBlkInfo{
		bySpecHeights: bySpecific,
		byHash:        false,
		from:          from,
		to:            to,
		heights:       heights,
		hashes:        [][]byte{},
	})
	Logger.Infof("[blkbyheight] Requesting block shard %v (by specific %v): from = %v to = %v; height: %v number of batching %v", shardID, bySpecific, from, to, heights, len(rangeBlks))
	res := [][]byte{}
	client := proto.NewHighwayServiceClient(c.conn)
	for _, rangeBlk := range rangeBlks {
		uuid := genUUID()
		Logger.Infof("[blkbyheight] Range blk Requesting block shard %v (by specific %v): from = %v to = %v; height: %v, uuid = %s", shardID, bySpecific, rangeBlk.from, rangeBlk.to, rangeBlk.heights, uuid)
		ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
		defer cancel()
		reply, err := client.GetBlockShardByHeight(
			ctx,
			&proto.GetBlockShardByHeightRequest{
				Shard:      shardID,
				Specific:   bySpecific,
				FromHeight: rangeBlk.from,
				ToHeight:   rangeBlk.to,
				Heights:    rangeBlk.heights,
				FromPool:   false,
				UUID:       uuid,
			},
			grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
		)
		if err != nil {
			Logger.Errorf("Request block shard %v by spec height %v (from %v to %v height %v) return error %v, uuid = %v", shardID, bySpecific, rangeBlk.from, rangeBlk.to, rangeBlk.heights, err, uuid)
			continue
		}
		Logger.Infof("[blkbyheight] Received block shard %v data %v, uuid = %v", shardID, len(reply.Data), uuid)
		res = append(res, reply.Data...)
	}
	return res, nil
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
	res := [][]byte{}
	blkHashBytes := [][]byte{}
	for _, hash := range hashes {
		blkHashBytes = append(blkHashBytes, hash.GetBytes())
	}
	Logger.Infof("[blkbyhash] Requesting shard block %v by hash: %v", shardID, hashes)
	rangeBlks := batchingBlkForSync(defaultMaxBlkReqPerPeer, syncBlkInfo{
		bySpecHeights: false,
		byHash:        true,
		from:          0,
		to:            0,
		heights:       []uint64{},
		hashes:        blkHashBytes,
	})
	client := proto.NewHighwayServiceClient(c.conn)
	for _, rangeBlk := range rangeBlks {
		uuid := genUUID()
		ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
		defer cancel()
		reply, err := client.GetBlockShardByHash(
			ctx,
			&proto.GetBlockShardByHashRequest{
				Shard:  shardID,
				Hashes: rangeBlk.hashes,
				UUID:   uuid,
			},
			grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
		)
		if err != nil {
			Logger.Errorf("Request block shard %v by hashes %v return error %v, uuid = %s", shardID, hashes, err, uuid)
			continue
		}
		res = append(res, reply.Data...)
		Logger.Infof("[blkbyhash] Received block shard % by hashes data %v, uuid = %s", shardID, len(reply.Data), uuid)
	}
	return res, nil
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
	res := [][]byte{}
	rangeBlks := batchingBlkForSync(defaultMaxBlkReqPerPeer, syncBlkInfo{
		bySpecHeights: bySpecific,
		byHash:        false,
		from:          from,
		to:            to,
		heights:       heights,
		hashes:        [][]byte{},
	})
	for _, rangeBlk := range rangeBlks {
		uuid := genUUID()
		ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
		defer cancel()
		reply, err := client.GetBlockBeaconByHeight(
			ctx,
			&proto.GetBlockBeaconByHeightRequest{
				Specific:   bySpecific,
				FromHeight: rangeBlk.from,
				ToHeight:   rangeBlk.to,
				Heights:    rangeBlk.heights,
				FromPool:   false,
				UUID:       uuid,
			},
			grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
		)
		if err != nil {
			Logger.Errorf("Request block beacon by spec height %v (from %v to %v height %v) return error %v, uuid = %s", bySpecific, rangeBlk.from, rangeBlk.to, rangeBlk.heights, err, uuid)
			continue
		} else if reply != nil {
			res = append(res, reply.Data...)
			Logger.Infof("[blkbyheight] Received block beacon data len: %v, uuid = %s", len(reply.Data), uuid)
		}
	}
	return res, nil
}

func (c *BlockRequester) GetBlockBeaconByHash(
	hashes []common.Hash,
) ([][]byte, error) {
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	res := [][]byte{}
	blkHashBytes := [][]byte{}
	for _, hash := range hashes {
		blkHashBytes = append(blkHashBytes, hash.GetBytes())
	}
	Logger.Infof("Requesting beacon block by hash: %v", hashes)
	rangeBlks := batchingBlkForSync(defaultMaxBlkReqPerPeer, syncBlkInfo{
		bySpecHeights: false,
		byHash:        true,
		from:          0,
		to:            0,
		heights:       []uint64{},
		hashes:        blkHashBytes,
	})
	client := proto.NewHighwayServiceClient(c.conn)
	for _, rangeBlk := range rangeBlks {
		uuid := genUUID()
		ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
		defer cancel()
		reply, err := client.GetBlockBeaconByHash(
			ctx,
			&proto.GetBlockBeaconByHashRequest{
				Hashes: rangeBlk.hashes,
				UUID:   uuid,
			},
			grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
		)
		if err != nil {
			Logger.Errorf("Request block beacon by hashes %v return error %v, uuid = %s", hashes, err, uuid)
			continue
		}
		Logger.Infof("Received block beacon data from get beacon by hash %v, uuid = %s", len(reply.Data), uuid)
		res = append(res, reply.Data...)
	}
	return res, nil
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
	res := [][]byte{}
	Logger.Infof("[sync] Requesting blkshdtobcn shard %v by specific height %v: from = %v to = %v; Heights: %v", shardID, bySpecific, from, to, heights)
	client := proto.NewHighwayServiceClient(c.conn)
	rangeBlks := batchingBlkForSync(defaultMaxBlkReqPerPeer, syncBlkInfo{
		bySpecHeights: bySpecific,
		byHash:        false,
		from:          from,
		to:            to,
		heights:       heights,
		hashes:        [][]byte{},
	})
	Logger.Infof("[syncblkinfo] shard %v original from %v to %v heights %v", shardID, from, to, heights)
	for _, rangeBlk := range rangeBlks {
		uuid := genUUID()
		Logger.Infof("[syncblkinfo] shard %v from %v to %v heights %v, uuid = %s", shardID, rangeBlk.from, rangeBlk.to, rangeBlk.heights, uuid)
		ctx, cancel := context.WithTimeout(context.Background(), MaxTimePerRequest)
		defer cancel()
		reply, err := client.GetBlockShardToBeaconByHeight(
			ctx,
			&proto.GetBlockShardToBeaconByHeightRequest{
				FromShard:  shardID,
				Specific:   bySpecific,
				FromHeight: rangeBlk.from,
				ToHeight:   rangeBlk.to,
				Heights:    rangeBlk.heights,
				FromPool:   false,
				UUID:       uuid,
			},
			grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
		)
		if err != nil {
			Logger.Infof("[sync] Received err: %v from = %v to = %v shard %v; Heights: %v, uuid = %s", err, from, to, shardID, heights, uuid)
			continue
		} else if reply != nil {
			res = append(res, reply.Data...)
			Logger.Infof("[sync] Received block s2b (shard %v) data len: %v from = %v to = %v; Heights: %v , uuid = %s", shardID, len(reply.Data), from, to, heights, uuid)
		}
	}
	return res, nil
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
	res := [][]byte{}
	Logger.Infof("Requesting block crossshard by height: shard %v to %v, height %v", fromShard, toShard, heights)
	client := proto.NewHighwayServiceClient(c.conn)
	rangeBlks := batchingBlkForSync(defaultMaxBlkReqPerPeer, syncBlkInfo{
		bySpecHeights: true,
		byHash:        false,
		from:          0,
		to:            0,
		heights:       heights,
		hashes:        [][]byte{},
	})
	for _, rangeBlk := range rangeBlks {
		uuid := genUUID()
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
				Heights:    rangeBlk.heights,
				FromPool:   getFromPool,
				UUID:       uuid,
			},
			grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
		)
		if err != nil {
			Logger.Errorf("Request block crossshard by spec height (height %v) return error %v, uuid = %s", rangeBlk.heights, err, uuid)
			continue
		} else if reply != nil {
			Logger.Infof("Received block s2b data len: %v, uuid = %s", len(reply.Data), uuid)
			res = append(res, reply.Data...)
		}
	}
	return res, nil
}

type syncBlkInfo struct {
	bySpecHeights bool
	byHash        bool
	from          uint64
	to            uint64
	heights       []uint64
	hashes        [][]byte
}
