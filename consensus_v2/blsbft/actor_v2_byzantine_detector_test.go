package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"reflect"
	"testing"
)

func TestByzantineDetector_checkBlackListValidator(t *testing.T) {
	type fields struct {
		blackList map[string]*rawdb_consensus.BlackListValidator
	}
	type args struct {
		bftVote *BFTVote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "not in blacklist",
			fields: fields{
				blackList: map[string]*rawdb_consensus.BlackListValidator{
					blsKeys[0]: &rawdb_consensus.BlackListValidator{
						Error: ErrDuplicateVoteInOneTimeSlot.Error(),
					},
				},
			},
			args: args{
				bftVote: &BFTVote{
					Validator: blsKeys[1],
				},
			},
			wantErr: false,
		},
		{
			name: "in blacklist",
			fields: fields{
				blackList: map[string]*rawdb_consensus.BlackListValidator{
					blsKeys[0]: &rawdb_consensus.BlackListValidator{
						Error: ErrDuplicateVoteInOneTimeSlot.Error(),
					},
					blsKeys[1]: &rawdb_consensus.BlackListValidator{
						Error: ErrDuplicateVoteInOneTimeSlot.Error(),
					},
				},
			},
			args: args{
				bftVote: &BFTVote{
					Validator: blsKeys[1],
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList: tt.fields.blackList,
			}
			if err := b.checkBlackListValidator(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("checkBlackListValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestByzantineDetector_voteMoreThanOneTimesInATimeSlot(t *testing.T) {
	type fields struct {
		blackList        map[string]*rawdb_consensus.BlackListValidator
		timeslot         map[string]map[int64]*BFTVote
		committeeHandler CommitteeChainHandler
	}
	type args struct {
		bftVote *BFTVote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "first vote in a specific timeslot",
			fields: fields{
				timeslot: map[string]map[int64]*BFTVote{
					blsKeys[0]: make(map[int64]*BFTVote),
				},
			},
			args: args{
				&BFTVote{
					Validator:       blsKeys[0],
					BlockHeight:     10,
					ProposeTimeSlot: 163394559,
				},
			},
			wantErr: false,
		},
		{
			name: "second vote is not the same vote as the first vote",
			fields: fields{
				timeslot: map[string]map[int64]*BFTVote{
					blsKeys[0]: {
						163394559: &BFTVote{
							Validator:       blsKeys[0],
							BlockHeight:     10,
							ProposeTimeSlot: 163394559,
						},
					},
				},
			},
			args: args{
				&BFTVote{
					Validator:       blsKeys[0],
					BlockHeight:     11,
					ProposeTimeSlot: 163394559,
				},
			},
			wantErr: true,
		},
		{
			name: "second vote is the same vote as first vote",
			fields: fields{
				timeslot: map[string]map[int64]*BFTVote{
					blsKeys[0]: {
						163394559: &BFTVote{
							Validator:       blsKeys[0],
							BlockHeight:     11,
							ProposeTimeSlot: 163394559,
						},
					},
				},
			},
			args: args{
				&BFTVote{
					Validator:       blsKeys[0],
					BlockHeight:     11,
					ProposeTimeSlot: 163394559,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList:      tt.fields.blackList,
				voteInTimeSlot: tt.fields.timeslot,
			}
			if err := b.voteMoreThanOneTimesInATimeSlot(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteMoreThanOneTimesInATimeSlot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestByzantineDetector_voteForHigherTimeSlotSameHeight(t *testing.T) {
	type fields struct {
		blackList        map[string]*rawdb_consensus.BlackListValidator
		voteInTimeSlot   map[string]map[int64]*BFTVote
		smallestTimeSlot map[string]map[uint64]*BFTVote
		committeeHandler CommitteeChainHandler
	}
	type args struct {
		bftVote *BFTVote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "first vote in a specific height",
			fields: fields{
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					blsKeys[0]: make(map[uint64]*BFTVote),
				},
			},
			args: args{
				bftVote: &BFTVote{
					Validator:       blsKeys[0],
					ProduceTimeSlot: 163394559,
					BlockHeight:     10,
				},
			},
			wantErr: false,
		},
		{
			name: "second vote smaller timeslot in a specific height",
			fields: fields{
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					blsKeys[0]: {
						10: &BFTVote{ProduceTimeSlot: 163394559},
					},
				},
			},
			args: args{
				bftVote: &BFTVote{
					Validator:       blsKeys[0],
					ProduceTimeSlot: 163394558,
					BlockHeight:     10,
				},
			},
			wantErr: false,
		},
		{
			name: "second vote higher timeslot in a specific height",
			fields: fields{
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					blsKeys[0]: {
						10: &BFTVote{ProduceTimeSlot: 163394559},
					},
				},
			},
			args: args{
				bftVote: &BFTVote{
					Validator:       blsKeys[0],
					ProduceTimeSlot: 163394560,
					BlockHeight:     10,
				},
			},
			wantErr: true,
		},
		{
			name: "second vote higher timeslot in a specific height but different committee",
			fields: fields{
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					blsKeys[0]: {
						10: &BFTVote{ProduceTimeSlot: 163394559,
							CommitteeFromBlock: common.HashH([]byte{0})},
					},
				},
			},
			args: args{
				bftVote: &BFTVote{
					Validator:          blsKeys[0],
					ProduceTimeSlot:    163394560,
					BlockHeight:        10,
					CommitteeFromBlock: common.HashH([]byte{1}),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList:                    tt.fields.blackList,
				voteInTimeSlot:               tt.fields.voteInTimeSlot,
				smallestBlockProduceTimeSlot: tt.fields.smallestTimeSlot,
			}
			if err := b.voteForHigherTimeSlotSameHeight(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteForHigherTimeSlotSameHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestByzantineDetector_addNewVote(t *testing.T) {

	key0VoteHeight10_1 := &BFTVote{
		Validator:          blsKeys[0],
		BlockHeight:        10,
		ProduceTimeSlot:    163394560,
		ProposeTimeSlot:    163394560,
		CommitteeFromBlock: common.HashH([]byte{0}),
	}
	key0VoteHeight10_2 := &BFTVote{
		Validator:          blsKeys[0],
		BlockHeight:        10,
		ProduceTimeSlot:    163394560,
		ProposeTimeSlot:    163394561,
		CommitteeFromBlock: common.HashH([]byte{0}),
	}
	key0VoteHeight10_3 := &BFTVote{
		Validator:          blsKeys[0],
		BlockHeight:        10,
		ProduceTimeSlot:    163394562,
		ProposeTimeSlot:    163394562,
		CommitteeFromBlock: common.HashH([]byte{1}),
	}
	key1VoteHeight10_1 := &BFTVote{
		Validator:          blsKeys[1],
		BlockHeight:        10,
		ProduceTimeSlot:    163394562,
		ProposeTimeSlot:    163394562,
		CommitteeFromBlock: common.HashH([]byte{0}),
	}

	type fields struct {
		blackList        map[string]*rawdb_consensus.BlackListValidator
		voteInTimeSlot   map[string]map[int64]*BFTVote
		smallestTimeSlot map[string]map[uint64]*BFTVote
		validRecentVote  map[string]*BFTVote
		committeeHandler CommitteeChainHandler
	}
	type args struct {
		bftVote      *BFTVote
		validatorErr error
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		fieldAfterProcess fields
	}{
		{
			name: "no init data",
			fields: fields{
				blackList:        nil,
				voteInTimeSlot:   nil,
				smallestTimeSlot: nil,
				committeeHandler: nil,
				validRecentVote:  make(map[string]*BFTVote),
			},
			args: args{
				bftVote:      key0VoteHeight10_1,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProduceTimeSlot: key0VoteHeight10_1,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1,
					},
				},
				committeeHandler: nil,
			},
		},
		{
			name: "add vote with error",
			fields: fields{
				blackList:        nil,
				voteInTimeSlot:   nil,
				smallestTimeSlot: nil,
				committeeHandler: nil,
				validRecentVote:  make(map[string]*BFTVote),
			},
			args: args{
				bftVote:      key0VoteHeight10_1,
				validatorErr: ErrDuplicateVoteInOneTimeSlot,
			},
			fieldAfterProcess: fields{
				blackList: map[string]*rawdb_consensus.BlackListValidator{
					key0VoteHeight10_1.Validator: &rawdb_consensus.BlackListValidator{
						Error: ErrDuplicateVoteInOneTimeSlot.Error(),
					},
				},
				voteInTimeSlot:   nil,
				smallestTimeSlot: nil,
				committeeHandler: nil,
			},
		},
		{
			name: "add new data with no error 1",
			fields: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProduceTimeSlot: key0VoteHeight10_1,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1,
					},
				},
				validRecentVote:  make(map[string]*BFTVote),
				committeeHandler: nil,
			},
			args: args{
				bftVote:      key0VoteHeight10_2,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1,
					},
				},
				validRecentVote: map[string]*BFTVote{
					key0VoteHeight10_1.Validator: key0VoteHeight10_2,
				},
				committeeHandler: nil,
			},
		},
		{
			name: "add new data with no error 2",
			fields: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1,
					},
				},
				validRecentVote:  make(map[string]*BFTVote),
				committeeHandler: nil,
			},
			args: args{
				bftVote:      key1VoteHeight10_1,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
					key1VoteHeight10_1.Validator: {
						key1VoteHeight10_1.ProposeTimeSlot: key1VoteHeight10_1,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1,
					},
					key1VoteHeight10_1.Validator: {
						key1VoteHeight10_1.BlockHeight: key1VoteHeight10_1,
					},
				},
				validRecentVote: map[string]*BFTVote{
					key0VoteHeight10_1.Validator: key1VoteHeight10_1,
				},
				committeeHandler: nil,
			},
		},
		{
			name: "add new data with no error 3_same height higher timeslot but different committee",
			fields: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1,
					},
				},
				validRecentVote: map[string]*BFTVote{
					key0VoteHeight10_1.Validator: key0VoteHeight10_2,
				},
				committeeHandler: nil,
			},
			args: args{
				bftVote:      key0VoteHeight10_3,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]*rawdb_consensus.BlackListValidator),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
						key0VoteHeight10_3.ProposeTimeSlot: key0VoteHeight10_3,
					},
				},
				smallestTimeSlot: map[string]map[uint64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_3,
					},
				},
				validRecentVote: map[string]*BFTVote{
					key0VoteHeight10_1.Validator: key0VoteHeight10_3,
				},
				committeeHandler: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &ByzantineDetector{
				blackList:                    tt.fields.blackList,
				voteInTimeSlot:               tt.fields.voteInTimeSlot,
				smallestBlockProduceTimeSlot: tt.fields.smallestTimeSlot,
				validRecentVote:              tt.fields.validRecentVote,
			}
			b.addNewVote(diskDB, tt.args.bftVote, tt.args.validatorErr)
			for k, v := range b.blackList {
				wantV := tt.fieldAfterProcess.blackList[k]
				if wantV.Error != v.Error {
					t.Errorf("addNewVote.blackList want %+v, got %+v", tt.fieldAfterProcess.blackList, b.blackList)
				}

			}
			if !reflect.DeepEqual(b.smallestBlockProduceTimeSlot, tt.fieldAfterProcess.smallestTimeSlot) {
				t.Errorf("addNewVote.smallestBlockProduceTimeSlot want %+v, got %+v", tt.fieldAfterProcess.smallestTimeSlot, b.smallestBlockProduceTimeSlot)
			}
			if !reflect.DeepEqual(b.voteInTimeSlot, tt.fieldAfterProcess.voteInTimeSlot) {
				t.Errorf("addNewVote.voteInTimeSlot want %+v, got %+v", tt.fieldAfterProcess.voteInTimeSlot, b.voteInTimeSlot)
			}
		})
	}
}

