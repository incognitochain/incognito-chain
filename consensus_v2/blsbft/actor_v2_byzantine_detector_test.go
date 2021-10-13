package blsbft

import (
	"reflect"
	"testing"
)

func TestByzantineDetector_checkBlackListValidator(t *testing.T) {
	type fields struct {
		blackList map[string]error
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
				blackList: map[string]error{
					blsKeys[0]: ErrInvalidSignature,
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
				blackList: map[string]error{
					blsKeys[0]: ErrInvalidSignature,
					blsKeys[1]: ErrInvalidVoteOwner,
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
		blackList        map[string]error
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
				blackList:        tt.fields.blackList,
				voteInTimeSlot:   tt.fields.timeslot,
				committeeHandler: tt.fields.committeeHandler,
			}
			if err := b.voteMoreThanOneTimesInATimeSlot(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteMoreThanOneTimesInATimeSlot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestByzantineDetector_voteForHigherTimeSlotSameHeight(t *testing.T) {
	type fields struct {
		blackList        map[string]error
		voteInTimeSlot   map[string]map[int64]*BFTVote
		smallestTimeSlot map[string]map[uint64]int64
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
				smallestTimeSlot: map[string]map[uint64]int64{
					blsKeys[0]: make(map[uint64]int64),
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
				smallestTimeSlot: map[string]map[uint64]int64{
					blsKeys[0]: {
						10: 163394559,
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
				smallestTimeSlot: map[string]map[uint64]int64{
					blsKeys[0]: {
						10: 163394559,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList:               tt.fields.blackList,
				voteInTimeSlot:          tt.fields.voteInTimeSlot,
				smallestProduceTimeSlot: tt.fields.smallestTimeSlot,
				committeeHandler:        tt.fields.committeeHandler,
			}
			if err := b.voteForHigherTimeSlotSameHeight(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteForHigherTimeSlotSameHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestByzantineDetector_addNewVote(t *testing.T) {

	key0VoteHeight10_1 := &BFTVote{
		Validator:       blsKeys[0],
		BlockHeight:     10,
		ProduceTimeSlot: 163394560,
		ProposeTimeSlot: 163394560,
	}
	key0VoteHeight10_2 := &BFTVote{
		Validator:       blsKeys[0],
		BlockHeight:     10,
		ProduceTimeSlot: 163394560,
		ProposeTimeSlot: 163394561,
	}
	key1VoteHeight10_3 := &BFTVote{
		Validator:       blsKeys[1],
		BlockHeight:     10,
		ProduceTimeSlot: 163394562,
		ProposeTimeSlot: 163394562,
	}
	type fields struct {
		blackList        map[string]error
		voteInTimeSlot   map[string]map[int64]*BFTVote
		smallestTimeSlot map[string]map[uint64]int64
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
			},
			args: args{
				bftVote:      key0VoteHeight10_1,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]error),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProduceTimeSlot: key0VoteHeight10_1,
					},
				},
				smallestTimeSlot: map[string]map[uint64]int64{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1.ProduceTimeSlot,
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
			},
			args: args{
				bftVote:      key0VoteHeight10_1,
				validatorErr: ErrDuplicateVoteInOneTimeSlot,
			},
			fieldAfterProcess: fields{
				blackList: map[string]error{
					key0VoteHeight10_1.Validator: ErrDuplicateVoteInOneTimeSlot,
				},
				voteInTimeSlot:   nil,
				smallestTimeSlot: nil,
				committeeHandler: nil,
			},
		},
		{
			name: "add new data with no error 1",
			fields: fields{
				blackList: make(map[string]error),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProduceTimeSlot: key0VoteHeight10_1,
					},
				},
				smallestTimeSlot: map[string]map[uint64]int64{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1.ProduceTimeSlot,
					},
				},
				committeeHandler: nil,
			},
			args: args{
				bftVote:      key0VoteHeight10_2,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]error),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
				},
				smallestTimeSlot: map[string]map[uint64]int64{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1.ProduceTimeSlot,
					},
				},
				committeeHandler: nil,
			},
		},
		{
			name: "add new data with no error 2",
			fields: fields{
				blackList: make(map[string]error),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
				},
				smallestTimeSlot: map[string]map[uint64]int64{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1.ProduceTimeSlot,
					},
				},
				committeeHandler: nil,
			},
			args: args{
				bftVote:      key1VoteHeight10_3,
				validatorErr: nil,
			},
			fieldAfterProcess: fields{
				blackList: make(map[string]error),
				voteInTimeSlot: map[string]map[int64]*BFTVote{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.ProposeTimeSlot: key0VoteHeight10_1,
						key0VoteHeight10_2.ProposeTimeSlot: key0VoteHeight10_2,
					},
					key1VoteHeight10_3.Validator: {
						key1VoteHeight10_3.ProposeTimeSlot: key1VoteHeight10_3,
					},
				},
				smallestTimeSlot: map[string]map[uint64]int64{
					key0VoteHeight10_1.Validator: {
						key0VoteHeight10_1.BlockHeight: key0VoteHeight10_1.ProduceTimeSlot,
					},
					key1VoteHeight10_3.Validator: {
						key1VoteHeight10_3.BlockHeight: key1VoteHeight10_3.ProduceTimeSlot,
					},
				},
				committeeHandler: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &ByzantineDetector{
				blackList:               tt.fields.blackList,
				voteInTimeSlot:          tt.fields.voteInTimeSlot,
				smallestProduceTimeSlot: tt.fields.smallestTimeSlot,
				committeeHandler:        tt.fields.committeeHandler,
			}
			b.addNewVote(tt.args.bftVote, tt.args.validatorErr)
			if !reflect.DeepEqual(b.blackList, tt.fieldAfterProcess.blackList) {
				t.Errorf("addNewVote.blackList want %+v, got %+v", tt.fieldAfterProcess.blackList, b.blackList)
			}
			if !reflect.DeepEqual(b.smallestProduceTimeSlot, tt.fieldAfterProcess.smallestTimeSlot) {
				t.Errorf("addNewVote.smallestProduceTimeSlot want %+v, got %+v", tt.fieldAfterProcess.smallestTimeSlot, b.smallestProduceTimeSlot)
			}
			if !reflect.DeepEqual(b.voteInTimeSlot, tt.fieldAfterProcess.voteInTimeSlot) {
				t.Errorf("addNewVote.voteInTimeSlot want %+v, got %+v", tt.fieldAfterProcess.voteInTimeSlot, b.voteInTimeSlot)
			}
		})
	}
}
