package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func TestMatchAddLiquidity_FromStringSlice(t *testing.T) {
	type fields struct {
		contribution  statedb.Pdexv3ContributionState
		newPoolPairID string
		nfctID        common.Hash
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
			m := &MatchAddLiquidity{
				contribution:  tt.fields.contribution,
				newPoolPairID: tt.fields.newPoolPairID,
				nfctID:        tt.fields.nfctID,
			}
			if err := m.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("MatchAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchAddLiquidity_StringSlice(t *testing.T) {
	type fields struct {
		contribution  statedb.Pdexv3ContributionState
		newPoolPairID string
		nfctID        common.Hash
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
			m := &MatchAddLiquidity{
				contribution:  tt.fields.contribution,
				newPoolPairID: tt.fields.newPoolPairID,
				nfctID:        tt.fields.nfctID,
			}
			got, err := m.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchAddLiquidity.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
