package connmanager

import (
	"fmt"
	"math"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	libpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/constant/bootnode/server"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/peer"
	"github.com/ninjadotorg/constant/wire"
)

var SHARD_NUMBER = 256

// ConnState represents the state of the requested connection.
type ConnState uint8

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.

type ConsensusState struct {
	sync.Mutex
	Role            string
	CurrentShard    *byte
	BeaconCommittee []string
	ShardCommittee  map[byte][]string
	UserPbk         string
	Committee       map[string]byte
	ShardNumber     int
}

func (self *ConsensusState) rebuild() {
	self.Committee = make(map[string]byte)
	for shard, committees := range self.ShardCommittee {
		for _, committee := range committees {
			self.Committee[committee] = shard
		}
	}
}

func (self *ConsensusState) GetBeaconCommittee() []string {
	self.Lock()
	defer self.Unlock()
	ret := make([]string, len(self.BeaconCommittee))
	copy(ret, self.BeaconCommittee)
	return ret
}

func (self *ConsensusState) GetShardCommittee(shard byte) []string {
	self.Lock()
	defer self.Unlock()
	committee, ok := self.ShardCommittee[shard]
	if ok {
		ret := make([]string, len(committee))
		copy(ret, committee)
		return ret
	}
	return make([]string, 0)
}

type ConnManager struct {
	connReqCount uint64
	start        int32
	stop         int32
	// Discover Peers
	discoveredPeers     map[string]*DiscoverPeerInfo
	discoverPeerAddress string
	// channel
	cQuit            chan struct{}
	cDiscoveredPeers chan struct{}

	Config Config

	ListeningPeers map[string]*peer.Peer

	randShards []byte
}

type Config struct {
	ExternalAddress    string
	MaxPeersSameShard  int
	MaxPeersOtherShard int
	MaxPeersOther      int
	MaxPeersNoShard    int
	MaxPeersBeacon     int
	// ListenerPeers defines a slice of listeners for which the connection
	// manager will take ownership of and accept connections.  When a
	// connection is accepted, the OnAccept handler will be invoked with the
	// connection.  Since the connection manager takes ownership of these
	// listeners, they will be closed when the connection manager is
	// stopped.
	//
	// This field will not have any effect if the OnAccept field is not
	// also specified.  It may be nil if the caller does not wish to listen
	// for incoming connections.
	ListenerPeers []*peer.Peer

	// OnInboundAccept is a callback that is fired when an inbound connection is accepted
	OnInboundAccept func(peerConn *peer.PeerConn)

	//OnOutboundConnection is a callback that is fired when an outbound connection is established
	OnOutboundConnection func(peerConn *peer.PeerConn)

	//OnOutboundDisconnection is a callback that is fired when an outbound connection is disconnected
	OnOutboundDisconnection func(peerConn *peer.PeerConn)

	DiscoverPeers        bool
	DiscoverPeersAddress string
	ConsensusState       *ConsensusState
}

type DiscoverPeerInfo struct {
	PublicKey  string
	RawAddress string
	PeerID     libpeer.ID
}

func (self *ConnManager) UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string) {
	self.Config.ConsensusState.Lock()
	defer self.Config.ConsensusState.Unlock()

	bChange := false
	if self.Config.ConsensusState.Role != role {
		self.Config.ConsensusState.Role = role
		bChange = true
	}
	if (self.Config.ConsensusState.CurrentShard != nil && currentShard == nil) ||
		(self.Config.ConsensusState.CurrentShard == nil && currentShard != nil) ||
		(self.Config.ConsensusState.CurrentShard != nil && currentShard != nil && *self.Config.ConsensusState.CurrentShard != *currentShard) {
		self.Config.ConsensusState.CurrentShard = currentShard
		bChange = true
	}
	if !common.CompareStringArray(self.Config.ConsensusState.BeaconCommittee, beaconCommittee) {
		self.Config.ConsensusState.BeaconCommittee = make([]string, len(beaconCommittee))
		copy(self.Config.ConsensusState.BeaconCommittee, beaconCommittee)
		bChange = true
	}
	if len(self.Config.ConsensusState.ShardCommittee) != len(shardCommittee) {
		for shardID, _ := range self.Config.ConsensusState.ShardCommittee {
			_, ok := shardCommittee[shardID]
			if !ok {
				delete(self.Config.ConsensusState.ShardCommittee, shardID)
			}
		}
		bChange = true
	}
	if self.Config.ConsensusState.ShardCommittee == nil {
		self.Config.ConsensusState.ShardCommittee = make(map[byte][]string)
	}
	for shardID, committee := range shardCommittee {
		_, ok := self.Config.ConsensusState.ShardCommittee[shardID]
		if ok {
			if !common.CompareStringArray(self.Config.ConsensusState.ShardCommittee[shardID], committee) {
				self.Config.ConsensusState.ShardCommittee[shardID] = make([]string, len(committee))
				copy(self.Config.ConsensusState.ShardCommittee[shardID], committee)
				bChange = true
			}
		} else {
			self.Config.ConsensusState.ShardCommittee[shardID] = make([]string, len(committee))
			copy(self.Config.ConsensusState.ShardCommittee[shardID], committee)
			bChange = true
		}
	}
	if self.Config.ConsensusState.UserPbk != userPbk {
		self.Config.ConsensusState.UserPbk = userPbk
		bChange = true
	}

	// update peer connection
	if bChange {
		self.Config.ConsensusState.rebuild()
		self.processDiscoverPeers()
	}

	return
}

