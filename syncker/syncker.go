package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"time"
)

type Server interface {
	GetChainParam() *blockchain.Params
	GetUserMiningState() (role string, chainID int)
	RequestBlocksViaStream(ctx context.Context, peerID string, fromSID int, _type proto.BlkType, fromBlockHeight uint64, finalBlockHeight uint64, toBlockheight uint64, toBlockHashString string) (blockCh chan common.BlockInterface, err error)
	PublishNodeState(userLayer string, shardID int) error

	FetchNextCrossShard(fromSID, toSID int, currentHeight uint64) uint64
	StoreBeaconHashConfirmCrossShardHeight(fromSID, toSID int, height uint64, beaconHash string) error
	FetchBeaconBlockConfirmCrossShardHeight(fromSID, toSID int, height uint64) (*blockchain.BeaconBlock, error)

	PushMessageGetBlockCrossShardByHash(fromShard byte, toShard byte, blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error
}

type BeaconChainInterface interface {
	Chain
	GetShardBestViewHash() map[byte]common.Hash
	GetShardBestViewHeight() map[byte]uint64
	GetCurrentCrossShardHeightToShard(toShard byte) map[byte]uint64 //must use final block
}
type Chain interface {
	GetBestViewHeight() uint64
	GetFinalViewHeight() uint64
	SetReady(bool)
	IsReady() bool
	GetBestViewHash() string
	GetFinalViewHash() string
	GetEpoch() uint64
	ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	GetCommittee() []incognitokey.CommitteePublicKey
	CurrentHeight() uint64
	InsertBlk(block common.BlockInterface) error
}

type SynckerConfig struct {
	Node       Server
	Blockchain *blockchain.BlockChain
}

type Syncker struct {
	IsEnabled             bool //0 > stop, 1: running
	config                *SynckerConfig
	BeaconSyncProcess     *BeaconSyncProcess
	S2BSyncProcess        *S2BSyncProcess
	ShardSyncProcess      map[int]*ShardSyncProcess
	CrossShardSyncProcess map[int]*CrossShardSyncProcess

	BeaconPool     *BlkPool
	ShardPool      map[int]*BlkPool
	S2BPool        *BlkPool
	CrossShardPool map[int]*CrossShardBlkPool
}

// Everytime beacon block is inserted after sync finish, we update shard committee (from beacon view)
func (s *Syncker) WatchCommitteeChange() {
	defer func() {
		time.AfterFunc(time.Second*3, s.WatchCommitteeChange)
	}()

	//check if enable
	if !s.IsEnabled || s.config == nil {
		fmt.Println("SYNCKER: enable", s.IsEnabled, s.config == nil)
		return
	}
	role, chainID := s.config.Node.GetUserMiningState()
	s.BeaconSyncProcess.Start(chainID)

	if role == common.CommitteeRole || role == common.PendingRole {
		if chainID == -1 {
			s.BeaconSyncProcess.IsCommittee = true
		} else {
			for sid, syncProc := range s.ShardSyncProcess {
				if int(sid) == chainID {
					syncProc.IsCommittee = true
					syncProc.Start()
				} else {
					syncProc.IsCommittee = false
					syncProc.Stop()
				}
			}
		}
	}

	if chainID == -1 {
		s.config.Node.PublishNodeState(common.BeaconRole, chainID)
	} else if chainID >= 0 {
		s.config.Node.PublishNodeState(common.ShardRole, chainID)
	} else {
		s.config.Node.PublishNodeState("", -2)
	}

}

func NewSyncker() *Syncker {
	s := &Syncker{
		ShardSyncProcess:      make(map[int]*ShardSyncProcess),
		ShardPool:             make(map[int]*BlkPool),
		CrossShardSyncProcess: make(map[int]*CrossShardSyncProcess),
		CrossShardPool:        make(map[int]*CrossShardBlkPool),
	}

	return s
}

func (s *Syncker) Init(config *SynckerConfig) {
	s.config = config
	//init beacon sync process
	s.BeaconSyncProcess = NewBeaconSyncProcess(s.config.Node, s.config.Blockchain.Chains["beacon"].(BeaconChainInterface))
	s.S2BSyncProcess = s.BeaconSyncProcess.S2BSyncProcess
	s.BeaconPool = s.BeaconSyncProcess.BeaconPool
	s.S2BPool = s.S2BSyncProcess.S2BPool

	//init shard sync process
	for chainName, chain := range s.config.Blockchain.Chains {
		if chainName != "beacon" {
			sid := chain.GetShardID()
			s.ShardSyncProcess[sid] = NewShardSyncProcess(sid, s.config.Node, s.config.Blockchain.Chains["beacon"].(BeaconChainInterface), chain)
			s.ShardPool[sid] = s.ShardSyncProcess[sid].ShardPool
			s.CrossShardSyncProcess[sid] = s.ShardSyncProcess[sid].CrossShardSyncProcess
			s.CrossShardPool[sid] = s.CrossShardSyncProcess[sid].CrossShardPool
		}
	}

	//watch commitee change
	go s.WatchCommitteeChange()

	//Publish node state to other peer
	go func() {
		t := time.NewTicker(time.Second * 3)
		for _ = range t.C {
			_, chainID := s.config.Node.GetUserMiningState()
			if chainID == -1 {
				_ = s.config.Node.PublishNodeState("beacon", chainID)
			}
			if chainID >= 0 {
				_ = s.config.Node.PublishNodeState("shard", chainID)
			}
		}
	}()
}

func (s *Syncker) ReceiveBlock(blk interface{}, peerID string) {
	switch blk.(type) {
	case *blockchain.BeaconBlock:
		beaconBlk := blk.(*blockchain.BeaconBlock)
		s.BeaconPool.AddBlock(beaconBlk)
		//create fake s2b pool peerstate
		s.BeaconSyncProcess.BeaconPeerStateCh <- &wire.MessagePeerState{
			Beacon: wire.ChainState{
				Timestamp: beaconBlk.Header.Timestamp,
				BlockHash: *beaconBlk.Hash(),
				Height:    beaconBlk.GetHeight(),
			},
		}

	case *blockchain.ShardBlock:
		shardBlk := blk.(*blockchain.ShardBlock)
		s.ShardPool[shardBlk.GetShardID()].AddBlock(shardBlk)

	case *blockchain.ShardToBeaconBlock:
		s2bBlk := blk.(*blockchain.ShardToBeaconBlock)
		s.S2BPool.AddBlock(s2bBlk)
		//fmt.Println("syncker AddBlock S2B", s2bBlk.Header.ShardID, s2bBlk.Header.Height)
		//create fake s2b pool peerstate
		s.S2BSyncProcess.S2BPeerStateCh <- &wire.MessagePeerState{
			SenderID:          time.Now().String(),
			ShardToBeaconPool: map[byte][]uint64{s2bBlk.Header.ShardID: []uint64{1, s2bBlk.GetHeight()}},
			Timestamp:         time.Now().Unix(),
		}
	case *blockchain.CrossShardBlock:
		csBlk := blk.(*blockchain.CrossShardBlock)
		s.CrossShardPool[int(csBlk.ToShardID)].AddBlock(csBlk)
	}

}

func (s *Syncker) ReceivePeerState(peerState *wire.MessagePeerState) {
	//b, _ := json.Marshal(peerState)
	//fmt.Println("SYNCKER: receive peer state", string(b))
	//beacon
	if peerState.Beacon.Height != 0 {
		s.BeaconSyncProcess.BeaconPeerStateCh <- peerState
	}
	//s2b
	if len(peerState.ShardToBeaconPool) != 0 {
		s.S2BSyncProcess.S2BPeerStateCh <- peerState
	}
	//shard
	for sid, _ := range peerState.Shards {
		s.ShardSyncProcess[int(sid)].ShardPeerStateCh <- peerState
	}
}

func (s *Syncker) Start() {
	s.IsEnabled = true
}

func (s *Syncker) Stop() {
	s.IsEnabled = false
	s.BeaconSyncProcess.Stop()
	for _, chain := range s.ShardSyncProcess {
		chain.Stop()
	}
}

func (s *Syncker) GetS2BBlocksForBeaconProducer() map[byte][]interface{} {
	bestViewShardHash := s.config.Blockchain.Chains["beacon"].(BeaconChainInterface).GetShardBestViewHash()
	res := make(map[byte][]interface{})

	//bypass first block
	if len(bestViewShardHash) == 0 {
		for i := 0; i < s.config.Node.GetChainParam().ActiveShards; i++ {
			bestViewShardHash[byte(i)] = common.Hash{}
		}
	}

	//fist beacon beststate dont have shard hash end => create one
	for i, v := range bestViewShardHash {
		fmt.Println("syncker: bestViewShardHash", i, v.String())
		if (&v).IsEqual(&common.Hash{}) {
			blk := *s.config.Node.GetChainParam().GenesisShardBlock
			blk.Header.ShardID = i
			v = *blk.Hash()
		}
		for _, v := range s.S2BPool.GetFinalBlockFromBlockHash(v.String()) {
			res[i] = append(res[i], v)
			fmt.Println("syncker: get block ", v.GetHeight(), v.Hash().String())
		}
	}
	//fmt.Println("syncker: GetS2BBlocksForBeaconProducer", res)
	return res
}

func (s *Syncker) GetS2BBlocksForBeaconValidator(ctx context.Context, needBlkHash []common.Hash) map[byte][]interface{} {

	return nil
}

func (s *Syncker) GetCrossShardBlocksForShardProducer(toShard byte) map[byte][]interface{} {

	return nil
}
