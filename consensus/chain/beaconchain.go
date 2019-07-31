package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wire"
)

type BeaconChain struct {
	Node            Node
	BlockGen        *blockchain.BlkTmplGenerator
	Blockchain      *blockchain.BlockChain
	ConsensusEngine ConsensusInterface
}

func (s *BeaconChain) GetConsensusEngine() ConsensusInterface {
	return s.ConsensusEngine
}

func (s *BeaconChain) PushMessageToValidator(msg wire.Message) error {
	return s.Node.PushMessageToBeacon(msg)
}

func (s *BeaconChain) GetLastBlockTimeStamp() uint64 {
	return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
}

func (s *BeaconChain) GetBlkMinTime() time.Duration {
	return time.Second * 5

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
	return common.IndexOfStr(base58.Base58Check{}.Encode(s.Blockchain.BestState.Beacon.BestBlock.Header.ProducerAddress.Pk, common.ZeroByte), s.Blockchain.BestState.Beacon.BeaconCommittee)
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

func (s *BeaconChain) ValidatePreSignBlock(block interface{}) error {
	_ = block.(*blockchain.BeaconBlock)
	return nil
}

func (s *BeaconChain) InsertBlk(block interface{}, isValid bool) {
	if isValid {
		s.Blockchain.InsertBeaconBlock(block.(*blockchain.BeaconBlock), true)
	}
}

func (s *BeaconChain) GetActiveShardNumber() int {
	return s.Blockchain.GetActiveShardNumber()
}
