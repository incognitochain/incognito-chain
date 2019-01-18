package connmanager

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync/atomic"
	"time"
	libpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/constant/bootnode/server"
	"github.com/ninjadotorg/constant/peer"
	"github.com/ninjadotorg/constant/wire"
	"github.com/ninjadotorg/constant/common"
	"math"
)

var MAX_PEERS_SAME_SHARD = 10
var MAX_PEERS_OTHER_SHARD = 2
var MAX_PEERS_OTHER = 100
var MAX_PEERS = 200
var MAX_PEERS_NOSHARD = 100
var MAX_PEERS_BEACON = 20
var SHARD_NUMBER = 256

// ConnState represents the state of the requested connection.
type ConnState uint8

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.

type CommitteePbk struct {
	Shards map[byte][]string
	Beacon []string
	//Committee map[string]byte
}

type ConnManager struct {
	start        int32
	stop         int32
	// Discover Peers
	discoveredPeers map[string]*DiscoverPeerInfo
	// channel
	cQuit chan struct{}

	Config Config

	ListeningPeers map[libpeer.ID]*peer.Peer

	CurrentShard *byte
	OtherShards  []byte

	CommitteePbk CommitteePbk
}

type Config struct {
	ExternalAddress   string
	MaxPeerSameShard  int
	MaxPeerOtherShard int
	MaxPeerOther      int
	MaxPeerNoShard    int
	MaxPeerBeacon     int
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

	//GetCurrentShard func() *byte
	//GetPbksOfShard  func(shard byte) []string
	//GetPbksOfBeacon func() []string
	GetCurrentPbk func() string
	//GetShardByPbk   func(pbk string) *byte
}

type DiscoverPeerInfo struct {
	PublicKey  string
	RawAddress string
	PeerID     libpeer.ID
}

// Stop gracefully shuts down the connection manager.
func (connManager *ConnManager) Stop() {
	if atomic.AddInt32(&connManager.stop, 1) != 1 {
		Logger.log.Error("Connection manager already stopped")
		return
	}
	Logger.log.Warn("Stopping connection manager")

	// Stop all the listeners.  There will not be any listeners if
	// listening is disabled.
	for _, listener := range connManager.Config.ListenerPeers {
		listener.Stop()
	}

	close(connManager.cQuit)
	Logger.log.Warn("Connection manager stopped")
}

func (connManager ConnManager) New(cfg *Config) *ConnManager {
	connManager.Config = *cfg
	connManager.cQuit = make(chan struct{})
	connManager.discoveredPeers = make(map[string]*DiscoverPeerInfo)

	connManager.ListeningPeers = map[libpeer.ID]*peer.Peer{}
	// set default config
	if connManager.Config.MaxPeerSameShard <= 0 {
		connManager.Config.MaxPeerSameShard = MAX_PEERS_SAME_SHARD
	}
	if connManager.Config.MaxPeerOtherShard <= 0 {
		connManager.Config.MaxPeerOtherShard = MAX_PEERS_OTHER_SHARD
	}
	if connManager.Config.MaxPeerOther <= 0 {
		connManager.Config.MaxPeerOther = MAX_PEERS_OTHER
	}
	if connManager.Config.MaxPeerNoShard <= 0 {
		connManager.Config.MaxPeerNoShard = MAX_PEERS_NOSHARD
	}
	if connManager.Config.MaxPeerBeacon <= 0 {
		connManager.Config.MaxPeerBeacon = MAX_PEERS_BEACON
	}

	return &connManager
}

func (connManager *ConnManager) GetPeerId(addr string) string {
	ipfsAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Print(err)
		return ""
	}
	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Print(err)
		return ""
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return ""
	}
	return peerId.Pretty()
}