// Stop gracefully shuts down the connection manager.
func (self *ConnManager) Stop() {
	if atomic.AddInt32(&self.stop, 1) != 1 {
		Logger.log.Error("Connection manager already stopped")
		return
	}
	Logger.log.Warn("Stopping connection manager")

	// Stop all the listeners.  There will not be any listeners if
	// listening is disabled.
	for _, listener := range self.Config.ListenerPeers {
		listener.Stop()
	}

	if self.cDiscoveredPeers != nil {
		close(self.cDiscoveredPeers)
	}

	close(self.cQuit)
	Logger.log.Warn("Connection manager stopped")
}

func (self ConnManager) New(cfg *Config) *ConnManager {
	self.Config = *cfg
	self.cQuit = make(chan struct{})
	self.discoveredPeers = make(map[string]*DiscoverPeerInfo)
	self.ListeningPeers = make(map[string]*peer.Peer)
	self.Config.ConsensusState = &ConsensusState{}
	return &self
}

func (self *ConnManager) GetPeerId(addr string) string {
	ipfsAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		Logger.log.Error(err)
		return EmptyString
	}
	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		Logger.log.Error(err)
		return EmptyString
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		Logger.log.Error(err)
		return EmptyString
	}
	return peerId.Pretty()
}

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (self *ConnManager) Connect(addr string, pubKey string) {
	if atomic.LoadInt32(&self.stop) != 0 {
		return
	}
	// The following code extracts target's peer Id from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>

	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", libpeer.IDB58Encode(peerId)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	for _, listen := range self.Config.ListenerPeers {
		listen.HandleConnected = self.handleConnected
		listen.HandleDisconnected = self.handleDisconnected
		listen.HandleFailed = self.handleFailed

		peer := peer.Peer{
			TargetAddress:      targetAddr,
			PeerID:             peerId,
			RawAddress:         addr,
			Config:             listen.Config,
			PeerConns:          make(map[string]*peer.PeerConn),
			PendingPeers:       make(map[string]*peer.Peer),
			HandleConnected:    self.handleConnected,
			HandleDisconnected: self.handleDisconnected,
			HandleFailed:       self.handleFailed,
		}

		if pubKey != EmptyString {
			peer.PublicKey = pubKey
		}

		listen.Host.Peerstore().AddAddr(peer.PeerID, peer.TargetAddress, pstore.PermanentAddrTTL)
		Logger.log.Info("DEBUG Connect to RemotePeer", peer.PublicKey)
		Logger.log.Info(listen.Host.Peerstore().Addrs(peer.PeerID))
		listen.PushConn(&peer, nil)
	}
}

func (self *ConnManager) Start(discoverPeerAddress string) {
	// Already started?
	if atomic.AddInt32(&self.start, 1) != 1 {
		return
	}

	Logger.log.Info("Connection manager started")
	// Start handler to listent channel from connection peer
	//go self.connHandler()

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if self.Config.OnInboundAccept != nil {
		for _, listner := range self.Config.ListenerPeers {
			listner.HandleConnected = self.handleConnected
			listner.HandleDisconnected = self.handleDisconnected
			listner.HandleFailed = self.handleFailed
			go self.listenHandler(listner)

			self.ListeningPeers[listner.PeerID.Pretty()] = listner
		}

		if self.Config.DiscoverPeers && self.Config.DiscoverPeersAddress != EmptyString {
			Logger.log.Infof("DiscoverPeers: true\n----------------------------------------------------------------\n|               Discover peer url: %s               |\n----------------------------------------------------------------", self.Config.DiscoverPeersAddress)
			go self.DiscoverPeers(discoverPeerAddress)
		}
	}
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (self *ConnManager) listenHandler(listen *peer.Peer) {
	listen.Start()
}

func (self *ConnManager) handleConnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.Pretty())
	if peerConn.GetIsOutbound() {
		Logger.log.Infof("handleConnected OUTBOUND %s", peerConn.RemotePeerID.Pretty())

		if self.Config.OnOutboundConnection != nil {
			self.Config.OnOutboundConnection(peerConn)
		}

	} else {
		Logger.log.Infof("handleConnected INBOUND %s", peerConn.RemotePeerID.Pretty())
	}
}

