package connmanager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	libpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/cash-prototype/bootnode/server"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/peer"
	"github.com/ninjadotorg/cash-prototype/wire"
)

const (
	// defaultTargetOutbound is the default number of outbound connections to
	// maintain.
	defaultTargetOutbound = uint32(8)
	defaultTargetInbound  = uint32(8)
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

	Config Config
	// Pending Connection
	Pending map[libpeer.ID]*peer.Peer

	// Connected Connection
	Connected map[libpeer.ID]*peer.Peer

	// Discover Peers
	DiscoveredPeers map[string]*DiscoverPeerInfo

	ListeningPeers map[libpeer.ID]*peer.Peer

	WaitGroup sync.WaitGroup

	// Request channel
	Requests chan interface{}
	// quit channel
	Quit chan struct{}

	FailedAttempts uint32
}

type Config struct {
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
	ListenerPeers []peer.Peer

	// OnInboundAccept is a callback that is fired when an inbound connection is accepted
	OnInboundAccept func(peerConn *peer.PeerConn)

	//OnOutboundConnection is a callback that is fired when an outbound connection is established
	OnOutboundConnection func(peerConn *peer.PeerConn)

	//OnOutboundDisconnection is a callback that is fired when an outbound connection is disconnected
	OnOutboundDisconnection func(peerConn *peer.PeerConn)

	// TargetOutbound is the number of outbound network connections to
	// maintain. Defaults to 8.
	TargetOutbound uint32
	TargetInbound  uint32

	DiscoverPeers bool
}

type DiscoverPeerInfo struct {
	PublicKey  string
	RawAddress string
	PeerId     libpeer.ID
}

// registerPending is used to register a pending connection attempt. By
// registering pending connection attempts we allow callers to cancel pending
// connection attempts before their successful or in the case they're not
// longer wanted.
//type registerPending struct {
//	connRequest *ConnReq
//	done        chan struct{}
//}
//
//// handleConnected is used to queue a successful connection.
//type handleConnected struct {
//	connRequest *ConnReq
//	Peer        peer.Peer
//}
//
//// handleDisconnected is used to remove a connection.
//type handleDisconnected struct {
//	id    uint64
//	retry bool
//}
//
//// handleFailed is used to remove a pending connection.
//type handleFailed struct {
//	c   *ConnReq
//	err error
//}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string, netType string) ([]common.SimpleAddr, error) {
	netAddrs := make([]common.SimpleAddr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, common.SimpleAddr{Net: netType + "4", Addr: addr})
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
		} else {
			netAddrs = append(netAddrs, common.SimpleAddr{Net: netType + "4", Addr: addr})
		}
	}
	return netAddrs, nil
}

// Stop gracefully shuts down the connection manager.
func (self ConnManager) Stop() {
	if atomic.AddInt32(&self.stop, 1) != 1 {
		log.Println("Connection manager already stopped")
		return
	}
	log.Println("Stop connection manager")

	// Stop all the listeners.  There will not be any listeners if
	// listening is disabled.
	for _, listener := range self.Config.ListenerPeers {
		listener.Stop()
	}

	close(self.Quit)
	log.Println("Connection manager stopped")
}

func (self ConnManager) New(cfg *Config) (*ConnManager, error) {
	if cfg.TargetOutbound == 0 {
		cfg.TargetOutbound = defaultTargetOutbound
	}
	if cfg.TargetInbound == 0 {
		cfg.TargetInbound = defaultTargetInbound
	}
	self.Config = *cfg
	self.Quit = make(chan struct{})
	self.Requests = make(chan interface{})
	self.DiscoveredPeers = make(map[string]*DiscoverPeerInfo)

	self.Pending = map[libpeer.ID]*peer.Peer{}
	self.Connected = map[libpeer.ID]*peer.Peer{}
	self.ListeningPeers = map[libpeer.ID]*peer.Peer{}

	return &self, nil
}

func (self ConnManager) GetPeerId(addr string) string {
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Print(err)
		return ""
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
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

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (self *ConnManager) Connect(addr string, pubKey string) {
	if atomic.LoadInt32(&self.stop) != 0 {
		return
	}
	// The following code extracts target's peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Print(err)
		return
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Print(err)
		return
	}

	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>

	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", libpeer.IDB58Encode(peerId)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	for _, listen := range self.Config.ListenerPeers {
		listen.MaxOutbound = int(self.Config.TargetOutbound)
		listen.MaxInbound = int(self.Config.TargetInbound)

		listen.HandleConnected = self.handleConnected
		listen.HandleDisconnected = self.handleDisconnected
		listen.HandleFailed = self.handleFailed

		peer := peer.Peer{
			TargetAddress:      targetAddr,
			PeerId:             peerId,
			RawAddress:         addr,
			Config:             listen.Config,
			PeerConns:          make(map[libpeer.ID]*peer.PeerConn),
			PendingPeers:       make(map[libpeer.ID]*peer.Peer),
			HandleConnected:    self.handleConnected,
			HandleDisconnected: self.handleDisconnected,
			HandleFailed:       self.handleFailed,
		}

		if pubKey != "" {
			peer.PublicKey = pubKey
		}

		listen.Host.Peerstore().AddAddr(peer.PeerId, peer.TargetAddress, pstore.PermanentAddrTTL)
		Logger.log.Info("DEBUG Connect to Peer")
		Logger.log.Info(listen.Host.Peerstore().Addrs(peer.PeerId))
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /peer/1.0.0 protocol

		go listen.NewPeerConnection(&peer)
	}
}

