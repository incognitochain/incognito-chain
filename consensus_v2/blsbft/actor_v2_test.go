package blsbft

import (
	"reflect"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	mockchain "github.com/incognitochain/incognito-chain/blockchain/mocks"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	mocktypes "github.com/incognitochain/incognito-chain/blockchain/types/mocks"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	mockmultiview "github.com/incognitochain/incognito-chain/multiview/mocks"
)

func Test_actorV2_handleProposeMsg(t *testing.T) {
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
			name:   "",
			fields: fields{},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "",
					Block:    nil,
					TimeSlot: 19,
				},
			},
			wantErr: true,
		},
		{
			name:   "",
			fields: fields{},
			args: args{
				proposeMsg: BFTPropose{
					PeerID:   "",
					Block:    nil,
					TimeSlot: 19,
				},
			},
			wantErr: true,
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
		name    string
		fields  fields
		args    args
		wantErr bool
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
			if err := actorV2.handleVoteMsg(tt.args.voteMsg); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.handleVoteMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_actorV2_proposeBlock(t *testing.T) {
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
		userMiningKey     signatureschemes2.MiningKey
		proposerPk        incognitokey.CommitteePublicKey
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
			got, err := actorV2.proposeBlock(tt.args.userMiningKey, tt.args.proposerPk, tt.args.block, tt.args.committees, tt.args.committeeViewHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("actorV2.proposeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actorV2.proposeBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actorV2_proposeBeaconBlock(t *testing.T) {
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
				votedTimeslot: map[int64]bool{},
				blockVersion:  1,
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
			want: []*ProposeBlockInfo{
				&ProposeBlockInfo{
					block:            &validBlock,
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
					t.Errorf("actorV2.getValidProposeBlocks() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func Test_actorV2_validateBlock(t *testing.T) {
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
			if err := actorV2.validateBlock(tt.args.bestView, tt.args.proposeBlockInfo); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.validateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_actorV2_voteForBlock(t *testing.T) {
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
			if err := actorV2.voteForBlock(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.voteForBlock() error = %v, wantErr %v", err, tt.wantErr)
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
			if err := actorV2.processWithEnoughVotes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("actorV2.processWithEnoughVotes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
