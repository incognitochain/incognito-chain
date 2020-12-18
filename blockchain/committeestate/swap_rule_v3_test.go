package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

func Test_swapRuleV3_GenInstructions(t *testing.T) {
	initLog()
	initTestParams()

	type args struct {
		shardID                 byte
		committees              []string
		substitutes             []string
		minCommitteeSize        int
		maxCommitteeSize        int
		typeIns                 int
		numberOfFixedValidators int
		dcsMaxCommitteeSize     int
		dcsMinCommitteeSize     int
		penalty                 map[string]signaturecounter.Penalty
	}
	tests := []struct {
		name  string
		s     *swapRuleV3
		args  args
		want  *instruction.SwapShardInstruction
		want1 []string
		want2 []string
		want3 []string
		want4 []string
	}{
		{
			name: "Valid input",
			s:    &swapRuleV3{},
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7,
					key8, key9, key10, key11, key12,
				},
				substitutes: []string{
					key13, key14, key15, key16, key17, key18, key19,
				},
				minCommitteeSize:        4,
				maxCommitteeSize:        8,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				dcsMaxCommitteeSize:     51,
				dcsMinCommitteeSize:     15,
				penalty: map[string]signaturecounter.Penalty{
					key0:  signaturecounter.Penalty{},
					key11: signaturecounter.Penalty{},
				},
			},
			want: &instruction.SwapShardInstruction{
				InPublicKeys: []string{
					key13, key14,
				},
				InPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey13, *incKey14,
				},
				OutPublicKeys: []string{
					key11,
				},
				OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey11,
				},
				ChainID: 0,
				Type:    instruction.SWAP_BY_END_EPOCH,
			},
			want1: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key8, key9, key10, key12, key13, key14,
			},
			want2: []string{
				key15, key16, key17, key18, key19,
			},
			want3: []string{
				key11,
			},
			want4: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1, got2, got3, got4 := s.GenInstructions(tt.args.shardID, tt.args.committees, tt.args.substitutes, tt.args.minCommitteeSize, tt.args.maxCommitteeSize, tt.args.typeIns, tt.args.numberOfFixedValidators, tt.args.dcsMaxCommitteeSize, tt.args.dcsMinCommitteeSize, tt.args.penalty)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV3.GenInstructions() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV3.GenInstructions() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapRuleV3.GenInstructions() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("swapRuleV3.GenInstructions() got3 = %v, want %v", got3, tt.want3)
			}
			if !reflect.DeepEqual(got4, tt.want4) {
				t.Errorf("swapRuleV3.GenInstructions() got4 = %v, want %v", got4, tt.want4)
			}
		})
	}
}