func (self *ConnManager) Start() {
	// Already started?
	if atomic.AddInt32(&self.start, 1) != 1 {
		return
	}

	Logger.log.Info("Connection manager started")
	self.WaitGroup.Add(1)
	// Start handler to listent channel from connection peer
	//go self.connHandler()

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if self.Config.OnInboundAccept != nil {
		for _, listner := range self.Config.ListenerPeers {
			self.WaitGroup.Add(1)
			listner.HandleConnected = self.handleConnected
			listner.HandleDisconnected = self.handleDisconnected
			listner.HandleFailed = self.handleFailed
			go self.listenHandler(listner)

			self.ListeningPeers[listner.PeerId] = &listner
		}

		if self.Config.DiscoverPeers {
			go self.DiscoverPeers()
		}
	}
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (self *ConnManager) listenHandler(listen peer.Peer) {
	listen.Start()
}

func (self *ConnManager) handleConnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.PeerId.String())
	if peerConn.IsOutbound {
		Logger.log.Infof("handleConnected OUTBOUND %s", peerConn.PeerId.String())

		if self.Config.OnOutboundConnection != nil {
			self.Config.OnOutboundConnection(peerConn)
		}

	} else {
		Logger.log.Infof("handleConnected INBOUND %s", peerConn.PeerId.String())
	}
}

func (p *ConnManager) handleDisconnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.PeerId.String())
}

func (self *ConnManager) handleFailed(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.PeerId.String())
}

func (self *ConnManager) SeedFromDNS(hosts []string, seedFn func(addrs []string)) {
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
}

func (self *ConnManager) DiscoverPeers() {
	Logger.log.Infof("Start Discove Peers")
	var client *rpc.Client
	var err error

listen:
	for {
		//Logger.log.Infof("Peers", self.DiscoveredPeers)
		if client == nil {
			// server bootnode 35.199.177.89:9339
			// local bootnode 127.0.0.1:9889
			client, err = rpc.Dial("tcp", "127.0.0.1:9889")
			if err != nil {
				Logger.log.Error("[Exchange Peers] re-connect:", err)
			}
		}
		if client != nil {
			for _, listener := range self.Config.ListenerPeers {
				//Logger.log.Infof("[Exchange Peers] Ping")
				var response []wire.RawPeer

				var publicKey string

				if listener.Config.SealerPrvKey != "" {
					keyPair := &cashec.KeyPair{}
					keyPair.Import(listener.Config.SealerPrvKey)
					publicKey = base64.StdEncoding.EncodeToString(keyPair.PublicKey)
				}

				Logger.log.Info("PublicKey", publicKey)

				// remove later
				rawAddress := listener.RawAddress

				externalAddress := os.Getenv("EXTERNAL_ADDRESS")
				if externalAddress != "" {
					host, _, err := net.SplitHostPort(externalAddress)
					if err == nil && host != "" {
						rawAddress = strings.Replace(rawAddress, "127.0.0.1", host, 1)
					}
				}

				args := &server.PingArgs{rawAddress, publicKey}
				Logger.log.Infof("[Exchange Peers] Ping", args)
				err := client.Call("Handler.Ping", args, &response)
				if err != nil {
					//Logger.log.Error("[Exchange Peers] Ping:", err)
					client = nil
					time.Sleep(time.Second * 2)

					goto listen
				}

				for _, rawPeer := range response {
					if rawPeer.PublicKey != "" && !strings.Contains(rawPeer.RawAddress, listener.PeerId.String()) {
						Logger.log.Info("Peer - ", rawPeer.PublicKey)
						_, exist := self.DiscoveredPeers[rawPeer.PublicKey]
						Logger.log.Info("Discovered peer", rawPeer.PublicKey, rawPeer.RawAddress, exist)
						if !exist {
							// The following code extracts target's peer ID from the
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

							self.DiscoveredPeers[rawPeer.PublicKey] = &DiscoverPeerInfo{rawPeer.PublicKey, rawPeer.RawAddress, peerId}
							//Logger.log.Info("Start connect to peer", rawPeer.PublicKey, rawPeer.RawAddress, exist)
							go self.Connect(rawPeer.RawAddress, rawPeer.PublicKey)
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 30)
	}
}

func (p *ConnManager) GetPeerConnsByPeerId(peerId libpeer.ID) []*peer.PeerConn {
	results := []*peer.PeerConn{}
	for _, listen := range p.ListeningPeers {
		for _, peerConn := range listen.PeerConns {
			if peerConn.PeerId == peerId {
				results = append(results, peerConn)
			}
		}
	}
	return results
}
