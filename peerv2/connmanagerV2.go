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
	id := common.RandInt()
	cm.hwLocker.RLock()
	if cm.currentHW == nil {
		cm.hwLocker.RUnlock()
		return
	}
	hwAddr := *cm.currentHW
	cm.hwLocker.RUnlock()
	addrInfo, err := getAddressInfo(hwAddr.Libp2pAddr)
	if err != nil {
		Logger.Errorf("Retry connect to HW %v failed, err: %v, id %v", err, id)
		return
	}
	Logger.Infof("Disconnected, network local peer %v, peer conn .local %v %v, .remote %v %v; hwAddr.libp2pAddr %v id %v", nw.LocalPeer().Pretty(), conn.LocalMultiaddr().String(), conn.LocalPeer().Pretty(), conn.RemoteMultiaddr().String(), conn.RemotePeer().Pretty(), addrInfo.ID.Pretty(), id)
	if conn.RemotePeer().Pretty() != addrInfo.ID.Pretty() {
		return
	}
	cm.CloseConnToCurHW(true)
	err = cm.tryToConnectHW(&hwAddr, id)
	if err != nil {
		Logger.Errorf("Cannot retry connection to HW %v id %v", addrInfo.ID.Pretty(), id)
		cm.keeper.IgnoreAddress(hwAddr)
		cm.reqPickHW <- nil
	}
}

func (cm *ConnManager) PickHighway() error {
	id := common.RandInt()
	Logger.Infof("[newpeerv2] start pick HW thread ID %v", id)
	defer Logger.Infof("[newpeerv2] pick HW done ID %v", id)
	newHW, err := cm.keeper.GetHighway(&cm.peerID)
	var hwAddrInfo *peer.AddrInfo
	gotNewHW := false
	if err == nil {
		Logger.Infof("[newpeerv2] Got new HW = %v %v", newHW.Libp2pAddr, id)
		cm.hwLocker.RLock()
		if (cm.currentHW != nil) && (newHW.Libp2pAddr == cm.currentHW.Libp2pAddr) {
			Logger.Infof("[newpeerv2] currentHW == new HW")
			cm.hwLocker.RUnlock()
			return nil
		}
		cm.hwLocker.RUnlock()
		hwAddrInfo, err = getAddressInfo(newHW.Libp2pAddr)
		Logger.Infof("[newpeerv2] get address info %v %v %v", hwAddrInfo, err, id)
		if err == nil {
			err = cm.tryToConnectHW(newHW, id)
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
		if newHW != nil {
			cm.keeper.IgnoreAddress(*newHW)
		}
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

func (cm *ConnManager) tryToConnectHW(hwAddr *rpcclient.HighwayAddr, id int) error {
	var err error
	hwAddrInfo, err := getAddressInfo(hwAddr.Libp2pAddr)
	if err != nil {
		return err
	}
	Logger.Infof("Try to connect new HW %v, thread id %v", hwAddr.Libp2pAddr, id)
	for i := 0; i < MaxConnectionRetry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
		if err = cm.LocalHost.Host.Connect(ctx, *hwAddrInfo); err != nil {
			Logger.Errorf("Could not connect to highway: %v %v, failed times %v id %v", err, hwAddrInfo, i, id)
		} else {
			err = cm.Requester.ConnectNewHW(hwAddrInfo, id)
			if err == nil {
				cm.newHighway <- hwAddr
				Logger.Infof("Connected to HW %v %v", hwAddrInfo.ID, id)
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
	cm.hwLocker.RLock()
	if cm.currentHW == nil {
		cm.hwLocker.RUnlock()
		return
	}
	addrInfo, err := getAddressInfo(cm.currentHW.Libp2pAddr)
	cm.hwLocker.RUnlock()
	if err != nil {
		Logger.Error(err)
		return
	}
	Logger.Infof("Closing connection to HW %v", addrInfo.ID.Pretty())
	if !isDisconnected {
		if err := cm.LocalHost.Host.Network().ClosePeer(addrInfo.ID); err != nil {
			Logger.Errorf("Failed closing connection to old highway: hwID = %s err = %v", addrInfo.ID.Pretty(), err)
		}
	} else {
		Logger.Infof("[debugdisconnect] Send signal stop to Requester -->")
		cm.Requester.CloseConnection(0)
		Logger.Infof("Send signal stop to Requester DONE")
	}
	cm.hwLocker.Lock()
	cm.currentHW = nil
	cm.keeper.setCurrHW(nil)
	cm.hwLocker.Unlock()
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
			cm.hwLocker.RLock()
			sameHW := true
			if cm.currentHW != nil {
				if cm.currentHW.Libp2pAddr != newHW.Libp2pAddr {
					sameHW = false
				} else {
					Logger.Infof("[debugGRPC] New HW and current HW is the same")
				}
			}
			cm.hwLocker.RUnlock()
			if !sameHW {
				cm.CloseConnToCurHW(false)
			}
			cm.hwLocker.Lock()
			cm.currentHW = newHW
			cm.hwLocker.Unlock()
			Logger.Info("[debugGRPC] Force subscribe %v", newHW)
			err := cm.Subscriber.Subscribe(true)
			if err != nil {
				Logger.Errorf("[debugGRPC] Subscribe to HW %v failed, ignore this HW and repick", newHW.Libp2pAddr)
				cm.keeper.IgnoreAddress(*newHW)
				cm.reqPickHW <- nil
			} else {
				cm.keeper.setCurrHW(newHW)
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
			cm.hwLocker.RLock()
			if cm.currentHW != nil {
				Logger.Debugf("Resubscriber to currentHW %v", cm.currentHW.Libp2pAddr)
				err = cm.Subscriber.Subscribe(false)
				if err != nil {
					Logger.Errorf("Subscribe failed: forced = %v hwID = %s err = %+v", forced, hwID.String(), err)
				}
			}
			cm.hwLocker.RUnlock()
		case <-cm.Requester.disconnectNoti:
			Logger.Infof("Received signal disconnected, repickHW")
			cm.reqPickHW <- nil
		case <-cm.stop:
			Logger.Info("Stop managing role subscription")
			break
		}
	}
}

func (cm *ConnManager) GetConnectionStatus() interface{} {
	return cm.keeper.exportStatus()
}

func (cm *ConnManager) IsReady() bool {
	if (cm == nil) || (cm.Requester == nil) {
		return false
	}
	return cm.Requester.IsReady()
}
