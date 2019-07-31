package connmanager

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/bootnode/server"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	libpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type ConnManager struct {
	start               int32
	stop                int32
	discoverPeerAddress string
	// channel
	cQuit            chan struct{}
	cDiscoveredPeers chan struct{}

	config Config

	listeningPeer *peer.Peer

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
	ListenerPeer *peer.Peer

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

func (connManager ConnManager) GetConfig() *Config {
	return &connManager.config
}

func (connManager ConnManager) GetListeningPeer() *peer.Peer {
	return connManager.listeningPeer
}

func (connManager *ConnManager) UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string) bool {
	connManager.config.ConsensusState.Lock()
	defer connManager.config.ConsensusState.Unlock()

	bChange := false
	if connManager.config.ConsensusState.role != role {
		connManager.config.ConsensusState.role = role
		bChange = true
	}
	// checkChangeCurrentShard
	var checkChangeCurrentShard = func(consensusStateCurrentShard *byte, currentShard *byte) bool {
		return (consensusStateCurrentShard != nil && currentShard == nil) ||
			(consensusStateCurrentShard == nil && currentShard != nil) ||
			(consensusStateCurrentShard != nil && currentShard != nil && *consensusStateCurrentShard != *currentShard)

	}
	if checkChangeCurrentShard(connManager.config.ConsensusState.currentShard, currentShard) {
		connManager.config.ConsensusState.currentShard = currentShard
		bChange = true
	}
	if !common.CompareStringArray(connManager.config.ConsensusState.beaconCommittee, beaconCommittee) {
		connManager.config.ConsensusState.beaconCommittee = make([]string, len(beaconCommittee))
		copy(connManager.config.ConsensusState.beaconCommittee, beaconCommittee)
		bChange = true
	}
	if len(connManager.config.ConsensusState.committeeByShard) != len(shardCommittee) {
		for shardID, _ := range connManager.config.ConsensusState.committeeByShard {
			if _, ok := shardCommittee[shardID]; !ok {
				delete(connManager.config.ConsensusState.committeeByShard, shardID)
			}
		}
		bChange = true
	}
	if connManager.config.ConsensusState.committeeByShard == nil {
		connManager.config.ConsensusState.committeeByShard = make(map[byte][]string)
	}
	for shardID, committee := range shardCommittee {
		if _, ok := connManager.config.ConsensusState.committeeByShard[shardID]; ok {
			if !common.CompareStringArray(connManager.config.ConsensusState.committeeByShard[shardID], committee) {
				connManager.config.ConsensusState.committeeByShard[shardID] = make([]string, len(committee))
				copy(connManager.config.ConsensusState.committeeByShard[shardID], committee)
				bChange = true
			}
		} else {
			connManager.config.ConsensusState.committeeByShard[shardID] = make([]string, len(committee))
			copy(connManager.config.ConsensusState.committeeByShard[shardID], committee)
			bChange = true
		}
	}
	if connManager.config.ConsensusState.userPublicKey != userPbk {
		connManager.config.ConsensusState.userPublicKey = userPbk
		bChange = true
	}

	// update peer connection
	if bChange {
		connManager.config.ConsensusState.rebuild()
		go connManager.processDiscoverPeers()
	}

	return bChange
}

// Stop gracefully shuts down the connection manager.
func (connManager *ConnManager) Stop() error {
	if atomic.AddInt32(&connManager.stop, 1) != 1 {
		Logger.log.Error("Connection manager already stopped")
		return NewConnManagerError(StopError, errors.New("Connection manager already stopped"))
	}
	Logger.log.Warn("Stopping connection manager")

	// Stop all the listeners.  There will not be any listeners if
	// listening is disabled.
	listener := connManager.config.ListenerPeer
	if listener != nil {
		listener.Stop()
	}

	if connManager.cDiscoveredPeers != nil {
		close(connManager.cDiscoveredPeers)
	}

	if connManager.cQuit != nil {
		close(connManager.cQuit)
	}
	Logger.log.Warn("Connection manager stopped")
	return nil
}

// New - init an object connManager and return pointer to object
func New(cfg *Config) *ConnManager {
	connManager := ConnManager{
		config:           *cfg,
		cQuit:            make(chan struct{}),
		listeningPeer:    nil,
		cDiscoveredPeers: make(chan struct{}),
	}
	connManager.config.ConsensusState = &ConsensusState{}
	return &connManager
}

