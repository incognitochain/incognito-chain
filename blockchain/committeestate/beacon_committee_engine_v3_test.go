package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

func TestBeaconCommitteeEngineV3_UpdateCommitteeState(t *testing.T) {
	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BeaconCommitteeStateHash
		want1   *CommitteeChange
		want2   [][]string
		wantErr bool
	}{
		/*{*/
		//name:    "Process Random Instruction",
		//fields:  fields{},
		//args:    args{},
		//want1:   &CommitteeChange{},
		//want2:   [][]string{},
		//wantErr: false,
		//},
		//{
		//name:    "Process Unstake Instruction",
		//fields:  fields{},
		//args:    args{},
		//want1:   &CommitteeChange{},
		//want2:   [][]string{},
		//wantErr: false,
		//},
		//{
		//name:    "Process Swap Shard Instruction",
		//fields:  fields{},
		//args:    args{},
		//want1:   &CommitteeChange{},
		//want2:   [][]string{},
		//wantErr: false,
		//},
		//{
		//name:    "Process Finish Sync Instruction",
		//fields:  fields{},
		//args:    args{},
		//want1:   &CommitteeChange{},
		//want2:   [][]string{},
		//wantErr: false,
		/*},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV3{
				beaconCommitteeEngineSlashingBase: tt.fields.beaconCommitteeEngineSlashingBase,
			}
			_, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestBeaconCommitteeEngineV3_UpdateCommitteeState_MultipleInstructions(t *testing.T) {
	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BeaconCommitteeStateHash
		want1   *CommitteeChange
		want2   [][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV3{
				beaconCommitteeEngineSlashingBase: tt.fields.beaconCommitteeEngineSlashingBase,
			}
			got, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestBeaconCommitteeEngineV3_GenerateFinishSyncInstructions(t *testing.T) {

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*instruction.FinishSyncInstruction
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2},
								1: []incognitokey.CommitteePublicKey{*incKey3, *incKey4, *incKey5},
							},
							finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
								1: []incognitokey.CommitteePublicKey{*incKey5},
							},
						},
					},
				},
			},
			want: []*instruction.FinishSyncInstruction{
				&instruction.FinishSyncInstruction{
					ChainID:          0,
					PublicKeys:       []string{key0, key},
					PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
				},
				&instruction.FinishSyncInstruction{
					ChainID:          1,
					PublicKeys:       []string{key5},
					PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey5},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := BeaconCommitteeEngineV3{
				beaconCommitteeEngineSlashingBase: tt.fields.beaconCommitteeEngineSlashingBase,
			}
			got, err := engine.GenerateFinishSyncInstructions()
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.GenerateFinishSyncInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, v := range got {
				if !reflect.DeepEqual(v, tt.want[i]) {
					t.Errorf("v = %v, want %v", v, tt.want[i])
					return
				}
			}
		})
	}
}

func TestBeaconCommitteeEngineV3_AddFinishedSyncValidators(t *testing.T) {

	initLog()
	initTestParams()

	finalMutex := &sync.RWMutex{}
	unCommittedMutex := &sync.RWMutex{}

	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	type args struct {
		committeePublicKeys []string
		shardID             byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 10,
						beaconHash:   common.Hash{},
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
								},
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2,
								},
							},
							finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommittedMutex,
								},
							},
						},
					},
				},
			},
			args: args{
				committeePublicKeys: []string{key0, key},
				shardID:             1,
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 10,
						beaconHash:   common.Hash{},
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
								},
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2,
								},
							},
							finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommittedMutex,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := BeaconCommitteeEngineV3{
				beaconCommitteeEngineSlashingBase: tt.fields.beaconCommitteeEngineSlashingBase,
			}
			if err := engine.AddFinishedSyncValidators(tt.args.committeePublicKeys, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.AddFinishedSyncValidators() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(engine.beaconCommitteeEngineSlashingBase.beaconCommitteeEngineBase.finalState, tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.beaconCommitteeEngineBase.finalState) {
				t.Errorf("finalState = %v, fieldsAfterProcess.finalState %v", engine.finalState, tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.finalState)
			}
			if !reflect.DeepEqual(engine.beaconCommitteeEngineSlashingBase.beaconCommitteeEngineBase.uncommittedState, tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.beaconCommitteeEngineBase.uncommittedState) {
				t.Errorf("uncommittedState = %v, fieldsAfterProcess.uncommittedState %v", engine.uncommittedState, tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.uncommittedState)
			}
		})
	}
}
