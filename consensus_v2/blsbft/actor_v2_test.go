package blsbft

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	mockchain "github.com/incognitochain/incognito-chain/blockchain/mocks"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	mocktypes "github.com/incognitochain/incognito-chain/blockchain/types/mocks"
	"github.com/incognitochain/incognito-chain/common"
	mockblsbft "github.com/incognitochain/incognito-chain/consensus_v2/blsbft/mocks"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	mockmultiview "github.com/incognitochain/incognito-chain/multiview/mocks"
	"github.com/stretchr/testify/mock"
)

func Test_actorV2_handleProposeMsg(t *testing.T) {

	initTestParams()
	common.TIMESLOT = 1
	hash1, _ := common.Hash{}.NewHashFromStr("123")
	hash2, _ := common.Hash{}.NewHashFromStr("456")
	logger := initLog()

	shardBlock := &types.ShardBlock{
		Header: types.ShardHeader{
			Height:                    10,
			CommitteeFromBlock:        *hash1,
			SubsetCommitteesFromBlock: *hash2,
			Producer:                  key0,
			ProposeTime:               10,
			Version:                   4,
			PreviousBlockHash:         *hash1,
		},
	}
	shardBlockData, _ := json.Marshal(shardBlock)

	errorUnmarshalChain := &mockchain.Chain{}
	errorUnmarshalChain.On("UnmarshalBlock", shardBlockData).Return(nil, errors.New("Errror"))

	errorCommitteeChain := &mockchain.Chain{}
	errorCommitteeChain.On("CommitteesFromViewHashForShard", *hash1, *hash2, byte(1), 2).
		Return([]incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, errors.New("Errror"))

	errorGetBestViewHeight := &mockchain.Chain{}
	errorGetBestViewHeight.On("UnmarshalBlock", shardBlockData).Return(shardBlock, nil)
	errorGetBestViewHeight.On("GetBestViewHeight").Return(uint64(11))
	errorGetBestViewHeight.On("IsBeaconChain").Return(false)
	errorGetBestViewHeight.On("GetShardID").Return(1)

	validCommitteeChain := &mockchain.Chain{}
	validCommitteeChain.On("CommitteesFromViewHashForShard", *hash1, *hash2, byte(1), 2).
		Return(
			[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
			[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
			nil,
		)

	node := &mockblsbft.NodeInterface{}
	node.On("RequestMissingViewViaStream",
		"1", mock.AnythingOfType("[][]uint8"), mock.AnythingOfType("int"), mock.AnythingOfType("string")).
		Return(nil)

	syncProposeViewChain := &mockchain.Chain{}
	syncProposeViewChain.On("UnmarshalBlock", shardBlockData).Return(shardBlock, nil)
	syncProposeViewChain.On("GetBestViewHeight").Return(uint64(9))
	syncProposeViewChain.On("IsBeaconChain").Return(false)
	syncProposeViewChain.On("GetShardID").Return(1)
	syncProposeViewChain.On("GetChainName").Return("shard")
	syncProposeViewChain.On("GetViewByHash", *hash1).Return(nil)

	shardBestState := &blockchain.ShardBestState{}

	normalChain := &mockchain.Chain{}
	normalChain.On("UnmarshalBlock", shardBlockData).Return(shardBlock, nil)
	normalChain.On("GetBestViewHeight").Return(uint64(9))
	normalChain.On("IsBeaconChain").Return(false)
	normalChain.On("GetShardID").Return(1)
	normalChain.On("GetViewByHash", *hash1).Return(shardBestState)

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		proposeMsg BFTPropose
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Block is nil",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  errorUnmarshalChain,
				},
			},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "1",
					Block:    shardBlockData,
					TimeSlot: 10,
				},
			},
			wantErr: true,
		},
		{
			name: "Can not get committees from block",
			fields: fields{
				actorBase: actorBase{
					chain:  errorGetBestViewHeight,
					logger: logger,
				},
				committeeChain: errorCommitteeChain,
			},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "1",
					Block:    shardBlockData,
					TimeSlot: 10,
				},
			},
			wantErr: true,
		},
		{
			name: "Receive block from old view",
			fields: fields{
				actorBase: actorBase{
					chain:  errorGetBestViewHeight,
					logger: logger,
				},
				committeeChain:       validCommitteeChain,
				receiveBlockByHash:   map[string]*ProposeBlockInfo{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
			},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "1",
					Block:    shardBlockData,
					TimeSlot: 10,
				},
			},
			wantErr: true,
		},
		{
			name: "Sync blocks to current proposed block",
			fields: fields{
				actorBase: actorBase{
					chain:  syncProposeViewChain,
					logger: logger,
					node:   node,
				},
				committeeChain:       validCommitteeChain,
				receiveBlockByHash:   map[string]*ProposeBlockInfo{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
			},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "1",
					Block:    shardBlockData,
					TimeSlot: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "Normal Work",
			fields: fields{
				actorBase: actorBase{
					chain:  normalChain,
					logger: logger,
				},
				committeeChain:       validCommitteeChain,
				receiveBlockByHash:   map[string]*ProposeBlockInfo{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
			},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "1",
					Block:    shardBlockData,
					TimeSlot: 10,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			if err := actorV2.handleProposeMsg(tt.args.proposeMsg); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.handleProposeMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_actorV2_handleVoteMsg(t *testing.T) {

	logger := initLog()
	initTestParams()

	blockHash, _ := common.Hash{}.NewHashFromStr("123456")

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		voteMsg BFTVote
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantTotalVotes int
		wantErr        bool
	}{
		{
			name: "Receive vote before receive block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				receiveBlockByHash: map[string]*ProposeBlockInfo{},
			},
			args: args{
				voteMsg: BFTVote{
					Validator: key0,
					BlockHash: blockHash.String(),
				},
			},
			wantErr:        false,
			wantTotalVotes: 1,
		},
		{
			name: "Receive wrong vote after receive block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					blockHash.String(): &ProposeBlockInfo{
						votes: map[string]*BFTVote{},
						signingCommittes: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3,
						},
					},
				},
			},
			args: args{
				voteMsg: BFTVote{
					Validator: key4,
					BlockHash: blockHash.String(),
				},
			},
			wantErr:        false,
			wantTotalVotes: 1,
		},
		{
			name: "Receive right vote after block and this node is proposer and not send vote",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					blockHash.String(): &ProposeBlockInfo{
						votes: map[string]*BFTVote{},
						signingCommittes: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3,
						},
					},
				},
			},
			args: args{
				voteMsg: BFTVote{
					Validator: key0,
					BlockHash: blockHash.String(),
				},
			},
			wantErr:        false,
			wantTotalVotes: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			err := actorV2.handleVoteMsg(tt.args.voteMsg)
			if (err != nil) != tt.wantErr {
				t.Errorf("actorV2.handleVoteMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(actorV2.receiveBlockByHash[tt.args.voteMsg.BlockHash].votes) != tt.wantTotalVotes {
					t.Errorf("actorV2.handleVoteMsg() totalVotes = %v, wantTotalVotes %v",
						len(actorV2.receiveBlockByHash[tt.args.voteMsg.BlockHash].votes), tt.wantTotalVotes)
				}
			}
		})
	}
}

