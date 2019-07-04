package chain

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
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

func (s *BeaconChain) GetNodePubKey() string {
	return s.Node.GetNodePubKey()
}

func (s *BeaconChain) GetLastBlockTimeStamp() uint64 {
	return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
}

func (s *BeaconChain) GetBlkMinTime() time.Duration {
	return time.Second * 5

}

func (s *BeaconChain) IsReady() bool {
	return true
	//return s.Blockchain.Synker.IsLatest(false, 0)
}

func (s *BeaconChain) GetHeight() uint64 {
	return s.Blockchain.BestState.Beacon.BestBlock.Header.Height
}

func (s *BeaconChain) GetCommitteeSize() int {
	return len(s.Blockchain.BestState.Beacon.BeaconCommittee)
}

func (s *BeaconChain) GetNodePubKeyIndex() int {
	pubkey := s.Node.GetNodePubKey()
	return common.IndexOfStr(pubkey, s.Blockchain.BestState.Beacon.BeaconCommittee)
}

func (s *BeaconChain) GetLastProposerIndex() int {
	return common.IndexOfStr(base58.Base58Check{}.Encode(s.Blockchain.BestState.Beacon.BestBlock.Header.ProducerAddress.Pk, common.ZeroByte), s.Blockchain.BestState.Beacon.BeaconCommittee)
}

func (s *BeaconChain) CreateNewBlock(round int) BlockInterface {
	userKeyset := s.Node.GetUserKeySet()
	paymentAddress := userKeyset.PaymentAddress
	newBlock, err := s.BlockGen.NewBlockBeacon(&paymentAddress, round, s.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil
	} else {
		err = s.BlockGen.FinalizeBeaconBlock(newBlock, userKeyset)
		if err != nil {
			return nil
		}
	}
	return newBlock
}

func (s *BeaconChain) ValidateBlock(interface{}) bool {
	return true
}

func (s *BeaconChain) ValidateSignature(interface{}, string) bool {
	return true
}

func (s *BeaconChain) InsertBlk(block interface{}, isValid bool) {
	if isValid {
		s.Blockchain.InsertShardBlock(block.(*blockchain.ShardBlock), true)
	}
}