func TestByzantineDetector_voteForSmallerBlockHeight(t *testing.T) {
	type fields struct {
		validRecentVote map[string]*BFTVote
	}
	type args struct {
		newVote *BFTVote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "first vote",
			fields: fields{
				validRecentVote: map[string]*BFTVote{},
			},
			args: args{
				newVote: &BFTVote{
					Validator: blsKeys[0],
				},
			},
			wantErr: false,
		},
		{
			name: "vote for another chain",
			fields: fields{
				validRecentVote: map[string]*BFTVote{
					blsKeys[0]: &BFTVote{
						Validator: blsKeys[0],
						ChainID:   0,
					},
				},
			},
			args: args{
				newVote: &BFTVote{
					Validator: blsKeys[0],
					ChainID:   1,
				},
			},
			wantErr: false,
		},
		{
			name: "vote block created by another committee",
			fields: fields{
				validRecentVote: map[string]*BFTVote{
					blsKeys[0]: &BFTVote{
						Validator:          blsKeys[0],
						ChainID:            0,
						CommitteeFromBlock: common.Hash{0},
					},
				},
			},
			args: args{
				newVote: &BFTVote{
					Validator:          blsKeys[0],
					ChainID:            0,
					CommitteeFromBlock: common.Hash{1},
				},
			},
			wantErr: false,
		},
		{
			name: "vote for higher block",
			fields: fields{
				validRecentVote: map[string]*BFTVote{
					blsKeys[0]: &BFTVote{
						Validator:          blsKeys[0],
						ChainID:            0,
						CommitteeFromBlock: common.Hash{1},
						BlockHeight:        10,
					},
				},
			},
			args: args{
				newVote: &BFTVote{
					Validator:          blsKeys[0],
					ChainID:            0,
					CommitteeFromBlock: common.Hash{1},
					BlockHeight:        11,
				},
			},
			wantErr: false,
		},
		{
			name: "vote for equal block",
			fields: fields{
				validRecentVote: map[string]*BFTVote{
					blsKeys[0]: &BFTVote{
						Validator:          blsKeys[0],
						ChainID:            0,
						CommitteeFromBlock: common.Hash{1},
						BlockHeight:        10,
					},
				},
			},
			args: args{
				newVote: &BFTVote{
					Validator:          blsKeys[0],
					ChainID:            0,
					CommitteeFromBlock: common.Hash{1},
					BlockHeight:        10,
				},
			},
			wantErr: false,
		},
		{
			name: "vote for smaller block",
			fields: fields{
				validRecentVote: map[string]*BFTVote{
					blsKeys[0]: &BFTVote{
						Validator:          blsKeys[0],
						ChainID:            0,
						CommitteeFromBlock: common.Hash{1},
						BlockHeight:        10,
					},
				},
			},
			args: args{
				newVote: &BFTVote{
					Validator:          blsKeys[0],
					ChainID:            0,
					CommitteeFromBlock: common.Hash{1},
					BlockHeight:        9,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				validRecentVote: tt.fields.validRecentVote,
			}
			if err := b.voteForSmallerBlockHeight(tt.args.newVote); (err != nil) != tt.wantErr {
				t.Errorf("voteForSmallerBlockHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
