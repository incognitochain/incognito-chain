package pdexv3

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/stretchr/testify/assert"
)

func TestWaitingAddLiquidity_FromStringSlice(t *testing.T) {

	contribution := *rawdbv2.NewPdexv3ContributionWithValue(
		"pool_pair_id", validOTAReceiver0, validOTAReceiver1,
		common.PRVCoinID, common.PRVCoinID, 100, metadataPdexv3.BaseAmplifier, 1,
	)
	data, err := json.Marshal(contribution)
	assert.Nil(t, err)

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
		{
			name: "Invalid length",
			fields: fields{
				contribution: *statedb.NewPdexv3ContributionState(),
			},
			args: args{
				source: []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid metaType",
			fields: fields{
				contribution: *statedb.NewPdexv3ContributionState(),
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					"",
					"",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Status",
			fields: fields{
				contribution: *statedb.NewPdexv3ContributionState(),
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					"",
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				contribution: *statedb.NewPdexv3ContributionState(),
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(data),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WaitingAddLiquidity{
				contribution: tt.fields.contribution,
			}
			if err := w.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("WaitingAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWaitingAddLiquidity_StringSlice(t *testing.T) {
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
			w := &WaitingAddLiquidity{
				contribution: tt.fields.contribution,
			}
			got, err := w.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("WaitingAddLiquidity.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WaitingAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
