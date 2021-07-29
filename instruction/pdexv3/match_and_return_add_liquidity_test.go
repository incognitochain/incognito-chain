package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func TestMatchAndReturnAddLiquidity_FromStringSlice(t *testing.T) {
	type fields struct {
		shareAmount              uint64
		contribution             statedb.Pdexv3ContributionState
		returnAmount             uint64
		existedTokenActualAmount uint64
		existedTokenReturnAmount uint64
		existedTokenID           common.Hash
		nfctID                   common.Hash
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
			m := &MatchAndReturnAddLiquidity{
				shareAmount:              tt.fields.shareAmount,
				contribution:             tt.fields.contribution,
				returnAmount:             tt.fields.returnAmount,
				existedTokenActualAmount: tt.fields.existedTokenActualAmount,
				existedTokenReturnAmount: tt.fields.existedTokenReturnAmount,
				existedTokenID:           tt.fields.existedTokenID,
				nfctID:                   tt.fields.nfctID,
			}
			if err := m.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("MatchAndReturnAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchAndReturnAddLiquidity_StringSlice(t *testing.T) {
	type fields struct {
		shareAmount              uint64
		contribution             statedb.Pdexv3ContributionState
		returnAmount             uint64
		existedTokenActualAmount uint64
		existedTokenReturnAmount uint64
		existedTokenID           common.Hash
		nfctID                   common.Hash
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
			m := &MatchAndReturnAddLiquidity{
				shareAmount:              tt.fields.shareAmount,
				contribution:             tt.fields.contribution,
				returnAmount:             tt.fields.returnAmount,
				existedTokenActualAmount: tt.fields.existedTokenActualAmount,
				existedTokenReturnAmount: tt.fields.existedTokenReturnAmount,
				existedTokenID:           tt.fields.existedTokenID,
				nfctID:                   tt.fields.nfctID,
			}
			got, err := m.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchAndReturnAddLiquidity.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchAndReturnAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
