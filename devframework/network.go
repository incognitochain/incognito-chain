package devframework

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/peerv2"
	"github.com/incognitochain/incognito-chain/wire"
)

const (
	MSG_TX = iota
	MSG_TX_PRIVACYTOKEN

	MSG_BLOCK_SHARD
	MSG_BLOCK_BEACON
	MSG_BLOCK_XSHARD

	MSG_PEER_STATE
)

type MessageListeners interface {
	onTx(p *peer.PeerConn, msg *wire.MessageTx)
	onTxPrivacyToken(p *peer.PeerConn, msg *wire.MessageTxPrivacyToken)
	onBlockShard(p *peer.PeerConn, msg *wire.MessageBlockShard)
	onBlockBeacon(p *peer.PeerConn, msg *wire.MessageBlockBeacon)
	onCrossShard(p *peer.PeerConn, msg *wire.MessageCrossShard)
	onGetBlockBeacon(p *peer.PeerConn, msg *wire.MessageGetBlockBeacon)
	onGetBlockShard(p *peer.PeerConn, msg *wire.MessageGetBlockShard)
	onGetCrossShard(p *peer.PeerConn, msg *wire.MessageGetCrossShard)
	onVersion(p *peer.PeerConn, msg *wire.MessageVersion)
	onVerAck(p *peer.PeerConn, msg *wire.MessageVerAck)
	onGetAddr(p *peer.PeerConn, msg *wire.MessageGetAddr)
	onAddr(p *peer.PeerConn, msg *wire.MessageAddr)

	//PBFT
	onBFTMsg(p *peer.PeerConn, msg wire.Message)
	onPeerState(p *peer.PeerConn, msg *wire.MessagePeerState)
}

type NetworkInterface interface {
	On(msgType int, f func(msg interface{}))
	GetBeaconBlock(from, to int) []*blockchain.BeaconBlock
	GetShardBlock(sid, from, to int) *blockchain.ShardBlock
	GetCrossShardBlock(fromsid, tosid, from, to int) *blockchain.CrossShardBlock
	SyncChain([]int)
	StopSync([]int)
	IsSyncChain(chainID int) bool
}

type HighwayConnection struct {
	config            HighwayConnectionConfig
	syncker           blockchain.Syncker
	conn              *peerv2.ConnManager
	listennerRegister map[int][]func(msg interface{})
}

type HighwayConnectionConfig struct {
	LocalIP         string
	LocalPort       int
	Version         string
	HighwayEndpoint string
	PrivateKey      string
	ConsensusEngine peerv2.ConsensusData
	syncker         blockchain.Syncker
}

func NewHighwayConnection(cfg HighwayConnectionConfig) *HighwayConnection {
	return &HighwayConnection{
		config:            cfg,
		listennerRegister: make(map[int][]func(msg interface{})),
	}
}

func (s *HighwayConnection) Connect() {
	host := peerv2.NewHost(s.config.Version, s.config.LocalIP, s.config.LocalPort, s.config.PrivateKey)
	dispatcher := &peerv2.Dispatcher{
		MessageListeners: &peerv2.MessageListeners{
			OnBlockShard:     s.onBlockShard,
			OnBlockBeacon:    s.onBlockBeacon,
			OnCrossShard:     s.onCrossShard,
			OnTx:             s.onTx,
			OnTxPrivacyToken: s.onTxPrivacyToken,
			OnVersion:        s.onVersion,
			OnGetBlockBeacon: s.onGetBlockBeacon,
			OnGetBlockShard:  s.onGetBlockShard,
			OnGetCrossShard:  s.onGetCrossShard,
			OnVerAck:         s.onVerAck,
			OnGetAddr:        s.onGetAddr,
			OnAddr:           s.onAddr,

			//mubft
			OnBFTMsg:    s.onBFTMsg,
			OnPeerState: s.onPeerState,
		},
		BC: nil,
	}

	s.conn = peerv2.NewConnManager(
		host,
		s.config.HighwayEndpoint,
		&incognitokey.CommitteePublicKey{},
		s.config.ConsensusEngine,
		dispatcher,
		"relay",
		[]byte{},
	)

	go s.conn.Start(nil)

}

//framework register function on message event
func (s *HighwayConnection) On(msgType int, f func(msg interface{})) {
	s.listennerRegister[msgType] = append(s.listennerRegister[msgType], f)
}

//implement dispatch to listenner
func (s *HighwayConnection) onTx(p *peer.PeerConn, msg *wire.MessageTx) {
	panic("implement me")
}

func (s *HighwayConnection) onTxPrivacyToken(p *peer.PeerConn, msg *wire.MessageTxPrivacyToken) {
	panic("implement me")
}

func (s *HighwayConnection) onBlockShard(p *peer.PeerConn, msg *wire.MessageBlockShard) {
	panic("implement me")
}

func (s *HighwayConnection) onBlockBeacon(p *peer.PeerConn, msg *wire.MessageBlockBeacon) {
	panic("implement me")
}

func (s *HighwayConnection) onCrossShard(p *peer.PeerConn, msg *wire.MessageCrossShard) {
	panic("implement me")
}

func (s *HighwayConnection) onGetBlockBeacon(p *peer.PeerConn, msg *wire.MessageGetBlockBeacon) {
	panic("implement me")
}

func (s *HighwayConnection) onGetBlockShard(p *peer.PeerConn, msg *wire.MessageGetBlockShard) {
	panic("implement me")
}

func (s *HighwayConnection) onGetCrossShard(p *peer.PeerConn, msg *wire.MessageGetCrossShard) {
	panic("implement me")
}

func (s *HighwayConnection) onVersion(p *peer.PeerConn, msg *wire.MessageVersion) {
	panic("implement me")
}

func (s *HighwayConnection) onVerAck(p *peer.PeerConn, msg *wire.MessageVerAck) {
	panic("implement me")
}

func (s *HighwayConnection) onGetAddr(p *peer.PeerConn, msg *wire.MessageGetAddr) {
	panic("implement me")
}

func (s *HighwayConnection) onAddr(p *peer.PeerConn, msg *wire.MessageAddr) {
	panic("implement me")
}

func (s *HighwayConnection) onBFTMsg(p *peer.PeerConn, msg wire.Message) {
	panic("implement me")
}

func (s *HighwayConnection) onPeerState(p *peer.PeerConn, msg *wire.MessagePeerState) {
	panic("implement me")
}

/*
	Framework Network interface
*/
func (s *HighwayConnection) GetBeaconBlock(from, to int) []*blockchain.BeaconBlock {
	panic("implement me")
}

func (s *HighwayConnection) GetShardBlock(sid, from, to int) *blockchain.ShardBlock {
	panic("implement me")
}

func (s *HighwayConnection) GetCrossShardBlock(fromsid, tosid, from, to int) *blockchain.CrossShardBlock {
	panic("implement me")
}

func (s *HighwayConnection) SyncChain(ints []int) {
	panic("implement me")
}

func (s *HighwayConnection) StopSync(ints []int) {
	panic("implement me")
}

func (s *HighwayConnection) IsSyncChain(chainID int) bool {
	panic("implement me")
}
