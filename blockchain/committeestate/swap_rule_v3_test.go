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
			name: "[valid input]",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				},
				substitutes: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				minCommitteeSize:        20,
				maxCommitteeSize:        20,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
				},
			},
			want: &instruction.SwapShardInstruction{
				InPublicKeys: []string{
					key0, key,
				},
				InPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey,
				},
				OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey12, *incKey8,
				},
				OutPublicKeys: []string{
					key12, key8,
				},
				ChainID: 1,
				Type:    instruction.SWAP_BY_END_EPOCH,
			},
			want1: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key9, key10, key11, key13, key14, key15, key16, key17, key18, key19,
				key0, key,
			},
			want2: []string{
				key2, key3, key4, key5, key6, key7, key8, key9,
			},
			want3: []string{
				key12,
			},
			want4: []string{
				key8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1, got2, got3, got4 := s.GenInstructions(tt.args.shardID, tt.args.committees, tt.args.substitutes, tt.args.minCommitteeSize, tt.args.maxCommitteeSize, tt.args.typeIns, tt.args.numberOfFixedValidators, tt.args.penalty)
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
				lenCommittees:           4,
				numberOfFixedValidators: 4,
			},
			want: 1,
		},
		{
			name: "lenCommittees >= max assign percent",
			s:    &swapRuleV3{},
			args: args{
				lenCommittees:           10,
				numberOfFixedValidators: 8,
			},
			want: 1,
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
		numberOfFixedValidators int
		maxCommitteeSize        int
		maxSwapInPercent        int
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
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				},
				substitutes: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				},
				maxCommitteeSize:        64,
				maxSwapInPercent:        8,
				numberOfFixedValidators: 8,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				key0, key, key2, key3,
			},
			want1: []string{
				key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
			},
			want2: []string{
				key0, key, key2, key3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1, got2 := s.swapInAfterSwapOut(tt.args.committees, tt.args.substitutes, tt.args.numberOfFixedValidators, tt.args.maxCommitteeSize, tt.args.maxSwapInPercent)
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
		maxCommitteeSize          int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want int
	}{
		{
			name: "substitutes < committees / 8 && committees + offset <= maxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 25,
				lenSubstitutes:            2,
				maxSwapInPercent:          8,
				maxCommitteeSize:          64,
			},
			want: 2,
		},
		{
			name: "substitutes < committees / 8 && committees + offset > maxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 60,
				lenSubstitutes:            5,
				maxSwapInPercent:          8,
				maxCommitteeSize:          64,
			},
			want: 4,
		},
		{
			name: "substitutes >= committees / 8 && committees + offset <= maxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 25,
				lenSubstitutes:            5,
				maxSwapInPercent:          8,
				maxCommitteeSize:          64,
			},
			want: 3,
		},
		{
			name: "substitutes >= committees / 8 && committees + offset > maxCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesAfterSwapOut: 60,
				lenSubstitutes:            20,
				maxSwapInPercent:          8,
				maxCommitteeSize:          64,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.getSwapInOffset(tt.args.lenCommitteesAfterSwapOut, tt.args.lenSubstitutes, tt.args.maxSwapInPercent, tt.args.maxCommitteeSize); got != tt.want {
				t.Errorf("swapRuleV3.getSwapInOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_swapRuleV3_normalSwapOut(t *testing.T) {
	type args struct {
		committees                 []string
		substitutes                []string
		lenBeforeSlashedCommittees int
		lenSlashedCommittees       int
		numberOfFixedValidators    int
		minCommitteeSize           int
		maxSwapOutPercent          int
	}
	tests := []struct {
		name  string
		s     *swapRuleV3
		args  args
		want  []string
		want1 []string
	}{
		{
			name: "[valid input]",
			s:    &swapRuleV3{},
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key0,
				},
				substitutes: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				lenBeforeSlashedCommittees: 64,
				lenSlashedCommittees:       3,
				maxSwapOutPercent:          8,
				numberOfFixedValidators:    8,
				minCommitteeSize:           64,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key13, key14, key15, key16, key17, key18, key19,
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				key0,
			},
			want1: []string{
				key8, key9, key10, key11, key12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1 := s.normalSwapOut(tt.args.committees, tt.args.substitutes, tt.args.lenBeforeSlashedCommittees, tt.args.lenSlashedCommittees, tt.args.numberOfFixedValidators, tt.args.minCommitteeSize, tt.args.maxSwapOutPercent)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV3.normalSwapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapRuleV3.normalSwapOut() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_swapRuleV3_getNormalSwapOutOffset(t *testing.T) {
	type args struct {
		lenCommitteesBeforeSlash int
		lenSubstitutes           int
		lenSlashedCommittees     int
		maxSwapOutPercent        int
		numberOfFixedValidators  int
		minCommitteeSize         int
	}
	tests := []struct {
		name string
		s    *swapRuleV3
		args args
		want int
	}{
		{
			name: "slashed nodes >= committees / 3",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash: 25,
				lenSubstitutes:           10,
				lenSlashedCommittees:     10,
				maxSwapOutPercent:        8,
				numberOfFixedValidators:  8,
				minCommitteeSize:         64,
			},
			want: 0,
		},
		{

			name: "slashed nodes >= committees / 8",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash: 25,
				lenSubstitutes:           10,
				lenSlashedCommittees:     3,
				maxSwapOutPercent:        8,
				numberOfFixedValidators:  8,
				minCommitteeSize:         64,
			},
			want: 0,
		},
		{

			name: "slashed nodes < committees / 8 && lenCommitteesBeforeSlash < minCommitteeSize",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash: 25,
				lenSubstitutes:           10,
				lenSlashedCommittees:     2,
				maxSwapOutPercent:        8,
				numberOfFixedValidators:  8,
				minCommitteeSize:         64,
			},
			want: 0,
		},
		{
			name: "slashed nodes < committees / 8 && lenCommitteesBeforeSlash >= minCommitteeSize && lenSubstitutes == 0",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash: 64,
				lenSubstitutes:           0,
				lenSlashedCommittees:     5,
				maxSwapOutPercent:        8,
				numberOfFixedValidators:  8,
				minCommitteeSize:         64,
			},
			want: 0,
		},
		{

			name: "slashed nodes < committees / 8 && lenCommitteesBeforeSlash >= minCommitteeSize && lenSubstitutes > 0 ",
			s:    &swapRuleV3{},
			args: args{
				lenCommitteesBeforeSlash: 64,
				lenSubstitutes:           10,
				lenSlashedCommittees:     5,
				maxSwapOutPercent:        8,
				numberOfFixedValidators:  8,
				minCommitteeSize:         64,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.getNormalSwapOutOffset(tt.args.lenCommitteesBeforeSlash, tt.args.lenSubstitutes, tt.args.lenSlashedCommittees, tt.args.maxSwapOutPercent, tt.args.numberOfFixedValidators, tt.args.minCommitteeSize); got != tt.want {
				t.Errorf("swapRuleV3.getNormalSwapOutOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_swapRuleV3_slashingSwapOut(t *testing.T) {
	type args struct {
		committees              []string
		penalty                 map[string]signaturecounter.Penalty
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
			got, got1 := s.slashingSwapOut(tt.args.committees, tt.args.penalty, tt.args.numberOfFixedValidators, tt.args.maxSlashOutPercent)
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
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			if got := s.getSlashingOffset(tt.args.lenCommittees, tt.args.numberOfFixedValidators, tt.args.maxSlashOutPercent); got != tt.want {
				t.Errorf("swapRuleV3.getSlashingOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}
