package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"sort"
)

func (blockchain *BlockChain) ValidateProposedBeaconBlock_V2(beaconProposedBlock *BeaconBlock) (err error) {
	//check prev view exit
	preViewHash := beaconProposedBlock.GetPrevHash()
	preView := blockchain.BeaconChain.GetViewByHash(preViewHash)
	if preView == nil {
		return fmt.Errorf("Cannot find view %+v", preViewHash.String())
	}

	//create block and compare content
	processState := &BeaconProcessState{
		curView:            preView.(*BeaconBestState),
		newView:            nil,
		blockchain:         blockchain,
		version:            beaconProposedBlock.Header.Version,
		proposer:           beaconProposedBlock.Header.Proposer,
		round:              1,
		startTime:          beaconProposedBlock.Header.Timestamp,
		newBlock:           NewBeaconBlock(),
		shardToBeaconBlock: make(map[byte][]*ShardToBeaconBlock),
	}

	if err = processState.PreValidateProcess(beaconProposedBlock); err != nil {
		Logger.log.Error(err)
		return err
	}
	if err := processState.BuildBody(); err != nil {
		return err
	}

	processState.newView, err = processState.curView.updateBeaconBestState(processState.newBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime, newCommitteeChange())
	if err != nil {
		return err
	}

	if err := processState.BuildHeader(); err != nil {
		return err
	}

	//TODO: postProcessing (optional strictInsert) -> check header with new view
	if err = isSameBeaconContent(beaconProposedBlock, processState.newBlock); err != nil {
		return err
	}

	return
}

func (createState *BeaconProcessState) PreValidateProcess(proposeBlock *BeaconBlock) error {
	//TODO: basic validation

	//validate block signature
	if err := createState.blockchain.config.ConsensusEngine.ValidateProducerPosition(proposeBlock, createState.curView.BeaconCommittee); err != nil {
		return err
	}

	if err := createState.blockchain.config.ConsensusEngine.ValidateProducerSig(proposeBlock, createState.blockchain.BeaconChain.GetConsensusType()); err != nil {
		return err
	}

	//get enough s2B blocks
	s2bRequired := make(map[byte][]common.Hash)
	var keys []int
	for k, shardstates := range proposeBlock.Body.ShardState {
		keys = append(keys, int(k))
		for _, state := range shardstates {
			s2bRequired[k] = append(s2bRequired[k], state.Hash)
		}
	}
	sort.Ints(keys)
	s2bBlocks, err := createState.blockchain.config.Syncker.GetS2BBlocksForBeaconValidator(createState.curView.BestShardHash, s2bRequired)
	if err != nil {
		return err
	}

	for sid, v := range s2bBlocks {
		for _, b := range v {
			createState.shardToBeaconBlock[sid] = append(createState.shardToBeaconBlock[sid], b.(*ShardToBeaconBlock))
		}
	}
	return nil
}

func isSameBeaconContent(src, target *BeaconBlock) (err error) {
	//compare header
	//Version
	//Height
	//Epoch
	//Round
	//Timestamp
	//PreviousBlockHash
	//InstructionHash
	//ShardStateHash
	//InstructionMerkleRoot
	//BeaconCommitteeAndValidatorRoot
	//BeaconCandidateRoot
	//ShardCandidateRoot
	//ShardCommitteeAndValidatorRoot
	//AutoStakingRoot
	//ConsensusType
	//Producer
	//ProducerPubKeyStr
	//Proposer    string `json:"Proposer"`
	//ProposeTime int64  `json:"ProposeTime"`

	//compare body
	//ShardState   map[byte][]ShardState
	//Instructions [][]string

	return nil
}
