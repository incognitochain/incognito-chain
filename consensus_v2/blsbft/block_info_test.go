package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

var (
	beacon0Proposer = incognitokey.CommitteePublicKey{}
	shard00Proposer = incognitokey.CommitteePublicKey{}
	shard01Proposer = incognitokey.CommitteePublicKey{}
	shard02Proposer = incognitokey.CommitteePublicKey{}
	shard03Proposer = incognitokey.CommitteePublicKey{}
	shard04Proposer = incognitokey.CommitteePublicKey{}
	shard05Proposer = incognitokey.CommitteePublicKey{}
	beacon1         = "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy"
	shard00         = "121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy"
	shard01         = "121VhftSAygpEJZ6i9jGkGco4dFKpqVXZA6nmGjRKYWR7Q5NngQSX1adAfYY3EGtS32c846sAxYSKGCpqouqmJghfjtYfHEPZTRXctAcc6bYhR3d1YpB6m3nNjEdTYWf85agBq5QnVShMjBRFf54dK25MAazxBSYmpowxwiaEnEikpQah2W4LY9P9vF9HJuLUZ4BnknoXXK3BVkGHsimy5RXtvNet2LqXZgZWHX5CDj31q7kQ2jUGJHr862MgsaHfT4Qq8o4u71nhgtzKBYgw9fvXqJUU6EVynqJCVdqaDXmUvjanGkaZb9vQjaXVoHyf6XRxVSbQBTS5G7eb4D4V3RucXRLQp34KTadmmNQUxnCoPQztVcuDQwNqy9zRXPPAdw7pWvv7P7p4HuQVAHKqvJskMNk3v971WBH5VpZA1XMkmtu"
	shard02         = "121VhftSAygpEJZ6i9jGkB6Dizgqq7pbFeDL2QEMpXrQHhLLnnCW7JqM1mvpwtvPShhao3HL22hLBznXV89tuHaZiuB1jfd7fE7uBTgpaW23gpQCN6xcmJ5tDipxqdDQ4qsYswGe2qfAy9z6SyAwihD23RukBE2JPoqwuzzHNdQgoaU3nFuZMj51ZxrBU1K3QrVT5Xs9rSZzQkf1AP16WyDXBS7xDYFVbLNRJ14STqRsTDnbpgtdNCuVB7NvpFeVNLFHF5FoxwyLr6iD4sUZNapF4XMcxH28abWD9Vxw4xjH6iDJkY2Ht5duMaqCASMB4YBn8sQzFoGLpAUQWqs49sH118Fi7uMRbKVymgaQRzC3zasNfxQDd3pkAfMHkNqW6XFW23S1mETyyft9ZYtuzWvzeo366eMRCAdVTJAKEp7g3zJ7"
	shard03         = "121VhftSAygpEJZ6i9jGkRjV8czErtzomv6v8WPf2FSkDkes6dqgqP1Y3ebAoEWtm97KFoScxbN8kmBpwQVRDFzqrdbuPeQZMaTMBoXiJteAC8ZrUuKbrLxQWEKgoJvqUkZg9u2Dd2EAyDoreD6W7qYTUUjSXdS9NroR5C7RAztUhQt6TrzvVLzzRtHv4qTWyfdhaHP5tkqPNGXarMZvDCoSBXnR4WXL1uWD872PPXBP2WF62wRhMQN4aA7FSBtbfUsxqvM2HuZZ8ryhCeXb6VyeogWUDxRwNDmhaUMK2sUgez9DJpQ8Lcy2cW7yqco6BR8aUVzME1LetYKp7htB74fRTmGwx7KJUzNH4hiEL7FzTthbes1KyNZabyDH8HHL1zxGqAnDX3R6jKYinsvXtJHGpX1SpHwXfGUuTWn3VqSL7NVv"
	shard04         = "121VhftSAygpEJZ6i9jGk4fjTw3t5Lfbd1hzFRQjseWMsHPvRsMJiPDJsExEEYBVYar24wCoHPTuo4gQZ4dLtjxshrgmQxrL12dR6FzBWS4d89DKrctXsN2iCearvg9sRyftttsiuNneyb1LGRFuEnZw95YoUXfVNkV6qX7AvGfVnhYUkVX9KCZXAFDYKRbGArd47AQ8iTHjchQRxGqmsZ61GAnCVYzi3XLaV8avQCTvWmcQB9GdzB2yeU9wy1Gzec6vs8vNBf11ryPhTBwEc3bJezoCqJixEp47CvkWuMUJh7e3a28CDnZCvU5538XubywAXtcUyG3yyHFQAvadsa9ejRUFrKCWPGPJ5CYxsP8uVyXLzKEw6bKfsAKMD6NyNYkeTcte2CskEdGTCuZPDi2aNEhvPchQxso9KGNQb4D5w63b"
	shard05         = "121VhftSAygpEJZ6i9jGkEsxj9J8yMyftfK9kP371U12E7C5TnmGKzVkT4sMHZHYKmmXggfmbWiiYxuj7KT9KuBy5kCztri3MKyCAuKhbf6kyPxQ66cigusK71PMQR645vKUY7e8P5PjfkQxMiQ9ppCu38JnbMMWMETfaKEVwLjY8tJ3N19x8Lg6swPWdPQMWdBRDynz6MGSbspvK1xqPXdBRWa1hz8U5bpPm3UAhFLYXwWymWspsfi4aTJsYorkmuYHHPUj2GSRnAiNqBTEKsunhNrKe53XYqp7pQyrmoku3Tue7zrjyQzbk6pqzsRFZCip4PWrTZyxJyMBwMUBtmCfY2sv2uNLQyBon62KCu55ijck2j4jogE12PgZA5K79sp6dsKRDys7eYMwRgMxFCNURVaNLKjNz9LuYuqWXnweWH76"
	shard06         = "121VhftSAygpEJZ6i9jGkAEcPKnP3AnswE4vuMUJ89n1V2BtriqaHvb7xsoa9SDux317vReMUmyeRMTdwx4W5xvsBwPbju37RcA9uL3BVwSbymevUyFo5LAeyq95xy9Ynti9KLMK99z1oo58Jo9fKxy9aDqx9hRjKu7f9uN47VYgnQYg6XbA1Bi2zkM8YxUS8W9vZQuW1nGreHv3rWUKryiA3qDpLvjNpcGBwg9UZeLJL49hVEhgwV2JHyBXH1nYL8Z367SEfMWSd6ZzkPWNDaTMdp4HptSuCjZ4w8ur2G25yGqtPy1VR9CX5vVR9tD4Ff99YZTjJueZLpKjztZYwca72z1XxNqCWUbrrKk98dKf8h6n9zeKRNqKgQzVzceiqRv34MTuHm5UxJXecbw3VKrMhSD8d22W1sPeqF8P5ffZEuLR"
	shard07         = "121VhftSAygpEJZ6i9jGk9drLMq7xTahJoDWsLvmjbj3XnrQGTiCM7FVYjqUxCSSWsD9b7Zs16Q1ArPKGVRV5izvGjeqzTGgYdDXbdtyjPd2zeDaWsc7SUeyqQzwhK4xziVJvc5uBupTq9wbDiv6r2KRQsYAtPgPRedcTRyJoTFR7WcVTEoUyMDkyX9x4ZUcaVZgWBs6QwsyUxMTL5rYCC3SBjBV99HJsnWTbbLk355C1YkwuqgpWwuVCvaFq3ZyWFTHz48YpoYxt9bsqQeRJBzdSTjj7T9jR5KagCDJ1LU6b1nPypsvQJb8fvzT5nDoJqqHZZrzrQVyUpkepkNQt33t7UQAdHGKCsf6H8viWdJMuPDHG5KEi4GkMQzdGvoihhciwHh13zVnoJp4HUWiNKg3a6CndTPwSUQ5Lh6m3obK9wwg"
)