func Test_swapRuleV3_AssignOffset(t *testing.T) {
	type args struct {
		lenShardSubstitute      int
		lenCommittees           int
		numberOfFixedValidators int
		minCommitteeSize        int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want int
	}{
		{
			name: "lenCommittees < max assign percent",
			s:    &swapRuleV3{},
			args: args{
				lenShardSubstitute:      5,
				lenCommittees:           4,
				numberOfFixedValidators: 4,
				minCommitteeSize:        4,
			},
			want: 1,
		},
		{
			name: "lenCommittees - numberOfFixedValidators < lenCommittees / 3",
			s:    &swapRuleV3{},
			args: args{
				lenShardSubstitute:      10,
				lenCommittees:           12,
				numberOfFixedValidators: 11,
				minCommitteeSize:        4,
			},
			want: 1,
		},
		{
			name: "lenCommittees - numberOfFixedValidators >= lenCommittees / 3",
			s:    &swapRuleV3{},
			args: args{
				lenShardSubstitute:      12,
				lenCommittees:           12,
				numberOfFixedValidators: 8,
				minCommitteeSize:        4,
			},
			want: 2,
		},
		{
			name: "lenCommittees = numberOfFixedValidators >= max asign per shard",
			s:    &swapRuleV3{},
			args: args{
				lenShardSubstitute:      12,
				lenCommittees:           8,
				numberOfFixedValidators: 8,
				minCommitteeSize:        4,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.AssignOffset(tt.args.lenShardSubstitute, tt.args.lenCommittees, tt.args.numberOfFixedValidators, tt.args.minCommitteeSize); got != tt.want {
				t.Errorf("swapRuleV3.AssignOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_swapRuleV3_swapInAfterSwapOut(t *testing.T) {
	type args struct {
		committees              []string
		substitutes             []string
		maxSwapInPercent        int
		numberOfFixedValidators int
		dcsMaxCommitteeSize     int
		dcsMinCommitteeSize     int
	}
	tests := []struct {
		name  string
		s     *swapRuleV3
		args  args
		want  []string
		want1 []string
		want2 []string
	}{
		{
			name: "Valid input",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key0, key, key2, key3, key4,
				},
				maxSwapInPercent:        6,
				numberOfFixedValidators: 8,
				dcsMaxCommitteeSize:     51,
				dcsMinCommitteeSize:     15,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11, key0, key,
			},
			want1: []string{
				key2, key3, key4,
			},
			want2: []string{
				key0, key,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1, got2 := s.swapInAfterSwapOut(tt.args.committees, tt.args.substitutes, tt.args.maxSwapInPercent, tt.args.numberOfFixedValidators, tt.args.dcsMaxCommitteeSize, tt.args.dcsMinCommitteeSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV3.swapInAfterSwapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV3.swapInAfterSwapOut() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapRuleV3.swapInAfterSwapOut() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_swapRuleV3_getSwapInOffset(t *testing.T) {
	type args struct {
		lenCommitteesAfterSwapOut int
		lenSubstitutes            int
		maxSwapInPercent          int
		numberOfFixedValidators   int
		dcsMaxCommitteeSize       int
		dcsMinCommitteeSize       int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want int
	}{
		{
			name: "substitutes < committees && committees > dcsMinCommitteeSize ",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 24,
				lenSubstitutes:            20,
				maxSwapInPercent:          6,
				numberOfFixedValidators:   8,
				dcsMaxCommitteeSize:       51,
				dcsMinCommitteeSize:       15,
			},
			want: 0,
		},
		{
			name: "substitutes < committees && committees <= dcsMinCommitteeSize && committees / 6 + committees <= dcsMaxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 12,
				lenSubstitutes:            5,
				maxSwapInPercent:          6,
				numberOfFixedValidators:   8,
				dcsMaxCommitteeSize:       51,
				dcsMinCommitteeSize:       15,
			},
			want: 2,
		},
		{
			name: "substitutes >= committees && committees / 6 + committees <= dcsMaxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 12,
				lenSubstitutes:            5,
				maxSwapInPercent:          6,
				numberOfFixedValidators:   8,
				dcsMaxCommitteeSize:       51,
				dcsMinCommitteeSize:       15,
			},
			want: 2,
		},
		{
			name: "substitutes >= committees && committees / 6 + committees > dcsMaxCommitteeSize ",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 48,
				lenSubstitutes:            60,
				maxSwapInPercent:          6,
				numberOfFixedValidators:   8,
				dcsMaxCommitteeSize:       51,
				dcsMinCommitteeSize:       15,
			},
			want: 3,
		},
		{
			name: "substitutes >= committees && committees / 6 + committees <= dcsMaxCommitteeSize ",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 24,
				lenSubstitutes:            40,
				maxSwapInPercent:          6,
				numberOfFixedValidators:   8,
				dcsMaxCommitteeSize:       51,
				dcsMinCommitteeSize:       15,
			},
			want: 4,
		},
		{
			name: "committees >= substitutes && committees < maxSwapInPercent",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 4,
				lenSubstitutes:            2,
				maxSwapInPercent:          6,
				numberOfFixedValidators:   4,
				dcsMaxCommitteeSize:       51,
				dcsMinCommitteeSize:       15,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.getSwapInOffset(tt.args.lenCommitteesAfterSwapOut, tt.args.lenSubstitutes, tt.args.maxSwapInPercent, tt.args.numberOfFixedValidators, tt.args.dcsMaxCommitteeSize, tt.args.dcsMinCommitteeSize); got != tt.want {
				t.Errorf("swapRuleV3.getSwapInOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_swapRuleV3_normalSwapOut(t *testing.T) {
	type args struct {
		committees                        []string
		substitutes                       []string
		lenBeforeSlashedCommittees        int
		lenSlashedCommittees              int
		maxSwapOutPercent                 int
		numberOfFixedValidators           int
		dcsMaxCommitteeSize               int
		dcsMinCommitteeSize               int
		maxCommitteeeSubstituteRangeTimes int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want []string
	}{
		{
			name: "Valid input",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
					key12, key13, key14, key15, key16, key17, key18, key19,
				},
				substitutes: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
					key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
					key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
					key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
					key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
					key12, key13, key14, key15, key16, key17, key18, key19,
					key12, key13, key14, key15, key16, key17, key18, key19,
				},
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              2,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.normalSwapOut(tt.args.committees, tt.args.substitutes, tt.args.lenBeforeSlashedCommittees, tt.args.lenSlashedCommittees, tt.args.maxSwapOutPercent, tt.args.numberOfFixedValidators, tt.args.dcsMaxCommitteeSize, tt.args.dcsMinCommitteeSize, tt.args.maxCommitteeeSubstituteRangeTimes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV3.normalSwapOut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_swapRuleV3_getNormalSwapOutOffset(t *testing.T) {
	type args struct {
		lenCommitteesBeforeSlash          int
		lenSubstitutes                    int
		lenSlashedCommittees              int
		maxSwapOutPercent                 int
		numberOfFixedValidators           int
		dcsMaxCommitteeSize               int
		dcsMinCommitteeSize               int
		maxCommitteeeSubstituteRangeTimes int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want int
	}{
		{
			name: "slashed nodes >= lenCommittees / 6",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          15,
				lenSubstitutes:                    10,
				lenSlashedCommittees:              5,
				maxSwapOutPercent:                 6,
				numberOfFixedValidators:           8,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 0,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes >= 4 * lenCommittees && lenCommittees >= dcsMaxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          51,
				lenSubstitutes:                    210,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              5,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 3,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes >= 4 * lenCommittees && lenCommittees < dcsMaxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          40,
				lenSubstitutes:                    210,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              4,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 0,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes < 4 * lenCommittees && lenCommittees < dcsMinCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          12,
				lenSubstitutes:                    30,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              1,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 0,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes < 4 * lenCommittees && lenCommittees >= dcsMinCommitteeSize && lenCommittees - slashed nodes - normal swap out nodes >= dcsMinCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          20,
				lenSubstitutes:                    60,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              2,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 1,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes < 4 * lenCommittees && lenCommittees >= dcsMinCommitteeSize && lenCommittees - slashed nodes - normal swap out nodes < dcsMinCommitteeSize && normal swap out nodes >= 0",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          20,
				lenSubstitutes:                    60,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              2,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 1,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes < 4 * lenCommittees && lenCommittees >= dcsMinCommitteeSize && lenCommittees - slashed nodes - normal swap out nodes < dcsMinCommitteeSize && normal swap out nodes < 0 - 1",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          15,
				lenSubstitutes:                    30,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              1,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 0,
		},
		{
			name: "slashed nodes < lenCommittees / 6 && lenSubstitutes < 4 * lenCommittees && lenCommittees >= dcsMinCommitteeSize && lenCommittees - slashed nodes - normal swap out nodes < dcsMinCommitteeSize && normal swap out nodes < 0 - 2",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash:          16,
				lenSubstitutes:                    30,
				numberOfFixedValidators:           8,
				lenSlashedCommittees:              1,
				maxSwapOutPercent:                 6,
				dcsMaxCommitteeSize:               51,
				dcsMinCommitteeSize:               15,
				maxCommitteeeSubstituteRangeTimes: 4,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.getNormalSwapOutOffset(tt.args.lenCommitteesBeforeSlash, tt.args.lenSubstitutes, tt.args.lenSlashedCommittees, tt.args.maxSwapOutPercent, tt.args.numberOfFixedValidators, tt.args.dcsMaxCommitteeSize, tt.args.dcsMinCommitteeSize, tt.args.maxCommitteeeSubstituteRangeTimes); got != tt.want {
				t.Errorf("swapRuleV3.getNormalSwapOutOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_swapRuleV3_slashingSwapOut(t *testing.T) {
	type args struct {
		committees              []string
		penalty                 map[string]signaturecounter.Penalty
		minCommitteeSize        int
		numberOfFixedValidators int
		maxSlashOutPercent      int
	}
	tests := []struct {
		name  string
		s     *swapRuleV3
		args  args
		want  []string
		want1 []string
	}{
		{
			name: "slashingOffset = 0",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7,
				},
				penalty: map[string]signaturecounter.Penalty{
					key0: signaturecounter.Penalty{},
					key:  signaturecounter.Penalty{},
				},
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
			},
			want1: []string{},
		},
		{
			name: "slashingOffset <= len(committees) / 3 && len(committees) - slashingOffset - numberOfFixedValidators < 0",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11, key12, key13, key14,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8:  signaturecounter.Penalty{},
					key9:  signaturecounter.Penalty{},
					key10: signaturecounter.Penalty{},
					key11: signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
				},
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key13, key14,
			},
			want1: []string{
				key8, key9, key10, key11, key12,
			},
		},
		{
			name: "slashingOffset <= len(committees) / 3 && len(committees) - slashingOffset - numberOfFixedValidators >= 0",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key0: signaturecounter.Penalty{},
					key:  signaturecounter.Penalty{},
					key8: signaturecounter.Penalty{},
					key9: signaturecounter.Penalty{},
				},
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
			},
			want1: []string{
				key8, key9,
			},
		},
		{
			name: "slashingOffset <= len(penalty)",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11, key12, key13, key14,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8:  signaturecounter.Penalty{},
					key9:  signaturecounter.Penalty{},
					key10: signaturecounter.Penalty{},
					key11: signaturecounter.Penalty{},
				},
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key12, key13, key14,
			},
			want1: []string{
				key8, key9, key10, key11,
			},
		},
		{
			name: "slashingOffset > len(penalty)",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11, key12, key13, key14,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8:  signaturecounter.Penalty{},
					key9:  signaturecounter.Penalty{},
					key10: signaturecounter.Penalty{},
					key11: signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key13: signaturecounter.Penalty{},
				},
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key13, key14,
			},
			want1: []string{
				key8, key9, key10, key11, key12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1 := s.slashingSwapOut(tt.args.committees, tt.args.penalty, tt.args.minCommitteeSize, tt.args.numberOfFixedValidators, tt.args.maxSlashOutPercent)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV3.slashingSwapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV3.slashingSwapOut() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_swapRuleV3_getSlashingOffset(t *testing.T) {
	type args struct {
		lenCommittees           int
		minCommitteeSize        int
		numberOfFixedValidators int
		maxSlashOutPercent      int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want int
	}{
		{
			name: "fixed validators = committees",
			s:    &swapRuleV3{},
			args: args{
				lenCommittees:           8,
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: 0,
		},
		{
			name: "fixed validators + lenCommittees / 3 > lenCommittees",
			s:    &swapRuleV3{},
			args: args{
				lenCommittees:           10,
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: 2,
		},
		{
			name: "fixed validators + lenCommittees / 3 <= lenCommittees",
			s:    &swapRuleV3{},
			args: args{
				lenCommittees:           15,
				minCommitteeSize:        4,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.getSlashingOffset(tt.args.lenCommittees, tt.args.minCommitteeSize, tt.args.numberOfFixedValidators, tt.args.maxSlashOutPercent); got != tt.want {
				t.Errorf("swapRuleV3.getSlashingOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}
