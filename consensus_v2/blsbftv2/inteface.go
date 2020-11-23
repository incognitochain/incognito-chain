package blsbftv2

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
)

type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error)
	GetSelfPeerID() peer.ID
}

type ChainInterface interface {
	GetFinalView() multiview.View
	GetBestView() multiview.View
	GetChainName() string
	IsReady() bool
	UnmarshalBlock(blockString []byte) (types.BlockInterface, error)
	CreateNewBlock(version int, proposer string, round int, startTime int64) (types.BlockInterface, error)
	CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64) (types.BlockInterface, error)
	InsertAndBroadcastBlock(block types.BlockInterface) error
	ValidatePreSignBlock(block types.BlockInterface) error
	GetShardID() int
	GetViewByHash(hash common.Hash) multiview.View
}
