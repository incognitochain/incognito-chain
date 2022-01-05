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

func TestMatchAndReturnAddLiquidity_FromStringSlice(t *testing.T) {
	initTestParam(t)
	tokenHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)

	contributionState := *statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"pool_pair_id", validOTAReceiver0,
			common.PRVCoinID, common.PRVCoinID, common.Hash{}, 100, metadataPdexv3.BaseAmplifier, 1,
			accessOTA.ToBytesS(), nil,
		), "pair_hash",
	)
	inst := NewMatchAndReturnAddLiquidityWithValue(
		contributionState,
		100, 100, 200, 100,
		*tokenHash, accessOTA.ToBytesS(),
	)
	data, err := json.Marshal(inst)
	assert.Nil(t, err)

	type fields struct {
		shareAmount              uint64
		contribution             statedb.Pdexv3ContributionState
		returnAmount             uint64
		existedTokenActualAmount uint64
		existedTokenReturnAmount uint64
		existedTokenID           common.Hash
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
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(data),
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid Input",
			fields: fields{},
			fieldsAfterProcess: fields{
				contribution:             contributionState,
				shareAmount:              100,
				returnAmount:             100,
				existedTokenActualAmount: 200,
				existedTokenReturnAmount: 100,
				existedTokenID:           *tokenHash,
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(data),
				},
			},
			wantErr: false,
		},
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
			}
			if err := m.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("MatchAndReturnAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchAndReturnAddLiquidity_StringSlice(t *testing.T) {
	initTestParam(t)
	tokenHash, _ := common.Hash{}.NewHashFromStr("abc")
	contributionState := *statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"pool_pair_id", validOTAReceiver0,
			common.PRVCoinID, common.PRVCoinID, common.Hash{}, 100, metadataPdexv3.BaseAmplifier, 1,
			accessOTA.ToBytesS(), nil,
		), "pair_hash",
	)
	inst := NewMatchAndReturnAddLiquidityWithValue(
		contributionState,
		100, 100, 200, 100,
		*tokenHash, accessOTA.ToBytesS(),
	)
	data, err := json.Marshal(inst)
	assert.Nil(t, err)

	type fields struct {
		shareAmount              uint64
		contribution             statedb.Pdexv3ContributionState
		returnAmount             uint64
		existedTokenActualAmount uint64
		existedTokenReturnAmount uint64
		existedTokenID           common.Hash
		nftID                    common.Hash
		accessOTA                []byte
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
				contribution:             contributionState,
				shareAmount:              100,
				returnAmount:             100,
				existedTokenActualAmount: 200,
				existedTokenReturnAmount: 100,
				existedTokenID:           *tokenHash,
				nftID:                    common.PRVCoinID,
				accessOTA:                accessOTA.ToBytesS(),
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
				common.PDEContributionMatchedNReturnedChainStatus,
				string(data),
			},
			wantErr: false,
		},
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
				accessOTA:                tt.fields.accessOTA,
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