// GetPeerId return peer id from connection address
func (connManager *ConnManager) GetPeerId(addr string) (string, error) {
	ipfsAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		Logger.log.Error(err)
		return common.EmptyString, NewConnManagerError(GetPeerIdError, err)
	}
	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		Logger.log.Error(err)
		return common.EmptyString, NewConnManagerError(GetPeerIdError, err)
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		Logger.log.Error(err)
		return common.EmptyString, NewConnManagerError(GetPeerIdError, err)
	}
	return peerId.Pretty(), nil
}

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (connManager *ConnManager) Connect(addr string, publicKey string, cConn chan *peer.PeerConn) error {
	if atomic.LoadInt32(&connManager.stop) != 0 {
		return NewConnManagerError(ConnectError, errors.New("Can not connect because connManager is stoped"))
	}
	// The following code extracts target's peer Id from the
	// given multiaddress
	ipfsAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		Logger.log.Error(err)
		return NewConnManagerError(ConnectError, err)
	}

	// decode to a peerID from ipfs address
	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		Logger.log.Error(err)
		return NewConnManagerError(ConnectError, err)
	}
	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		Logger.log.Error(err)
		return NewConnManagerError(ConnectError, err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	// Create a Peer object
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", libpeer.IDB58Encode(peerId)))
	targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)

	listeningPeer := connManager.config.ListenerPeer
	listeningPeer.HandleConnected = connManager.handleConnected
	listeningPeer.HandleDisconnected = connManager.handleDisconnected
	listeningPeer.HandleFailed = connManager.handleFailed

	peer := peer.Peer{
		HandleConnected:    connManager.handleConnected,
		HandleDisconnected: connManager.handleDisconnected,
		HandleFailed:       connManager.handleFailed,
	}
	peer.SetPeerConns(nil)
	peer.SetPendingPeers(nil)
	peer.SetPeerID(peerId)
	peer.SetRawAddress(addr)
	peer.SetTargetAddress(targetAddr)
	peer.SetConfig(listeningPeer.GetConfig())

	// if we can get an pubbic key from params?
	if publicKey != common.EmptyString {
		// use public key to detect role in network
		peer.PublicKey = publicKey
	}

	// add remote address peer into our listening node peer
	listeningPeer.GetHost().Peerstore().AddAddr(peer.GetPeerID(), peer.GetTargetAddress(), pstore.PermanentAddrTTL)
	Logger.log.Debug("DEBUG Connect to RemotePeer", peer.PublicKey)
	Logger.log.Debug(listeningPeer.GetHost().Peerstore().Addrs(peer.GetPeerID()))
	listeningPeer.PushConn(&peer, cConn)
	return nil
}

func (connManager *ConnManager) Start(discoverPeerAddress string) error {
	// Already started?
	if atomic.AddInt32(&connManager.start, 1) != 1 {
		return NewConnManagerError(StartError, errors.New("ConnManager already started"))
	}

	Logger.log.Info("Connection manager started")

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if connManager.config.OnInboundAccept != nil {
		listenner := connManager.config.ListenerPeer
		listenner.HandleConnected = connManager.handleConnected
		listenner.HandleDisconnected = connManager.handleDisconnected
		listenner.HandleFailed = connManager.handleFailed
		go connManager.listenHandler(listenner)
		connManager.listeningPeer = listenner

		if connManager.config.DiscoverPeers && connManager.config.DiscoverPeersAddress != common.EmptyString {
			Logger.log.Debugf("DiscoverPeers: true\n----------------------------------------------------------------"+
				"\n|               Discover peer url: %s               |"+
				"\n----------------------------------------------------------------",
				connManager.config.DiscoverPeersAddress)
			go connManager.discoverPeers(discoverPeerAddress)
		}
	}
	return nil
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (connManager *ConnManager) listenHandler(listen *peer.Peer) {
	listen.Start()
}

func (connManager *ConnManager) handleConnected(peerConn *peer.PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.Pretty())
	if peerConn.GetIsOutbound() {
		Logger.log.Infof("handleConnected OUTBOUND %s", peerConn.RemotePeerID.Pretty())

		if connManager.config.OnOutboundConnection != nil {
			connManager.config.OnOutboundConnection(peerConn)
		}

	} else {
		Logger.log.Infof("handleConnected INBOUND %s", peerConn.RemotePeerID.Pretty())
	}
}

