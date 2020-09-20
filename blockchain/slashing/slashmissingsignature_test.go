package slashing

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestSignatureCounter_AddMissingSignature(t *testing.T) {
	type fields struct {
		missingSignature map[string]uint
		penalties        []Penalty
	}
	type args struct {
		data       string
		committees []incognitokey.CommitteePublicKey
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
			s := &SignatureCounter{
				missingSignature: tt.fields.missingSignature,
				penalties:        tt.fields.penalties,
			}
			if err := s.AddMissingSignature(tt.args.data, tt.args.committees); (err != nil) != tt.wantErr {
				t.Errorf("AddMissingSignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
