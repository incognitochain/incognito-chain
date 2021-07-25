package pdex

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/metadata"
)

func Test_stateProducerV2_addLiquidity(t *testing.T) {
	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs                  []metadata.Transaction
		beaconHeight         uint64
		poolPairs            map[string]PoolPairState
		waitingContributions map[string]Contribution
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               [][]string
		want1              map[string]PoolPairState
		want2              map[string]Contribution
		wantErr            bool
	}{
		{
			name:    "Wrong metadata",
			fields:  fields{},
			args:    args{},
			want:    [][]string{},
			want1:   map[string]PoolPairState{},
			want2:   map[string]Contribution{},
			wantErr: true,
		},
		{
			name:    "Success add to waiting list",
			fields:  fields{},
			args:    args{},
			want:    [][]string{},
			want1:   map[string]PoolPairState{},
			want2:   map[string]Contribution{},
			wantErr: true,
		},
		{
			name:    "Refund contributions",
			fields:  fields{},
			args:    args{},
			want:    [][]string{},
			want1:   map[string]PoolPairState{},
			want2:   map[string]Contribution{},
			wantErr: true,
		},
		{
			name:    "Match contributions new pool",
			fields:  fields{},
			args:    args{},
			want:    [][]string{},
			want1:   map[string]PoolPairState{},
			want2:   map[string]Contribution{},
			wantErr: true,
		},
		{
			name:    "Match contributions old pool",
			fields:  fields{},
			args:    args{},
			want:    [][]string{},
			want1:   map[string]PoolPairState{},
			want2:   map[string]Contribution{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, got2, err := sp.addLiquidity(tt.args.txs, tt.args.beaconHeight, tt.args.poolPairs, tt.args.waitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.addLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.addLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.addLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProducerV2.addLiquidity() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
