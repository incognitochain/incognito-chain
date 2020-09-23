package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/common"
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