func (connManager *ConnManager) GetPeerIDStr(addr string) (string, error) {
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		Logger.log.Error(err)
		return "", err
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		Logger.log.Error(err)
		return "", err
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		Logger.log.Error(err)
		return "", err
	}
	return peerId.Pretty(), nil
}

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (connManager *ConnManager) Connect(addr string, pubKey string) {
	if atomic.LoadInt32(&connManager.stop) != 0 {
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

	for _, listen := range connManager.Config.ListenerPeers {
		listen.HandleConnected = connManager.handleConnected
		listen.HandleDisconnected = connManager.handleDisconnected
		listen.HandleFailed = connManager.handleFailed

		peer := peer.Peer{
			TargetAddress:      targetAddr,
			PeerID:             peerId,
			RawAddress:         addr,
			Config:             listen.Config,
			PeerConns:          make(map[string]*peer.PeerConn),
			PendingPeers:       make(map[string]*peer.Peer),
			HandleConnected:    connManager.handleConnected,
			HandleDisconnected: connManager.handleDisconnected,
			HandleFailed:       connManager.handleFailed,
		}

		if pubKey != "" {
			peer.PublicKey = pubKey
		}

		listen.Host.Peerstore().AddAddr(peer.PeerID, peer.TargetAddress, pstore.PermanentAddrTTL)
		Logger.log.Info("DEBUG Connect to RemotePeer", peer.PublicKey)
		Logger.log.Info(listen.Host.Peerstore().Addrs(peer.PeerID))
		go listen.PushConn(&peer, nil)
	}
}

func (connManager *ConnManager) Start(discoverPeerAddress string) {
	// Already started?
	if atomic.AddInt32(&connManager.start, 1) != 1 {
		return
	}

	Logger.log.Info("Connection manager started")
	// Start handler to listent channel from connection peer
	//go connManager.connHandler()

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if connManager.Config.OnInboundAccept != nil {
		for _, listner := range connManager.Config.ListenerPeers {
			listner.HandleConnected = connManager.handleConnected
			listner.HandleDisconnected = connManager.handleDisconnected
			listner.HandleFailed = connManager.handleFailed
			go connManager.listenHandler(listner)

			connManager.ListeningPeers[listner.PeerID] = listner
		}

		if connManager.Config.DiscoverPeers && connManager.Config.DiscoverPeersAddress != "" {
			Logger.log.Infof("DiscoverPeers: true\n----------------------------------------------------------------\n|               Discover peer url: %s               |\n----------------------------------------------------------------", connManager.Config.DiscoverPeersAddress)
			go connManager.DiscoverPeers(discoverPeerAddress)
		}
	}
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (connManager *ConnManager) listenHandler(listen *peer.Peer) {
	listen.Start()
}

func (connManager *ConnManager) handleConnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.Pretty())
	if peerConn.IsOutbound {
		Logger.log.Infof("handleConnected OUTBOUND %s", peerConn.RemotePeerID.Pretty())

		if connManager.Config.OnOutboundConnection != nil {
			connManager.Config.OnOutboundConnection(peerConn)
		}

	} else {
		Logger.log.Infof("handleConnected INBOUND %s", peerConn.RemotePeerID.Pretty())
	}
}

func (p *ConnManager) handleDisconnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.RemotePeerID.Pretty())
}

func (connManager *ConnManager) handleFailed(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.RemotePeerID.Pretty())
}

/*func (connManager *ConnManager) SeedFromDNS(hosts []string, seedFn func(addrs []string)) {
	addrs := []string{}
	for _, host := range hosts {
		request, err := http.NewRequest("GET", host, nil)
		if err != nil {
			Logger.log.Info(err)
			continue
		}
		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			Logger.log.Info(err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.log.Info(err)
			continue
		}
		results := map[string]interface{}{}
		err = json.Unmarshal(bodyBytes, &results)
		if err != nil {
			Logger.log.Info(err)
			continue
		}
		dataT, ok := results["data"]
		if !ok {
			continue
		}
		data, ok := dataT.([]string)
		if !ok {
			continue
		}
		for _, peer := range data {
			addrs = append(addrs, peer)
		}
	}
	seedFn(addrs)
}*/

