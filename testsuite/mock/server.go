package mock

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-core/peer"
	peer2 "github.com/libp2p/go-libp2p-peer"
)

type Server struct {
	BlockChain *blockchain.BlockChain
	TxPool     *mempool.TxPool
}

func (s *Server) PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error {
	return nil
}

func (s *Server) PushMessageToBeacon(msg wire.Message, exclusivePeerIDs map[libp2p.ID]bool) error {
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
	if config.Param().TxPoolVersion == 1 {
		sid := common.GetShardIDFromLastByte(msg.Transaction.GetSenderAddrLastByte())
		s.BlockChain.ShardChain[sid].TxPool.GetInbox() <- msg.Transaction
	} else {
		s.TxPool.MaybeAcceptTransaction(msg.Transaction, int64(s.BlockChain.BeaconChain.GetFinalViewHeight()))
	}

}
func (s *Server) OnTxPrivacyToken(p *peer.PeerConn, msg *wire.MessageTxPrivacyToken) {
	if config.Param().TxPoolVersion == 1 {
		sid := common.GetShardIDFromLastByte(msg.Transaction.GetSenderAddrLastByte())
		s.BlockChain.ShardChain[sid].TxPool.GetInbox() <- msg.Transaction
	} else {
		s.TxPool.MaybeAcceptTransaction(msg.Transaction, int64(s.BlockChain.BeaconChain.GetFinalViewHeight()))
	}
}