func Test_actorV2_proposeBeaconBlock(t *testing.T) {
	initTestParams()
	logger := initLog()
	hash, _ := common.Hash{}.NewHashFromStr("123456")
	block := &types.BeaconBlock{}

	invalidChain := &mockchain.Chain{}
	invalidChain.On(
		"CreateNewBlock",
		4, key0, 1, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(nil, errors.New("Error"))

	invalidChain.On(
		"CreateNewBlockFromOldBlock",
		block, key0, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(nil, errors.New("Error"))

	validChain := &mockchain.Chain{}
	validChain.On(
		"CreateNewBlock",
		4, key0, 1, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(block, nil)

	validChain.On(
		"CreateNewBlockFromOldBlock",
		block, key0, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(block, nil)

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		b58Str            string
		block             types.BlockInterface
		committees        []incognitokey.CommitteePublicKey
		committeeViewHash common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.BlockInterface
		wantErr bool
	}{
		{
			name: "Invalid Create New block",
			fields: fields{
				blockVersion: 4,
				actorBase: actorBase{
					logger: logger,
					chain:  invalidChain,
				},
				currentTime: 10,
			},
			args: args{
				b58Str: key0,
				committees: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey, *incKey2, *incKey3,
				},
				committeeViewHash: *hash,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Create New Block From Old Block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  invalidChain,
				},
				currentTime: 10,
			},
			args: args{
				block:  block,
				b58Str: key0,
				committees: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey, *incKey2, *incKey3,
				},
				committeeViewHash: *hash,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create new valid block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  validChain,
				},
				currentTime:  10,
				blockVersion: 4,
			},
			args: args{
				b58Str: key0,
				committees: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey, *incKey2, *incKey3,
				},
				committeeViewHash: *hash,
			},
			want:    block,
			wantErr: false,
		},
		{
			name: "Create new valid block from old block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  validChain,
				},
				currentTime:  10,
				blockVersion: 4,
			},
			args: args{
				block:  block,
				b58Str: key0,
				committees: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey, *incKey2, *incKey3,
				},
				committeeViewHash: *hash,
			},
			want:    block,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			got, err := actorV2.proposeBeaconBlock(tt.args.b58Str, tt.args.block, tt.args.committees, tt.args.committeeViewHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("actorV2.proposeBeaconBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actorV2.proposeBeaconBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actorV2_proposeShardBlock(t *testing.T) {

	initTestParams()
	logger := initLog()
	hash, _ := common.Hash{}.NewHashFromStr("123456")
	block := &types.ShardBlock{
		Header: types.ShardHeader{
			CommitteeFromBlock:        *hash,
			SubsetCommitteesFromBlock: *hash,
		},
	}

	invalidChain := &mockchain.Chain{}
	invalidChain.On(
		"CreateNewBlock",
		4, key0, 1, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(nil, errors.New("Error"))

	invalidChain.On(
		"CreateNewBlockFromOldBlock",
		block, key0, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(nil, errors.New("Error"))
	invalidChain.On("GetShardID").Return(1)

	validChain := &mockchain.Chain{}
	validChain.On(
		"CreateNewBlock",
		4, key0, 1, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(block, nil)

	validChain.On(
		"CreateNewBlockFromOldBlock",
		block, key0, int64(10),
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3,
		},
		*hash,
	).Return(block, nil)

	invalidCommitteeChain := &mockchain.Chain{}
	invalidCommitteeChain.On("CommitteesFromViewHashForShard", *hash, *hash, byte(1), 2).Return(
		[]incognitokey.CommitteePublicKey{},
		[]incognitokey.CommitteePublicKey{},
		errors.New("Error"),
	)

	validCommitteeChain := &mockchain.Chain{}
	validCommitteeChain.On("CommitteesFromViewHashForShard", *hash, *hash, byte(1), 2).Return(
		[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
		[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
		nil,
	)
	validChain.On("GetShardID").Return(1)

	invalidChain.On("IsBeaconChain").Return(false)
	validChain.On("IsBeaconChain").Return(false)

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		b58Str            string
		block             types.BlockInterface
		committees        []incognitokey.CommitteePublicKey
		committeeViewHash common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.BlockInterface
		wantErr bool
	}{
		{
			name: "Can't get committees for current block",
			fields: fields{
				actorBase: actorBase{
					chain:  invalidChain,
					logger: logger,
				},
				committeeChain: invalidCommitteeChain,
			},
			args: args{
				b58Str:            key0,
				block:             block,
				committees:        []incognitokey.CommitteePublicKey{},
				committeeViewHash: *hash,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "CreateNewBlock invalid",
			fields: fields{
				actorBase: actorBase{
					chain:  invalidChain,
					logger: logger,
				},
				currentTime:    10,
				committeeChain: validCommitteeChain,
				blockVersion:   4,
			},
			args: args{
				b58Str:            key0,
				block:             nil,
				committees:        []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				committeeViewHash: *hash,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "CreateNewBlockFromOldBlock invalid",
			fields: fields{
				actorBase: actorBase{
					chain:  invalidChain,
					logger: logger,
				},
				committeeChain: validCommitteeChain,
				currentTime:    10,
				blockVersion:   4,
			},
			args: args{
				b58Str:            key0,
				block:             block,
				committees:        []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				committeeViewHash: *hash,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "CreateNewBlock valid",
			fields: fields{
				actorBase: actorBase{
					chain:  validChain,
					logger: logger,
				},
				committeeChain: validCommitteeChain,
				currentTime:    10,
				blockVersion:   4,
			},
			args: args{
				b58Str:            key0,
				block:             nil,
				committees:        []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				committeeViewHash: *hash,
			},
			want:    block,
			wantErr: false,
		},
		{
			name: "CreateNewBlockFromOldBlock valid",
			fields: fields{
				actorBase: actorBase{
					chain:  validChain,
					logger: logger,
				},
				committeeChain: validCommitteeChain,
				currentTime:    10,
				blockVersion:   4,
			},
			args: args{
				b58Str:            key0,
				block:             block,
				committees:        []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				committeeViewHash: *hash,
			},
			want:    block,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			got, err := actorV2.proposeShardBlock(tt.args.b58Str, tt.args.block, tt.args.committees, tt.args.committeeViewHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("actorV2.proposeShardBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actorV2.proposeShardBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actorV2_getValidProposeBlocks(t *testing.T) {
	common.TIMESLOT = 1
	tempHash, _ := common.Hash{}.NewHashFromStr("123456")
	tempView := mockmultiview.View{}
	tempView.On("GetHash").Return(tempHash)
	tempView.On("GetHeight").Return(uint64(5))

	hash, _ := common.Hash{}.NewHashFromStr("123")
	blockHeightGreaterThanValidView := mocktypes.BlockInterface{}
	blockHeightGreaterThanValidView.On("Hash").Return(hash)
	blockHashDifFromCurHash := blockHeightGreaterThanValidView

	blockHeightGreaterThanValidView.On("GetHeight").Return(uint64(7))

	blockHashDifFromCurHash.On("GetHeight").Return(uint64(5))

	validBlock := mocktypes.BlockInterface{}
	validBlock.On("Hash").Return(hash)
	validBlock.On("GetHeight").Return(uint64(6))

	blockOutOfValidateTime := validBlock
	validBlock.On("GetProposeTime").Return(int64(3))
	blockProposerTimeDifCurrTimeSlot := validBlock
	validBlock.On("GetProduceTime").Return(int64(2))

	blockProposeTimeSmallerProduceTime := mocktypes.BlockInterface{}
	blockProposeTimeSmallerProduceTime.On("Hash").Return(hash)
	blockProposeTimeSmallerProduceTime.On("GetProposeTime").Return(int64(3))
	blockProposeTimeSmallerProduceTime.On("GetHeight").Return(uint64(6))
	blockProposeTimeSmallerProduceTime.On("GetProduceTime").Return(int64(4))

	blockTimeSlotHasBeenVoted := mocktypes.BlockInterface{}
	blockTimeSlotHasBeenVoted.On("Hash").Return(hash)
	blockTimeSlotHasBeenVoted.On("GetProposeTime").Return(int64(3))
	blockTimeSlotHasBeenVoted.On("GetHeight").Return(uint64(6))
	blockTimeSlotHasBeenVoted.On("GetProduceTime").Return(int64(2))

	tempView1 := mockmultiview.View{}
	tempView1.On("GetHeight").Return(uint64(4))
	tempChain := mockchain.Chain{}
	tempChain.On("GetFinalView").Return(&tempView1)

	receiveTime := time.Now().Add(-time.Second * 3)
	lastValidateTime := time.Now().Add(-time.Second * 2)

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}

	type args struct {
		bestView multiview.View
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*ProposeBlockInfo
	}{
		{
			name: "Block is nil",
			fields: fields{
				actorBase:            actorBase{},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      1,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            nil,
						receiveTime:      time.Now(),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now().Add(time.Second * 3),
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "blockHeight is larger than validViewHeight",
			fields: fields{
				actorBase:            actorBase{},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      1,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &blockHeightGreaterThanValidView,
						receiveTime:      time.Now(),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now().Add(time.Second * 3),
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "blockHeight == currentBestViewHeight && blockHash != currentBestViewHash",
			fields: fields{
				actorBase:            actorBase{},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      1,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &blockHashDifFromCurHash,
						receiveTime:      time.Now(),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now().Add(time.Second * 3),
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "block is out of validate time",
			fields: fields{
				actorBase:            actorBase{},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      1,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &blockOutOfValidateTime,
						receiveTime:      time.Now(),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now(),
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "block proposer time is different from current time slot",
			fields: fields{
				actorBase:            actorBase{},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      4,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &blockProposerTimeDifCurrTimeSlot,
						receiveTime:      time.Now().Add(-time.Second * 3),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now().Add(-time.Second * 2),
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "Block propose time is smaller than produce time",
			fields: fields{
				actorBase:            actorBase{},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      3,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &blockProposeTimeSmallerProduceTime,
						receiveTime:      time.Now().Add(-time.Second * 3),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now().Add(-time.Second * 2),
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "Block Time Slot Has Been Voted",
			fields: fields{
				actorBase: actorBase{
					chain: &tempChain,
				},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      3,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &blockTimeSlotHasBeenVoted,
						receiveTime:      time.Now().Add(-time.Second * 3),
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: time.Now().Add(-time.Second * 2),
					},
				},
				votedTimeslot: map[int64]bool{
					3: true,
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{},
		},
		{
			name: "Valid Block",
			fields: fields{
				actorBase: actorBase{
					chain: &tempChain,
				},
				committeeChain:       nil,
				currentTime:          1,
				currentTimeSlot:      3,
				proposeHistory:       &lru.Cache{},
				receiveBlockByHeight: map[uint64][]*ProposeBlockInfo{},
				receiveBlockByHash: map[string]*ProposeBlockInfo{
					"hash": &ProposeBlockInfo{
						block:            &validBlock,
						receiveTime:      receiveTime,
						committees:       []incognitokey.CommitteePublicKey{},
						signingCommittes: []incognitokey.CommitteePublicKey{},
						userKeySet:       []signatureschemes.MiningKey{},
						votes:            map[string]*BFTVote{},
						isValid:          false,
						hasNewVote:       false,
						sendVote:         false,
						isVoted:          false,
						isCommitted:      false,
						errVotes:         2,
						validVotes:       5,
						proposerSendVote: false,
						lastValidateTime: lastValidateTime,
					},
				},
				blockVersion: 1,
			},
			args: args{
				bestView: &tempView,
			},
			want: []*ProposeBlockInfo{
				&ProposeBlockInfo{
					block:            &validBlock,
					receiveTime:      receiveTime,
					committees:       []incognitokey.CommitteePublicKey{},
					signingCommittes: []incognitokey.CommitteePublicKey{},
					userKeySet:       []signatureschemes.MiningKey{},
					votes:            map[string]*BFTVote{},
					isValid:          false,
					hasNewVote:       false,
					sendVote:         false,
					isVoted:          false,
					isCommitted:      false,
					errVotes:         2,
					validVotes:       5,
					proposerSendVote: false,
					lastValidateTime: lastValidateTime,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			got := actorV2.getValidProposeBlocks(tt.args.bestView)
			for i, v := range got {
				if !reflect.DeepEqual(*v, *tt.want[i]) {
					t.Errorf("actorV2.getValidProposeBlocks() = %v, want %v", *v, *tt.want[i])
					return
				}
			}
		})
	}
}

func Test_actorV2_validateBlock(t *testing.T) {

	logger := initLog()

	common.TIMESLOT = 1

	hash1, _ := common.Hash{}.NewHashFromStr("123")
	hash2, _ := common.Hash{}.NewHashFromStr("456")
	blockHash1, _ := common.Hash{}.NewHashFromStr("100")
	//blockHash2, _ := common.Hash{}.NewHashFromStr("200")

	lastVotedBlk := &mocktypes.BlockInterface{}
	lastVotedBlk.On("GetProduceTime").Return(int64(3))
	lastVotedBlk.On("GetProposeTime").Return(int64(3))
	lastVotedBlk.On("CommitteeFromBlock").Return(*hash1)

	view := &mockmultiview.View{}
	view.On("GetHeight").Return(uint64(5))

	//valid blocks
	blkProducerTimeSmallerThanVotedLastBlk := &mocktypes.BlockInterface{}
	blkProducerTimeSmallerThanVotedLastBlk.On("GetProduceTime").Return(int64(2))
	blkProducerTimeSmallerThanVotedLastBlk.On("GetPrevHash").Return(*hash2)
	blkProducerTimeSmallerThanVotedLastBlk.On("GetHeight").Return(uint64(6))
	blkProducerTimeSmallerThanVotedLastBlk.On("BodyHash").Return(*hash2)
	blkProducerTimeSmallerThanVotedLastBlk.On("Hash").Return(blockHash1)
	blkProducerTimeSmallerThanVotedLastBlk.On("GetPrevHash").Return(*hash2)

	blkReproposeWithLargerTimeslot := &mocktypes.BlockInterface{}
	blkReproposeWithLargerTimeslot.On("GetProduceTime").Return(int64(3))
	blkReproposeWithLargerTimeslot.On("GetProposeTime").Return(int64(4))
	blkReproposeWithLargerTimeslot.On("GetHeight").Return(uint64(6))
	blkReproposeWithLargerTimeslot.On("Hash").Return(blockHash1)
	blkReproposeWithLargerTimeslot.On("GetPrevHash").Return(*blockHash1)

	blkWithDifCommittees := &mocktypes.BlockInterface{}
	blkWithDifCommittees.On("GetProduceTime").Return(int64(4))
	blkWithDifCommittees.On("GetProposeTime").Return(int64(4))
	blkWithDifCommittees.On("CommitteeFromBlock").Return(*hash2)
	blkWithDifCommittees.On("GetPrevHash").Return(*hash2)
	blkWithDifCommittees.On("GetHeight").Return(uint64(6))
	blkWithDifCommittees.On("BodyHash").Return(*blockHash1)

	blkNormal := &mocktypes.BlockInterface{}
	blkNormal.On("GetProduceTime").Return(int64(4))
	blkNormal.On("GetProposeTime").Return(int64(4))
	blkNormal.On("GetHeight").Return(uint64(6))
	blkNormal.On("BodyHash").Return(*hash2)
	blkNormal.On("GetPrevHash").Return(*hash2)
	blkNormal.On("Hash").Return(hash2)

	//
	inValidBlock := &mocktypes.BlockInterface{}
	inValidBlock.On("GetProduceTime").Return(int64(4))
	inValidBlock.On("GetProposeTime").Return(int64(4))
	inValidBlock.On("CommitteeFromBlock").Return(*hash1)
	inValidBlock.On("GetHeight").Return(uint64(6))
	inValidBlock.On("Hash").Return(blockHash1)

	tempView := &mockmultiview.View{}
	tempChain := &mockchain.Chain{}
	tempChain.On("GetViewByHash", *blockHash1).Return(nil)
	tempChain.On("GetViewByHash", *hash2).Return(tempView)
	tempChain.On("ValidatePreSignBlock",
		blkProducerTimeSmallerThanVotedLastBlk,
		[]incognitokey.CommitteePublicKey{},
		[]incognitokey.CommitteePublicKey{}).Return(errors.New("Error"))

	tempChain.On("ValidatePreSignBlock",
		blkNormal,
		[]incognitokey.CommitteePublicKey{},
		[]incognitokey.CommitteePublicKey{}).Return(nil)

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		bestView         multiview.View
		proposeBlockInfo *ProposeBlockInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Proposetime and Producetime is not valid for voting",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block: inValidBlock,
				},
			},
			wantErr: true,
		},
		{
			name: "sendVote == true",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:    blkReproposeWithLargerTimeslot,
					sendVote: true,
				},
			},
			wantErr: true,
		},
		{
			name: "isVoted == true",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:    blkReproposeWithLargerTimeslot,
					sendVote: false,
					isVoted:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "proposeBlockInfo is valid",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:    blkReproposeWithLargerTimeslot,
					sendVote: false,
					isVoted:  false,
					isValid:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "prev hash of block is not valid",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  tempChain,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:    blkReproposeWithLargerTimeslot,
					sendVote: false,
					isVoted:  false,
					isValid:  false,
				},
			},
			wantErr: true,
		},
		{
			name: "Body block has been verified",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  tempChain,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
				bodyHashes: map[uint64]map[string]bool{
					6: map[string]bool{
						blockHash1.String(): true,
					},
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:    blkWithDifCommittees,
					sendVote: false,
					isVoted:  false,
					isValid:  false,
				},
			},
			wantErr: false,
		},
		{
			name: "Verify valid block FAIL",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  tempChain,
				},
				voteHistory: map[uint64]types.BlockInterface{
					6: lastVotedBlk,
				},
				bodyHashes: map[uint64]map[string]bool{
					6: map[string]bool{
						blockHash1.String(): true,
					},
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:            blkProducerTimeSmallerThanVotedLastBlk,
					sendVote:         false,
					isVoted:          false,
					isValid:          false,
					committees:       []incognitokey.CommitteePublicKey{},
					signingCommittes: []incognitokey.CommitteePublicKey{},
				},
			},
			wantErr: true,
		},
		{
			name: "Verify valid block SUCCESS",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
					chain:  tempChain,
				},
				voteHistory: map[uint64]types.BlockInterface{},
				bodyHashes: map[uint64]map[string]bool{
					6: map[string]bool{
						blockHash1.String(): true,
					},
				},
			},
			args: args{
				bestView: view,
				proposeBlockInfo: &ProposeBlockInfo{
					block:            blkNormal,
					sendVote:         false,
					isVoted:          false,
					isValid:          false,
					committees:       []incognitokey.CommitteePublicKey{},
					signingCommittes: []incognitokey.CommitteePublicKey{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			if err := actorV2.validateBlock(tt.args.bestView, tt.args.proposeBlockInfo); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.validateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_actorV2_processIfBlockGetEnoughVote(t *testing.T) {
	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		blockHash string
		v         *ProposeBlockInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			actorV2.processIfBlockGetEnoughVote(tt.args.blockHash, tt.args.v)
		})
	}
}

func Test_actorV2_processWithEnoughVotes(t *testing.T) {

	logger := initLog()
	common.TIMESLOT = 1
	initTestParams()
	prevHash, _ := common.Hash{}.NewHashFromStr("12345")
	validationData := consensustypes.ValidationData{
		ValidatiorsIdx: []int{1, 2, 3, 4},
	}
	validationDataStr, _ := consensustypes.EncodeValidationData(validationData)

	errShardBlock := &types.ShardBlock{
		Header: types.ShardHeader{
			PreviousBlockHash: *prevHash,
		},
		ValidationData: validationDataStr,
	}

	errReplaceValidationDataShardChain := &mockchain.Chain{}
	//errReplaceValidationDataShardChain.On("ReplacePreviousValidationData", *prevHash, )

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		v *ProposeBlockInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "Fail to insert shard block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
			},
			args: args{
				v: &ProposeBlockInfo{
					block: errShardBlock,
				},
			},
			wantErr: true,
		},
		{
			name: "Fail to insert shard block",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
			},
			args: args{
				v: &ProposeBlockInfo{
					block: errShardBlock,
				},
			},
			wantErr: true,
		},
		/*{*/
		//name:    "Valid Input beacon block",
		//fields:  fields{},
		//args:    args{},
		//wantErr: true,
		/*},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			if err := actorV2.processWithEnoughVotes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.processWithEnoughVotes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_actorV2_sendVote(t *testing.T) {

	common.TIMESLOT = 1

	initTestParams()
	logger := initLog()

	hash, _ := common.Hash{}.NewHashFromStr("123456")
	prevHash, _ := common.Hash{}.NewHashFromStr("12345")

	block := &mocktypes.BlockInterface{}
	block.On("Hash").Return(hash)
	block.On("GetPrevHash").Return(*prevHash)
	block.On("GetInstructions").Return([][]string{})
	block.On("GetHeight").Return(uint64(11))

	node := &mockblsbft.NodeInterface{}
	node.On("PushMessageToChain",
		mock.AnythingOfType("wire.Message"), mock.AnythingOfType("common.ChainInterface")).
		Return(nil)

	committees := []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3}

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		userKey    *signatureschemes2.MiningKey
		block      types.BlockInterface
		committees []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				actorBase: actorBase{
					logger:   logger,
					node:     node,
					chainKey: "shard",
				},
				voteHistory:     map[uint64]types.BlockInterface{},
				currentTime:     10,
				currentTimeSlot: 10,
			},
			args: args{
				userKey:    miningKey0,
				block:      block,
				committees: committees,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			if err := actorV2.sendVote(tt.args.userKey, tt.args.block, tt.args.committees); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.sendVote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createVote(t *testing.T) {

	initTestParams()

	hash, _ := common.Hash{}.NewHashFromStr("123456")
	prevHash, _ := common.Hash{}.NewHashFromStr("12345")

	block := &mocktypes.BlockInterface{}
	block.On("Hash").Return(hash)
	block.On("GetPrevHash").Return(*prevHash)
	block.On("GetInstructions").Return([][]string{})

	committees := []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3}

	type args struct {
		userKey    *signatureschemes2.MiningKey
		block      types.BlockInterface
		committees []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name    string
		args    args
		want    *BFTVote
		wantErr bool
	}{
		{
			name: "Valid Input",
			args: args{
				userKey:    miningKey0,
				block:      block,
				committees: committees,
			},
			want: &BFTVote{
				RoundKey:      "",
				PrevBlockHash: prevHash.String(),
				BlockHash:     hash.String(),
				Validator:     miningKey0.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus),
				IsValid:       0,
				TimeSlot:      0,
				Bls: []byte{
					134, 242, 97, 208, 116, 253, 189, 250, 248, 188, 242, 62, 204, 133, 185, 97, 233, 3, 20, 1, 164, 67, 220, 253, 146, 24, 43, 245, 156, 53, 123, 236,
				},
				Bri: []byte{},
				Confirmation: []byte{
					81, 158, 170, 152, 127, 70, 139, 153, 9, 176, 2, 160, 33, 213, 231, 172, 246, 175, 86, 131, 10, 112, 252, 42, 188, 15, 53, 38, 253, 157, 51, 173, 57, 174, 39, 68, 118, 23, 7, 51, 174, 111, 181, 209, 115, 20, 53, 105, 99, 29, 138, 202, 29, 70, 174, 86, 130, 178, 22, 247, 216, 9, 143, 94, 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createVote(tt.args.userKey, tt.args.block, tt.args.committees)
			if (err != nil) != tt.wantErr {
				t.Errorf("createVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createVote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actorV2_createBLSAggregatedSignatures(t *testing.T) {
	logger := initLog()
	common.TIMESLOT = 1
	initTestParams()
	prevHash, _ := common.Hash{}.NewHashFromStr("12345")
	validationData := consensustypes.ValidationData{
		ValidatiorsIdx: []int{1, 2, 3, 4},
	}
	validationDataStr, _ := consensustypes.EncodeValidationData(validationData)
	shardBlock := &types.ShardBlock{
		Header: types.ShardHeader{
			PreviousBlockHash: *prevHash,
		},
		ValidationData: validationDataStr,
	}

	wantValidationData := consensustypes.ValidationData{
		ProducerBLSSig: nil,
		ProducerBriSig: nil,
		ValidatiorsIdx: []int{0},
		AggSig:         []byte{134, 242, 97, 208, 116, 253, 189, 250, 248, 188, 242, 62, 204, 133, 185, 97, 233, 3, 20, 1, 164, 67, 220, 253, 146, 24, 43, 245, 156, 53, 123, 236},
		BridgeSig:      [][]byte{[]byte{}},
	}
	wantValidationDataBytes, err := json.Marshal(wantValidationData)
	if err != nil {
		panic(err)
	}

	type fields struct {
		actorBase            actorBase
		committeeChain       blockchain.Chain
		currentTime          int64
		currentTimeSlot      int64
		proposeHistory       *lru.Cache
		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
		receiveBlockByHash   map[string]*ProposeBlockInfo
		voteHistory          map[uint64]types.BlockInterface
		bodyHashes           map[uint64]map[string]bool
		votedTimeslot        map[int64]bool
		blockVersion         int
	}
	type args struct {
		committees         []incognitokey.CommitteePublicKey
		tempValidationData string
		votes              map[string]*BFTVote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				actorBase: actorBase{
					logger: logger,
				},
			},
			args: args{
				votes: map[string]*BFTVote{
					incKey0.GetMiningKeyBase58(common.BlsConsensus): &BFTVote{
						RoundKey:      "",
						PrevBlockHash: prevHash.String(),
						BlockHash:     shardBlock.Hash().String(),
						Validator:     miningKey0.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus),
						IsValid:       1,
						TimeSlot:      10,
						Bls: []byte{
							134, 242, 97, 208, 116, 253, 189, 250, 248, 188, 242, 62, 204, 133, 185, 97, 233, 3, 20, 1, 164, 67, 220, 253, 146, 24, 43, 245, 156, 53, 123, 236,
						},
						Bri: []byte{},
						Confirmation: []byte{
							81, 158, 170, 152, 127, 70, 139, 153, 9, 176, 2, 160, 33, 213, 231, 172, 246, 175, 86, 131, 10, 112, 252, 42, 188, 15, 53, 38, 253, 157, 51, 173, 57, 174, 39, 68, 118, 23, 7, 51, 174, 111, 181, 209, 115, 20, 53, 105, 99, 29, 138, 202, 29, 70, 174, 86, 130, 178, 22, 247, 216, 9, 143, 94, 1,
						},
					},
				},
				committees:         []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				tempValidationData: validationDataStr,
			},
			want:    string(wantValidationDataBytes),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorV2 := &actorV2{
				actorBase:            tt.fields.actorBase,
				committeeChain:       tt.fields.committeeChain,
				currentTime:          tt.fields.currentTime,
				currentTimeSlot:      tt.fields.currentTimeSlot,
				proposeHistory:       tt.fields.proposeHistory,
				receiveBlockByHeight: tt.fields.receiveBlockByHeight,
				receiveBlockByHash:   tt.fields.receiveBlockByHash,
				voteHistory:          tt.fields.voteHistory,
				bodyHashes:           tt.fields.bodyHashes,
				votedTimeslot:        tt.fields.votedTimeslot,
				blockVersion:         tt.fields.blockVersion,
			}
			got, err := actorV2.createBLSAggregatedSignatures(tt.args.committees, tt.args.tempValidationData, tt.args.votes)
			if (err != nil) != tt.wantErr {
				t.Errorf("actorV2.createBLSAggregatedSignatures() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("actorV2.createBLSAggregatedSignatures() = %v, want %v", got, tt.want)
			}
		})
	}
}
