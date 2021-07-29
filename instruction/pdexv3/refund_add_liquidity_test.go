package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func TestRefundAddLiquidity_FromStringSlice(t *testing.T) {
	type fields struct {
		contribution statedb.Pdexv3ContributionState
	}
	type args struct {
		source []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RefundAddLiquidity{
				contribution: tt.fields.contribution,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RefundAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefundAddLiquidity_StringSlice(t *testing.T) {
	type fields struct {
		contribution statedb.Pdexv3ContributionState
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RefundAddLiquidity{
				contribution: tt.fields.contribution,
			}
			got, err := r.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("RefundAddLiquidity.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RefundAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
