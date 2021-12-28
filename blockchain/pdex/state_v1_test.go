package pdex

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func Test_stateV1_BuildInstructions(t *testing.T) {
	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]*rawdbv2.PDEContribution
		deletedWaitingContributions map[string]*rawdbv2.PDEContribution
		poolPairs                   map[string]*rawdbv2.PDEPoolForPair
		shares                      map[string]uint64
		tradingFees                 map[string]uint64
		producer                    stateProducerV1
		processor                   stateProcessorV1
	}
	type args struct {
		env StateEnvironment
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
			s := &stateV1{
				stateBase:                   tt.fields.stateBase,
				waitingContributions:        tt.fields.waitingContributions,
				deletedWaitingContributions: tt.fields.deletedWaitingContributions,
				poolPairs:                   tt.fields.poolPairs,
				shares:                      tt.fields.shares,
				tradingFees:                 tt.fields.tradingFees,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			got, err := s.BuildInstructions(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateV1.BuildInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateV1.BuildInstructions() = %v, want %v", got, tt.want)
			}
		})
	}
}