func (connManager *ConnManager) handleDisconnected(peerConn *peer.PeerConn) {
	Logger.log.Warnf("handleDisconnected %s", peerConn.RemotePeerID.Pretty())
}

func (connManager *ConnManager) handleFailed(peerConn *peer.PeerConn) {
	Logger.log.Warnf("handleFailed %s", peerConn.RemotePeerID.Pretty())
}

// DiscoverPeers - connect to bootnode
// create a rpc client to ping to bootnode
// this is a private func
func (connManager *ConnManager) discoverPeers(discoverPeerAddress string) {
	Logger.log.Infof("Start Discover Peers : %s", discoverPeerAddress)
	connManager.randShards = connManager.makeRandShards(common.MAX_SHARD_NUMBER)
	connManager.discoverPeerAddress = discoverPeerAddress
	for {
		// main process of discover peer
		// connect RPC server of boot node
		// -> get response of peers
		// -> use to make peer connection
		err := connManager.processDiscoverPeers()
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		select {
		case <-connManager.cDiscoveredPeers:
			// receive channel stop
			Logger.log.Warn("Stop Discover Peers")
			return
		case <-time.NewTimer(intervalDiscoverPeer).C:
			// every IntervalDiscoverPeer, (const = 60 second)
			// call processDiscoverPeers func to reconnect RPC server of boot node
			// and process data
			continue
		}
	}
}

// processDiscoverPeers - create a connection to
// RPC server of bootnode with golang RPC client
// after receive a response which contains data
// of peers(connectable peers) from bootnode
// conneManager should use this data to make connections with
// node peers are beacon committee
// node peers are shard commttee
// other role of other peers
func (connManager *ConnManager) processDiscoverPeers() error {
	discoverPeerAddress := connManager.discoverPeerAddress
	if discoverPeerAddress == common.EmptyString {
		// we dont have config to make discover peer
		// so we dont need to do anything here
		Logger.log.Debug("Not config discovery peer")
		return nil
	}

	// create a rpc client object,
	// connect to boot node with URL
	// get from discoverPeerAddress
	// in conf of our node
	client, err := rpc.Dial("tcp", discoverPeerAddress)
	if err != nil {
		// can not create connection to rpc server with
		// provided "discover peer address" in config
		Logger.log.Error("[Exchange Peers] re-connect:")
		Logger.log.Error(err)
		return err
	}
	if client != nil {
		defer client.Close()

		// get data about our current node peer
		listener := connManager.config.ListenerPeer
		var response []wire.RawPeer

		externalAddress := connManager.config.ExternalAddress
		Logger.log.Info("Start Process Discover Peers ExternalAddress", externalAddress)

		// remove later
		rawAddress := listener.GetRawAddress()
		rawPort := listener.GetPort()
		if externalAddress == common.EmptyString {
			externalAddress = os.Getenv("EXTERNAL_ADDRESS")
		}
		if externalAddress != common.EmptyString {
			host, port, err := net.SplitHostPort(externalAddress)
			if err == nil && host != common.EmptyString {
				rawAddress = strings.Replace(rawAddress, "127.0.0.1", host, 1)
				rawAddress = strings.Replace(rawAddress, "0.0.0.0", host, 1)
				rawAddress = strings.Replace(rawAddress, "localhost", host, 1)
				rawAddress = strings.Replace(rawAddress, fmt.Sprintf("/%s/", rawPort), fmt.Sprintf("/%s/", port), 1)
			}
		} else {
			rawAddress = common.EmptyString
		}

		// In case WE run a node look like  committee of shard or beacon
		// we need TO Generate a signature with base58check format string
		// and send to boot node like a notice from us that
		// we live and we send info about us to bootnode(peerID, node rol, ...)
		publicKeyInBase58CheckEncode := common.EmptyString
		signDataInBase58CheckEncode := common.EmptyString
		if listener.GetConfig().UserKeySet != nil {
			publicKeyInBase58CheckEncode = listener.GetConfig().UserKeySet.GetPublicKeyB58()
			Logger.log.Info("Start Process Discover Peers", publicKeyInBase58CheckEncode)
			// sign data
			signDataInBase58CheckEncode, err = listener.GetConfig().UserKeySet.SignDataB58([]byte(rawAddress))
			if err != nil {
				Logger.log.Error(err)
			}
		}

		// packing in a object PingArgs
		args := &server.PingArgs{
			RawAddress: rawAddress,
			PublicKey:  publicKeyInBase58CheckEncode,
			SignData:   signDataInBase58CheckEncode,
		}
		Logger.log.Debugf("[Exchange Peers] Ping %+v", args)

		err := client.Call("Handler.Ping", args, &response)
		if err != nil {
			// can not call method PING to rpc server of boot node
			Logger.log.Error("[Exchange Peers] Ping:")
			Logger.log.Error(err)
			client = nil
			return err
		}
		// make models
		responsePeers := make(map[string]*wire.RawPeer)
		for _, rawPeer := range response {
			p := rawPeer
			responsePeers[rawPeer.PublicKey] = &p
		}
		// connect to relay nodes
		connManager.handleRelayNode(responsePeers)
		// connect to beacon peers
		connManager.handleRandPeersOfBeacon(connManager.config.MaxPeersBeacon, responsePeers)
		// connect to same shard peers
		connManager.handleRandPeersOfShard(connManager.config.ConsensusState.currentShard, connManager.config.MaxPeersSameShard, responsePeers)
		// connect to other shard peers
		connManager.handleRandPeersOfOtherShard(connManager.config.ConsensusState.currentShard, connManager.config.MaxPeersOtherShard, connManager.config.MaxPeersOther, responsePeers)
		// connect to no shard peers
		connManager.handleRandPeersOfNoShard(connManager.config.MaxPeersNoShard, responsePeers)
	}
	return nil
}

