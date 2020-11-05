package devframework

import (
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
	OnTx(p *peer.PeerConn, msg *wire.MessageTx)
	OnTxPrivacyToken(p *peer.PeerConn, msg *wire.MessageTxPrivacyToken)
	OnBlockShard(p *peer.PeerConn, msg *wire.MessageBlockShard)
	OnBlockBeacon(p *peer.PeerConn, msg *wire.MessageBlockBeacon)
	OnCrossShard(p *peer.PeerConn, msg *wire.MessageCrossShard)
	OnGetBlockBeacon(p *peer.PeerConn, msg *wire.MessageGetBlockBeacon)
	OnGetBlockShard(p *peer.PeerConn, msg *wire.MessageGetBlockShard)
	OnGetCrossShard(p *peer.PeerConn, msg *wire.MessageGetCrossShard)
	OnVersion(p *peer.PeerConn, msg *wire.MessageVersion)
	OnVerAck(p *peer.PeerConn, msg *wire.MessageVerAck)
	OnGetAddr(p *peer.PeerConn, msg *wire.MessageGetAddr)
	OnAddr(p *peer.PeerConn, msg *wire.MessageAddr)

	//PBFT
	OnBFTMsg(p *peer.PeerConn, msg wire.Message)
	OnPeerState(p *peer.PeerConn, msg *wire.MessagePeerState)
}

type HighwayConnection struct {
	config            HighwayConnectionConfig
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
}

func NewHighwayConnection(cfg HighwayConnectionConfig) *HighwayConnection {
	return &HighwayConnection{
		config:            cfg,
		listennerRegister: make(map[int][]func(msg interface{})),
	}
}

func (s *HighwayConnection) ConnectHighway() {
	host := peerv2.NewHost(s.config.Version, s.config.LocalIP, s.config.LocalPort, s.config.PrivateKey)
	dispatcher := &peerv2.Dispatcher{
		MessageListeners: &peerv2.MessageListeners{
			OnBlockShard:     s.OnBlockShard,
			OnBlockBeacon:    s.OnBlockBeacon,
			OnCrossShard:     s.OnCrossShard,
			OnTx:             s.OnTx,
			OnTxPrivacyToken: s.OnTxPrivacyToken,
			OnVersion:        s.OnVersion,
			OnGetBlockBeacon: s.OnGetBlockBeacon,
			OnGetBlockShard:  s.OnGetBlockShard,
			OnGetCrossShard:  s.OnGetCrossShard,
			OnVerAck:         s.OnVerAck,
			OnGetAddr:        s.OnGetAddr,
			OnAddr:           s.OnAddr,

			//mubft
			OnBFTMsg:    s.OnBFTMsg,
			OnPeerState: s.OnPeerState,
		},
		BC: nil,
	}

	s.conn = peerv2.NewConnManager(
		host,
		s.config.HighwayEndpoint,
		&incognitokey.CommitteePublicKey{},
		s.config.ConsensusEngine,
		dispatcher,
		"relays",
		[]byte{},
	)

	go s.conn.Start(nil)

}

//register function on message event
func (s *HighwayConnection) On(msgType int, f func(msg interface{})) {
	s.listennerRegister[msgType] = append(s.listennerRegister[msgType], f)
}

//implement dispatch to listenner
func (s *HighwayConnection) OnTx(p *peer.PeerConn, msg *wire.MessageTx) {
	panic("implement me")
}

func (s *HighwayConnection) OnTxPrivacyToken(p *peer.PeerConn, msg *wire.MessageTxPrivacyToken) {
	panic("implement me")
}

func (s *HighwayConnection) OnBlockShard(p *peer.PeerConn, msg *wire.MessageBlockShard) {
	panic("implement me")
}

func (s *HighwayConnection) OnBlockBeacon(p *peer.PeerConn, msg *wire.MessageBlockBeacon) {
	panic("implement me")
}

func (s *HighwayConnection) OnCrossShard(p *peer.PeerConn, msg *wire.MessageCrossShard) {
	panic("implement me")
}

func (s *HighwayConnection) OnGetBlockBeacon(p *peer.PeerConn, msg *wire.MessageGetBlockBeacon) {
	panic("implement me")
}

func (s *HighwayConnection) OnGetBlockShard(p *peer.PeerConn, msg *wire.MessageGetBlockShard) {
	panic("implement me")
}

func (s *HighwayConnection) OnGetCrossShard(p *peer.PeerConn, msg *wire.MessageGetCrossShard) {
	panic("implement me")
}

func (s *HighwayConnection) OnVersion(p *peer.PeerConn, msg *wire.MessageVersion) {
	panic("implement me")
}

func (s *HighwayConnection) OnVerAck(p *peer.PeerConn, msg *wire.MessageVerAck) {
	panic("implement me")
}

func (s *HighwayConnection) OnGetAddr(p *peer.PeerConn, msg *wire.MessageGetAddr) {
	panic("implement me")
}

func (s *HighwayConnection) OnAddr(p *peer.PeerConn, msg *wire.MessageAddr) {
	panic("implement me")
}

func (s *HighwayConnection) OnBFTMsg(p *peer.PeerConn, msg wire.Message) {
	panic("implement me")
}

func (s *HighwayConnection) OnPeerState(p *peer.PeerConn, msg *wire.MessagePeerState) {
	panic("implement me")
}
