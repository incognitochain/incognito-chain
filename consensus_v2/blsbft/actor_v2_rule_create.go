package blsbft

import (
	"context"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"time"
)

type ICreateNewBlockRule interface {
	CreateBlock(
		b58Str string,
		block types.BlockInterface,
		committees []incognitokey.CommitteePublicKey,
		committeeViewHash common.Hash,
		isValidRePropose bool,
		consensusName string,
		blockVersion int,
		currentTime int64,
		isRePropose bool,
	) (types.BlockInterface, error)
}

type NormalCreateBlockRule struct {
	logger common.Logger
	chain  Chain
}

func (n NormalCreateBlockRule) CreateBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
	isValidRePropose bool,
	consensusName string,
	blockVersion int,
	currentTime int64,
	isRePropose bool,
) (types.BlockInterface, error) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
	defer cancel()

	if !isRePropose {
		proposerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{b58Str})
		proposerKeySet := proposerCommitteePK[0].GetMiningKeyBase58(consensusName)
		proposerKeySetIndex, proposerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, proposerKeySet, consensusName)
		newBlock, err := n.chain.CreateNewBlock(blockVersion, b58Str, 1, currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
		n.logger.Infof("CreateNewBlock, Block Height %+v, Block Hash %+v | "+
			"Producer Index %+v, Producer SubsetID %+v", newBlock.GetHeight(), newBlock.FullHashString(),
			proposerKeySetIndex, proposerKeySetSubsetID)

		return newBlock, nil
	} else {
		proposerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{b58Str})
		proposerKeySet := proposerCommitteePK[0].GetMiningKeyBase58(consensusName)
		proposerKeySetIndex, proposerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, proposerKeySet, consensusName)
		producerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{block.GetProducer()})
		producerKeySet := producerCommitteePK[0].GetMiningKeyBase58(consensusName)
		producerKeySetIndex, producerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, producerKeySet, consensusName)

		n.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v hash %+v | "+
			"Producer Index %+v, Producer SubsetID %+v | "+
			"Proposer Index %+v, Proposer SubsetID %+v ", block.GetHeight(), block.FullHashString(),
			producerKeySetIndex, producerKeySetSubsetID, proposerKeySetIndex, proposerKeySetSubsetID)
		newBlock, err := n.chain.CreateNewBlockFromOldBlock(block, b58Str, currentTime, isValidRePropose)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
		return newBlock, nil
	}
}

func NewNormalCreateBlockRule(logger common.Logger, chain Chain) *NormalCreateBlockRule {
	return &NormalCreateBlockRule{logger: logger, chain: chain}
}

type OnlyCreateBlockRule struct {
	logger common.Logger
	chain  Chain
}

func NewOnlyCreateBlockRule(logger common.Logger, chain Chain) *OnlyCreateBlockRule {
	return &OnlyCreateBlockRule{logger: logger, chain: chain}
}

func (n OnlyCreateBlockRule) CreateBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
	isValidRePropose bool,
	consensusName string,
	blockVersion int,
	currentTime int64,
	isRePropose bool,
) (types.BlockInterface, error) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
	defer cancel()

	proposerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{b58Str})
	proposerKeySet := proposerCommitteePK[0].GetMiningKeyBase58(consensusName)
	proposerKeySetIndex, proposerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, proposerKeySet, consensusName)
	newBlock, err := n.chain.CreateNewBlock(blockVersion, b58Str, 1, currentTime, committees, committeeViewHash)
	if err != nil {
		return nil, NewConsensusError(BlockCreationError, err)
	}
	n.logger.Infof("CreateNewBlock, Block Height %+v, Block Hash %+v | "+
		"Producer Index %+v, Producer SubsetID %+v", newBlock.GetHeight(), newBlock.FullHashString(),
		proposerKeySetIndex, proposerKeySetSubsetID)

	return newBlock, nil
}

type IInsertBlockRule interface {
	InsertBlock(block types.BlockInterface) error
	InsertWithPrevValidationData(types.BlockInterface, string) error
}

type InsertOnlyRule struct {
	chain  Chain
	logger common.Logger
}

func (i InsertOnlyRule) InsertWithPrevValidationData(blockInterface types.BlockInterface, s string) error {
	return i.chain.InsertWithPrevValidationData(blockInterface, s)
}

func NewInsertOnlyRule(chain Chain, logger common.Logger) *InsertOnlyRule {
	return &InsertOnlyRule{chain: chain, logger: logger}
}

func (i InsertOnlyRule) InsertBlock(block types.BlockInterface) error {
	return i.chain.InsertBlock(block, true)
}

type InsertAndBroadcastRule struct {
	chain  Chain
	logger common.Logger
}

func (i InsertAndBroadcastRule) InsertWithPrevValidationData(blockInterface types.BlockInterface, s string) error {
	return i.chain.InsertAndBroadcastBlockWithPrevValidationData(blockInterface, s)
}

func NewInsertAndBroadcastRule(chain Chain, logger common.Logger) *InsertAndBroadcastRule {
	return &InsertAndBroadcastRule{chain: chain, logger: logger}
}

func (i InsertAndBroadcastRule) InsertBlock(block types.BlockInterface) error {
	return i.chain.InsertAndBroadcastBlock(block)
}