// getPeerConnOfShard - return connection which you connect in shard
func (connManager *ConnManager) getPeerConnOfShard(shard *byte) []*peer.PeerConn {
	c := make([]*peer.PeerConn, 0)
	listener := connManager.config.ListenerPeer
	allPeers := listener.GetPeerConnOfAll()
	for _, peerConn := range allPeers {
		sh := connManager.getShardOfPublicKey(peerConn.RemotePeer.PublicKey)
		if (shard == nil && sh == nil) || (sh != nil && shard != nil && *sh == *shard) {
			c = append(c, peerConn)
		}
	}
	return c
}

// countPeerConnOfShard - count peer connection which you connect in shard
func (connManager *ConnManager) countPeerConnOfShard(shard *byte) int {
	count := 0
	listener := connManager.config.ListenerPeer
	if listener != nil {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			sh := connManager.getShardOfPublicKey(peerConn.RemotePeer.PublicKey)
			if (shard == nil && sh == nil) || (sh != nil && shard != nil && *sh == *shard) {
				count++
			}
		}
	}
	return count
}

// checkPeerConnOfPublicKey - check peer connection contain which can contain public key
func (connManager *ConnManager) checkPeerConnOfPublicKey(publicKey string) bool {
	listener := connManager.config.ListenerPeer
	if listener != nil {
		pcs := listener.GetPeerConnOfAll()
		for _, peerConn := range pcs {
			if peerConn.RemotePeer.PublicKey == publicKey {
				return true
			}
		}
	}
	return false
}

// checkBeaconOfPbk - check a public key is beacon committee?
func (connManager *ConnManager) checkBeaconOfPbk(pbk string) bool {
	bestState := blockchain.GetBestStateBeacon()
	beaconCommitteeList := bestState.BeaconCommittee
	isInBeaconCommittee := common.IndexOfStr(pbk, beaconCommitteeList) != -1
	return isInBeaconCommittee
}

// closePeerConnOfShard
func (connManager *ConnManager) closePeerConnOfShard(shard byte) {
	cPeers := connManager.getPeerConnOfShard(&shard)
	for _, p := range cPeers {
		p.ForceClose()
	}
}

