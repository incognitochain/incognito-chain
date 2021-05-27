package peerv2

import (
	"context"
	"time"

	pubsub "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/rpcclient"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc"
)

func (cm *ConnManager) StartV2(ns NetSync) {
	// Pubsub
	var err error
	cm.ps, err = pubsub.NewFloodSub(
		context.Background(),
		cm.LocalHost.Host,
		pubsub.WithMaxMessageSize(common.MaxPSMsgSize),
		pubsub.WithPeerOutboundQueueSize(1024),
		pubsub.WithValidateQueueSize(1024),
	)

	if err != nil {
		panic(err)
	}
	cm.messages = make(chan *pubsub.Message, 1000)

	go cm.keeper.Start(cm.LocalHost, cm.rttService, cm.discoverer)

	// NOTE: must Connect after creating FloodSub
	cm.Requester = NewRequesterV2(cm.LocalHost.GRPC)
	cm.Subscriber = NewSubManager(cm.info, cm.ps, cm.Requester, cm.messages)

	go cm.manageHighwayConnection()

	cm.Provider = NewBlockProvider(cm.LocalHost.GRPC, ns)
	go cm.keepConnectionAlive()
	cm.process()
}

func (cm *ConnManager) disconnectAction(nw network.Network, conn network.Conn) {
	Logger.Infof("Disconnected, network local peer %v, peer conn .local %v %v, .remote %v %v", nw.LocalPeer().Pretty(), conn.LocalMultiaddr().String(), conn.LocalPeer().Pretty(), conn.RemoteMultiaddr().String(), conn.RemotePeer().Pretty())
	if conn.RemotePeer().Pretty() != cm.currentHW.Libp2pAddr {
		return
	}
	addrInfo, err := getAddressInfo(cm.currentHW.Libp2pAddr)
	if err != nil {
		Logger.Errorf("Retry connect to HW %v failed, err: %v", err)
		return
	}
	for i := 0; i < MaxConnectionRetry; i++ {
		ctx := context.Background()
		if err := cm.LocalHost.Host.Connect(ctx, *addrInfo); err != nil {
			Logger.Errorf("Could not connect to highway: %v %v", err, addrInfo)
			time.Sleep(ReconnectHighwayTimestep)
			continue
		}
		return
	}
	cm.keeper.IgnoreAddress(*cm.currentHW)
	err = cm.PickHighway()
	if err != nil {
		Logger.Error(err)
		//TODO dosomething
	}
}

func (cm *ConnManager) PickHighway() error {
	Logger.Infof("[newpeerv2] start pick HW")
	defer Logger.Infof("[newpeerv2] pick HW done")
	newHW, err := cm.keeper.GetHighway(&cm.peerID)
	var hwAddrInfo *peer.AddrInfo
	if err == nil {
		time.Sleep(2 * time.Second)
		// cm.keeper.IgnoreAddress(*newHW)
		Logger.Infof("[newpeerv2] Got new HW = %v", newHW.Libp2pAddr)
		// cm.reqPickHW <- nil
		if (cm.currentHW != nil) && (newHW.Libp2pAddr == cm.currentHW.Libp2pAddr) {
			// time.Sleep(2 * time.Second)
			// cm.keeper.IgnoreAddress(*newHW)
			Logger.Infof("[newpeerv2] currentHW == new HW")
			// cm.reqPickHW <- nil
			return nil
		}
		hwAddrInfo, err = getAddressInfo(newHW.Libp2pAddr)
		Logger.Infof("[newpeerv2] get address info %v %v", hwAddrInfo, err)
		if err == nil {
			err = tryToConnect(cm, hwAddrInfo)
			if err == nil {
				var conn *grpc.ClientConn
				conn, err = cm.Requester.tryToDial(hwAddrInfo)
				if err == nil {
					cm.Requester.closeConnection()
					cm.Requester.Lock()
					cm.Requester.conn = conn
					cm.Requester.Unlock()
					go cm.Requester.watchConnection(context.Background(), hwAddrInfo.ID)
					cm.newHighway <- newHW
				}
			}
		}
	}
	if err != nil {
		Logger.Error(err)
		time.Sleep(2 * time.Second)
		cm.keeper.IgnoreAddress(*newHW)
		cm.reqPickHW <- nil
		return err
	}
	return nil
}

