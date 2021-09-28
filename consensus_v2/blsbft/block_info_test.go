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
		{
			name: "testcase 2, 48 re-propose proof",
			fields: fields{
				ReProposeHashSignature: []string{
					"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",
				},
			},
			args: args{
				previousBlockHash: common.Hash{}.NewHashFromStr2("c508aba84bb65bb4ee1524e50df8bad054c67bc1b0159e7d9e071a781ecc669b"),
				producer:          shard00,
				beginTimeSlot:     163274040,
				proposers: []string{
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
					shard00, shard01, shard02, shard03, shard04, shard05, shard06, shard07,
				},
				rootHash: common.Hash{}.NewHashFromStr2("ee62ba679f7a611bf9d5a2fc51e27637e7040a9e8a7dd006953db07b59ac20dc"),
			},
		},
		{
			name: "testcase 3, 62 re-propose proof",
			fields: fields{
				ReProposeHashSignature: []string{
					"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",
				},
			},
			args: args{
				previousBlockHash: common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"),
				producer:          shard06,
				beginTimeSlot:     163275326,
				proposers: []string{
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03, shard04, shard05,
					shard06, shard07, shard00, shard01, shard02, shard03,
				},
				rootHash: common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"),
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