func (connManager *ConnManager) handleRandPeersOfShard(shard *byte, maxPeers int, mPeers map[string]*wire.RawPeer) int {
	if shard == nil {
		return 0
	}
	//Logger.log.Info("handleRandPeersOfShard", *shard)
	countPeerShard := connManager.countPeerConnOfShard(shard)
	if countPeerShard >= maxPeers {
		// close if over max conn
		if countPeerShard > maxPeers {
			cPeers := connManager.getPeerConnOfShard(shard)
			lPeers := len(cPeers)
			for idx := maxPeers; idx < lPeers; idx++ {
				cPeers[idx].ForceClose()
			}
		}
		return maxPeers
	}
	pBKs := connManager.config.ConsensusState.getCommitteeByShard(*shard)
	for len(pBKs) > 0 {
		randN := common.RandInt() % len(pBKs)
		pbk := pBKs[randN]
		pBKs = append(pBKs[:randN], pBKs[randN+1:]...)
		if peerI, ok := mPeers[pbk]; ok {
			cPbk := connManager.config.ConsensusState.userPublicKey
			// if existed conn then not append to array
			if cPbk != pbk && !connManager.checkPeerConnOfPublicKey(pbk) {
				go connManager.Connect(peerI.RawAddress, peerI.PublicKey, nil)
				countPeerShard++
			}
			if countPeerShard >= maxPeers {
				return countPeerShard
			}
		}
	}
	return countPeerShard
}