var _ = func() (_ struct{}) {
	_ = beacon0Proposer.FromString(beacon1)
	_ = shard00Proposer.FromString(shard00)
	_ = shard01Proposer.FromString(shard01)
	_ = shard02Proposer.FromString(shard02)
	_ = shard03Proposer.FromString(shard03)
	_ = shard04Proposer.FromString(shard04)
	_ = shard05Proposer.FromString(shard05)
	return
}()

func TestReProposeBlockInfo_VerifySignature(t *testing.T) {

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
				publicKey: beacon0Proposer.MiningPubKey[common.BridgeConsensus],
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
				publicKey: shard00Proposer.MiningPubKey[common.BridgeConsensus],
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
				publicKey: shard04Proposer.MiningPubKey[common.BridgeConsensus],
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
				publicKey: shard05Proposer.MiningPubKey[common.BridgeConsensus],
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "success 5",
			fields: fields{
				PreviousBlockHash: common.Hash{}.NewHashFromStr2("1dc43704ebd3ffd6d35be478f86740fc0cbf5b2f825e66fe3c83f31e3e4d18d8"),
				Producer:          shard00,
				ProducerTimeSlot:  163271008,
				Proposer:          shard03,
				ProposerTimeSlot:  163271035,
				RootHash:          common.Hash{}.NewHashFromStr2("874615a2fcacb79434fb851a137da03055eddd9bd086d7b60e6f74a565c887b9"),
			},
			args: args{
				sigBase58: "1N5um21mSoQ2KEBejXofYhuYoBKpwyqdA5SbmErVH8DL5QXMFD15HzxausBMR54FQAT5hZZLV6mM9rMnBK54Nnpvz5b1JWr",
				publicKey: shard03Proposer.MiningPubKey[common.BridgeConsensus],
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

	shard00Prk := []byte{138, 126, 157, 106, 5, 216, 24, 227, 7, 231, 55, 225, 6, 128, 93, 0, 177, 210, 100, 44, 213, 101, 8, 228, 153, 167, 10, 77, 167, 237, 133, 79}
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
				privateKey: shard00Prk,
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
				privateKey: shard00Prk,
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

func TestFinalityProof_Verify(t *testing.T) {
	type fields struct {
		ReProposeHashSignature []string
	}
	type args struct {
		previousBlockHash common.Hash
		producer          string
		beginTimeSlot     int64
		proposers         []string
		rootHash          common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "testcase 1, 28 re-propose proof",
			fields: fields{
				ReProposeHashSignature: []string{
					"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",
				},
			},
			args: args{
				previousBlockHash: common.Hash{}.NewHashFromStr2("1dc43704ebd3ffd6d35be478f86740fc0cbf5b2f825e66fe3c83f31e3e4d18d8"),
				producer:          shard00,
				beginTimeSlot:     163271008,
				proposers: []string{
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02,
				},
				rootHash: common.Hash{}.NewHashFromStr2("874615a2fcacb79434fb851a137da03055eddd9bd086d7b60e6f74a565c887b9"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FinalityProof{
				ReProposeHashSignature: tt.fields.ReProposeHashSignature,
			}
			if err := f.Verify(tt.args.previousBlockHash, tt.args.producer, tt.args.beginTimeSlot, tt.args.proposers, tt.args.rootHash); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