func (self *ConnManager) handleDisconnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.RemotePeerID.Pretty())
}

func (self *ConnManager) handleFailed(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.RemotePeerID.Pretty())
}

func (self *ConnManager) DiscoverPeers(discoverPeerAddress string) {
	Logger.log.Infof("Start Discover Peers : %s", discoverPeerAddress)
	self.randShards = self.makeRandShards(SHARD_NUMBER)
	self.discoverPeerAddress = discoverPeerAddress
	self.cDiscoveredPeers = make(chan struct{})
	for {
		self.processDiscoverPeers()
		select {
		case <-self.cDiscoveredPeers:
			Logger.log.Info("Stop Discover Peers")
			return
		case <-time.NewTimer(60 * time.Second).C:
			continue
		}
	}
}

func (self *ConnManager) processDiscoverPeers() {
	discoverPeerAddress := self.discoverPeerAddress
	if discoverPeerAddress == "" {
		return
	}
	client, err := rpc.Dial("tcp", discoverPeerAddress)
	if err != nil {
		Logger.log.Error("[Exchange Peers] re-connect:")
		Logger.log.Error(err)
		return
	}
	if client != nil {
		for _, listener := range self.Config.ListenerPeers {
			var response []wire.RawPeer

			var pbkB58 string
			signDataB58 := ""
			nowNano := time.Now().UnixNano()
			if listener.Config.UserKeySet != nil {
				pbkB58 = listener.Config.UserKeySet.GetPublicKeyB58()
				Logger.log.Info("Start Process Discover Peers", pbkB58)
				// sign data
				signDataB58, err = listener.Config.UserKeySet.SignDataB58(common.Int64ToBytes(nowNano))
				if err != nil {
					Logger.log.Error(err)
				}
			}

			externalAddress := self.Config.ExternalAddress
			Logger.log.Info("Start Process Discover Peers ExternalAddress", externalAddress)

			// remove later
			rawAddress := listener.RawAddress
			if externalAddress == EmptyString {
				externalAddress = os.Getenv("EXTERNAL_ADDRESS")
			}
			if externalAddress != EmptyString {
				host, _, err := net.SplitHostPort(externalAddress)
				if err == nil && host != EmptyString {
					rawAddress = strings.Replace(rawAddress, "127.0.0.1", host, 1)
				}
			} else {
				rawAddress = ""
			}

			args := &server.PingArgs{
				RawAddress: rawAddress,
				PublicKey:  pbkB58,
				SignData:   signDataB58,
				Timestamp:  nowNano,
			}
			Logger.log.Infof("[Exchange Peers] Ping %+v", args)

			Logger.log.Info("Dump PeerConns", len(listener.PeerConns))
			for pubK, info := range self.discoveredPeers {
				var result []string
				for _, peerConn := range listener.PeerConns {
					if peerConn.RemotePeer.PublicKey == pubK {
						result = append(result, peerConn.RemotePeer.PeerID.Pretty())
					}
				}
				Logger.log.Infof("Public PubKey %s, %s, %s", pubK, info.PeerID.Pretty(), result)
			}

			for _, peerConn := range listener.PeerConns {
				Logger.log.Info("PeerConn state %s %s %s", peerConn.ConnState(), peerConn.GetIsOutbound(), peerConn.RemotePeerID.Pretty(), peerConn.RemotePeer.RawAddress)
			}

			err := client.Call("Handler.Ping", args, &response)
			if err != nil {
				Logger.log.Error("[Exchange Peers] Ping:")
				Logger.log.Error(err)
				client = nil
				return
			}
			// make models
			mPeers := make(map[string]*wire.RawPeer)
			for _, rawPeer := range response {
				p := rawPeer
				mPeers[rawPeer.PublicKey] = &p
				fmt.Println(p)
			}
			// connect to beacon peers
			self.handleRandPeersOfBeacon(self.Config.MaxPeersBeacon, mPeers)
			// connect to same shard peers
			self.handleRandPeersOfShard(self.Config.ConsensusState.CurrentShard, self.Config.MaxPeersSameShard, mPeers)
			// connect to other shard peers
			self.handleRandPeersOfOtherShard(self.Config.ConsensusState.CurrentShard, self.Config.MaxPeersOtherShard, self.Config.MaxPeersOther, mPeers)

		}
	}
}

