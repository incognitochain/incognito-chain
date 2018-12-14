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
)

// ConnState represents the state of the requested connection.
type ConnState uint8

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.

type ConnManager struct {
	connReqCount uint64
	start        int32
	stop         int32
	// Discover Peers
	discoveredPeers map[string]*DiscoverPeerInfo
	// channel
	cQuit chan struct{}

	Config Config

	ListeningPeers map[libpeer.ID]*peer.Peer
}

type Config struct {
	ExternalAddress string
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
}

type DiscoverPeerInfo struct {
	PublicKey  string
	RawAddress string
	PeerID     libpeer.ID
}

// Stop gracefully shuts down the connection manager.
func (self ConnManager) Stop() {
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

	close(self.cQuit)
	Logger.log.Warn("Connection manager stopped")
}

func (self ConnManager) New(cfg *Config) *ConnManager {
	self.Config = *cfg
	self.cQuit = make(chan struct{})
	self.discoveredPeers = make(map[string]*DiscoverPeerInfo)

	self.ListeningPeers = map[libpeer.ID]*peer.Peer{}

	return &self
}

func (self ConnManager) GetPeerId(addr string) string {
	ipfsAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Print(err)
		return EmptyString
	}
	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Print(err)
		return EmptyString
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return EmptyString
	}
	return peerId.Pretty()
}

func (self ConnManager) GetPeerIDStr(addr string) (string, error) {
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		Logger.log.Error(err)
		return EmptyString, err
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		Logger.log.Error(err)
		return EmptyString, err
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		Logger.log.Error(err)
		return EmptyString, err
	}
	return peerId.String(), nil
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
		go listen.PushConn(&peer, nil)
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

			self.ListeningPeers[listner.PeerID] = listner
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
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.String())
	if peerConn.IsOutbound {
		Logger.log.Infof("handleConnected OUTBOUND %s", peerConn.RemotePeerID.String())

		if self.Config.OnOutboundConnection != nil {
			self.Config.OnOutboundConnection(peerConn)
		}

	} else {
		Logger.log.Infof("handleConnected INBOUND %s", peerConn.RemotePeerID.String())
	}
}

func (p *ConnManager) handleDisconnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.RemotePeerID.String())
}

func (self *ConnManager) handleFailed(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.RemotePeerID.String())
}

/*func (self *ConnManager) SeedFromDNS(hosts []string, seedFn func(addrs []string)) {
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

func (self *ConnManager) DiscoverPeers(discoverPeerAddress string) {
	Logger.log.Info("Start Discover Peers")
	var client *rpc.Client
	var err error

listen:
	for {
		//Logger.log.Infof("Peers", self.discoveredPeers)
		if client == nil {
			client, err = rpc.Dial("tcp", discoverPeerAddress)
			if err != nil {
				Logger.log.Error("[Exchange Peers] re-connect:")
				Logger.log.Error(err)
			}
		}
		if client != nil {
			for _, listener := range self.Config.ListenerPeers {
				var response []wire.RawPeer

				var pbkB58 string
				signDataB58 := ""
				if listener.Config.ProducerKeySet != nil {
					pbkB58 = listener.Config.ProducerKeySet.GetPublicKeyB58()
					Logger.log.Info("Start Discover Peers", pbkB58)
					// sign data
					signDataB58, err = listener.Config.ProducerKeySet.SignDataB58([]byte{byte(0x00)})
					if err != nil {
						Logger.log.Error(err)
					}
					//sig, err := listener.Config.ProducerKeySet.Sign([]byte{byte(0x00)})
					//if err != nil {
					//	Logger.log.Error(err)
					//}
					//ok, err := listener.Config.ProducerKeySet.Verify([]byte{byte(0x00)}, sig)
					//if err != nil {
					//	Logger.log.Error(err)
					//}
					//if ok {
					//	Logger.log.Info("Verify OK")
					//} else {
					//	Logger.log.Info("Verify Not OK")
					//}
				}
				// remove later
				rawAddress := listener.RawAddress

				externalAddress := self.Config.ExternalAddress
				if externalAddress == EmptyString {
					externalAddress = os.Getenv("EXTERNAL_ADDRESS")
				}
				if externalAddress != EmptyString {
					host, _, err := net.SplitHostPort(externalAddress)
					if err == nil && host != EmptyString {
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
				for _, rawPeer := range response {
					if rawPeer.PublicKey != EmptyString && !strings.Contains(rawPeer.RawAddress, listener.PeerID.String()) {
						_, exist := self.discoveredPeers[rawPeer.PublicKey]
						//Logger.log.Info("Discovered peer", rawPeer.PaymentAddress, rawPeer.RemoteRawAddress, exist)
						if !exist {
							// The following code extracts target's peer Id from the
							// given multiaddress
							ipfsaddr, err := ma.NewMultiaddr(rawPeer.RawAddress)
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

							self.discoveredPeers[rawPeer.PublicKey] = &DiscoverPeerInfo{rawPeer.PublicKey, rawPeer.RawAddress, peerId}
							//Logger.log.Info("Start connect to peer", rawPeer.PaymentAddress, rawPeer.RemoteRawAddress, exist)
							go self.Connect(rawPeer.RawAddress, rawPeer.PublicKey)
						} else {
							peerIds := self.getPeerIdsFromPublicKey(rawPeer.PublicKey)
							if len(peerIds) == 0 {
								go self.Connect(rawPeer.RawAddress, rawPeer.PublicKey)
							}
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 60)
	}
}

func (self *ConnManager) getPeerIdsFromPublicKey(pubKey string) []libpeer.ID {
	result := []libpeer.ID{}

	for _, listener := range self.Config.ListenerPeers {
		for _, peerConn := range listener.PeerConns {
			// Logger.log.Info("Test PeerConn", peerConn.RemotePeer.PaymentAddress)
			if peerConn.RemotePeer.PublicKey == pubKey {
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
