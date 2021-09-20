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

	tc2ProposerBase58 := "121VhftSAygpEJZ6i9jGkGco4dFKpqVXZA6nmGjRKYWR7Q5NngQSX1adAfYY3EGtS32c846sAxYSKGCpqouqmJghfjtYfHEPZTRXctAcc6bYhR3d1YpB6m3nNjEdTYWf85agBq5QnVShMjBRFf54dK25MAazxBSYmpowxwiaEnEikpQah2W4LY9P9vF9HJuLUZ4BnknoXXK3BVkGHsimy5RXtvNet2LqXZgZWHX5CDj31q7kQ2jUGJHr862MgsaHfT4Qq8o4u71nhgtzKBYgw9fvXqJUU6EVynqJCVdqaDXmUvjanGkaZb9vQjaXVoHyf6XRxVSbQBTS5G7eb4D4V3RucXRLQp34KTadmmNQUxnCoPQztVcuDQwNqy9zRXPPAdw7pWvv7P7p4HuQVAHKqvJskMNk3v971WBH5VpZA1XMkmtu,"
	tc2Proposer := incognitokey.CommitteePublicKey{}
	_ = tc2Proposer.FromString(tc2ProposerBase58)

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
		{
			name: "test 2",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("3db501f8679a8a706d121251c2d15b3e36aec18da6985260030d8ad6c55e2051"),
				Producer:          "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
				ProducerTimeSlot:  163212312,
				Proposer:          tc2ProposerBase58,
				ProposerTimeSlot:  163212313,
				RootHash:          common.Hash{}.NewHashFromStr2("0f6a7f9d1bfc1557d6ca21edbd9c0db57f3db48f1ae4e8dbab8ae83272553ab9"),
			},
			args: args{
				sigBase58: "1KExn61mpKXRK7GaddXtS3fVBKkCouFyQhJiewmTVd8Tr8tq4VmescS3SWjTjPR2Ji4gFpLVZYAxqf9zPUHfRq9avDtJc1s",
				publicKey: tc2Proposer.MiningPubKey[common.BridgeConsensus],
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
