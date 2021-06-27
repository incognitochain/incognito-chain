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
	isRunning bool

	disconnectNoti chan bool
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

		disconnectNoti: make(chan bool, 10),
	}
	go req.keepConnection()
	return req
}

func (c *BlockRequester) closeConnection() {
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

func (c *BlockRequester) keepConnection() {
	currentHWID := peer.ID("")
	watchTimestep := time.NewTicker(RequesterDialTimestep)
	defer watchTimestep.Stop()

	for {
		select {
		case <-watchTimestep.C:
			ready := c.IsReady()
			if ready {
				continue
			}

			Logger.Warn("BlockRequester is not ready, dialing")
			c.closeConnection()
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
				c.closeConnection()
			}
			if currentHWID == "" && c.conn == nil {
				ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
				if conn, err := c.prtc.Dial(
					ctx,
					hwID,
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
			}

			currentHWID = hwID
		case <-c.stop:
			Logger.Info("Stop keeping blockrequester connection to highway")
			c.closeConnection()
			return
		}
	}
}

func (c *BlockRequester) IsReady() bool {
	c.RLock()
	ready := c.ready()
	c.RUnlock()
	return ready
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
		return nil, nil, errors.New("requester still not ready")
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

func (c *BlockRequester) GetBlockShardByHash(
	shardID int32,
	hashes []common.Hash,
) ([][]byte, error) {
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester still not ready")
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

func (c *BlockRequester) StreamBlockByHeight(
	ctx context.Context,
	req *proto.BlockByHeightRequest,
) (proto.HighwayService_StreamBlockByHeightClient, error) {

	uuid := genUUID()
	Logger.Infof("[stream] Requesting stream block type %v, spec %v, height [%v..%v] len %v, from %v to %v, uuid = %s", req.Type, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights), req.From, req.To, uuid)
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	req.UUID = uuid
	client := proto.NewHighwayServiceClient(c.conn)
	stream, err := client.StreamBlockByHeight(ctx, req, grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize))
	if err != nil {
		Logger.Infof("[stream] This client not return stream for this request %v, got error %v ", req, err)
		return nil, err
	}
	return stream, nil
}

func (c *BlockRequester) StreamBlockByHash(
	ctx context.Context,
	req *proto.BlockByHashRequest,
) (proto.HighwayService_StreamBlockByHashClient, error) {

	uuid := genUUID()
	Logger.Infof("[stream] Requesting stream block type %v, hashes [%v..%v] len %v, from %v to %v, uuid = %s", req.Type, req.Hashes[0], req.Hashes[len(req.Hashes)-1], len(req.Hashes), req.From, req.To, uuid)
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester not ready")
	}
	req.UUID = uuid
	client := proto.NewHighwayServiceClient(c.conn)
	stream, err := client.StreamBlockByHash(ctx, req, grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize))
	if err != nil {
		Logger.Infof("[stream] This client not return stream for this request %v, got error %v ", req, err)
		return nil, err
	}
	return stream, nil
}

func (c *BlockRequester) GetBlockBeaconByHash(
	hashes []common.Hash,
) ([][]byte, error) {
	c.RLock()
	defer c.RUnlock()
	if !c.ready() {
		return nil, errors.New("requester still not ready")
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

type syncBlkInfo struct {
	bySpecHeights bool
	byHash        bool
	from          uint64
	to            uint64
	heights       []uint64
	hashes        [][]byte
}

func NewRequesterV2(prtc GRPCDialer) *BlockRequester {
	req := &BlockRequester{
		prtc:    prtc,
		peerIDs: make(chan peer.ID, 100),
		conn:    nil,
		stop:    make(chan int, 1),
		RWMutex: sync.RWMutex{},

		disconnectNoti: make(chan bool, 10),
	}
	return req
}

func (c *BlockRequester) tryToDial(hwAddrInfo *peer.AddrInfo) (conn *grpc.ClientConn, err error) {
	for i := 0; i < MaxConnectionRetry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
		Logger.Infof("Dial to new HW %v", hwAddrInfo.ID.Pretty())
		if conn, err = c.prtc.Dial(
			ctx,
			hwAddrInfo.ID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    RequesterKeepaliveTime,
				Timeout: RequesterKeepaliveTimeout,
			}),
		); err != nil {
			Logger.Error("Could not dial to highway grpc server:", err, hwAddrInfo.ID)
		}
		cancel()

		if (conn != nil) && (conn.GetState() != connectivity.Ready) {
			time.Sleep(2 * time.Second)
		}
		if (conn != nil) && (conn.GetState() == connectivity.Ready) {
			return conn, nil
		}
	}
	Logger.Error("Could not dial to highway grpc server:", err, hwAddrInfo.ID)
	return nil, err
}