func (connManager *ConnManager) DiscoverPeers(discoverPeerAddress string) {
	Logger.log.Info("Start Discover Peers")
	var client *rpc.Client
	var err error

	connManager.CurrentShard = connManager.GetCurrentShard()
	connManager.OtherShards = connManager.randShards(SHARD_NUMBER)

listen:
	for {
		//Logger.log.Infof("Peers", connManager.discoveredPeers)
		if client == nil {
			client, err = rpc.Dial("tcp", discoverPeerAddress)
			if err != nil {
				Logger.log.Error("[Exchange Peers] re-connect:")
				Logger.log.Error(err)
			}
		}
		if client != nil {
			for _, listener := range connManager.Config.ListenerPeers {
				var response []wire.RawPeer

				var pbkB58 string
				signDataB58 := ""
				if listener.Config.ProducerKeySet != nil {
					pbkB58 = listener.Config.ProducerKeySet.GetPublicKeyB58()
					Logger.log.Info("Start Discover Peers", pbkB58)
					// sign data
					signDataB58, err = listener.Config.ProducerKeySet.SignDataB58([]byte{common.ZeroByte})
					if err != nil {
						Logger.log.Error(err)
					}
				}
				// remove later
				rawAddress := listener.RawAddress

				externalAddress := connManager.Config.ExternalAddress
				if externalAddress == "" {
					externalAddress = os.Getenv("EXTERNAL_ADDRESS")
				}
				if externalAddress != "" {
					host, _, err := net.SplitHostPort(externalAddress)
					if err == nil && host != "" {
						rawAddress = strings.Replace(rawAddress, "127.0.0.1", host, 1)
					}
				}

				args := &server.PingArgs{
					RawAddress: rawAddress,
					PublicKey:  pbkB58,
					SignData:   signDataB58,
				}
				Logger.log.Infof("[Exchange Peers] Ping", args)

				Logger.log.Info("Dump PeerConns", len(listener.PeerConns))
				for pubK, info := range connManager.discoveredPeers {
					var result []string
					for _, peerConn := range listener.PeerConns {
						if peerConn.RemotePeer.PublicKey == pubK {
							result = append(result, peerConn.RemotePeer.PeerID.Pretty())
						}
					}
					Logger.log.Infof("Public PubKey %s, %s, %s", pubK, info.PeerID.Pretty(), result)
				}

				for _, peerConn := range listener.PeerConns {
					Logger.log.Info("PeerConn state %s %s %s", peerConn.ConnState(), peerConn.IsOutbound, peerConn.RemotePeerID.Pretty(), peerConn.RemotePeer.RawAddress)
				}

				err := client.Call("Handler.Ping", args, &response)
				if err != nil {
					Logger.log.Error("[Exchange Peers] Ping:")
					Logger.log.Error(err)
					client = nil
					time.Sleep(time.Second * 2)

					goto listen
				}
				// make models
				mPeers := make(map[string]*wire.RawPeer)
				for _, rawPeer := range response {
					p := rawPeer
					mPeers[rawPeer.PublicKey] = &p
				}
				//for _, rawPeer := range response {
				//	if rawPeer.PublicKey != "" && !strings.Contains(rawPeer.RawAddress, listener.PeerID.Pretty()) {
				//		_, exist := connManager.discoveredPeers[rawPeer.PublicKey]
				//		//Logger.log.Info("Discovered peer", rawPeer.PaymentAddress, rawPeer.RemoteRawAddress, exist)
				//		if !exist {
				//			// The following code extracts target's peer Id from the
				//			// given multiaddress
				//			ipfsaddr, err := ma.NewMultiaddr(rawPeer.RawAddress)
				//			if err != nil {
				//				Logger.log.Error(err)
				//				return
				//			}
				//
				//			pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
				//			if err != nil {
				//				Logger.log.Error(err)
				//				return
				//			}
				//
				//			peerId, err := libpeer.IDB58Decode(pid)
				//			if err != nil {
				//				Logger.log.Error(err)
				//				return
				//			}
				//
				//			connManager.discoveredPeers[rawPeer.PublicKey] = &DiscoverPeerInfo{rawPeer.PublicKey, rawPeer.RawAddress, peerId}
				//			//Logger.log.Info("Start connect to peer", rawPeer.PaymentAddress, rawPeer.RemoteRawAddress, exist)
				//			go connManager.Connect(rawPeer.RawAddress, rawPeer.PublicKey)
				//		} else {
				//			peerIds := connManager.getPeerIdsFromPublicKey(rawPeer.PublicKey)
				//			if len(peerIds) == 0 {
				//				go connManager.Connect(rawPeer.RawAddress, rawPeer.PublicKey)
				//			}
				//		}
				//	}
				//}

				// connect to beacon peers
				connManager.handleRandPeersOfBeacon(connManager.Config.MaxPeerBeacon, mPeers)
				// connect to same shard peers
				connManager.handleRandPeersOfShard(connManager.CurrentShard, connManager.Config.MaxPeerSameShard, mPeers)
				// connect to other shard peers
				connManager.handleRandPeersOfOtherShard(connManager.CurrentShard, connManager.Config.MaxPeerOtherShard, connManager.Config.MaxPeerOther, mPeers)
			}
		}
		time.Sleep(time.Second * 60)
	}
}

