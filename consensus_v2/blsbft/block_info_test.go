package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestReProposeBlockInfo_VerifySignature(t *testing.T) {
	beacon1 := "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy"
	tc1Proposer := incognitokey.CommitteePublicKey{}
	_ = tc1Proposer.FromString(beacon1)

	shard00 := "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy"
	tc2Proposer := incognitokey.CommitteePublicKey{}
	_ = tc2Proposer.FromString(shard00)

	shard04 := "121VhftSAygpEJZ6i9jGk4fjTw3t5Lfbd1hzFRQjseWMsHPvRsMJiPDJsExEEYBVYar24wCoHPTuo4gQZ4dLtjxshrgmQxrL12dR6FzBWS4d89DKrctXsN2iCearvg9sRyftttsiuNneyb1LGRFuEnZw95YoUXfVNkV6qX7AvGfVnhYUkVX9KCZXAFDYKRbGArd47AQ8iTHjchQRxGqmsZ61GAnCVYzi3XLaV8avQCTvWmcQB9GdzB2yeU9wy1Gzec6vs8vNBf11ryPhTBwEc3bJezoCqJixEp47CvkWuMUJh7e3a28CDnZCvU5538XubywAXtcUyG3yyHFQAvadsa9ejRUFrKCWPGPJ5CYxsP8uVyXLzKEw6bKfsAKMD6NyNYkeTcte2CskEdGTCuZPDi2aNEhvPchQxso9KGNQb4D5w63b"
	tc3Proposer := incognitokey.CommitteePublicKey{}
	_ = tc3Proposer.FromString(shard04)

	shard05 := "121VhftSAygpEJZ6i9jGkEsxj9J8yMyftfK9kP371U12E7C5TnmGKzVkT4sMHZHYKmmXggfmbWiiYxuj7KT9KuBy5kCztri3MKyCAuKhbf6kyPxQ66cigusK71PMQR645vKUY7e8P5PjfkQxMiQ9ppCu38JnbMMWMETfaKEVwLjY8tJ3N19x8Lg6swPWdPQMWdBRDynz6MGSbspvK1xqPXdBRWa1hz8U5bpPm3UAhFLYXwWymWspsfi4aTJsYorkmuYHHPUj2GSRnAiNqBTEKsunhNrKe53XYqp7pQyrmoku3Tue7zrjyQzbk6pqzsRFZCip4PWrTZyxJyMBwMUBtmCfY2sv2uNLQyBon62KCu55ijck2j4jogE12PgZA5K79sp6dsKRDys7eYMwRgMxFCNURVaNLKjNz9LuYuqWXnweWH76"
	tc4Proposer := incognitokey.CommitteePublicKey{}
	_ = tc4Proposer.FromString(shard05)

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
				Proposer:          beacon1,
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
			name: "success 2",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"),
				Producer:          "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
				ProducerTimeSlot:  163214440,
				Proposer:          shard00,
				ProposerTimeSlot:  163214440,
				RootHash:          common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"),
			},
			args: args{
				sigBase58: "12EGPejMDU3wwnAo2drdnWw14TmnFtozW1LBDtMA4oZR4FVfhxwRqMqkXoWMQymEy2tvMbcgkeVF9woiZV8dCHk1Fi8R8Wat",
				publicKey: tc2Proposer.MiningPubKey[common.BridgeConsensus],
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "success 3",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("b620978d408c0a1895475e8c0b6d16d952842b5d62b45b7a114cc8cae72a9514"),
				Producer:          "121VhftSAygpEJZ6i9jGk4fjTw3t5Lfbd1hzFRQjseWMsHPvRsMJiPDJsExEEYBVYar24wCoHPTuo4gQZ4dLtjxshrgmQxrL12dR6FzBWS4d89DKrctXsN2iCearvg9sRyftttsiuNneyb1LGRFuEnZw95YoUXfVNkV6qX7AvGfVnhYUkVX9KCZXAFDYKRbGArd47AQ8iTHjchQRxGqmsZ61GAnCVYzi3XLaV8avQCTvWmcQB9GdzB2yeU9wy1Gzec6vs8vNBf11ryPhTBwEc3bJezoCqJixEp47CvkWuMUJh7e3a28CDnZCvU5538XubywAXtcUyG3yyHFQAvadsa9ejRUFrKCWPGPJ5CYxsP8uVyXLzKEw6bKfsAKMD6NyNYkeTcte2CskEdGTCuZPDi2aNEhvPchQxso9KGNQb4D5w63b",
				ProducerTimeSlot:  163214732,
				Proposer:          shard04,
				ProposerTimeSlot:  163214732,
				RootHash:          common.Hash{}.NewHashFromStr2("66ab7de3472847c26ed727ab99713416c9a0fc8cc005b921f2651588f607a85d"),
			},
			args: args{
				sigBase58: "13LJzhF413S1ZPKGM7e7hVpPx7CJbKMt7Z9oRzm2xQsmDBoWtAXsjmmxjtHGFe8vLp1aJEHxa1oCk4F3uYfUTjpXLctAjDU7",
				publicKey: tc3Proposer.MiningPubKey[common.BridgeConsensus],
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "success 4",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("b620978d408c0a1895475e8c0b6d16d952842b5d62b45b7a114cc8cae72a9514"),
				Producer:          "121VhftSAygpEJZ6i9jGk4fjTw3t5Lfbd1hzFRQjseWMsHPvRsMJiPDJsExEEYBVYar24wCoHPTuo4gQZ4dLtjxshrgmQxrL12dR6FzBWS4d89DKrctXsN2iCearvg9sRyftttsiuNneyb1LGRFuEnZw95YoUXfVNkV6qX7AvGfVnhYUkVX9KCZXAFDYKRbGArd47AQ8iTHjchQRxGqmsZ61GAnCVYzi3XLaV8avQCTvWmcQB9GdzB2yeU9wy1Gzec6vs8vNBf11ryPhTBwEc3bJezoCqJixEp47CvkWuMUJh7e3a28CDnZCvU5538XubywAXtcUyG3yyHFQAvadsa9ejRUFrKCWPGPJ5CYxsP8uVyXLzKEw6bKfsAKMD6NyNYkeTcte2CskEdGTCuZPDi2aNEhvPchQxso9KGNQb4D5w63b",
				ProducerTimeSlot:  163214732,
				Proposer:          shard05,
				ProposerTimeSlot:  163214733,
				RootHash:          common.Hash{}.NewHashFromStr2("66ab7de3472847c26ed727ab99713416c9a0fc8cc005b921f2651588f607a85d"),
			},
			args: args{
				sigBase58: "12CD6B9WuUUBYfcfsMNGeQNZMNhKjSEHtwM36do7fGPJ7CZBmz2eW7WFQLXFjJQZc7EvRb9ToxVbhwqcGuBgm4rKjPJK7EBc",
				publicKey: tc4Proposer.MiningPubKey[common.BridgeConsensus],
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

func TestReProposeBlockInfo_Sign(t *testing.T) {

	shard00 := []byte{138, 126, 157, 106, 5, 216, 24, 227, 7, 231, 55, 225, 6, 128, 93, 0, 177, 210, 100, 44, 213, 101, 8, 228, 153, 167, 10, 77, 167, 237, 133, 79}
	previousHashTemp1 := []byte{161, 185, 95, 21, 205, 156, 22, 41, 170, 104, 1, 124, 47, 3, 86, 65, 195, 149, 76, 16, 161, 128, 174, 225, 4, 148, 171, 89, 0, 134, 48, 193}
	previousHash1 := common.Hash{}
	previousHash1.NewHash2(previousHashTemp1)
	rootHashTemp1 := []byte{197, 6, 188, 217, 215, 91, 92, 192, 114, 21, 251, 35, 161, 114, 212, 189, 101, 22, 41, 238, 116, 51, 174, 89, 216, 228, 188, 179, 213, 86, 197, 82}
	rootHash1 := common.Hash{}
	rootHash1.NewHash2(rootHashTemp1)
	type fields struct {
		PreviousBlockHash common.Hash
		Producer          string
		ProducerTimeSlot  int64
		Proposer          string
		ProposerTimeSlot  int64
		RootHash          common.Hash
	}
	type args struct {
		privateKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			fields: fields{
				PreviousBlockHash: previousHash1,
				Producer:          "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
				ProducerTimeSlot:  163214376,
				Proposer:          "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
				ProposerTimeSlot:  163214376,
				RootHash:          rootHash1,
			},
			args: args{
				privateKey: shard00,
			},
			want:    "12mBW7WPkYN8hqV4HrzQ59GjKYwrgXYQELkfG5jC6X6yFoMpxcfo8NELM1sUAHHUChUgNgibGVyDymUTJ9DRcySrPhSd5bkj",
			wantErr: false,
		},
		{
			name: "test 2",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"),
				Producer:          "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
				ProducerTimeSlot:  163214440,
				Proposer:          "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
				ProposerTimeSlot:  163214440,
				RootHash:          common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"),
			},
			args: args{
				privateKey: shard00,
			},
			want:    "12EGPejMDU3wwnAo2drdnWw14TmnFtozW1LBDtMA4oZR4FVfhxwRqMqkXoWMQymEy2tvMbcgkeVF9woiZV8dCHk1Fi8R8Wat",
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
			got, err := r.Sign(tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Sign() got = %v, want %v", got, tt.want)
			}
		})
	}
}
