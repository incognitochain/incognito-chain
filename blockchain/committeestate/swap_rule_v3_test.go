package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
)

func Test_swapRuleV3_Process(t *testing.T) {
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
		name                    string
		s                       *swapRuleV3
		args                    args
		want                    *instruction.SwapShardInstruction
		newCommittees           []string
		newSubstitutes          []string
		slashingCommittees      []string
		normalSwapOutCommittees []string
	}{
		//TODO: @hung add testcase
		// Testcase 2: SL = C/3 && NS = 0, NO SWAP_IN
		// Testcase 4: SL = C/3 && NS = 0, SWAP_IN = C/8, C_old = MAX_COMMITTEE_SIZE
		// Testcase 4: SL = C/3 && NS = 0, SWAP_IN = C/8, C_old < MAX_COMMITTEE_SIZE
		// Testcase 5: SL < C/3 && SL >= C/8 && NS = 0, SWAP_IN = c/8
		// Testcase 12: 0 < SL < C/8 && C < MAX_COMMITTEE_SIZE && NS = 0, SWAP_IN = c/8
		// Testcase 13: 0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && SUB = 0 && NS = 0, SWAP_IN = 0
		// Testcase 14: 0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && NS = C/8-SL, SUB >= c/8, SWAP_IN = c/8
		// Testcase 14: 0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && NS = C/8-SL && 0 < SUB < MAX_NS && NS = SUB
		// Testcase 15: 0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && NS = C - FixedValidator - SL && 0 < SUB < MAX_NS && NS = SUB
		// Testcase 16: 0 < SL < C/3 && SL < C/8 && C >= MAX_COMMITTEE_SIZE && NS = C/8-SL && SUB > MAX_NS && NS = MAX_NS
		// Testcase 17: 0 < SL < C/3 && SL < C/8 && C >= MAX_COMMITTEE_SIZE && NS = C - FixedValidator - SL && SUB > MAX_NS && NS = MAX_NS
		// Testcase 6: SL < C/3 && SL = 0 && C < MAX_COMMITTEE_SIZE && NS = 0, SWAP_IN = c/8
		// Testcase 7: SL < C/3 && SL = 0 && C >= MAX_COMMITTEE_SIZE && NS = C/8-SL && 0 < SUB < MAX_NS && NS = SUB
		// Testcase 8: SL < C/3 && SL = 0 && C >= MAX_COMMITTEE_SIZE && NS = C - FixedValidator - SL && 0 < SUB < MAX_NS && NS = SUB
		// Testcase 9: SL < C/3 && SL = 0 && C >= MAX_COMMITTEE_SIZE && SUB = 0 && NS = 0
		// Testcase 10: SL < C/3 && SL = 0 && C >= MAX_COMMITTEE_SIZE && NS = C/8-SL && SUB > MAX_NS && NS = MAX_NS
		// Testcase 11: SL < C/3 && SL = 0 && C >= MAX_COMMITTEE_SIZE && NS = C - FixedValidator - SL && SUB > MAX_NS && NS = MAX_NS
		// Add SI = c/8 or MAX_COMMITTEE_SIZE/8
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
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key0, key},
				[]string{key12, key8},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key9, key10, key11, key13, key14, key15, key16, key17, key18, key19,
				key0, key,
			},
			newSubstitutes: []string{
				key2, key3, key4, key5, key6, key7, key8, key9,
			},
			slashingCommittees: []string{
				key12,
			},
			normalSwapOutCommittees: []string{
				key8,
			},
		},
		{
			name: "SL = C/3 && NS = 0 && SWAP_IN = 0",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				},
				substitutes:             []string{},
				minCommitteeSize:        8,
				maxCommitteeSize:        20,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key10: signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key14: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
					key16: signaturecounter.Penalty{},
					key18: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{},
				[]string{key8, key10, key12, key14, key16, key18},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key9, key11, key13, key15, key17, key19,
			},
			newSubstitutes: []string{},
			slashingCommittees: []string{
				key8, key10, key12, key14, key16, key18,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "SL = C/3 && NS = 0, SWAP_IN C/8, C_old = MAX_COMMITTEE_SIZE",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23,
				},
				substitutes: []string{
					key24, key25, key26, key27, key28,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key10: signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key14: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
					key16: signaturecounter.Penalty{},
					key18: signaturecounter.Penalty{},
					key23: signaturecounter.Penalty{},
					key20: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key24, key25},
				[]string{key8, key10, key12, key14, key16, key18, key20, key23},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key9, key11, key13, key15, key17, key19,
				key21, key22, key24, key25,
			},
			newSubstitutes: []string{
				key26, key27, key28,
			},
			slashingCommittees: []string{
				key8, key10, key12, key14, key16, key18, key20, key23,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "SL = C/3 && NS = 0, SWAP_IN C/8, C_old < MAX_COMMITTEE_SIZE",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key23,
				},
				substitutes: []string{
					key24, key25, key26, key27, key28,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key10: signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key14: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
					key16: signaturecounter.Penalty{},
					key18: signaturecounter.Penalty{},
					key23: signaturecounter.Penalty{},
					key20: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key24},
				[]string{key8, key10, key12, key14, key16, key18, key20},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key9, key11, key13, key15, key17, key19, key23, key24,
			},
			newSubstitutes: []string{
				key25, key26, key27, key28,
			},
			slashingCommittees: []string{
				key8, key10, key12, key14, key16, key18, key20,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "SL < C/3 && SL > C/8 && NS = 0, SWAP_IN = c/8",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23,
				},
				substitutes: []string{
					key24, key25, key26, key27, key28,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
					key16: signaturecounter.Penalty{},
					key23: signaturecounter.Penalty{},
					key20: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key24, key25},
				[]string{key8, key12, key16, key20, key23},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key9,
				key10, key11, key13, key14, key15, key17, key18, key19,
				key21, key22, key24, key25,
			},
			newSubstitutes: []string{
				key26, key27, key28,
			},
			slashingCommittees: []string{
				key8, key12, key16, key20, key23,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "SL < C/3 && SL = C/8 && NS = 0, SWAP_IN = c/8",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23,
				},
				substitutes: []string{
					key24, key25, key26, key27, key28,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
					key16: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key24, key25},
				[]string{key8, key12, key16},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key9,
				key10, key11, key13, key14, key15, key17, key18, key19,
				key20, key21, key22, key23, key24, key25,
			},
			newSubstitutes: []string{
				key26, key27, key28,
			},
			slashingCommittees: []string{
				key8, key12, key16,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "0 < SL < C/8 && C < MAX_COMMITTEE_SIZE && NS = 0, SWAP_IN = c/8",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21,
				},
				substitutes: []string{
					key24, key25, key26, key27, key28,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key24, key25},
				[]string{key8, key12},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key9,
				key10, key11, key13, key14, key15, key16, key17, key18, key19,
				key20, key21, key24, key25,
			},
			newSubstitutes: []string{
				key26, key27, key28,
			},
			slashingCommittees: []string{
				key8, key12,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && SUB = 0 && NS = 0, SWAP_IN = 0",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23,
				},
				substitutes:             []string{},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{},
				[]string{key8, key12},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key9,
				key10, key11, key13, key14, key15, key16, key17, key18, key19,
				key20, key21, key22, key23,
			},
			newSubstitutes: []string{},
			slashingCommittees: []string{
				key8, key12,
			},
			normalSwapOutCommittees: []string{},
		},
		{
			name: "0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && NS = C/8-SL, SUB >= c/8, SWAP_IN = c/8",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23,
				},
				substitutes: []string{
					key24, key25, key26, key27, key28,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        24,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
					key8:  signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key24, key25},
				[]string{key8, key12, key9},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key10, key11, key13, key14, key15, key16, key17, key18, key19,
				key20, key21, key22, key23, key24, key25,
			},
			newSubstitutes: []string{
				key26, key27, key28,
			},
			slashingCommittees: []string{
				key8, key12,
			},
			normalSwapOutCommittees: []string{
				key9,
			},
		},
		{
			name: "0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && NS = C/8-SL && 0 < SUB < MAX_NS && NS = SUB",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
					key30, key31,
				},
				substitutes: []string{
					key32, key33,
				},
				minCommitteeSize:        8,
				maxCommitteeSize:        32,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 8,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key12: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key32, key33},
				[]string{key12, key8, key9},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7,
				key10, key11, key13, key14, key15, key16, key17, key18, key19,
				key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
				key30, key31, key32, key33,
			},
			newSubstitutes: []string{},
			slashingCommittees: []string{
				key12,
			},
			normalSwapOutCommittees: []string{
				key8, key9,
			},
		},
		{
			//TODO: fix get normal offset to pass this testcase
			name: "0 < SL < C/8 && C = MAX_COMMITTEE_SIZE && NS = C - FixedValidator - SL && 0 < SUB < MAX_NS && NS = SUB",
			s:    &swapRuleV3{},
			args: args{
				shardID: 1,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
					key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
					key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
					key30, key31,
				},
				substitutes: []string{
					key32, key33, key34,
				},
				minCommitteeSize:        29,
				maxCommitteeSize:        32,
				typeIns:                 instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidators: 29,
				penalty: map[string]signaturecounter.Penalty{
					key2:  signaturecounter.Penalty{},
					key3:  signaturecounter.Penalty{},
					key4:  signaturecounter.Penalty{},
					key30: signaturecounter.Penalty{},
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{key32},
				[]string{key30, key29, key31},
				1,
				instruction.SWAP_BY_END_EPOCH,
			),
			newCommittees: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
				key20, key21, key22, key23, key24, key25, key26, key27, key28, key32, key33,
			},
			newSubstitutes: []string{
				key34,
			},
			slashingCommittees: []string{
				key30,
			},
			normalSwapOutCommittees: []string{
				key29, key31,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &swapRuleV3{}
			got, got1, got2, got3, got4 := s.Process(tt.args.shardID, tt.args.committees, tt.args.substitutes, tt.args.minCommitteeSize, tt.args.maxCommitteeSize, tt.args.typeIns, tt.args.numberOfFixedValidators, tt.args.penalty)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapRuleV3.Process() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.newCommittees) {
				t.Errorf("swapRuleV3.Process() got1 = %v, want %v", got1, tt.newCommittees)
			}
			if !reflect.DeepEqual(got2, tt.newSubstitutes) {
				t.Errorf("swapRuleV3.Process() got2 = %v, want %v", got2, tt.newSubstitutes)
			}
			if !reflect.DeepEqual(got3, tt.slashingCommittees) {
				t.Errorf("swapRuleV3.Process() got3 = %v, want %v", got3, tt.slashingCommittees)
			}
			if !reflect.DeepEqual(got4, tt.normalSwapOutCommittees) {
				t.Errorf("swapRuleV3.Process() got4 = %v, want %v", got4, tt.normalSwapOutCommittees)
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
			if got := s.CalculateAssignOffset(tt.args.lenShardSubstitute, tt.args.lenCommittees, tt.args.numberOfFixedValidators, tt.args.minCommitteeSize); got != tt.want {
				t.Errorf("swapRuleV3.CalculateAssignOffset() = %v, want %v", got, tt.want)
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
		//TODO: @hung add testcase
		// Testcase 1: Max_Swap_In = 8  && Max_Swap_In > SUB => SI = SUB
		// Testcase 2: Max_Swap_In = 8  && Max_Swap_In < SUB => SI = Max_Swap_In
		// Testcase 3: Max_Swap_In = C/8 && Max_Swap_In > SUB => SI = SUB
		// Testcase 4: Max_Swap_In = C/8  && Max_Swap_In < SUB => SI = Max_Swap_In
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
			got, got1, got2 := s.swapInAfterSwapOut(tt.args.committees, tt.args.substitutes, tt.args.maxCommitteeSize, tt.args.maxSwapInPercent, 0, 0)
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
		//TODO: @hung add testcase
		// Testcase 1: Max_Swap_In = 8  && Max_Swap_In > SUB => SI = SUB
		// Testcase 2: Max_Swap_In = 8  && Max_Swap_In < SUB => SI = Max_Swap_In
		// Testcase 3: Max_Swap_In = C/8 && Max_Swap_In > SUB => SI = SUB
		// Testcase 4: Max_Swap_In = C/8  && Max_Swap_In < SUB => SI = Max_Swap_In
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
			if got := s.getSwapInOffset(tt.args.lenCommitteesAfterSwapOut, tt.args.lenSubstitutes, tt.args.maxSwapInPercent, tt.args.maxCommitteeSize, 0, 0); got != tt.want {
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
		//TODO: @hung add testcase
		// Testcase 1: SL > C/3 => NS = 0
		// Testcase 2: SL > C/8 => NS = 0
		// Testcase 3: SL < C/8 && SUB = 0 => NS = 0
		// Testcase 4: SL < C/8 && MAX_NS = C/8 - SL && SUB < MAX_NS => NS = SUB
		// Testcase 5: SL < C/8 && MAX_NS = C - FixedValidator - SL && SUB < MAX_NS => NS = SUB
		// Testcase 6: SL < C/8 && MAX_NS = C/8 - SL && SUB >= MAX_NS => NS = MAX_NS
		// Testcase 7: SL < C/8 && MAX_NS = C - FixedValidator - SL && SUB >= MAX_NS => NS = MAX_NS
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
			got, got1 := s.normalSwapOut(tt.args.committees, tt.args.substitutes, tt.args.lenBeforeSlashedCommittees, tt.args.lenSlashedCommittees, tt.args.numberOfFixedValidators, tt.args.minCommitteeSize, tt.args.maxSwapOutPercent, 0)
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
		//TODO: @hung add testcase
		// Testcase 1: SL > C/3 => NS = 0
		// Testcase 2: SL > C/8 => NS = 0
		// Testcase 3: SL < C/8 && SUB = 0 => NS = 0
		// Testcase 4: SL < C/8 && MAX_NS = C/8 - SL && SUB < MAX_NS => NS = SUB
		// Testcase 5: SL < C/8 && MAX_NS = C - FixedValidator - SL && SUB < MAX_NS => NS = SUB
		// Testcase 6: SL < C/8 && MAX_NS = C/8 - SL && SUB >= MAX_NS => NS = MAX_NS
		// Testcase 7: SL < C/8 && MAX_NS = C - FixedValidator - SL && SUB >= MAX_NS => NS = MAX_NS
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
			if got := s.getNormalSwapOutOffset(tt.args.lenCommitteesBeforeSlash, tt.args.lenSubstitutes, tt.args.lenSlashedCommittees, tt.args.maxSwapOutPercent, tt.args.numberOfFixedValidators, tt.args.minCommitteeSize, 0); got != tt.want {
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
		//TODO: @hung add testcase
		// Testcase 1: NO SLASH: fixed validators = committees, penalty > 0
		// Testcase 2: NO SLASH: fixed validators < committees, penalty = 0
		// Testcase 2: fixed validators + max_slashing_offset > lenCommittees, penalty > 0
		// Testcase 3: fixed validators + max_slashing_offset <= lenCommittees, penalty > 0
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
			name: "slashingOffset <= len(committees) / 3 && len(committees) > slashingOffset + numberOfFixedValidators",
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
			name: "slashingOffset <= len(committees) / 3 && len(committees) < slashingOffset + numberOfFixedValidators ",
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
			name: "fixed validators + max_slashing_offset > lenCommittees",
			s:    &swapRuleV3{},
			args: args{
				lenCommittees:           10,
				numberOfFixedValidators: 8,
				maxSlashOutPercent:      3,
			},
			want: 2,
		},
		{
			name: "fixed validators + max_slashing_offset <= lenCommittees",
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
			if got := s.getMaxSlashingOffset(tt.args.lenCommittees, tt.args.numberOfFixedValidators, tt.args.maxSlashOutPercent); got != tt.want {
				t.Errorf("swapRuleV3.getMaxSlashingOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateNewSubstitutePosition(t *testing.T) {
	type args struct {
		candidate string
		rand      int64
		total     int
	}
	tests := []struct {
		name    string
		args    args
		wantPos int
	}{
		//TODO: @hung add testcase
		// testcase 1: this function must be deterministic with the same parameters
		// testcase 2: make sure random offset is in valid range from 0 to len(substitute)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPos := calculateNewSubstitutePosition(tt.args.candidate, tt.args.rand, tt.args.total); gotPos != tt.wantPos {
				t.Errorf("calculateNewSubstitutePosition() = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}
