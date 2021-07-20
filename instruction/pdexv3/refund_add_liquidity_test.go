package pdexv3

import (
	"reflect"
	"testing"
)

func TestRefundAddLiquidity_FromStringArr(t *testing.T) {
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
			r := &RefundAddLiquidity{
				pairHash:    tt.fields.pairHash,
				otaReceiver: tt.fields.otaReceiver,
				tokenID:     tt.fields.tokenID,
				tokenAmount: tt.fields.tokenAmount,
				txReqID:     tt.fields.txReqID,
				shardID:     tt.fields.shardID,
			}
			if err := r.FromStringArr(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RefundAddLiquidity.FromStringArr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefundAddLiquidity_StringArr(t *testing.T) {
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
			r := &RefundAddLiquidity{
				pairHash:    tt.fields.pairHash,
				otaReceiver: tt.fields.otaReceiver,
				tokenID:     tt.fields.tokenID,
				tokenAmount: tt.fields.tokenAmount,
				txReqID:     tt.fields.txReqID,
				shardID:     tt.fields.shardID,
			}
			if got := r.StringArr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RefundAddLiquidity.StringArr() = %v, want %v", got, tt.want)
			}
		})
	}
}