func (connManager *ConnManager) countPeerConnByShard(shard *byte) int {
	if shard == nil {
		return 0
	}
	c := 0
	for _, listener := range connManager.Config.ListenerPeers {
		c += listener.CountPeerConnOfShard(shard)
	}
	return c
}

func (connManager *ConnManager) checkPeerConnByPbk(pubKey string) bool {
	for _, listener := range connManager.Config.ListenerPeers {
		for _, peerConn := range listener.PeerConns {
			if peerConn.RemotePeer.PublicKey == pubKey {
				return true
			}
		}
	}
	return false
}

func (connManager *ConnManager) closePeerConnOfShard(shard byte) {
	cPeers := connManager.GetPeerConnOfShard(&shard)
	for _, p := range cPeers {
		p.ForceClose()
	}
}

func (connManager *ConnManager) GetPeerConnOfShard(shard *byte) []*peer.PeerConn {
	c := make([]*peer.PeerConn, 0)
	for _, listener := range connManager.Config.ListenerPeers {
		cT := listener.GetPeerConnOfShard(shard)
		c = append(c, cT...)
	}
	return c
}

func (connManager *ConnManager) handleRandPeersOfShard(shard *byte, maxPeers int, mPeers map[string]*wire.RawPeer) int {
	if shard == nil {
		return 0
	}
	Logger.log.Info("handleRandPeersOfShard", *shard)
	countPeerShard := connManager.countPeerConnByShard(shard)
	if countPeerShard >= maxPeers {
		// close if over max conn
		if countPeerShard > maxPeers {
			cPeers := connManager.GetPeerConnOfShard(shard)
			lPeers := len(cPeers)
			for idx := maxPeers; idx < lPeers; idx++ {
				cPeers[idx].ForceClose()
			}
		}
		return maxPeers
	}
	pBKs := connManager.GetPbksOfShard(*shard)
	for len(pBKs) > 0 {
		randN := common.RandInt() % len(pBKs)
		pbk := pBKs[randN]
		pBKs = append(pBKs[:randN], pBKs[randN+1:]...)
		peerI, ok := mPeers[pbk]
		if ok {
			cPbk := connManager.Config.GetCurrentPbk()
			// if existed conn then not append to array
			if !connManager.checkPeerConnByPbk(pbk) && (cPbk == "" || cPbk != pbk) {
				go connManager.Connect(peerI.RawAddress, peerI.PublicKey)
				countPeerShard ++
			}
			if countPeerShard >= maxPeers {
				return countPeerShard
			}
		}
	}
	return countPeerShard
}

