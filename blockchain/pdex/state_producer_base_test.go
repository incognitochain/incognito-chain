package pdex

import (
	"reflect"
	"testing"
)

func Test_stateProducerBase_feeWithdrawal(t *testing.T) {
	type args struct {
		actions      [][]string
		beaconHeight uint64
		tradingFees  map[string]uint64
	}
	tests := []struct {
		name    string
		sp      *stateProducerBase
		args    args
		want    [][]string
		wantErr bool
	}{
		/*{*/
		//name:    "",
		//sp:      &stateProducerBase{},
		//args:    args{},
		//want:    [][]string{},
		//wantErr: false,
		/*},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerBase{}
			got, err := sp.feeWithdrawal(tt.args.actions, tt.args.beaconHeight, tt.args.tradingFees)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerBase.feeWithdrawal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerBase.feeWithdrawal() = %v, want %v", got, tt.want)
			}
		})
	}
}
