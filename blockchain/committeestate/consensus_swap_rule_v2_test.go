package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
)

var samplePenalty = signaturecounter.Penalty{
	MinPercent: 50,
	// Time:         0,
	ForceUnstake: true,
}

func Test_swapRuleV2_GenInstructions(t *testing.T) {
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
		s     *swapRuleV2
		args  args
		want  *instruction.SwapShardInstruction
		want1 []string
		want2 []string
		want3 []string
		want4 []string
	}{
		{
			name: "max committee size 8, one slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5,
				},
				substitutes: []string{
					key6, key7, key8, key9,
				},
				minCommitteeSize:        4,
				maxCommitteeSize:        8,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
				},
			},
			s: &swapRuleV2{},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key6, key7, key8, key9,
				},
				[]string{
					key5, key4,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{key0, key, key2, key3, key6, key7, key8, key9},
			want2: []string{},
			want3: []string{key5},
			want4: []string{key4},
		},
		{
			name: "max committee size 6, one slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5,
				},
				substitutes: []string{
					key6, key7, key8, key9,
				},
				minCommitteeSize:        4,
				maxCommitteeSize:        6,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
				},
			},
			s: &swapRuleV2{},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key6, key7,
				},
				[]string{
					key5, key4,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key6, key7,
			},
			want2: []string{
				key8, key9,
			},
			want3: []string{key5},
			want4: []string{key4},
		},
		{
			name: "max committee size 9, two slash, spare one slash in fixed nodes, no normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key8,
				},
				substitutes: []string{
					key6, key7, key9, key10, key11, key12,
				},
				minCommitteeSize:        4,
				maxCommitteeSize:        9,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			s: &swapRuleV2{},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key6, key7, key9, key10,
				},
				[]string{
					key5, key8,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key4, key6, key7, key9, key10,
			},
			want2: []string{key11, key12},
			want3: []string{key5, key8},
			want4: []string{},
		},
		{
			name: "max committee size 12, swap offset 4 - 1, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12,
				},
				minCommitteeSize:        10,
				maxCommitteeSize:        12,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			s: &swapRuleV2{},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key12,
				},
				[]string{
					key5, key8, key4,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key6, key7, key9, key10, key11, key12,
			},
			want2: []string{},
			want3: []string{key5, key8},
			want4: []string{key4},
		},
		{
			name: "max committee size 12, swap offset 4, two slash, spare one slash in fixed nodes, two normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12, key13,
				},
				minCommitteeSize:        10,
				maxCommitteeSize:        12,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			s: &swapRuleV2{},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key12, key13,
				},
				[]string{
					key5, key8, key4, key6,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key7, key9, key10, key11, key12, key13,
			},
			want2: []string{},
			want3: []string{key5, key8},
			want4: []string{key4, key6},
		},
		{
			name: "max committee size 12, swap offset 4 (push max), two slash, spare one slash in fixed nodes, two normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12, key13, key14, key15, key16, key17,
				},
				minCommitteeSize:        10,
				maxCommitteeSize:        12,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			s: &swapRuleV2{},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key12, key13, key14, key15,
				},
				[]string{
					key5, key8, key4, key6,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key7, key9, key10, key11, key12, key13, key14, key15,
			},
			want2: []string{key16, key17},
			want3: []string{key5, key8},
			want4: []string{key4, key6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV2{}
			got, got1, got2, got3, got4 := s.Process(tt.args.shardID, tt.args.committees, tt.args.substitutes, tt.args.minCommitteeSize, tt.args.maxCommitteeSize, tt.args.typeIns, tt.args.numberOfFixedValidators, tt.args.penalty)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV2.Process() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV2.Process() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapRuleV2.Process() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("swapRuleV2.Process() got3 = %v, want %v", got3, tt.want3)
			}
			if !reflect.DeepEqual(got4, tt.want4) {
				t.Errorf("swapRuleV2.Process() got4 = %v, want %v", got4, tt.want4)
			}
		})
	}
}

