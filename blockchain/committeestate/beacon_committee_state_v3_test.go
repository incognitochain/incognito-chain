package committeestate

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
)

func TestBeaconCommitteeStateV3_processSwapShardInstruction(t *testing.T) {
	initTestParams()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreShardStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3, *incKey6,
			*incKey8, *incKey9, *incKey10, *incKey11,
			*incKey21, *incKey22, *incKey23, *incKey24, *incKey25, *incKey26, *incKey27, *incKey28, *incKey29,
			*incKey31, *incKey32, *incKey33, *incKey34, *incKey35, *incKey36,
			*incKey52, *incKey53, *incKey54, *incKey55, *incKey56, *incKey57, *incKey58, *incKey59},
		map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
			incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
			incKey31.GetIncKeyBase58(): paymentAddress, incKey32.GetIncKeyBase58(): paymentAddress, incKey33.GetIncKeyBase58(): paymentAddress, incKey34.GetIncKeyBase58(): paymentAddress, incKey35.GetIncKeyBase58(): paymentAddress, incKey36.GetIncKeyBase58(): paymentAddress,
			incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
		map[string]bool{
			key0:  true,
			key:   true,
			key2:  true,
			key3:  true,
			key6:  true,
			key8:  true,
			key9:  true,
			key10: false,
			key11: true,
			key21: true,
			key22: false,
			key23: false,
			key24: true,
			key25: true,
			key26: true,
			key27: true,
			key28: true,
			key29: true,
			key31: true,
			key32: false,
			key33: false,
			key34: true,
			key35: true,
			key36: true,
			key52: true,
			key53: true,
			key54: true,
			key55: true,
			key56: true,
			key57: true,
			key58: true,
			key59: true,
		},
		map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
			key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
			key31: *hash, key32: *hash, key33: *hash, key34: *hash, key35: *hash, key36: *hash,
			key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash})
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		swapShardInstruction     *instruction.SwapShardInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
		newState                 BeaconCommitteeState
	}
	tests := []struct {
		name                  string
		fields                fields
		fieldsAfterProcess    *BeaconCommitteeStateV3
		args                  args
		wantCommitteeChange   *CommitteeChange
		wantReturnInstruction *instruction.ReturnStakeInstruction
		wantErr               bool
	}{
		{
			name: "3 Slashing, 5 Normal Swap Out, 7 Swap In, 3 Back to Sub, 5 Return Inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key54, key56, key58, key59,
								key60, key61, key62, key63,
								key70, key71, key72, key73, key74, key75, key76,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								key21, key24, key25,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
							incKey21.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress,
							incKey54.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
							key21: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash,
							key54: *hash, key56: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key54: true,
							key56: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
				},
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key70, key71, key72, key73, key74, key75, key76},
					[]string{key53, key55, key57, key21, key22, key23, key24, key25},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				committeeChange: NewCommitteeChange(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardCommitteeRemoved(0, []string{key53, key55, key57, key21, key22, key23, key24, key25}).
				AddShardSubstituteAdded(0, []string{key21, key24, key25}).
				AddSlashingCommittees(0, []string{key53, key55, key57}).
				AddRemovedStakers([]string{key22, key23, key53, key55, key57}),
			wantReturnInstruction: instruction.NewReturnStakeInsWithValue(
				[]string{key22, key23, key53, key55, key57},
				[]string{hash.String(), hash.String(), hash.String(), hash.String(), hash.String()},
			),
			wantErr: false,
		},
		{
			name: "8 Slashing, 0 Normal Swap Out, 7 Swap In, 0 Back to Sub, 8 Return Inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key58, key59,
								key60, key61, key62, key63,
								key70, key71, key72, key73, key74, key75, key76,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
							incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
							incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
							key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
							key52: *hash, key53: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key8:  signaturecounter.NewPenalty(),
						key9:  signaturecounter.NewPenalty(),
						key10: signaturecounter.NewPenalty(),
						key11: signaturecounter.NewPenalty(),
						key21: signaturecounter.NewPenalty(),
						key22: signaturecounter.NewPenalty(),
						key23: signaturecounter.NewPenalty(),
						key24: signaturecounter.NewPenalty(),
						key54: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key56: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
				},
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key70, key71, key72, key73, key74, key75, key76},
					[]string{key21, key22, key23, key24, key54, key55, key56, key57},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				committeeChange: NewCommitteeChange(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardCommitteeRemoved(0, []string{key21, key22, key23, key24, key54, key55, key56, key57}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddSlashingCommittees(0, []string{key21, key22, key23, key24, key54, key55, key56, key57}).
				AddRemovedStakers([]string{key21, key22, key23, key24, key54, key55, key56, key57}),
			wantReturnInstruction: instruction.NewReturnStakeInsWithValue(
				[]string{key21, key22, key23, key24, key54, key55, key56, key57},
				[]string{hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String()},
			),
			wantErr: false,
		},
		{
			name: "0 Slashing, 8 Normal Swap Out, 8 Swap In, 6 Back to Sub, 2 Return Inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
								key70, key71, key72, key73, key74, key75, key76, key77,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								key21, key24, key25, key26, key27, key28,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
							incKey21.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
							incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
							key21: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
							key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key10: signaturecounter.NewPenalty(),
						key11: signaturecounter.NewPenalty(),
						key12: signaturecounter.NewPenalty(),
						key13: signaturecounter.NewPenalty(),
					},
				},
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key70, key71, key72, key73, key74, key75, key76, key77},
					[]string{key21, key22, key23, key24, key25, key26, key27, key28},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				committeeChange: NewCommitteeChange(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76, key77}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76, key77}).
				AddShardSubstituteAdded(0, []string{key21, key24, key25, key26, key27, key28}).
				AddShardCommitteeRemoved(0, []string{key21, key22, key23, key24, key25, key26, key27, key28}).
				AddRemovedStakers([]string{key22, key23}),
			wantReturnInstruction: instruction.NewReturnStakeInsWithValue(
				[]string{key22, key23},
				[]string{hash.String(), hash.String()},
			),
			wantErr: false,
		},
		{
			name: "0 Slashing, 0 Normal Swap Out, 7 Swap In, 0 Back to Sub, 0 Return Inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
								key70, key71, key72, key73, key74, key75, key76,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key77, key78, key79,
								key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
							incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
							incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
							key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
							key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key10: signaturecounter.NewPenalty(),
						key11: signaturecounter.NewPenalty(),
						key12: signaturecounter.NewPenalty(),
						key13: signaturecounter.NewPenalty(),
					},
				},
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key70, key71, key72, key73, key74, key75, key76},
					[]string{},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				committeeChange: NewCommitteeChange(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76}),
			wantReturnInstruction: instruction.NewReturnStakeIns(),
			wantErr:               false,
		},
		{
			name: "max 21 Slashing, 0 Normal Swap Out, 0 Swap In, 0 Back to Sub, 21 Return Inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
							incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
							incKey31.GetIncKeyBase58(): paymentAddress, incKey32.GetIncKeyBase58(): paymentAddress, incKey33.GetIncKeyBase58(): paymentAddress, incKey34.GetIncKeyBase58(): paymentAddress, incKey35.GetIncKeyBase58(): paymentAddress, incKey36.GetIncKeyBase58(): paymentAddress,
							incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
							key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
							key31: *hash, key32: *hash, key33: *hash, key34: *hash, key35: *hash, key36: *hash,
							key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key31: true,
							key32: false,
							key33: false,
							key34: true,
							key35: true,
							key36: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20,
								key30, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
							incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
							key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key21: signaturecounter.NewPenalty(), key22: signaturecounter.NewPenalty(), key23: signaturecounter.NewPenalty(), key24: signaturecounter.NewPenalty(), key25: signaturecounter.NewPenalty(), key26: signaturecounter.NewPenalty(), key27: signaturecounter.NewPenalty(), key28: signaturecounter.NewPenalty(), key29: signaturecounter.NewPenalty(),
						key31: signaturecounter.NewPenalty(), key32: signaturecounter.NewPenalty(), key33: signaturecounter.NewPenalty(), key34: signaturecounter.NewPenalty(), key35: signaturecounter.NewPenalty(), key36: signaturecounter.NewPenalty(),
						key52: signaturecounter.NewPenalty(), key53: signaturecounter.NewPenalty(), key54: signaturecounter.NewPenalty(), key55: signaturecounter.NewPenalty(), key56: signaturecounter.NewPenalty(), key57: signaturecounter.NewPenalty(),
						key61: signaturecounter.NewPenalty(), key62: signaturecounter.NewPenalty(), key63: signaturecounter.NewPenalty(),
					},
				},
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{},
					[]string{
						key21, key22, key23, key24, key25, key26, key27, key28, key29,
						key31, key32, key33, key34, key35, key36,
						key52, key53, key54, key55, key56, key57,
					},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				committeeChange: NewCommitteeChange(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key21, key22, key23, key24, key25, key26, key27, key28, key29,
					key31, key32, key33, key34, key35, key36,
					key52, key53, key54, key55, key56, key57}).
				AddRemovedStakers([]string{key21, key22, key23, key24, key25, key26, key27, key28, key29,
					key31, key32, key33, key34, key35, key36,
					key52, key53, key54, key55, key56, key57}).
				AddSlashingCommittees(0, []string{
					key21, key22, key23, key24, key25, key26, key27, key28, key29,
					key31, key32, key33, key34, key35, key36,
					key52, key53, key54, key55, key56, key57}),
			wantReturnInstruction: instruction.NewReturnStakeInsWithValue(
				[]string{
					key21, key22, key23, key24, key25, key26, key27, key28, key29,
					key31, key32, key33, key34, key35, key36,
					key52, key53, key54, key55, key56, key57},
				[]string{hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(),
					hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(),
					hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String(), hash.String()},
			),
			wantErr: false,
		},
		{
			name: "0 Slashing, 0 Normal Swap Out (max committee), 0 Swap In, 0 Back to Sub, 0 Return Inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
								key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
								key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
								key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
								key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
								key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
								key60, key61, key62, key63,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
						stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  true,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: false,
							key11: true,
							key21: true,
							key22: false,
							key23: false,
							key24: true,
							key25: true,
							key26: true,
							key27: true,
							key28: true,
							key29: true,
							key52: true,
							key53: true,
							key54: true,
							key55: true,
							key56: true,
							key57: true,
							key58: true,
							key59: true,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					MissingSignaturePenalty:          map[string]signaturecounter.Penalty{},
				},
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{},
					[]string{},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				committeeChange: NewCommitteeChange(),
			},
			wantCommitteeChange:   NewCommitteeChange(),
			wantReturnInstruction: &instruction.ReturnStakeInstruction{},
			wantErr:               false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			tt.args.env.shardCommittee = b.shardCommittee
			tt.args.env.shardSubstitute = b.shardSubstitute
			got, got1, err := b.processSwapShardInstruction(tt.args.swapShardInstruction, tt.args.env, tt.args.committeeChange, tt.args.returnStakingInstruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantCommitteeChange) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() got = %v, want %v", got, tt.wantCommitteeChange)
			}
			if !reflect.DeepEqual(got1, tt.wantReturnInstruction) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() got1 = %v, want %v", got1, tt.wantReturnInstruction)
			}
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() fieldsAfterProcess got = %v, want %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAssignWithRandomInstruction(t *testing.T) {

	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		rand              int64
		numberOfValidator []int
		committeeChange   *CommitteeChange
		oldState          BeaconCommitteeState
		beaconHeight      uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "valid input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
					},
					shardCommonPool: []string{
						key0, key, key2, key3,
					},
					numberOfAssignedCandidates: 2,
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
					},
					shardCommonPool: []string{
						key2, key3,
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]string{
					0: []string{
						key,
					},
					1: []string{
						key0,
					},
				},
			},
			args: args{
				rand:              1000,
				numberOfValidator: []int{8, 8},
				committeeChange: &CommitteeChange{
					SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{},
				},
				beaconHeight: 1000,
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey,
				},
				SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey},
					1: []incognitokey.CommitteePublicKey{*incKey0},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.processAssignWithRandomInstruction(tt.args.rand, tt.args.numberOfValidator, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processAssignWithRandomInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignRandomlyToSubstituteList(t *testing.T) {

	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		candidates      []string
		rand            int64
		shardID         byte
		committeeChange *CommitteeChange
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "substitute list is empty",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key2},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}),
		},
		{
			name: "substitute list is not empty",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key3},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key2, key3},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}),
		},
		{
			name: "substitute list is empty, > 1 candidate",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key6, key4, key2},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2, key4, key6},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2, key4, key6}),
		}, {
			name: "substitute list is not empty, > 1 candidate",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key5},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key2, key6, key4, key5},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2, key4, key6},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2, key4, key6}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignRandomlyToSubstituteList(tt.args.candidates, tt.args.rand, tt.args.shardID, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToPending() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardSubstitute, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute) {
				t.Errorf("BeaconCommitteeStateV3.assignToPending() b = %v, tt.fieldsAfterProcess %v", b.shardSubstitute, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignToSync(t *testing.T) {
	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		shardID         byte
		candidates      []string
		committeeChange *CommitteeChange
		beaconHeight    uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "not empty list candidates",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
				},
				syncPool: map[byte][]string{
					0: []string{
						key0, key, key2, key3,
					},
					1: []string{
						key4, key5, key6, key7,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
				},
				syncPool: map[byte][]string{
					0: []string{
						key0, key, key2, key3,
					},
					1: []string{
						key4, key5, key6, key7, key8, key9, key10,
					},
				},
			},
			args: args{
				shardID:    1,
				candidates: []string{key8, key9, key10},
				committeeChange: &CommitteeChange{
					SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
						1: []incognitokey.CommitteePublicKey{},
					},
				},
				beaconHeight: 1000,
			},
			want: &CommitteeChange{
				SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey8, *incKey9, *incKey10,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignToSyncPool(tt.args.shardID, tt.args.candidates, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToSyncPool() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.assignToSyncPool() b = %v, tt.fieldsAfterProcess %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_clone(t *testing.T) {
	initTestParams()

	mutex := &sync.RWMutex{}
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	txHash, _ := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
		finishedSyncValidators           map[byte][]string
	}
	tests := []struct {
		name   string
		fields fields
		want   *BeaconCommitteeStateV3
	}{
		{
			name: "[Valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{
							key0, key, key2, key3,
						},
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						autoStake: map[string]bool{
							key:  true,
							key0: false,
							key2: true,
							key3: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():  paymentAddress,
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey3.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *txHash,
							key0: *txHash,
							key2: *txHash,
							key3: *txHash,
						},
						hashes: NewBeaconCommitteeStateHash(),
						mu:     mutex,
					},
					shardCommonPool: []string{
						key, key0, key2, key3,
					},
					numberOfAssignedCandidates: 1,
					swapRule:                   NewSwapRuleV3(),
				},
				syncPool: map[byte][]string{
					0: []string{
						key, key0, key2, key3,
					},
					1: []string{
						key, key0, key2, key3,
					},
				},
				finishedSyncValidators: map[byte][]string{
					0: []string{
						key, key0, key2, key3,
					},
					1: []string{
						key, key0, key2, key3,
					},
				},
			},
			want: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{
							key0, key, key2, key3,
						},
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						autoStake: map[string]bool{
							key:  true,
							key0: false,
							key2: true,
							key3: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():  paymentAddress,
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey3.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *txHash,
							key0: *txHash,
							key2: *txHash,
							key3: *txHash,
						},
						hashes: NewBeaconCommitteeStateHash(),
						mu:     mutex,
					},
					shardCommonPool: []string{
						key, key0, key2, key3,
					},
					numberOfAssignedCandidates: 1,
					swapRule:                   NewSwapRuleV3(),
				},
				syncPool: map[byte][]string{
					0: []string{
						key, key0, key2, key3,
					},
					1: []string{
						key, key0, key2, key3,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processUnstakeInstruction(t *testing.T) {

	initTestParams()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreShardStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey, *incKey0, *incKey2, *incKey4, *incKey6, *incKey8, *incKey10, *incKey12,
		},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():   paymentAddress,
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey4.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key:   true,
			key0:  true,
			key2:  true,
			key4:  true,
			key6:  true,
			key8:  true,
			key10: true,
			key12: true,
		},
		map[string]common.Hash{
			key:   *hash,
			key0:  *hash,
			key2:  *hash,
			key4:  *hash,
			key6:  *hash,
			key8:  *hash,
			key10: *hash,
			key12: *hash,
		},
	)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
		finishedSyncValidators           map[byte][]string
	}
	type args struct {
		unstakeInstruction       *instruction.UnstakeInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
	}
	tests := []struct {
		name                string
		fields              fields
		fieldsAfterProcess  *BeaconCommitteeStateV3
		args                args
		wantCommitteeChange *CommitteeChange
		wantReturnStakeInst *instruction.ReturnStakeInstruction
		wantErr             bool
	}{
		{
			name: "Unstake next candidate > 0, current candidate > 0, turn off auto stake flag > 0",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule:      NewAssignRuleV2(),
					shardCommonPool: []string{key2, key, key0},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key:   true,
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():   paymentAddress,
							incKey0.GetIncKeyBase58():  paymentAddress,
							incKey2.GetIncKeyBase58():  paymentAddress,
							incKey4.GetIncKeyBase58():  paymentAddress,
							incKey6.GetIncKeyBase58():  paymentAddress,
							incKey8.GetIncKeyBase58():  paymentAddress,
							incKey10.GetIncKeyBase58(): paymentAddress,
							incKey12.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:   *hash,
							key0:  *hash,
							key2:  *hash,
							key4:  *hash,
							key6:  *hash,
							key8:  *hash,
							key10: *hash,
							key12: *hash,
						},
					},
					numberOfAssignedCandidates: 1,
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule:      NewAssignRuleV2(),
					shardCommonPool: []string{key2},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key2:  false,
							key4:  false,
							key6:  false,
							key8:  false,
							key10: false,
							key12: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey2.GetIncKeyBase58():  paymentAddress,
							incKey4.GetIncKeyBase58():  paymentAddress,
							incKey6.GetIncKeyBase58():  paymentAddress,
							incKey8.GetIncKeyBase58():  paymentAddress,
							incKey10.GetIncKeyBase58(): paymentAddress,
							incKey12.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key2:  *hash,
							key4:  *hash,
							key6:  *hash,
							key8:  *hash,
							key10: *hash,
							key12: *hash,
						},
					},
					numberOfAssignedCandidates: 1,
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			args: args{
				unstakeInstruction: instruction.NewUnstakeInstructionWithValue([]string{key, key0, key2, key4, key6, key8, key10, key12}),
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				committeeChange:          NewCommitteeChange(),
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddStopAutoStakes([]string{key2, key4, key6, key8, key10, key12}).
				AddNextEpochShardCandidateRemoved([]string{key, key0}).
				AddRemovedStakers([]string{key, key0}),
			wantReturnStakeInst: instruction.NewReturnStakeInsWithValue(
				[]string{key, key0},
				[]string{hash.String(), hash.String()},
			),
		},
		{
			name: "Unstake next candidate = 0, current candidate = 0, turn off auto stake flag > 0",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule:      NewAssignRuleV2(),
					shardCommonPool: []string{key2, key, key0},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key:   true,
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():   paymentAddress,
							incKey0.GetIncKeyBase58():  paymentAddress,
							incKey2.GetIncKeyBase58():  paymentAddress,
							incKey4.GetIncKeyBase58():  paymentAddress,
							incKey6.GetIncKeyBase58():  paymentAddress,
							incKey8.GetIncKeyBase58():  paymentAddress,
							incKey10.GetIncKeyBase58(): paymentAddress,
							incKey12.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:   *hash,
							key0:  *hash,
							key2:  *hash,
							key4:  *hash,
							key6:  *hash,
							key8:  *hash,
							key10: *hash,
							key12: *hash,
						},
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule:      NewAssignRuleV2(),
					shardCommonPool: []string{key2, key, key0},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key:   true,
							key0:  true,
							key2:  true,
							key4:  false,
							key6:  false,
							key8:  false,
							key10: false,
							key12: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():   paymentAddress,
							incKey0.GetIncKeyBase58():  paymentAddress,
							incKey2.GetIncKeyBase58():  paymentAddress,
							incKey4.GetIncKeyBase58():  paymentAddress,
							incKey6.GetIncKeyBase58():  paymentAddress,
							incKey8.GetIncKeyBase58():  paymentAddress,
							incKey10.GetIncKeyBase58(): paymentAddress,
							incKey12.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:   *hash,
							key0:  *hash,
							key2:  *hash,
							key4:  *hash,
							key6:  *hash,
							key8:  *hash,
							key10: *hash,
							key12: *hash,
						},
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			args: args{
				unstakeInstruction: instruction.NewUnstakeInstructionWithValue([]string{key4, key6, key8, key10, key12}),
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				committeeChange:          NewCommitteeChange(),
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddStopAutoStakes([]string{key4, key6, key8, key10, key12}),
			wantReturnStakeInst: instruction.NewReturnStakeIns(),
		},
		{
			name: "Unstake next candidate > 0, current candidate = 0, turn off auto stake flag = 0",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule:      NewAssignRuleV2(),
					shardCommonPool: []string{key2, key, key0},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key:   true,
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():   paymentAddress,
							incKey0.GetIncKeyBase58():  paymentAddress,
							incKey2.GetIncKeyBase58():  paymentAddress,
							incKey4.GetIncKeyBase58():  paymentAddress,
							incKey6.GetIncKeyBase58():  paymentAddress,
							incKey8.GetIncKeyBase58():  paymentAddress,
							incKey10.GetIncKeyBase58(): paymentAddress,
							incKey12.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:   *hash,
							key0:  *hash,
							key2:  *hash,
							key4:  *hash,
							key6:  *hash,
							key8:  *hash,
							key10: *hash,
							key12: *hash,
						},
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule:      NewAssignRuleV2(),
					shardCommonPool: []string{key},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key:   true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():   paymentAddress,
							incKey4.GetIncKeyBase58():  paymentAddress,
							incKey6.GetIncKeyBase58():  paymentAddress,
							incKey8.GetIncKeyBase58():  paymentAddress,
							incKey10.GetIncKeyBase58(): paymentAddress,
							incKey12.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:   *hash,
							key4:  *hash,
							key6:  *hash,
							key8:  *hash,
							key10: *hash,
							key12: *hash,
						},
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			args: args{
				unstakeInstruction: instruction.NewUnstakeInstructionWithValue([]string{key0, key2}),
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				committeeChange:          NewCommitteeChange(),
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddNextEpochShardCandidateRemoved([]string{key0, key2}).
				AddRemovedStakers([]string{key0, key2}),
			wantReturnStakeInst: instruction.NewReturnStakeInsWithValue(
				[]string{key0, key2},
				[]string{hash.String(), hash.String()},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			b.addDataToEnvironment(tt.args.env)
			got, got1, err := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange, tt.args.returnStakingInstruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("processUnstakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantCommitteeChange) {
				t.Errorf("BeaconCommitteeStateV3.processUnstakeInstruction() gotCommitteeChange = %v, want %v", got, tt.wantCommitteeChange)
			}
			if !reflect.DeepEqual(got1, tt.wantReturnStakeInst) {
				t.Errorf("BeaconCommitteeStateV3.processUnstakeInstruction() gotReturnStakeInst = %v, want %v", got1, tt.wantReturnStakeInst)
			}
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("BeaconCommitteeStateV3.processUnstakeInstruction() fieldsAfterProcess = %v, want %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_UpdateCommitteeState(t *testing.T) {
	initTestParams()

	finalMutex := &sync.RWMutex{}

	type fields struct {
		BeaconCommitteeStateV3 *BeaconCommitteeStateV3
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	wl := wallet.KeyWallet{}
	wl.KeySet.PaymentAddress = paymentAddress
	paymentAddressString := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreShardStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3, *incKey6,
			*incKey8, *incKey9, *incKey10, *incKey11,
			*incKey21, *incKey22, *incKey23, *incKey24, *incKey25, *incKey26, *incKey27, *incKey28, *incKey29,
			*incKey31, *incKey32, *incKey33, *incKey34, *incKey35, *incKey36,
			*incKey52, *incKey53, *incKey54, *incKey55, *incKey56, *incKey57, *incKey58, *incKey59},
		map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
			incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
			incKey31.GetIncKeyBase58(): paymentAddress, incKey32.GetIncKeyBase58(): paymentAddress, incKey33.GetIncKeyBase58(): paymentAddress, incKey34.GetIncKeyBase58(): paymentAddress, incKey35.GetIncKeyBase58(): paymentAddress, incKey36.GetIncKeyBase58(): paymentAddress,
			incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
		map[string]bool{
			key0:  true,
			key:   true,
			key2:  true,
			key3:  true,
			key6:  true,
			key8:  true,
			key9:  true,
			key10: false,
			key11: true,
			key21: true,
			key22: false,
			key23: false,
			key24: true,
			key25: true,
			key26: true,
			key27: true,
			key28: true,
			key29: true,
			key31: true,
			key32: false,
			key33: false,
			key34: true,
			key35: true,
			key36: true,
			key52: true,
			key53: true,
			key54: true,
			key55: true,
			key56: true,
			key57: true,
			key58: true,
			key59: true,
		},
		map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
			key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
			key31: *hash, key32: *hash, key33: *hash, key34: *hash, key35: *hash, key36: *hash,
			key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash})

	randomInstructionCommitteeChange := NewCommitteeChange()
	randomInstructionCommitteeChange.NextEpochShardCandidateRemoved =
		append(randomInstructionCommitteeChange.NextEpochShardCandidateRemoved, []incognitokey.CommitteePublicKey{*incKey0, *incKey}...)
	randomInstructionCommitteeChange.SyncingPoolAdded[0] = []incognitokey.CommitteePublicKey{*incKey}
	randomInstructionCommitteeChange.SyncingPoolAdded[1] = []incognitokey.CommitteePublicKey{*incKey0}

	tests := []struct {
		name                  string
		fields                fields
		fieldsAfterProcess    fields
		args                  args
		want                  *BeaconCommitteeStateHash
		wantCommitteeChange   *CommitteeChange
		wantReturnInstruction [][]string
		wantErr               bool
	}{
		{
			name: "Process Random Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 2,
					},
					syncPool: map[byte][]string{
						0: []string{},
						1: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key},
						1: []string{key0},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"800000",
							"120000",
							"350000",
							"190000",
						},
					},
				},
			},
			wantCommitteeChange:   randomInstructionCommitteeChange,
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Unstake Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							autoStake: map[string]bool{
								key:   true,
								key10: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58():   paymentAddress,
								incKey10.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key:   *hash,
								key10: *hash,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3, key4, key5, key6, key7, key8,
							key9, key10, key11, key12, key13, key14, key15,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:             finalMutex,
							autoStake:      map[string]bool{},
							stakingTx:      map[string]common.Hash{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
						},
						shardCommonPool: []string{
							key0, key2, key3, key4, key5, key6, key7, key8,
							key9, key11, key12, key13, key14, key15,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards:     2,
					ConsensusStateDB: sDB,
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							strings.Join([]string{key, key10}, instruction.SPLITTER),
						},
					},
					newUnassignedCommonPool: []string{
						key0, key, key2, key3, key4, key5, key6, key7, key8,
						key9, key10, key11, key12, key13, key14, key15,
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddNextEpochShardCandidateRemoved([]string{key, key10}).
				AddRemovedStakers([]string{key, key10}),
			wantReturnInstruction: [][]string{
				instruction.NewReturnStakeInsWithValue([]string{key, key10}, []string{hash.String(), hash.String()}).ToString(),
			},
			wantErr: false,
		},
		{
			name: "Process Finish Sync Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key15, key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						instruction.NewFinishSyncInstructionWithValue(1, []string{key15}).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddFinishedSyncValidators(1, []string{key15}).
				AddSyncingPoolRemoved(1, []string{key15}).
				AddShardSubstituteAdded(1, []string{key15}),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Stop Auto Stake Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						instruction.NewStopAutoStakeInstructionWithValue([]string{key7}).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddStopAutoStake(key7),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Stake Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:   true,
								key7:  false,
								key20: true,
								key21: false,
								key22: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey20.GetIncKeyBase58(): paymentAddress,
								incKey21.GetIncKeyBase58(): paymentAddress,
								incKey22.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key20: *hash,
								key21: *hash,
								key22: *hash,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3, key20, key21, key22,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						instruction.NewStakeInstructionWithValue(
							[]string{key20, key21, key22},
							instruction.SHARD_INST,
							[]string{hash.String(), hash.String(), hash.String()},
							[]string{paymentAddressString, paymentAddressString, paymentAddressString},
							[]bool{true, false, true},
							[]string{"", "", ""},
						).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddNextEpochShardCandidateAdded([]string{key20, key21, key22}),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Swap Shard Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key54, key56, key58, key59,
									key60, key61, key62, key63,
									key70, key71, key72, key73, key74, key75, key76,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
									key21, key24, key25,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
								incKey21.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress,
								incKey54.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
							stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
								key21: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash,
								key54: *hash, key56: *hash, key58: *hash, key59: *hash},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key54: true,
								key56: true,
								key58: true,
								key59: true,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewSwapShardInstructionWithValue(
							[]string{key70, key71, key72, key73, key74, key75, key76},
							[]string{key53, key55, key57, key21, key22, key23, key24, key25},
							0,
							instruction.SWAP_BY_END_EPOCH,
						).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardCommitteeRemoved(0, []string{key53, key55, key57, key21, key22, key23, key24, key25}).
				AddShardSubstituteAdded(0, []string{key21, key24, key25}).
				AddSlashingCommittees(0, []string{key53, key55, key57}).
				AddRemovedStakers([]string{key22, key23, key53, key55, key57}),
			wantReturnInstruction: [][]string{
				instruction.NewReturnStakeInsWithValue(
					[]string{key22, key23, key53, key55, key57},
					[]string{hash.String(), hash.String(), hash.String(), hash.String(), hash.String()},
				).ToString(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.BeaconCommitteeStateV3
			_, got1, got2, err := b.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.wantCommitteeChange) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.wantCommitteeChange)
			}
			if !reflect.DeepEqual(got2, tt.wantReturnInstruction) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.wantReturnInstruction)
			}
			gotCommitteeState := b
			wantCommitteeState := tt.fieldsAfterProcess.BeaconCommitteeStateV3
			if !reflect.DeepEqual(gotCommitteeState, wantCommitteeState) {
				t.Errorf("fieldsAfterProcess, got = %v, want %v", gotCommitteeState, wantCommitteeState)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_UpdateCommitteeState_MultipleInstructions(t *testing.T) {

	initTestParams()

	finalMutex := &sync.RWMutex{}

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	wl := wallet.KeyWallet{}
	wl.KeySet.PaymentAddress = paymentAddress
	paymentAddressString := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreShardStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3, *incKey6,
			*incKey8, *incKey9, *incKey10, *incKey11,
			*incKey21, *incKey22, *incKey23, *incKey24, *incKey25, *incKey26, *incKey27, *incKey28, *incKey29,
			*incKey31, *incKey32, *incKey33, *incKey34, *incKey35, *incKey36,
			*incKey52, *incKey53, *incKey54, *incKey55, *incKey56, *incKey57, *incKey58, *incKey59,
			*incKey91, *incKey92, *incKey93, *incKey94, *incKey95},
		map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
			incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
			incKey31.GetIncKeyBase58(): paymentAddress, incKey32.GetIncKeyBase58(): paymentAddress, incKey33.GetIncKeyBase58(): paymentAddress, incKey34.GetIncKeyBase58(): paymentAddress, incKey35.GetIncKeyBase58(): paymentAddress, incKey36.GetIncKeyBase58(): paymentAddress,
			incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress,
			incKey91.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
		map[string]bool{
			key0:  true,
			key:   true,
			key2:  true,
			key3:  true,
			key6:  true,
			key8:  true,
			key9:  true,
			key10: false,
			key11: true,
			key21: true,
			key22: false,
			key23: false,
			key24: true,
			key25: true,
			key26: true,
			key27: true,
			key28: true,
			key29: true,
			key31: true,
			key32: false,
			key33: false,
			key34: true,
			key35: true,
			key36: true,
			key52: true,
			key53: true,
			key54: true,
			key55: true,
			key56: true,
			key57: true,
			key58: true,
			key59: true,
			key91: true,
			key92: true,
			key93: true,
			key94: true,
			key95: true,
		},
		map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
			key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
			key31: *hash, key32: *hash, key33: *hash, key34: *hash, key35: *hash, key36: *hash,
			key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash,
			key91: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash})

	type fields struct {
		BeaconCommitteeStateV3 *BeaconCommitteeStateV3
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name                  string
		fields                fields
		fieldsAfterProcess    fields
		args                  args
		want                  *BeaconCommitteeStateHash
		wantCommitteeChange   *CommitteeChange
		wantReturnInstruction [][]string
		wantErr               bool
	}{
		{
			name: "Process Random,Assign and Stop Auto Stake Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
								key92: true,
								key93: true,
								key94: true,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93, key94, key95, key96, key97, key98, key99,
							key100, key101, key102, key103, key104, key105, key106, key107, key108, key109,
						},
						numberOfAssignedCandidates: 9,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113},
						1: []string{key114, key115},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
								key92: false,
								key93: false,
								key94: false,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key99,
							key100, key101, key102, key103, key104, key105, key106, key107, key108, key109,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113, key91},
						1: []string{key114, key115, key90, key92, key93, key94, key95, key96, key97, key98},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewRandomInstructionWithValue(1000).ToString(),
						instruction.NewStopAutoStakeInstructionWithValue([]string{key92, key94, key93}).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddNextEpochShardCandidateRemoved([]string{key90, key91, key92, key93, key94, key95, key96, key97, key98}).
				AddSyncingPoolAdded(0, []string{key91}).
				AddSyncingPoolAdded(1, []string{key90, key92, key93, key94, key95, key96, key97, key98}).
				AddStopAutoStakes([]string{key92, key94, key93}),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Finish Sync and Stop Auto Stake Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key91: *hash, key92: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey91.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
								key91: true,
								key92: true,
								key93: true,
								key94: true,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key99,
							key100, key101, key102, key103, key104, key105, key106, key107, key108, key109,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113, key93, key98},
						1: []string{key114, key115, key90, key91, key92, key94, key95, key96, key97},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key98, key71, key72, key73, key74, key75, key112, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key97, key114, key12, key13, key91, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key91: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey91.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
								key91: false,
								key92: true,
								key93: false,
								key94: false,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key99,
							key100, key101, key102, key103, key104, key105, key106, key107, key108, key109,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key113, key93},
						1: []string{key115, key90, key92, key94, key95, key96},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewFinishSyncInstructionWithValue(0, []string{key112, key98}).ToString(),
						instruction.NewFinishSyncInstructionWithValue(1, []string{key114, key97, key91}).ToString(),
						instruction.NewStopAutoStakeInstructionWithValue([]string{key91, key94, key93}).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key112, key98}).
				AddFinishedSyncValidators(0, []string{key112, key98}).
				AddShardSubstituteAdded(0, []string{key112, key98}).
				AddSyncingPoolRemoved(1, []string{key114, key97, key91}).
				AddFinishedSyncValidators(1, []string{key114, key97, key91}).
				AddShardSubstituteAdded(1, []string{key114, key97, key91}).
				AddStopAutoStakes([]string{key91, key94, key93}),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Swap (Slash + Normal SO + SI), Stop Auto Stake",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key54, key56, key58, key59,
									key60, key61, key62, key63,
									key70, key71, key72, key73, key74, key75, key76,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
									key21, key24, key25,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
								incKey21.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress,
								incKey54.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
							stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash,
								key21: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash,
								key54: *hash, key56: *hash, key58: *hash, key59: *hash},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: false,
								key24: false,
								key25: false,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key54: true,
								key56: true,
								key58: true,
								key59: true,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewSwapShardInstructionWithValue(
							[]string{key70, key71, key72, key73, key74, key75, key76},
							[]string{key53, key55, key57, key21, key22, key23, key24, key25},
							0,
							instruction.SWAP_BY_END_EPOCH,
						).ToString(),
						instruction.NewStopAutoStakeInstructionWithValue([]string{key21, key25, key24}).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardCommitteeRemoved(0, []string{key53, key55, key57, key21, key22, key23, key24, key25}).
				AddShardSubstituteAdded(0, []string{key21, key24, key25}).
				AddSlashingCommittees(0, []string{key53, key55, key57}).
				AddRemovedStakers([]string{key22, key23, key53, key55, key57}).
				AddStopAutoStakes([]string{key21, key25, key24}),
			wantReturnInstruction: [][]string{
				instruction.NewReturnStakeInsWithValue(
					[]string{key22, key23, key53, key55, key57},
					[]string{hash.String(), hash.String(), hash.String(), hash.String(), hash.String()},
				).ToString(),
			},
			wantErr: false,
		},
		{
			name: "Process Finish Sync and Swap Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113, key93, key98},
						1: []string{key114, key115, key90, key91, key92, key94, key95, key96, key97},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key54, key56, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key98, key112},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key54: *hash, key56: *hash, key58: *hash, key59: *hash},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key54: true,
								key56: true,
								key58: true,
								key59: true,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key113, key93},
						1: []string{key114, key115, key90, key91, key92, key94, key95, key96, key97},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1000,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewFinishSyncInstructionWithValue(0, []string{key112, key98}).ToString(),
						instruction.NewSwapShardInstructionWithValue(
							[]string{},
							[]string{key53, key55, key57},
							0,
							instruction.SWAP_BY_END_EPOCH,
						).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key53, key55, key57}).
				AddShardSubstituteAdded(0, []string{key112, key98}).
				AddSlashingCommittees(0, []string{key53, key55, key57}).
				AddRemovedStakers([]string{key53, key55, key57}).
				AddFinishedSyncValidators(0, []string{key112, key98}).
				AddSyncingPoolRemoved(0, []string{key112, key98}),
			wantReturnInstruction: [][]string{
				instruction.NewReturnStakeInsWithValue(
					[]string{key53, key55, key57},
					[]string{hash.String(), hash.String(), hash.String()},
				).ToString(),
			},
			wantErr: false,
		},
		{
			name: "Process Stake and Snapshot",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
								key92: true,
								key93: true,
								key94: true,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113},
						1: []string{key114, key115},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash,
								key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash,
								key92: *hash, key93: *hash, key94: *hash, key95: *hash,
								key116: *hash, key117: *hash, key118: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
								incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress,
								incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress,
								incKey116.GetIncKeyBase58(): paymentAddress, incKey117.GetIncKeyBase58(): paymentAddress, incKey118.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:   true,
								key:    true,
								key2:   true,
								key3:   true,
								key6:   true,
								key8:   true,
								key9:   true,
								key10:  false,
								key11:  true,
								key21:  true,
								key22:  false,
								key23:  false,
								key24:  true,
								key25:  true,
								key26:  true,
								key27:  true,
								key28:  true,
								key29:  true,
								key52:  true,
								key53:  true,
								key54:  true,
								key55:  true,
								key56:  true,
								key57:  true,
								key58:  true,
								key59:  true,
								key92:  true,
								key93:  true,
								key94:  true,
								key95:  true,
								key116: true,
								key117: false,
								key118: true,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93, key116, key117, key118,
						},
						numberOfAssignedCandidates: 4,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113},
						1: []string{key114, key115},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					IsBeaconRandomTime:               true,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewStakeInstructionWithValue(
							[]string{key116, key117, key118},
							instruction.SHARD_INST,
							[]string{hash.String(), hash.String(), hash.String()},
							[]string{paymentAddressString, paymentAddressString, paymentAddressString},
							[]bool{true, false, true},
							[]string{"", "", ""},
						).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddNextEpochShardCandidateAdded([]string{key116, key117, key118}),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Unstake and Finish Sync",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:   true,
								key:    true,
								key2:   true,
								key3:   true,
								key6:   true,
								key8:   true,
								key9:   true,
								key10:  false,
								key11:  true,
								key21:  true,
								key22:  false,
								key23:  false,
								key24:  true,
								key25:  true,
								key26:  true,
								key27:  true,
								key28:  true,
								key29:  true,
								key52:  true,
								key53:  true,
								key54:  true,
								key55:  true,
								key56:  true,
								key57:  true,
								key58:  true,
								key59:  true,
								key92:  true,
								key93:  true,
								key94:  true,
								key95:  true,
								key116: true,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113, key116},
						1: []string{key114, key115},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key116, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:   true,
								key:    true,
								key2:   true,
								key3:   true,
								key6:   true,
								key8:   true,
								key9:   true,
								key10:  false,
								key11:  true,
								key21:  true,
								key22:  false,
								key23:  false,
								key24:  true,
								key25:  true,
								key26:  true,
								key27:  true,
								key28:  true,
								key29:  true,
								key52:  true,
								key53:  true,
								key54:  true,
								key55:  true,
								key56:  true,
								key57:  true,
								key58:  true,
								key59:  true,
								key92:  true,
								key93:  true,
								key94:  true,
								key95:  true,
								key116: false,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113},
						1: []string{key114, key115},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					IsBeaconRandomTime:               false,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewUnstakeInstructionWithValue(
							[]string{key116},
						).ToString(),
						instruction.NewFinishSyncInstructionWithValue(
							0,
							[]string{key116},
						).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key116}).
				AddSyncingPoolRemoved(0, []string{key116}).
				AddFinishedSyncValidators(0, []string{key116}).
				AddStopAutoStake(key116),
			wantReturnInstruction: [][]string{},
			wantErr:               false,
		},
		{
			name: "Process Unstake and Swap",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
									key60, key61, key62, key63,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx:      map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key22: *hash, key23: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key53: *hash, key54: *hash, key55: *hash, key56: *hash, key57: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress, incKey21.GetIncKeyBase58(): paymentAddress, incKey22.GetIncKeyBase58(): paymentAddress, incKey23.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress, incKey52.GetIncKeyBase58(): paymentAddress, incKey53.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey55.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey57.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key22: false,
								key23: false,
								key24: true,
								key25: true,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key53: true,
								key54: true,
								key55: true,
								key56: true,
								key57: true,
								key58: true,
								key59: true,
								key92: true,
								key93: true,
								key94: true,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113},
						1: []string{key114, key115},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: finalMutex,
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
									key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
									key20, key26, key27, key28, key29,
									key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
									key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
									key50, key51, key52, key54, key56, key58, key59,
									key60, key61, key62, key63,
									key70, key71, key72, key73, key74, key75, key76,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key77, key78, key79,
									key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
									key21, key24, key25,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							stakingTx: map[string]common.Hash{key0: *hash, key: *hash, key2: *hash, key3: *hash, key6: *hash, key8: *hash, key9: *hash, key10: *hash, key11: *hash, key21: *hash, key24: *hash, key25: *hash, key26: *hash, key27: *hash, key28: *hash, key29: *hash, key52: *hash, key54: *hash, key56: *hash, key58: *hash, key59: *hash, key92: *hash, key93: *hash, key94: *hash, key95: *hash},
							rewardReceiver: map[string]privacy.PaymentAddress{incKey0.GetIncKeyBase58(): paymentAddress, incKey.GetIncKeyBase58(): paymentAddress, incKey2.GetIncKeyBase58(): paymentAddress, incKey3.GetIncKeyBase58(): paymentAddress, incKey6.GetIncKeyBase58(): paymentAddress, incKey8.GetIncKeyBase58(): paymentAddress, incKey9.GetIncKeyBase58(): paymentAddress, incKey10.GetIncKeyBase58(): paymentAddress, incKey11.GetIncKeyBase58(): paymentAddress,
								incKey21.GetIncKeyBase58(): paymentAddress, incKey24.GetIncKeyBase58(): paymentAddress, incKey25.GetIncKeyBase58(): paymentAddress, incKey26.GetIncKeyBase58(): paymentAddress, incKey27.GetIncKeyBase58(): paymentAddress, incKey28.GetIncKeyBase58(): paymentAddress, incKey29.GetIncKeyBase58(): paymentAddress,
								incKey52.GetIncKeyBase58(): paymentAddress, incKey54.GetIncKeyBase58(): paymentAddress, incKey56.GetIncKeyBase58(): paymentAddress, incKey58.GetIncKeyBase58(): paymentAddress, incKey59.GetIncKeyBase58(): paymentAddress, incKey92.GetIncKeyBase58(): paymentAddress, incKey93.GetIncKeyBase58(): paymentAddress, incKey94.GetIncKeyBase58(): paymentAddress, incKey95.GetIncKeyBase58(): paymentAddress},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  true,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: false,
								key11: true,
								key21: true,
								key24: false,
								key25: false,
								key26: true,
								key27: true,
								key28: true,
								key29: true,
								key52: true,
								key54: true,
								key56: true,
								key58: true,
								key59: true,
								key92: true,
								key93: true,
								key94: true,
								key95: true,
							},
						},
						shardCommonPool: []string{
							key90, key91, key92, key93,
						},
						numberOfAssignedCandidates: 0,
						swapRule:                   &swapRuleV3{},
					},
					syncPool: map[byte][]string{
						0: []string{key112, key113},
						1: []string{key114, key115},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					IsBeaconRandomTime:               false,
					MaxShardCommitteeSize:            64,
					NumberOfFixedShardBlockValidator: 21,
					MinShardCommitteeSize:            21,
					ShardID:                          0,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.NewPenalty(),
						key20: signaturecounter.NewPenalty(),
						key53: signaturecounter.NewPenalty(),
						key55: signaturecounter.NewPenalty(),
						key57: signaturecounter.NewPenalty(),
					},
					BeaconInstructions: [][]string{
						instruction.NewUnstakeInstructionWithValue(
							[]string{key24, key25},
						).ToString(),
						instruction.NewSwapShardInstructionWithValue(
							[]string{key70, key71, key72, key73, key74, key75, key76},
							[]string{key53, key55, key57, key21, key22, key23, key24, key25},
							0,
							instruction.SWAP_BY_END_EPOCH,
						).ToString(),
					},
				},
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardSubstituteRemoved(0, []string{key70, key71, key72, key73, key74, key75, key76}).
				AddShardCommitteeRemoved(0, []string{key53, key55, key57, key21, key22, key23, key24, key25}).
				AddShardSubstituteAdded(0, []string{key21, key24, key25}).
				AddRemovedStakers([]string{key22, key23, key53, key55, key57}).
				AddSlashingCommittees(0, []string{key53, key55, key57}).
				AddStopAutoStakes([]string{key24, key25}),
			wantReturnInstruction: [][]string{
				instruction.NewReturnStakeInsWithValue(
					[]string{key22, key23, key53, key55, key57},
					[]string{hash.String(), hash.String(), hash.String(), hash.String(), hash.String()}).ToString(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.BeaconCommitteeStateV3
			_, got1, got2, err := b.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.wantCommitteeChange) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.wantCommitteeChange)
			}
			if !reflect.DeepEqual(got2, tt.wantReturnInstruction) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.wantReturnInstruction)
			}
			gotCommitteeState := b
			wantCommitteeState := tt.fieldsAfterProcess.BeaconCommitteeStateV3
			if !reflect.DeepEqual(gotCommitteeState, wantCommitteeState) {
				t.Errorf("fieldsAfterProcess, got = %v, want %v", gotCommitteeState, wantCommitteeState)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processFinishSyncInstruction(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		finishSyncInstruction *instruction.FinishSyncInstruction
		env                   *BeaconCommitteeStateEnvironment
		committeeChange       *CommitteeChange
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess *BeaconCommitteeStateV3
		args               args
		want               *CommitteeChange
	}{
		{
			name: "remove one validator, sync pool not empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key4, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key4}),
		},
		{
			name: "remove one validator, sync pool is empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key4},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key4, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key4}),
		},
		{
			name: "remove multiple validator, sync pool is empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4, key5},
					1: []string{key13, key14},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key3, key0, key4, key5, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key13, key14},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4, key3, key5}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddSyncingPoolRemoved(0, []string{key3}).
				AddSyncingPoolRemoved(0, []string{key5}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key3}).
				AddFinishedSyncValidators(0, []string{key5}).
				AddShardSubstituteAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key3}).
				AddShardSubstituteAdded(0, []string{key5}),
		},
		{
			name: "remove multiple validator, sync pool not empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4, key5},
					1: []string{key13, key14},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key3, key0, key4, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key5},
					1: []string{key13, key14},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4, key3}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddSyncingPoolRemoved(0, []string{key3}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key3}).
				AddShardSubstituteAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key3}),
		},
		{
			name: "remove multiple validator, sync pool not empty => pending is empty before assign",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4, key5},
					1: []string{key13, key14},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key3, key5, key4},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key13, key14},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4, key3, key5}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddSyncingPoolRemoved(0, []string{key3}).
				AddSyncingPoolRemoved(0, []string{key5}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key3}).
				AddFinishedSyncValidators(0, []string{key5}).
				AddShardSubstituteAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key3}).
				AddShardSubstituteAdded(0, []string{key5}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.processFinishSyncInstruction(tt.args.finishSyncInstruction, tt.args.env, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processFinishSyncInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("processFinishSyncInstruction() fieldsAfterProcess got = %v, want %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_removeValidatorsFromSyncPool(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		validators []string
		shardID    byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess *BeaconCommitteeStateV3
		args               args
		wantErr            bool
	}{
		{
			name: "remove 2 validators, 1 syncPool error",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key5, key0},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key0},
					1: []string{key10, key11, key12},
				},
			},
			wantErr: true,
		},
		{
			name: "remove 2 validators, 1 syncPool no error",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key0},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key10, key11, key12},
				},
			},
			wantErr: false,
		},
		{
			name: "remove 1 validators, 2 syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key0},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key5},
					1: []string{key10, key11, key12},
				},
			},
			wantErr: false,
		},
		{
			name: "remove validators not in syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key6},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key0, key5},
					1: []string{key10, key11, key12},
				},
			},
			wantErr: true,
		},
		{
			name: "remove all validators must be removed from syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key, key3, key2, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key5, key0, key2, key, key3},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key10, key11, key12},
				},
			},
			wantErr: false,
		},
		{
			name: "remove 3 validators, 5 sync pool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key, key3, key2, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key5, key0, key3},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key, key2},
					1: []string{key10, key11, key12},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			err := b.removeValidatorsFromSyncPool(tt.args.validators, tt.args.shardID)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.removeValidatorsFromSyncPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
					t.Errorf("removeValidatorsFromSyncPool() got %+v, want %+v", b, tt.fieldsAfterProcess)
				}
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAfterNormal(t *testing.T) {

	initTestParams()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreShardStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey2, *incKey4, *incKey6, *incKey8, *incKey10, *incKey12,
		},
		map[string]privacy.PaymentAddress{
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey4.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key0:  true,
			key2:  true,
			key4:  true,
			key6:  false,
			key8:  false,
			key10: false,
			key12: false,
		},
		map[string]common.Hash{
			key0:  *hash,
			key2:  *hash,
			key4:  *hash,
			key6:  *hash,
			key8:  *hash,
			key10: *hash,
			key12: *hash,
		},
	)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		env                      *BeaconCommitteeStateEnvironment
		outPublicKeys            []string
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
	}
	tests := []struct {
		name                        string
		fields                      fields
		fieldsAfterProcess          *BeaconCommitteeStateV3
		args                        args
		wantCommitteeChange         *CommitteeChange
		want1ReturnStakeInstruction *instruction.ReturnStakeInstruction
		wantErr                     bool
	}{
		{
			name: "1 stop auto stake = false, 1 return stake, no assign back",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						stakingTx: map[string]common.Hash{
							key6: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey6.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key6: false,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						stakingTx:      map[string]common.Hash{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						autoStake:      map[string]bool{},
					},
				},
				syncPool: make(map[byte][]string),
			},
			args: args{
				committeeChange: NewCommitteeChange(),
				env: &BeaconCommitteeStateEnvironment{
					ShardID:          0,
					ConsensusStateDB: sDB,
				},
				outPublicKeys:            []string{key6},
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddRemovedStaker(key6),
			want1ReturnStakeInstruction: instruction.NewReturnStakeInsWithValue([]string{key6}, []string{hash.String()}),
			wantErr:                     false,
		},
		{
			name: "stop auto stake = true, no return stake, assign back",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key2: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey2.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key2: true,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key, key2},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key2: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey2.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key2: true,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			args: args{
				committeeChange: NewCommitteeChange(),
				env: &BeaconCommitteeStateEnvironment{
					ShardID:          0,
					ConsensusStateDB: sDB,
				},
				outPublicKeys:            []string{key2},
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}),
			want1ReturnStakeInstruction: instruction.NewReturnStakeIns(),
			wantErr:                     false,
		},
		{
			name: "both stop auto stake = true, false, has assign back and return inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key0: *hash,
							key2: *hash,
							key6: *hash,
							key8: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey6.GetIncKeyBase58(): paymentAddress,
							incKey8.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key0: true,
							key2: true,
							key6: false,
							key8: false,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key, key2, key0},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key0: *hash,
							key2: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key0: true,
							key2: true,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			args: args{
				committeeChange: NewCommitteeChange(),
				env: &BeaconCommitteeStateEnvironment{
					ShardID:          0,
					ConsensusStateDB: sDB,
				},
				outPublicKeys:            []string{key2, key0, key6, key8},
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddRemovedStaker(key6).
				AddRemovedStaker(key8),
			want1ReturnStakeInstruction: instruction.NewReturnStakeInsWithValue([]string{key6, key8}, []string{hash.String(), hash.String()}),
			wantErr:                     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			got, got1, err := b.processAfterNormalSwap(tt.args.env, tt.args.outPublicKeys, tt.args.committeeChange, tt.args.returnStakingInstruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("processAfterNormalSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantCommitteeChange) {
				t.Errorf("processAfterNormalSwap() got = %v, want %v", got, tt.wantCommitteeChange)
			}
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("processAfterNormalSwap() fieldsAfterProcess got = %v, want %v", b, tt.fieldsAfterProcess)
			}
			if !reflect.DeepEqual(got1, tt.want1ReturnStakeInstruction) {
				t.Errorf("processAfterNormalSwap() got1 = %v, want %v", got1, tt.want1ReturnStakeInstruction)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignBackToSubstituteList(t *testing.T) {
	testcase3CommitteeChange := NewCommitteeChange().AddShardSubstituteAdded(0, []string{key0, key})
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		candidates      []string
		shardID         byte
		committeeChange *CommitteeChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CommitteeChange
	}{
		{
			name: "one candidate, empty committee change",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: make(map[byte][]string),
					},
				},
			},
			args: args{
				candidates: []string{
					key0,
				},
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().AddShardSubstituteAdded(0, []string{key0}),
		},
		{
			name: "two candidate, empty committee change",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: make(map[byte][]string),
					},
				},
			},
			args: args{
				candidates: []string{
					key0, key,
				},
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().AddShardSubstituteAdded(0, []string{key0, key}),
		},
		{
			name: "two candidate, not empty committee change",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: make(map[byte][]string),
					},
				},
			},
			args: args{
				candidates: []string{
					key5, key6,
				},
				shardID:         0,
				committeeChange: testcase3CommitteeChange,
			},
			want: testcase3CommitteeChange.AddShardSubstituteAdded(0, []string{key5, key6}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignBackToSubstituteList(tt.args.candidates, tt.args.shardID, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("assignBackToSubstituteList() = %v, want %v", got, tt.want)
			}
		})
	}
}