func tryToConnect(cm *ConnManager, hwAddrInfo *peer.AddrInfo) error {
	var err error
	for i := 0; i < MaxConnectionRetry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
		defer cancel()
		if err = cm.LocalHost.Host.Connect(ctx, *hwAddrInfo); err != nil {
			Logger.Errorf("Could not connect to highway: %v %v", err, hwAddrInfo)
		} else {
			Logger.Infof("Connected to HW %v", hwAddrInfo.ID)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return err
}

func (cm *ConnManager) CloseConnToCurHW() {
	pID, err := peer.IDB58Decode(cm.currentHW.Libp2pAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
	// cm.Requester.closeConnection()
	if err := cm.LocalHost.Host.Network().ClosePeer(pID); err != nil {
		Logger.Errorf("Failed closing connection to old highway: hwID = %s err = %v", pID.String(), err)
	}
}

func (cm *ConnManager) checkConnectionStatus(addrInfo *peer.AddrInfo) bool {
	net := cm.LocalHost.Host.Network()
	net.Notify(&network.NotifyBundle{})
	// Reconnect if not connected
	if net.Connectedness(addrInfo.ID) != network.Connected {
		cm.disconnected++
		cm.registered = false // Next time we connect to highway, we need to register again
		Logger.Info("Not connected to highway, connecting")
		ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
		defer cancel()
		if err := cm.LocalHost.Host.Connect(ctx, *addrInfo); err != nil {
			Logger.Errorf("Could not connect to highway: %v %v", err, addrInfo)
		}
		if cm.disconnected > MaxConnectionRetry {
			Logger.Error("Retry maxed out")
			cm.disconnected = 0 // Retry N times for next chosen highway
			return true
		}
	}

	if !cm.registered && net.Connectedness(addrInfo.ID) == network.Connected {
		// Register again since this might be a new highway
		Logger.Info("Connected to highway, sending register request")
		cm.registerRequests <- addrInfo.ID
		cm.disconnected = 0
		cm.registered = true
	}
	return false
}

func (cm *ConnManager) manageHighwayConnection() {
	for _, dpa := range cm.DiscoverPeersAddress {
		cm.keeper.Add(
			rpcclient.HighwayAddr{
				Libp2pAddr: "",
				RPCUrl:     dpa,
			},
		)
	}

	go func(cm *ConnManager) {
		for {
			cm.reqPickHW <- nil
			time.Sleep(10 * time.Minute)
		}
	}(cm)
	for {
		select {
		case <-cm.reqPickHW:
			Logger.Info("Received request repick HW")
			err := cm.PickHighway()
			if err != nil {
				Logger.Error(err)
			}
		case newHW := <-cm.newHighway:
			Logger.Info("Received newHW %v", newHW)
			if cm.currentHW != nil {
				Logger.Info("newHW %v current %v", cm.currentHW.Libp2pAddr, newHW.Libp2pAddr)
				if cm.currentHW.Libp2pAddr == newHW.Libp2pAddr {
					continue
				}
				cm.CloseConnToCurHW()
			}
			Logger.Info("newHW1 %v current1 %v", cm.currentHW, newHW)
			cm.currentHW = newHW
			for i := 0; i < MaxConnectionRetry; i++ {
				err := cm.Subscriber.Subscribe(true)
				if err != nil {
					cm.keeper.IgnoreAddress(*cm.currentHW)
					cm.reqPickHW <- nil
				}
			}
		case <-cm.stop:
			Logger.Info("Stop keeping connection to highway")
			break
		}
	}
}

// manageRoleSubscription: polling current role periodically and subscribe to relevant topics
func (cm *ConnManager) keepConnectionAlive() {
	forced := false // only subscribe when role changed or last forced subscribe failed
	hwID := peer.ID("")
	var err error
	subsTimestep := time.NewTicker(CheckSubsTimestep)
	defer subsTimestep.Stop()
	for {
		select {
		case <-subsTimestep.C:
			if cm.currentHW != nil {
				err = cm.Subscriber.Subscribe(false)
				if err != nil {
					Logger.Errorf("Subscribe failed: forced = %v hwID = %s err = %+v", forced, hwID.String(), err)
				}
			}
		case <-cm.Requester.disconnectNoti:
			cm.reqPickHW <- nil
		case <-cm.stop:
			Logger.Info("Stop managing role subscription")
			break
		}
	}
}
