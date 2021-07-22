package pdexv3

import (
	"reflect"
	"testing"
)

func TestMatchAddLiquidity_FromStringArr(t *testing.T) {
	type fields struct {
		pairHash    string
		otaReceiver string
		tokenID     string
		tokenAmount uint64
		txReqID     string
		shardID     byte
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
				pairHash:    tt.fields.pairHash,
				otaReceiver: tt.fields.otaReceiver,
				tokenID:     tt.fields.tokenID,
				tokenAmount: tt.fields.tokenAmount,
				txReqID:     tt.fields.txReqID,
				shardID:     tt.fields.shardID,
			}
			if err := m.FromStringArr(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("MatchAddLiquidity.FromStringArr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchAddLiquidity_StringArr(t *testing.T) {
	type fields struct {
		pairHash    string
		otaReceiver string
		tokenID     string
		tokenAmount uint64
		txReqID     string
		shardID     byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MatchAddLiquidity{
				pairHash:    tt.fields.pairHash,
				otaReceiver: tt.fields.otaReceiver,
				tokenID:     tt.fields.tokenID,
				tokenAmount: tt.fields.tokenAmount,
				txReqID:     tt.fields.txReqID,
				shardID:     tt.fields.shardID,
			}
			if got := m.StringArr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchAddLiquidity.StringArr() = %v, want %v", got, tt.want)
			}
		})
	}
}
