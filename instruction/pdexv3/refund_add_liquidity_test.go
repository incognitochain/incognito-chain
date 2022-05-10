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

func TestRefundAddLiquidity_FromStringSlice(t *testing.T) {
	initTestParam(t)
	contributionState := *statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"pool_pair_id", validOTAReceiver0,
			common.PRVCoinID, common.PRVCoinID, common.Hash{}, 100, metadataPdexv3.BaseAmplifier, 1,
			accessOTA.ToBytesS(), nil,
		), "pair_hash",
	)
	inst := NewRefundAddLiquidityWithValue(contributionState)
	data, err := json.Marshal(inst)
	assert.Nil(t, err)

	type fields struct {
		contribution statedb.Pdexv3ContributionState
	}
	type args struct {
		source []string
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name:    "Invalid length",
			fields:  fields{},
			args:    args{},
			wantErr: true,
		},
		{
			name:   "Invalid metadata type",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityResponseMeta),
					common.PDEContributionRefundChainStatus,
					string(data),
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid status",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityResponseMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(data),
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid Input",
			fields: fields{},
			fieldsAfterProcess: fields{
				contribution: contributionState,
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(data),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RefundAddLiquidity{
				contribution: tt.fields.contribution,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RefundAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(r.contribution, tt.fieldsAfterProcess.contribution) {
				t.Errorf("fieldsAfterProcess expect = %v, but get %v", tt.fieldsAfterProcess.contribution, r.contribution)
				return
			}
		})
	}
}

func TestRefundAddLiquidity_StringSlice(t *testing.T) {
	initTestParam(t)
	contributionState := *statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"pool_pair_id", validOTAReceiver0,
			common.PRVCoinID, common.PRVCoinID, common.Hash{}, 100, metadataPdexv3.BaseAmplifier, 1,
			accessOTA.ToBytesS(), nil,
		), "pair_hash",
	)
	inst := NewRefundAddLiquidityWithValue(contributionState)
	data, err := json.Marshal(inst)
	assert.Nil(t, err)

	type fields struct {
		contribution statedb.Pdexv3ContributionState
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				contribution: contributionState,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
				common.PDEContributionRefundChainStatus,
				string(data),
			},
			wantErr: false,
		},
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
