package peerv2

import (
	"context"
	"time"

	pubsub "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/rpcclient"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

func (cm *ConnManager) StartV2(bg BlockGetter) {
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

	go cm.keeper.Start(cm.LocalHost, cm.rttService, cm.discoverer, cm.DiscoverPeersAddress)

	// NOTE: must Connect after creating FloodSub
	cm.Requester = NewRequesterV2(cm.LocalHost.GRPC)
	cm.Subscriber = NewSubManager(cm.info, cm.ps, cm.Requester, cm.messages, cm.disp)

	go cm.manageHighwayConnection()

	cm.Provider = NewBlockProvider(cm.LocalHost.GRPC, bg)
	go cm.keepConnectionAlive()
	cm.process()
}

func (cm *ConnManager) disconnectAction(nw network.Network, conn network.Conn) {
	addrInfo, err := getAddressInfo(cm.currentHW.Libp2pAddr)
	if err != nil {
		Logger.Errorf("Retry connect to HW %v failed, err: %v", err)
		return
	}
	Logger.Infof("Disconnected, network local peer %v, peer conn .local %v %v, .remote %v %v; currentHW.libp2pAddr %v", nw.LocalPeer().Pretty(), conn.LocalMultiaddr().String(), conn.LocalPeer().Pretty(), conn.RemoteMultiaddr().String(), conn.RemotePeer().Pretty(), addrInfo.ID.Pretty())
	if conn.RemotePeer().Pretty() != addrInfo.ID.Pretty() {
		return
	}
	hwAddr := *cm.currentHW
	cm.CloseConnToCurHW(true)
	err = cm.tryToConnectHW(&hwAddr)
	if err != nil {
		Logger.Errorf("Cannot retry connection to HW %v", addrInfo.ID.Pretty())
		cm.keeper.IgnoreAddress(hwAddr)
		cm.reqPickHW <- nil
	}
}

func (cm *ConnManager) PickHighway() error {
	Logger.Infof("[newpeerv2] start pick HW")
	defer Logger.Infof("[newpeerv2] pick HW done")
	newHW, err := cm.keeper.GetHighway(&cm.peerID)
	var hwAddrInfo *peer.AddrInfo
	gotNewHW := false
	if err == nil {
		// time.Sleep(2 * time.Second)
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
			err = cm.tryToConnectHW(newHW)
			if err == nil {
				gotNewHW = true
			}
		}
	}
	if err != nil {
		Logger.Error(err)
	}
	if !gotNewHW {
		time.Sleep(2 * time.Second)
		cm.keeper.IgnoreAddress(*newHW)
		Logger.Info("Can not pick HW, repick")
		cm.reqPickHW <- nil
		return errors.Errorf("Can not pick new Highway, err %v", err)
	}
	return nil
}

func (cm *ConnManager) tryToConnect(hwAddrInfo *peer.AddrInfo) error {
	var err error
	Logger.Infof("Start tryToConnect")
	for i := 0; i < MaxConnectionRetry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
		defer cancel()
		Logger.Infof("TryToConnect to hw %v", hwAddrInfo.ID.Pretty())
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

func (cm *ConnManager) tryToConnectHW(hwAddr *rpcclient.HighwayAddr) error {
	var err error
	hwAddrInfo, err := getAddressInfo(hwAddr.Libp2pAddr)
	if err != nil {
		return err
	}
	Logger.Infof("Try to connect new HW %v", hwAddr.Libp2pAddr)
	for i := 0; i < MaxConnectionRetry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
		if err = cm.LocalHost.Host.Connect(ctx, *hwAddrInfo); err != nil {
			Logger.Errorf("Could not connect to highway: %v %v, failed times %v", err, hwAddrInfo, i)
		} else {
			err = cm.Requester.ConnectNewHW(hwAddrInfo)
			if err == nil {
				cm.newHighway <- hwAddr
				Logger.Infof("Connected to HW %v", hwAddrInfo.ID)
				cancel()
				return nil
			}
		}
		cancel()
		time.Sleep(2 * time.Second)
	}
	return err
}

func (cm *ConnManager) CloseConnToCurHW(isDisconnected bool) {
	addrInfo, err := getAddressInfo(cm.currentHW.Libp2pAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
	Logger.Infof("Closing connection to HW %v", addrInfo.ID.Pretty())
	Logger.Infof("[debugdisconnect] Send signal stop to Requester -->")
	if cm.Requester.isRunning {
		cm.Requester.stop <- 0
	}
	Logger.Infof("Send signal stop to Requester DONE")
	if !isDisconnected {
		if err := cm.LocalHost.Host.Network().ClosePeer(addrInfo.ID); err != nil {
			Logger.Errorf("Failed closing connection to old highway: hwID = %s err = %v", addrInfo.ID.Pretty(), err)
		}
	}
	cm.currentHW = nil
	cm.keeper.setCurrHW(nil)
	Logger.Infof("[debugdisconnect] CloseConnToCurHW DONE")
}

func (cm *ConnManager) manageHighwayConnection() {
	Logger.Infof("manageHighwayConnection")

	go func(cm *ConnManager) {
		for {
			Logger.Infof("Pick HW intervally")
			cm.reqPickHW <- nil
			time.Sleep(10 * time.Minute)
		}
	}(cm)
	for {
		select {
		case <-cm.reqPickHW:
			Logger.Info("[debugGRPC] Received request repick HW")
			err := cm.PickHighway()
			if err != nil {
				Logger.Error(err)
			}
			Logger.Info("[debugGRPC] Pick HW Done")
		case newHW := <-cm.newHighway:
			Logger.Info("[debugGRPC] Received newHW %v", newHW)
			if cm.currentHW != nil {
				Logger.Info("[debugGRPC] newHW %v current %v", newHW.Libp2pAddr, cm.currentHW.Libp2pAddr)
				if cm.currentHW.Libp2pAddr == newHW.Libp2pAddr {
					Logger.Infof("[debugGRPC] New HW and current HW is the same")
					continue
				}
				cm.CloseConnToCurHW(false)
			}
			cm.currentHW = newHW
			cm.keeper.currentHW = newHW
			Logger.Info("[debugGRPC] Force subscribe %v", newHW)
			err := cm.Subscriber.Subscribe(true)
			if err != nil {
				Logger.Errorf("[debugGRPC] Subscribe to HW %v failed, ignore this HW and repick", cm.currentHW.Libp2pAddr)
				cm.keeper.IgnoreAddress(*cm.currentHW)
				cm.reqPickHW <- nil
			} else {
				cm.keeper.setCurrHW(cm.currentHW)
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
	Logger.Infof("keepConnectionAlive")
	hwID := peer.ID("")
	var err error
	subsTimestep := time.NewTicker(CheckSubsTimestep)
	defer subsTimestep.Stop()
	for {
		select {
		case <-subsTimestep.C:
			if cm.currentHW != nil {
				Logger.Infof("Resubscriber to currentHW %v", cm.currentHW.Libp2pAddr)
				err = cm.Subscriber.Subscribe(false)
				if err != nil {
					Logger.Errorf("Subscribe failed: forced = %v hwID = %s err = %+v", forced, hwID.String(), err)
				}
			}
		case <-cm.Requester.disconnectNoti:
			Logger.Infof("Received signal disconnected, repickHW")
			cm.reqPickHW <- nil
		case <-cm.stop:
			Logger.Info("Stop managing role subscription")
			break
		}
	}
}
