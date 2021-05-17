package mock

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	peer2 "github.com/libp2p/go-libp2p-peer"
)

type Server struct {
	BlockChain *blockchain.BlockChain
}

func (s *Server) PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error {
	return nil
}

func (s *Server) PushMessageToShard(message wire.Message, shardID byte) error {
	return nil
}

func (s *Server) PushMessageToAll(message wire.Message) error {
	return nil
}

func (s *Server) PushMessageToPeer(message wire.Message, id peer2.ID) error {
	return nil
}

func (s *Server) GetNodeRole() string {
	return ""
}
func (s *Server) EnableMining(enable bool) error {
	return nil
}
func (s *Server) IsEnableMining() bool {
	return true
}
func (s *Server) GetChainMiningStatus(chain int) string {
	return ""
}
func (s *Server) GetPublicKeyRole(publicKey string, keyType string) (int, int) {
	return -1, -1
}
func (s *Server) GetIncognitoPublicKeyRole(publicKey string) (int, bool, int) {
	return 0, true, 0
}
func (s *Server) GetMinerIncognitoPublickey(publicKey string, keyType string) []byte {
	return nil
}

func (s *Server) OnTx(p *peer.PeerConn, msg *wire.MessageTx) {
	sid := common.GetShardIDFromLastByte(msg.Transaction.GetSenderAddrLastByte())
	s.BlockChain.GetChain(int(sid)).(*blockchain.ShardChain).TxPool.GetInbox() <- msg.Transaction
}
func (s *Server) OnTxPrivacyToken(p *peer.PeerConn, msg *wire.MessageTxPrivacyToken) {
	sid := common.GetShardIDFromLastByte(msg.Transaction.GetSenderAddrLastByte())
	s.BlockChain.GetChain(int(sid)).(*blockchain.ShardChain).TxPool.GetInbox() <- msg.Transaction
}