func Test_swapRuleV2_slashingSwapOut(t *testing.T) {
	type args struct {
		committees             []string
		substitutes            []string
		penalty                map[string]signaturecounter.Penalty
		minCommitteeSize       int
		numberOfFixedValidator int
	}
	tests := []struct {
		name  string
		s     *swapRuleV2
		args  args
		want  []string
		want1 []string
		want2 []string
	}{
		{
			name: "Length of committees is min",
			s:    &swapRuleV2{},
			args: args{
				committees: []string{
					key, key0, key2, key3,
				},
				substitutes: []string{},
				penalty: map[string]signaturecounter.Penalty{
					key0: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key, key0, key2, key3,
			},
			want1: []string{},
			want2: []string{},
		},
		{
			name: "swap offset 3, one slash, spare one slash in fixed nodes, two normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				substitutes: []string{},
				penalty: map[string]signaturecounter.Penalty{
					key8: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key6, key7, key9,
			},
			want1: []string{
				key8,
			},
			want2: []string{
				key4, key5,
			},
			s: &swapRuleV2{},
		},
		{
			name: "swap offset 3, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8: samplePenalty,
					key6: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key5, key7, key9,
			},
			want1: []string{
				key6, key8,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 3, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8: samplePenalty,
					key6: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key5, key7, key9,
			},
			want1: []string{
				key6, key8,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 2, one slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5,
				},
				substitutes: []string{
					key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key:  samplePenalty,
					key8: samplePenalty,
					key5: samplePenalty,
				},
				minCommitteeSize:       5,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3,
			},
			want1: []string{
				key5,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 4 - 1, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12,
				},
				penalty: map[string]signaturecounter.Penalty{
					key:  samplePenalty,
					key8: samplePenalty,
					key5: samplePenalty,
				},
				minCommitteeSize:       10,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key6, key7, key9, key10, key11,
			},
			want1: []string{
				key5, key8,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 4 - 0, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12, key13,
				},
				penalty: map[string]signaturecounter.Penalty{
					key:  samplePenalty,
					key8: samplePenalty,
					key5: samplePenalty,
				},
				minCommitteeSize:       10,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key7, key9, key10, key11,
			},
			want1: []string{
				key5, key8,
			},
			want2: []string{
				key4, key6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV2{}
			got, got1, got2 := s.slashingSwapOut(tt.args.committees, tt.args.substitutes, tt.args.penalty, tt.args.minCommitteeSize, tt.args.numberOfFixedValidator)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV2.slashingSwapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV2.slashingSwapOut() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapRuleV2.slashingSwapOut() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_swapRuleV2_swapInAfterSwapOut(t *testing.T) {
	type args struct {
		committees       []string
		substitutes      []string
		maxCommitteeSize int
	}
	tests := []struct {
		name  string
		s     *swapRuleV2
		args  args
		want  []string
		want1 []string
		want2 []string
	}{
		{
			name: "push to max committee size",
			args: args{
				committees: []string{
					key0, key, key2, key3,
				},
				substitutes: []string{
					key4, key5, key6, key7, key8, key9,
				},
				maxCommitteeSize: 9,
			},
			s: &swapRuleV2{},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8,
			},
			want1: []string{
				key9,
			},
			want2: []string{
				key4, key5, key6, key7, key8,
			},
		},
		{
			name: "push all substitute list but not max committee size",
			args: args{
				committees: []string{
					key0, key, key2, key3,
				},
				substitutes: []string{
					key4, key5, key6, key7, key8, key9,
				},
				maxCommitteeSize: 11,
			},
			s: &swapRuleV2{},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
			},
			want1: []string{},
			want2: []string{
				key4, key5, key6, key7, key8, key9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV2{}
			got, got1, got2 := s.swapInAfterSwapOut(tt.args.committees, tt.args.substitutes, tt.args.maxCommitteeSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV2.swapInAfterSwapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV2.swapInAfterSwapOut() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapRuleV2.swapInAfterSwapOut() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_swapRuleV2_getSwapOutOffset(t *testing.T) {
	type args struct {
		numberOfSubstitutes    int
		numberOfCommittees     int
		numberOfFixedValidator int
		minCommitteeSize       int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test 1",
			args: args{
				numberOfCommittees:     6,
				numberOfSubstitutes:    0,
				numberOfFixedValidator: 4,
				minCommitteeSize:       4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV2{}
			if got := s.getSwapOutOffset(tt.args.numberOfSubstitutes, tt.args.numberOfCommittees, tt.args.numberOfFixedValidator, tt.args.minCommitteeSize); got != tt.want {
				t.Errorf("getSwapOutOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}