func (connManager *ConnManager) handleRandPeersOfOtherShard(cShard *byte, maxShardPeers int, maxPeers int, mPeers map[string]*wire.RawPeer) int {
	//Logger.log.Info("handleRandPeersOfOtherShard", maxShardPeers, maxPeers)
	countPeers := 0
	for _, shard := range connManager.randShards {
		if cShard == nil || (cShard != nil && *cShard != shard) {
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

func (connManager *ConnManager) handleRandPeersOfBeacon(maxBeaconPeers int, mPeers map[string]*wire.RawPeer) int {
	Logger.log.Info("handleRandPeersOfBeacon")
	countPeerShard := 0
	pBKs := connManager.config.ConsensusState.getBeaconCommittee()
	for len(pBKs) > 0 {
		randN := common.RandInt() % len(pBKs)
		pbk := pBKs[randN]
		pBKs = append(pBKs[:randN], pBKs[randN+1:]...)
		peerI, ok := mPeers[pbk]
		if ok {
			cPbk := connManager.config.ConsensusState.userPublicKey
			// if existed conn then not append to array
			if cPbk != pbk && !connManager.checkPeerConnOfPublicKey(pbk) {
				go connManager.Connect(peerI.RawAddress, peerI.PublicKey, nil)
			}
			countPeerShard++
			if countPeerShard >= maxBeaconPeers {
				return countPeerShard
			}
		}
	}
	return countPeerShard
}

func (connManager *ConnManager) handleRandPeersOfNoShard(maxPeers int, mPeers map[string]*wire.RawPeer) int {
	countPeers := 0
	shardByCommittee := connManager.config.ConsensusState.getShardByCommittee()
	for _, peer := range mPeers {
		publicKey := peer.PublicKey
		if !connManager.checkPeerConnOfPublicKey(publicKey) {
			pBKs := connManager.config.ConsensusState.getBeaconCommittee()
			if common.IndexOfStr(publicKey, pBKs) >= 0 {
				continue
			}
			_, ok := shardByCommittee[publicKey]
			if ok {
				continue
			}
			go connManager.Connect(peer.RawAddress, peer.PublicKey, nil)
			countPeers++
			if countPeers >= maxPeers {
				return countPeers
			}
		}
	}
	return countPeers
}

func (connManager *ConnManager) makeRandShards(maxShards int) []byte {
	shardBytes := make([]byte, 0)
	for i := 0; i < common.MAX_SHARD_NUMBER; i++ {
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

// CheckForAcceptConn - return true if our connection manager can accept a new connection from new peer
func (connManager *ConnManager) CheckForAcceptConn(peerConn *peer.PeerConn) (bool, error) {
	if peerConn == nil {
		return false, NewConnManagerError(NotAcceptConnectionError, errors.New("peerConn is nil"))
	}
	// check max shard conn
	shardID := connManager.getShardOfPublicKey(peerConn.RemotePeer.PublicKey)
	currentShard := connManager.config.ConsensusState.currentShard
	if shardID != nil && currentShard != nil && *shardID == *currentShard {
		//	same shard
		countPeerShard := connManager.countPeerConnOfShard(shardID)
		if countPeerShard > connManager.config.MaxPeersSameShard {
			return false, NewConnManagerError(NotAcceptConnectionError, errors.New("same shard but countPeerShard > connManager.config.MaxPeersSameShard"))
		}
	} else if shardID != nil {
		//	other shard
		countPeerShard := connManager.countPeerConnOfShard(shardID)
		if countPeerShard > connManager.config.MaxPeersOtherShard {
			return false, NewConnManagerError(NotAcceptConnectionError, errors.New("other shard but countPeerShard > connManager.config.MaxPeersOtherShard"))
		}
	} else if shardID == nil {
		// none shard
		countPeerShard := connManager.countPeerConnOfShard(nil)
		if countPeerShard > connManager.config.MaxPeersNoShard {
			return false, NewConnManagerError(NotAcceptConnectionError, errors.New("none shard but countPeerShard > connManager.config.MaxPeersNoShard"))
		}
	}
	return true, nil
}

//getShardOfPublicKey - return shardID of public key of peer connection
func (connManager *ConnManager) getShardOfPublicKey(publicKey string) *byte {
	bestState := blockchain.GetBestStateBeacon()
	shardCommitteeList := bestState.GetShardCommittee()
	for shardID, committees := range shardCommitteeList {
		isInShardCommittee := common.IndexOfStr(publicKey, committees) != -1
		if isInShardCommittee {
			return &shardID
		}
	}
	return nil
}

// GetCurrentRoleShard - return current role in shard of connected peer
func (connManager *ConnManager) GetCurrentRoleShard() (string, *byte) {
	return connManager.config.ConsensusState.role, connManager.config.ConsensusState.currentShard
}

// GetPeerConnOfShard - return peer connection of shard
func (connManager *ConnManager) GetPeerConnOfShard(shard byte) []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	listener := connManager.config.ListenerPeer
	if listener != nil {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			shardT := connManager.getShardOfPublicKey(peerConn.RemotePeer.PublicKey)
			if shardT != nil && *shardT == shard {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

// GetPeerConnOfBeacon - return peer connection of nodes which are beacon committee
func (connManager *ConnManager) GetPeerConnOfBeacon() []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	listener := connManager.config.ListenerPeer
	if listener != nil {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			pbk := peerConn.RemotePeer.PublicKey
			if pbk != common.EmptyString && connManager.checkBeaconOfPbk(pbk) {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

// GetPeerConnOfPublicKey - return PeerConn from public key
func (connManager *ConnManager) GetPeerConnOfPublicKey(publicKey string) []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	if publicKey == common.EmptyString {
		return peerConns
	}
	listener := connManager.config.ListenerPeer
	if listener != nil {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			if publicKey == peerConn.RemotePeer.PublicKey {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

// GetPeerConnOfAll - return all Peer connection of node
func (connManager *ConnManager) GetPeerConnOfAll() []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	listener := connManager.config.ListenerPeer
	if listener != nil {
		peerConns = append(peerConns, listener.GetPeerConnOfAll()...)
	}
	return peerConns
}

// GetConnOfRelayNode - return connection of relay nodes
func (connManager *ConnManager) GetConnOfRelayNode() []*peer.PeerConn {
	peerConns := make([]*peer.PeerConn, 0)
	listener := connManager.config.ListenerPeer
	if listener != nil {
		allPeers := listener.GetPeerConnOfAll()
		for _, peerConn := range allPeers {
			pbk := peerConn.RemotePeer.PublicKey
			if pbk != common.EmptyString && common.IndexOfStr(pbk, relayNode) != -1 {
				peerConns = append(peerConns, peerConn)
			}
		}
	}
	return peerConns
}

// handleRelayNode - handle connect to relay node
func (connManager *ConnManager) handleRelayNode(mPeers map[string]*wire.RawPeer) {
	for _, p := range mPeers {
		publicKey := p.PublicKey
		if connManager.checkPeerConnOfPublicKey(publicKey) ||
			common.IndexOfStr(publicKey, relayNode) == -1 {
			continue
		}

		go connManager.Connect(p.RawAddress, p.PublicKey, nil)
	}
}
