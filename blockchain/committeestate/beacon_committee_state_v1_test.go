package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
)

func TestBeaconCommitteeEngine_BuildIncurredInstructions(t *testing.T) {

	initPublicKey()
	initStateDB()
	initLog()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	rewardReceiverkey := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfoV1(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
		},
		map[string]bool{
			key: true,
		},
		map[string]common.Hash{
			key: *hash,
		},
	)

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		beaconCommitteeStateV1            *BeaconCommitteeStateV1
		uncommittedBeaconCommitteeStateV1 *BeaconCommitteeStateV1
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name: "Environment Is Null",
			fields: fields{
				beaconCommitteeStateV1:            &BeaconCommitteeStateV1{},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args:    args{},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Length Of Beacon Instructions Is 0",
			fields: fields{
				beaconCommitteeStateV1:            &BeaconCommitteeStateV1{},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{},
				},
			},
			want:    [][]string{},
			wantErr: false,
		},
		{
			name: "Subtitutes Candidates Public Key's Format Is Not Valid",
			fields: fields{
				beaconCommitteeStateV1: &BeaconCommitteeStateV1{
					nextEpochShardCandidate: []incognitokey.CommitteePublicKey{},
				},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args:    args{},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Validators Public Key's Format Is Not Valid",
			fields: fields{
				beaconCommitteeStateV1: &BeaconCommitteeStateV1{
					nextEpochShardCandidate: []incognitokey.CommitteePublicKey{},
				},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args:    args{},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Invalid Unstake Instruction Format",
			fields: fields{
				beaconCommitteeStateV1: &BeaconCommitteeStateV1{
					currentEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
				},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{instruction.UNSTAKE_ACTION},
					},
				},
			},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Error In Processing Unstake Instruction",
			fields: fields{
				beaconCommitteeStateV1: &BeaconCommitteeStateV1{
					nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey2},
				},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key2,
						},
					},
					ConsensusStateDB:     sDB,
					unassignedCommonPool: []string{key2},
				},
			},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeStateV1: &BeaconCommitteeStateV1{
					nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
				},
				uncommittedBeaconCommitteeStateV1: &BeaconCommitteeStateV1{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key,
						},
					},
					ConsensusStateDB:     sDB,
					unassignedCommonPool: []string{key},
				},
			},
			want: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key,
					"0",
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := BeaconCommitteeEngineV1{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				beaconCommitteeStateV1:            tt.fields.beaconCommitteeStateV1,
				uncommittedBeaconCommitteeStateV1: tt.fields.uncommittedBeaconCommitteeStateV1,
			}
			got, err := engine.BuildIncurredInstructions(tt.args.env)

			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV1.BuildIncurredInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngineV1.BuildIncurredInstructions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_processUnstakeInstruction(t *testing.T) {

	// Init data for testcases
	initStateDB()
	initPublicKey()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	rewardReceiverkey := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	validSDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfoV1(
		validSDB,
		[]incognitokey.CommitteePublicKey{*incKey},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
		},
		map[string]bool{
			key: true,
		},
		map[string]common.Hash{
			key: *hash,
		},
	)
	committeePublicKeyWrongFormat := incognitokey.CommitteePublicKey{}
	committeePublicKeyWrongFormat.MiningPubKey = nil
	//

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	type args struct {
		unstakeInstruction *instruction.UnstakeInstruction
		env                *BeaconCommitteeStateEnvironment
		committeeChange    *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   [][]string
		wantErr bool
	}{
		{
			name: "[Subtitute List] Invalid Format Of Committee Public Key In Unstake Instruction",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{"123"},
				},
				env: &BeaconCommitteeStateEnvironment{
					unassignedCommonPool: []string{"123"},
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			wantErr: true,
		},
		{
			name: "[Subtitute List] Can't find staker info in database",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
				autoStake: map[string]bool{
					key: true,
				},
				rewardReceiver: map[string]privacy.PaymentAddress{
					rewardReceiverkey: paymentAddress,
				},
				stakingTx: map[string]common.Hash{
					key: *hash,
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key2},
				},
				env: &BeaconCommitteeStateEnvironment{
					unassignedCommonPool: []string{key2},
					ConsensusStateDB:     sDB,
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			wantErr: true,
		},
		{
			name: "Valid Input Key In Subtitutes List",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
				autoStake: map[string]bool{
					key: true,
				},
				rewardReceiver: map[string]privacy.PaymentAddress{
					rewardReceiverkey: paymentAddress,
				},
				stakingTx: map[string]common.Hash{
					key: *hash,
				},
			},
			args: args{
				committeeChange: &CommitteeChange{},
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					unassignedCommonPool: []string{key},
					ConsensusStateDB:     validSDB,
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{*incKey},
			},
			want1: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key,
					"0",
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Input Key In Validators List",
			fields: fields{
				currentEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					allSubstituteCommittees: []string{key},
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			want1:   [][]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, got1, err := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange)

			sDB.ClearObjects() // Clear Object From StateDB

			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_getSubtituteCandidates(t *testing.T) {

	initPublicKey()

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			want:    []string{key},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, err := b.getSubstituteCandidates()
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.unassignedCommonPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.unassignedCommonPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_getAllSubstituteCommittees(t *testing.T) {

	initPublicKey()

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				currentEpochShardCandidate: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			want:    []string{key},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, err := b.getAllSubstituteCommittees()
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.getValidators() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.getValidators() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_processUnstakeChange(t *testing.T) {
	initStateDB()
	initPublicKey()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	rewardReceiver := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	hash, err := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	type args struct {
		committeeChange *CommitteeChange
		env             *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		wantErr bool
	}{
		{
			name:   "Invalid Format Of Public Key In List Unstake Of CommitteeChange",
			fields: fields{},
			args: args{
				committeeChange: &CommitteeChange{
					Unstake: []string{"123"},
				},
				env: &BeaconCommitteeStateEnvironment{},
			},
			want: &CommitteeChange{
				Unstake: []string{"123"},
			},
			wantErr: true,
		},
		{
			name:   "Error When Store Staker Info",
			fields: fields{},
			args: args{
				committeeChange: &CommitteeChange{
					Unstake: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
			},
			want: &CommitteeChange{
				Unstake: []string{key},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				autoStake: map[string]bool{
					key: true,
				},
				rewardReceiver: map[string]privacy.PaymentAddress{
					rewardReceiver: paymentAddress,
				},
				stakingTx: map[string]common.Hash{
					key: *hash,
				},
			},
			args: args{
				committeeChange: &CommitteeChange{
					Unstake: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
			},
			want: &CommitteeChange{
				Unstake: []string{key},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, err := b.processUnstakeChange(tt.args.committeeChange, tt.args.env)

			sDB.ClearObjects() // Clear objects for next test case

			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeChange() = %v, want %v", got, tt.want)
			}
		})
	}
}
