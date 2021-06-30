package pdex

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func Test_stateProducerV1_trade(t *testing.T) {
	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		actions      [][]string
		beaconHeight uint64
		poolPairs    map[string]*rawdbv2.PDEPoolForPair
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV1{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, err := sp.trade(tt.args.actions, tt.args.beaconHeight, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV1.trade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV1.trade() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateProducerV1_withdrawal(t *testing.T) {
	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		actions      [][]string
		beaconHeight uint64
		poolPairs    map[string]*rawdbv2.PDEPoolForPair
		shares       map[string]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV1{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, err := sp.withdrawal(tt.args.actions, tt.args.beaconHeight, tt.args.poolPairs, tt.args.shares)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV1.withdrawal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV1.withdrawal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateProducerV1_contribution(t *testing.T) {
	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		actions              [][]string
		beaconHeight         uint64
		isPRVRequired        bool
		metaType             int
		waitingContributions map[string]*rawdbv2.PDEContribution
		poolPairs            map[string]*rawdbv2.PDEPoolForPair
		shares               map[string]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV1{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, err := sp.contribution(tt.args.actions, tt.args.beaconHeight, tt.args.isPRVRequired, tt.args.metaType, tt.args.waitingContributions, tt.args.poolPairs, tt.args.shares)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV1.contribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV1.contribution() = %v, want %v", got, tt.want)
			}
		})
	}
}
