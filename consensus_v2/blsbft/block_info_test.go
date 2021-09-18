package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestReProposeBlockInfo_VerifySignature(t *testing.T) {
	tc1ProposerBase58 := "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy"
	tc1Proposer := incognitokey.CommitteePublicKey{}
	_ = tc1Proposer.FromString(tc1ProposerBase58)
	type fields struct {
		PreviousBlockHash common.Hash
		Producer          string
		ProducerTimeSlot  int64
		Proposer          string
		ProposerTimeSlot  int64
		RootHash          common.Hash
	}
	type args struct {
		sigBase58 string
		publicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "success 1",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("c644ce267479ab3085a607f344b6fd4a5e2ac0e73aadd4eac5755d57acdc7e49"),
				Producer:          "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy",
				ProducerTimeSlot:  163193381,
				Proposer:          tc1ProposerBase58,
				ProposerTimeSlot:  163193381,
				RootHash:          common.Hash{}.NewHashFromStr2("32109ef0c83b6d2b6ae23e165a8c920161d282f41244736f2c43f06ade04b04f"),
			},
			args: args{
				sigBase58: "1sfuYPVAjFPuG7uHReskCSqN8eBvN33usJ1RZCMfbLGD6rYh4WnfsdULBMP4Wp2kw3Hw5nFSEYBY8cQjoA2vJxX5PJjWbuF",
				publicKey: tc1Proposer.MiningPubKey[common.BridgeConsensus],
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReProposeBlockInfo{
				PreviousBlockHash: tt.fields.PreviousBlockHash,
				Producer:          tt.fields.Producer,
				ProducerTimeSlot:  tt.fields.ProducerTimeSlot,
				Proposer:          tt.fields.Proposer,
				ProposerTimeSlot:  tt.fields.ProposerTimeSlot,
				RootHash:          tt.fields.RootHash,
			}
			got, err := r.VerifySignature(tt.args.sigBase58, tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifySignature() got = %v, want %v", got, tt.want)
			}
		})
	}
}