func (self *ConnManager) getPeerIdsFromPbk(pbk string) []libpeer.ID {
	result := make([]libpeer.ID, 0)
	for _, listener := range self.Config.ListenerPeers {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			// Logger.log.Info("Test PeerConn", peerConn.RemotePeer.PaymentAddress)
			if peerConn.RemotePeer.PublicKey == pbk {
				exist := false
				for _, item := range result {
					if item.Pretty() == peerConn.RemotePeer.PeerID.Pretty() {
						exist = true
					}
				}
				if !exist {
					result = append(result, peerConn.RemotePeer.PeerID)
				}
			}
		}
	}
	return result
}

func (self *ConnManager) getPeerConnOfShard(shard *byte) []*peer.PeerConn {
	c := make([]*peer.PeerConn, 0)
	for _, listener := range self.Config.ListenerPeers {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			sh := self.getShardOfPbk(peerConn.RemotePeer.PublicKey)
			if (shard == nil && sh == nil) || (sh != nil && shard != nil && *sh == *shard) {
				c = append(c, peerConn)
			}
		}
	}
	return c
}

func (self *ConnManager) countPeerConnOfShard(shard *byte) int {
	c := 0
	for _, listener := range self.Config.ListenerPeers {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			sh := self.getShardOfPbk(peerConn.RemotePeer.PublicKey)
			if (shard == nil && sh == nil) || (sh != nil && shard != nil && *sh == *shard) {
				c++
			}
		}
	}
	return c
}

func (self *ConnManager) checkPeerConnOfPbk(pbk string) bool {
	for _, listener := range self.Config.ListenerPeers {
		pcs := listener.GetPeerConnOfAll()
		for _, peerConn := range pcs {
			if peerConn.RemotePeer.PublicKey == pbk {
				return true
			}
		}
	}
	return false
}

func (self *ConnManager) checkBeaconOfPbk(pbk string) bool {
	beaconCommittee := self.Config.ConsensusState.GetBeaconCommittee()
	if pbk != "" && common.IndexOfStr(pbk, beaconCommittee) >= 0 {
		return true
	}
	return false
}

func (self *ConnManager) closePeerConnOfShard(shard byte) {
	cPeers := self.getPeerConnOfShard(&shard)
	for _, p := range cPeers {
		p.ForceClose()
	}
}

func (self *ConnManager) handleRandPeersOfShard(shard *byte, maxPeers int, mPeers map[string]*wire.RawPeer) int {
	if shard == nil {
		return 0
	}
	//Logger.log.Info("handleRandPeersOfShard", *shard)
	countPeerShard := self.countPeerConnOfShard(shard)
	if countPeerShard >= maxPeers {
		// close if over max conn
		if countPeerShard > maxPeers {
			cPeers := self.getPeerConnOfShard(shard)
			lPeers := len(cPeers)
			for idx := maxPeers; idx < lPeers; idx++ {
				cPeers[idx].ForceClose()
			}
		}
		return maxPeers
	}
	pBKs := self.Config.ConsensusState.GetShardCommittee(*shard)
	for len(pBKs) > 0 {
		randN := common.RandInt() % len(pBKs)
		pbk := pBKs[randN]
		pBKs = append(pBKs[:randN], pBKs[randN+1:]...)
		peerI, ok := mPeers[pbk]
		if ok {
			cPbk := self.Config.ConsensusState.UserPbk
			// if existed conn then not append to array
			if cPbk != pbk && !self.checkPeerConnOfPbk(pbk) {
				go self.Connect(peerI.RawAddress, peerI.PublicKey)
				countPeerShard++
			}
			if countPeerShard >= maxPeers {
				return countPeerShard
			}
		}
	}
	return countPeerShard
}

