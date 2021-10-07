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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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

func TestByzantineDetector_voteForSmallerTimeSlotSameHeight(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByzantineDetector{
				blackList:        tt.fields.blackList,
				voteInTimeSlot:   tt.fields.voteInTimeSlot,
				smallestTimeSlot: tt.fields.smallestTimeSlot,
				committeeHandler: tt.fields.committeeHandler,
			}
			if err := b.voteForSmallerTimeSlotSameHeight(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteForSmallerTimeSlotSameHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestByzantineDetector_voteForSmallerBlockHeight(t *testing.T) {
	type fields struct {
		blackList         map[string]error
		voteInTimeSlot    map[string]map[int64]*BFTVote
		smallestTimeSlot  map[string]map[uint64]int64
		latestBlockHeight map[string]uint64
		committeeHandler  CommitteeChainHandler
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
				blackList:         tt.fields.blackList,
				voteInTimeSlot:    tt.fields.voteInTimeSlot,
				smallestTimeSlot:  tt.fields.smallestTimeSlot,
				latestBlockHeight: tt.fields.latestBlockHeight,
				committeeHandler:  tt.fields.committeeHandler,
			}
			if err := b.voteForSmallerBlockHeight(tt.args.bftVote); (err != nil) != tt.wantErr {
				t.Errorf("voteForSmallerBlockHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