func (connManager *ConnManager) handleRandPeersOfOtherShard(cShard *byte, maxShardPeers int, maxPeers int, mPeers map[string]*wire.RawPeer) int {
	Logger.log.Info("handleRandPeersOfOtherShard", maxShardPeers, maxPeers)
	countPeers := 0
	for _, shard := range connManager.OtherShards {
		if cShard != nil && *cShard != shard {
			if countPeers < maxPeers {
				mP := int(math.Min(float64(maxShardPeers), float64(maxPeers-countPeers)))
				cPeer := connManager.handleRandPeersOfShard(&shard, mP, mPeers)
				countPeers += cPeer
				if countPeers >= maxPeers {
					continue
				}
			}
			if countPeers >= maxPeers {
				connManager.closePeerConnOfShard(shard)
			}
		}
	}
	return countPeers
}

func (connManager *ConnManager) randShards(maxShards int) []byte {
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

func (connManager *ConnManager) handleRandPeersOfBeacon(maxBeaconPeers int, mPeers map[string]*wire.RawPeer) int {
	Logger.log.Info("handleRandPeersOfBeacon")
	countPeerShard := 0
	pBKs := connManager.GetPbksOfBeacon()
	for len(pBKs) > 0 {
		randN := common.RandInt() % len(pBKs)
		pbk := pBKs[randN]
		pBKs = append(pBKs[:randN], pBKs[randN+1:]...)
		peerI, ok := mPeers[pbk]
		if ok {
			cPbk := connManager.Config.GetCurrentPbk()
			// if existed conn then not append to array
			if !connManager.checkPeerConnByPbk(pbk) && (cPbk == "" || cPbk != pbk) {
				go connManager.Connect(peerI.RawAddress, peerI.PublicKey)
			}
			countPeerShard ++
			if countPeerShard >= maxBeaconPeers {
				return countPeerShard
			}
		}
	}
	return countPeerShard
}

func (connManager *ConnManager) CheckAcceptConn(peerConn *peer.PeerConn) bool {
	if peerConn == nil {
		return false
	}
	// check max conn
	// check max shard conn
	sh := connManager.GetShardByPbk(peerConn.RemotePeer.PublicKey)
	if sh != nil && connManager.CurrentShard != nil && *sh == *connManager.CurrentShard {
		//	same shard
		countPeerShard := connManager.countPeerConnByShard(sh)
		if countPeerShard > connManager.Config.MaxPeerSameShard {
			return false
		}
	} else if sh != nil {
		//	order shard
		countPeerShard := connManager.countPeerConnByShard(sh)
		if countPeerShard > connManager.Config.MaxPeerOtherShard {
			return false
		}
	} else if sh == nil {
		// none shard
		countPeerShard := connManager.countPeerConnByShard(sh)
		if countPeerShard > connManager.Config.MaxPeerNoShard {
			return false
		}
	}
	return true
}

func (connManager *ConnManager) GetCurrentShard() *byte {
	pbk := connManager.Config.GetCurrentPbk()
	if pbk == "" {
		return nil
	}
	return connManager.GetShardByPbk(pbk)
}
func (connManager *ConnManager) GetPbksOfShard(shard byte) []string {
	committee, ok := connManager.CommitteePbk.Shards[shard]
	if ok {
		return committee
	}
	return make([]string, 0)
}
func (connManager *ConnManager) GetPbksOfBeacon() []string {
	return connManager.CommitteePbk.Beacon
}
func (connManager *ConnManager) GetCurrentPbk() string {
	return connManager.Config.GetCurrentPbk()
}
func (connManager *ConnManager) GetShardByPbk(pbk string) *byte {
	if pbk == "" {
		return nil
	}
	for shard, committee := range connManager.CommitteePbk.Shards {
		for _, v := range committee {
			if v == pbk {
				return &shard
			}
		}
	}
	return nil
}
