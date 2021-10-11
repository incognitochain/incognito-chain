package blsbft

import "testing"

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

func TestByzantineDetector_voteOwnerSignature(t *testing.T) {
	type fields struct {
		blackList        map[string]error
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList:        tt.fields.blackList,
				committeeHandler: tt.fields.committeeHandler,
			}
			if err := b.voteOwnerSignature(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteOwnerSignature() error = %v, wantErr %v", err, tt.wantErr)
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
					Validator:   blsKeys[0],
					BlockHeight: 10,
					TimeSlot:    163394559,
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
							Validator:   blsKeys[0],
							BlockHeight: 10,
							TimeSlot:    163394559,
						},
					},
				},
			},
			args: args{
				&BFTVote{
					Validator:   blsKeys[0],
					BlockHeight: 11,
					TimeSlot:    163394559,
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
							Validator:   blsKeys[0],
							BlockHeight: 11,
							TimeSlot:    163394559,
						},
					},
				},
			},
			args: args{
				&BFTVote{
					Validator:   blsKeys[0],
					BlockHeight: 11,
					TimeSlot:    163394559,
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
					Validator:   blsKeys[0],
					TimeSlot:    163394559,
					BlockHeight: 10,
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
					Validator:   blsKeys[0],
					TimeSlot:    163394558,
					BlockHeight: 10,
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
					Validator:   blsKeys[0],
					TimeSlot:    163394560,
					BlockHeight: 10,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList:        tt.fields.blackList,
				voteInTimeSlot:   tt.fields.voteInTimeSlot,
				smallestTimeSlot: tt.fields.smallestTimeSlot,
				committeeHandler: tt.fields.committeeHandler,
			}
			if err := b.voteForHigherTimeSlotSameHeight(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteForHigherTimeSlotSameHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
