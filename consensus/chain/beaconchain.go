package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
)

type BeaconChain struct {
	Node            Node
	BlockGen        *blockchain.BlockGenerator
	Blockchain      *blockchain.BlockChain
	Consensus       ConsensusInterface
	ConsensusEngine ConsensusEngineInterface
}

func (s *BeaconChain) GetConsensusEngine() ConsensusEngineInterface {
	return s.ConsensusEngine
}

func (s *BeaconChain) PushMessageToValidator(msg wire.Message) error {
	return s.Node.PushMessageToBeacon(msg, map[peer.ID]bool{})
}

func (s *BeaconChain) GetLastBlockTimeStamp() uint64 {
	return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
}

func (s *BeaconChain) GetBlkMinTime() time.Duration {
	return common.MinShardBlkInterval

}

func (s *BeaconChain) IsReady() bool {
	return s.Blockchain.Synker.IsLatest(false, 0)
}

func (s *BeaconChain) GetHeight() uint64 {
	return s.Blockchain.BestState.Beacon.BestBlock.Header.Height
}

func (s *BeaconChain) GetCommitteeSize() int {
	return len(s.Blockchain.BestState.Beacon.BeaconCommittee)
}

func (s *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	return common.IndexOfStr(pubkey, s.Blockchain.BestState.Beacon.BeaconCommittee)
}

func (s *BeaconChain) GetLastProposerIndex() int {
	return s.Blockchain.BestState.Beacon.BeaconProposerIndex
}

func (s *BeaconChain) CreateNewBlock(round int) BlockInterface {
	newBlock, err := s.BlockGen.NewBlockBeacon(round, s.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil
	}
	// err = s.BlockGen.FinalizeBeaconBlock(newBlock, userKeyset)
	// if err != nil {
	// 	return nil
	// }
	return newBlock
}

func (s *BeaconChain) ValidateBlock(block interface{}) error {
	_ = block.(*blockchain.BeaconBlock)
	return nil
}

// func (s *BeaconChain) ValidatePreSignBlock(block interface{}) error {
// 	_ = block.(*blockchain.BeaconBlock)
// 	return nil
// }

func (s *BeaconChain) InsertBlk(block interface{}, isValid bool) {
	if isValid {
		s.Blockchain.InsertBeaconBlock(block.(*blockchain.BeaconBlock), true)
	}
}

func (s *BeaconChain) GetActiveShardNumber() int {
	return s.Blockchain.GetActiveShardNumber()
}
