package pdex

import (
	"reflect"
	"testing"
)

func Test_stateProducerBase_feeWithdrawal(t *testing.T) {
	initLog()
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
		{
			name: "Invalid action content format",
			sp:   &stateProducerBase{},
			args: args{
				actions: [][]string{
					[]string{
						"",
						"12312",
					},
				},
			},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Invalid action content type",
			sp:   &stateProducerBase{},
			args: args{
				actions: [][]string{
					[]string{
						"",
						buildFeeWithdrawalRequestActionForTest(
							"contributorAddress2",
							"tokenID1",
							"tokenID2",
							10,
						)[1],
					},
				},
			},
			want:    [][]string{},
			wantErr: true,
		},
		/*{*/
		//name: "Reject",
		//sp:   &stateProducerBase{},
		//args: args{
		//beaconHeight: 10,
		//actions: [][]string{
		//[]string{
		//"",
		//buildFeeWithdrawalRequestActionForTest(
		//paymentAddress0,
		//common.PRVCoinID.String(),
		//tempPToken,
		//10,
		//)[1],
		//},
		//},
		//tradingFees: map[string]uint64{},
		//},
		//want:    [][]string{},
		//wantErr: false,
		//},
		//{
		//name: "Accept",
		//sp:   &stateProducerBase{},
		//args: args{
		//beaconHeight: 10,
		//actions: [][]string{
		//[]string{
		//"",
		//buildFeeWithdrawalRequestActionForTest(
		//paymentAddress0,
		//common.PRVCoinID.String(),
		//tempPToken,
		//10,
		//)[1],
		//},
		//},
		//tradingFees: map[string]uint64{},
		//},
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
