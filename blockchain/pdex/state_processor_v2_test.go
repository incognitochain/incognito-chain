package pdex

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func Test_stateProcessorV2_addLiquidity(t *testing.T) {
	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		poolPairs                   map[string]PoolPairState
		waitingContributions        map[string]Contribution
		deletedWaitingContributions map[string]Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]PoolPairState
		want1   map[string]Contribution
		want2   map[string]Contribution
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, err := sp.addLiquidity(tt.args.stateDB, tt.args.inst, tt.args.beaconHeight, tt.args.poolPairs, tt.args.waitingContributions, tt.args.deletedWaitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.addLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.addLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.addLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.addLiquidity() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_stateProcessorV2_waitingContribution(t *testing.T) {
	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		waitingContributions        map[string]Contribution
		deletedWaitingContributions map[string]Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]Contribution
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, err := sp.waitingContribution(tt.args.stateDB, tt.args.inst, tt.args.waitingContributions, tt.args.deletedWaitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.waitingContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.waitingContribution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateProcessorV2_refundContribution(t *testing.T) {
	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		waitingContributions        map[string]Contribution
		deletedWaitingContributions map[string]Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]Contribution
		want1   map[string]Contribution
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, err := sp.refundContribution(tt.args.stateDB, tt.args.inst, tt.args.waitingContributions, tt.args.deletedWaitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.refundContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.refundContribution() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.refundContribution() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProcessorV2_matchContribution(t *testing.T) {
	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		waitingContributions        map[string]Contribution
		deletedWaitingContributions map[string]Contribution
		poolPairs                   map[string]PoolPairState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]Contribution
		want1   map[string]Contribution
		want2   map[string]PoolPairState
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, err := sp.matchContribution(tt.args.stateDB, tt.args.inst, tt.args.beaconHeight, tt.args.waitingContributions, tt.args.deletedWaitingContributions, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.matchContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.matchContribution() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.matchContribution() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.matchContribution() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_stateProcessorV2_matchAndReturnContribution(t *testing.T) {
	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		waitingContributions        map[string]Contribution
		deletedWaitingContributions map[string]Contribution
		poolPairs                   map[string]PoolPairState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]Contribution
		want1   map[string]Contribution
		want2   map[string]PoolPairState
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, err := sp.matchAndReturnContribution(tt.args.stateDB, tt.args.inst, tt.args.beaconHeight, tt.args.waitingContributions, tt.args.deletedWaitingContributions, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