func (c *BlockRequester) ConnectNewHW(hwAddrInfo *peer.AddrInfo) (err error) {
	Logger.Infof("[debugGRPC] Connecting to new HW %v", hwAddrInfo.ID.Pretty())
	var conn *grpc.ClientConn
	conn, err = c.tryToDial(hwAddrInfo)
	if err == nil {
		Logger.Infof("[debugGRPC] Connected to new HW %v", hwAddrInfo.ID.Pretty())
		c.closeConnection()
		Logger.Infof("[debugGRPC] Closed old connection")
		c.Lock()
		c.conn = conn
		c.Unlock()
		Logger.Infof("[debugGRPC] ~~> Set new conn done, conn state %v", c.conn.GetState().String())
		go c.WatchConnection(hwAddrInfo.ID)
	}
	return err
}

func (c *BlockRequester) retryDial(retryFailedCounter int, hwAddrInfo *peer.AddrInfo) (conn *grpc.ClientConn, retryFailedCounterN int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	Logger.Infof("retry dial to new HW %v", hwAddrInfo.ID.Pretty())
	if conn, err = c.prtc.Dial(
		ctx,
		hwAddrInfo.ID,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    RequesterKeepaliveTime,
			Timeout: RequesterKeepaliveTimeout,
		}),
	); err != nil {
		Logger.Error("Could not dial to highway grpc server:", err, hwAddrInfo.ID)
	}
	cancel()
	if (conn != nil) && (conn.GetState() != connectivity.Ready) {
		time.Sleep(2 * time.Second)
	}
	if (conn != nil) && (conn.GetState() == connectivity.Ready) {
		return conn, 0, nil
	}
	Logger.Error("Could not dial to highway grpc server:", err, hwAddrInfo.ID)
	return nil, retryFailedCounter + 1, err
}

func (c *BlockRequester) WatchConnection(currentHW peer.ID) {
	x := common.RandInt()
	c.isRunning = true
	Logger.Infof("[debugGRPC] Start WatchConnection to HW %v - %v", currentHW.Pretty(), x)
	defer Logger.Infof("[debugGRPC] End WatchConnection to HW %v - %v", currentHW.Pretty(), x)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	reqRetry := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	Logger.Infof("%v", ctx)
	Logger.Infof("%v", currentHW)
	Logger.Infof("%v", reqRetry)
	Logger.Infof("%v", c.conn)
	go c.watchConnection(ctx, currentHW, reqRetry, x)
	needToRetry := false
	failedCounter := 0
	for {
		select {
		case <-ticker.C:
			if needToRetry {
				Logger.Infof("[debugGRPC] needtoRetry: retry dial to new HW %v", currentHW.Pretty())
				conn, failedCounter, err := c.retryDial(failedCounter, &peer.AddrInfo{ID: currentHW})
				if failedCounter >= MaxConnectionRetry {
					if c.disconnectNoti != nil {
						c.disconnectNoti <- true
					}
					cancel()
					c.isRunning = false
					return
				} else {
					if (err == nil) && (conn != nil) {
						Logger.Infof("[debugGRPC]  needtoRetry: retry ok, connected to new HW %v, conn %v", currentHW.Pretty(), conn)
						c.Lock()
						c.conn = conn
						c.Unlock()
						go c.watchConnection(ctx, currentHW, reqRetry, x)
						needToRetry = false
					} else {
						Logger.Errorf("Retry connection failed %v times, error %v", failedCounter, err)
					}
				}

			}
		case <-c.stop:
			Logger.Infof("[debugGRPC] Received stop WatchConnection signal, id %v", x)
			cancel()
			c.closeConnection()
			c.isRunning = false
			Logger.Infof("[debugGRPC] Received stop WatchConnection signal, id %v DONE", x)
			return
		case <-reqRetry:
			Logger.Infof("[debugGRPC] Received requestRetry WatchConnection signal, id %v", x)
			c.closeConnection()
			needToRetry = true
		}
	}

}

func (c *BlockRequester) watchConnection(ctx context.Context, currentHW peer.ID, reqRetry chan bool, id int) {
	c.RLock()
	Logger.Infof("Start watchConnection to HW %v; ID %v", currentHW.Pretty(), id)
	defer Logger.Infof("End watchConnection to HW %v; ID %v", currentHW.Pretty(), id)
	Logger.Infof("---> %v", c.conn)
	Logger.Infof("---> %v", reqRetry)
	Logger.Infof("---> %v", ctx)
	disconnected := c.conn.WaitForStateChange(ctx, connectivity.Ready)
	c.RUnlock()
	Logger.Infof("[debugGRPC] BlockRequester disconnected, send to reqRetry id %v", id)
	reqRetry <- disconnected
}