func (self *ConnManager) handleRandPeersOfOtherShard(cShard *byte, maxShardPeers int, maxPeers int, mPeers map[string]*wire.RawPeer) int {
	//Logger.log.Info("handleRandPeersOfOtherShard", maxShardPeers, maxPeers)
	countPeers := 0
	for _, shard := range self.randShards {
		if cShard == nil || (cShard != nil && *cShard != shard) {
			if countPeers < maxPeers {
				mP := int(math.Min(float64(maxShardPeers), float64(maxPeers-countPeers)))
				cPeer := self.handleRandPeersOfShard(&shard, mP, mPeers)
				countPeers += cPeer
				if countPeers >= maxPeers {
					continue
				}
			}
			if countPeers >= maxPeers {
				self.closePeerConnOfShard(shard)
			}
		}
	}
	return countPeers
}

func (self *ConnManager) handleRandPeersOfBeacon(maxBeaconPeers int, mPeers map[string]*wire.RawPeer) int {
	Logger.log.Info("handleRandPeersOfBeacon")

	countPeerShard := 0
	pBKs := self.Config.ConsensusState.GetBeaconCommittee()
	for len(pBKs) > 0 {
		randN := common.RandInt() % len(pBKs)
		pbk := pBKs[randN]
		pBKs = append(pBKs[:randN], pBKs[randN+1:]...)
		peerI, ok := mPeers[pbk]
		if ok {
			cPbk := self.Config.ConsensusState.UserPbk
			// if existed conn then not append to array
			if cPbk != pbk && !self.checkPeerConnOfPbk(pbk) {
				go self.Connect(peerI.RawAddress, peerI.PublicKey)
			}
			countPeerShard++
			if countPeerShard >= maxBeaconPeers {
				return countPeerShard
			}
		}
	}
	return countPeerShard
}

func (self *ConnManager) makeRandShards(maxShards int) []byte {
	shardBytes := make([]byte, 0)
	for i := 0; i < SHARD_NUMBER; i++ {
		shardBytes = append(shardBytes, byte(i))
	}
	shardsRet := make([]byte, 0)
	for len(shardsRet) < maxShards && len(shardBytes) > 0 {
		randN := common.RandInt() % len(shardBytes)
		shardV := shardBytes[randN]
		shardBytes = append(shardBytes[:randN], shardBytes[randN+1:]...)
		shardsRet = append(shardsRet, shardV)
	}
	return shardsRet
}

func (self *ConnManager) CheckForAcceptConn(peerConn *peer.PeerConn) bool {
	if peerConn == nil {
		return false
	}
	// check max shard conn
	sh := self.getShardOfPbk(peerConn.RemotePeer.PublicKey)
	currentShard := self.Config.ConsensusState.CurrentShard
	if sh != nil && currentShard != nil && *sh == *currentShard {
		//	same shard
		countPeerShard := self.countPeerConnOfShard(sh)
		if countPeerShard > self.Config.MaxPeersSameShard {
			return false
		}
	} else if sh != nil {
		//	order shard
		countPeerShard := self.countPeerConnOfShard(sh)
		if countPeerShard > self.Config.MaxPeersOtherShard {
			return false
		}
	} else if sh == nil {
		// none shard
		countPeerShard := self.countPeerConnOfShard(nil)
		if countPeerShard > self.Config.MaxPeersNoShard {
			return false
		}
	}
	return true
}

func (self *ConnManager) getShardOfPbk(pbk string) *byte {
	shard, ok := self.Config.ConsensusState.Committee[pbk]
	if ok {
		return &shard
	}
	return nil
}

func (self *ConnManager) GetCurrentRoleShard() (string, *byte) {
	return self.Config.ConsensusState.Role, self.Config.ConsensusState.CurrentShard
}

func (self *ConnManager) GetPeerConnOfShard(shard byte) []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	for _, listener := range self.Config.ListenerPeers {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			shardT := self.getShardOfPbk(peerConn.RemotePeer.PublicKey)
			if shardT != nil && *shardT == shard {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

func (self *ConnManager) GetPeerConnOfBeacon() []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	for _, listener := range self.Config.ListenerPeers {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			pbk := peerConn.RemotePeer.PublicKey
			if pbk != "" && self.checkBeaconOfPbk(pbk) {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

func (self *ConnManager) GetPeerConnOfPbk(pbk string) []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	if pbk == "" {
		return peerConns
	}
	for _, listener := range self.Config.ListenerPeers {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			if pbk == peerConn.RemotePeer.PublicKey {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

func (self *ConnManager) GetPeerConnOfAll() []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	for _, listener := range self.Config.ListenerPeers {
		peerConns = append(peerConns, listener.GetPeerConnOfAll()...)
	}
	return peerConns
}
