package signaturecounter

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"sync"
	"testing"
)

var (
	committeePublicKeys = []string{
		"121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM",
		"121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy",
		"121VhftSAygpEJZ6i9jGkGLcYhJBeaJTGY5aFjqQA2WwyxU69Utrviuy9AJ3ATkeEyigVGScQUZw22cD1HeFKiyASYAs82WEamujt3nefYA9FPhURBpRTn6jDmGKUdb4QNbs7HVCJkRRaL9aktg1yaQaZE8TJFg2UeE9tBqUdmvD8fy36aDCYM5W86jaTVCXeEJQWPxUunP2EEL3e283PJ8zqPeBkpoFvkvhB28Hk3oRDeCCTC7QhbaV18ayKeToYqAxoUMBBihanfA33ixeX1daeKpajLCgDZ6jrfphwdYwQbf7dMcZ2NVvQ1a5JUCTJUZypwgKRt8tnTAKCowt2L1KNGP4NJJZm61cfHAGbKRyG9QxCJgK2SdMKsKPVefZSc9LbVaB7VeBby5LHxvMoCD7bN7g1HYRp4BX9n1fZJUeEkVa",
		"121VhftSAygpEJZ6i9jGkDjJj7e2cfgQvrLsPsmLhGMmGD9U9Knffa1MZAw79EijnpueVfTStN2VYt5jRqEr2DTjVqzUinwHVKWH4Tg4szHUntiBdWeqzNC4E8iiwC9Y2KtcRr3hBkpfqvyuBvchigatrigRvFVWu8H2RQqjvopLL51DQ4LFD87L9Zgj9HhasMeyr6f37yirs47JgtGs4BM7EhhpM5zD3TCsFabPphtwDKnfuLMaGzoAw5fM8zEXvdLMuohk96oayjdYothncdtZom17DxB1Mmw535eEjxBwz9ELoZRKk3LYiheSd4xGN9QsxrT2WnZCTd8B5QktARte5S91QYvRMixKC8UEuovQhXt8jMZNkq7CmMeXoybfYdmNaAHuqbY1QeUT2AgaqPho4ay3z5eeKRhnB28H18RGWQ1L",
	}
	samplePenaltyRule = []Penalty{
		{
			MinRange:     800,
			Time:         302400,
			ForceUnstake: false,
		},
		{
			MinRange:     1500,
			Time:         302400 * 2,
			ForceUnstake: false,
		},
		{
			MinRange:     3001,
			Time:         302400 * 2,
			ForceUnstake: true,
		},
	}
	committeePublicKeyStructs = []incognitokey.CommitteePublicKey{}
)

var _ = func() (_ struct{}) {
	committeePublicKeyStructs, _ = incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	return
}()

func TestSignatureCounter_AddMissingSignature(t *testing.T) {
	type fields struct {
		missingSignature           map[string]uint
		aggregatedMissingSignature map[string]uint
	}
	type args struct {
		data       string
		committees []incognitokey.CommitteePublicKey
	}

	aggregatedMissingSignature := make(map[string]uint)

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFields fields
		wantErr    bool
	}{
		{
			name: "invalid input 1",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"1I6pNHXngYdBKspO08xZvk3fasdklaw;dkl;alwkd;lawkdl;kawl;dkkAaRQ9VpD+GhmwfT2b8p3PIYzouW4q/BFDxinllrIwUqq+XpugEiDjdmpfsHCAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"LuFMS0uCziQOC/AL83xZb0Mortu+3lvx5mZ/kCtyJWE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{},
			},
			wantErr: true,
		},
		{
			name: "invalid input 2",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"1I6pNHXngYdBKspO08xZvk3fkAaRQ9VpD+GhmwfT2b8p3PIYzouW4q/BFDxinllrIwUqq+XpugEiDjdmpfsHCAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"LuFMS0uCziQOC/AL83xZb0Mortuawdawdawd+3lvx5mZ/kCtyJWE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{},
			},
			wantErr: true,
		},
		{
			name: "valid input, committee slot 3 miss 1 signature",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"1I6pNHXngYdBKspO08xZvk3fkAaRQ9VpD+GhmwfT2b8p3PIYzouW4q/BFDxinllrIwUqq+XpugEiDjdmpfsHCAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"LuFMS0uCziQOC/AL83xZb0Mortu+3lvx5mZ/kCtyJWE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[3]: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 3 miss 1 signature",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"4lEXt6Z5RwRJmG7vK/6q2pLwGc0EcWi3Pw2D+rYvwBM/3YwgDjElAnH8Qb2OrAX4Lx3APk0Wo3oHYp1eO9hj7gA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"B93JfdZq3Q110tbR4fC7BWQim3NYICJRG/DZ3xlHw04=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[3]: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 2 miss 1 signature",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"LGcjV69UWOBv90wEVFgeq8pMNRWXaxqVPr82g1wqWA5XMmbdq7TZzECtPJl8pCkrSyzQnGVduAVaODGQrykTNQE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,3],\"AggSig\":\"Flod04E7A67JW4uPp43RGGLJR6j5ZnS8ZMrmz7MdE/A=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[2]: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"HrpGEaXOUzydou9S9YE96OD48dSAtgI3zzIC2eisytQJJhtj0MgEwqU9MP1HswRk87NW3msE8w7Uyi7C+npWogA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"HTraoh3hx22W3iRl3SB9a7kv+p1N+ESGodAp28yjRDk=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[1]: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"uSpMynim78XpsufwR6imkWcNKT6c5wwz4Nyb1GR+d3FplCfBwSQXNCd3bCgNGhieBuwGqSg5C5KG+zThOpY4rAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"ocFaeoEmrzq0Ivg1N5gAvkuW4xsyDnC+NQiDUnYqQPE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[1]: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: make(map[string]uint),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"5fp+nanu4VJoVIU5ZpA+uRASzkrjJgZMZ5eZOfYY5kwRWfnhWW4HlZhZdJ+dw2nzVzoR0KTyiG4Hno+TfMPvewE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"idOzTlb8oEoL6VsZ7UsQdPiFVf8HUX4Pad+8xxlE1/0=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[1]: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignatureCounter{
				missingSignature: tt.fields.missingSignature,
				lock:             new(sync.RWMutex),
			}
			if err := s.AddMissingSignature(tt.args.data, tt.args.committees); (err != nil) != tt.wantErr {
				t.Errorf("AddMissingSignature() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(s.missingSignature, tt.wantFields.missingSignature) {
					t.Errorf("AddMissingSignature() missingSignature = got %v, want %v", s.missingSignature, tt.wantFields.missingSignature)
				}
			}
			for k, _ := range s.missingSignature {
				aggregatedMissingSignature[k] += 1
			}
		})
	}
	wantAggregatedMissingSignature := map[string]uint{
		committeePublicKeys[1]: 3,
		committeePublicKeys[2]: 1,
		committeePublicKeys[3]: 2,
	}
	for wantK, wantV := range wantAggregatedMissingSignature {
		if gotV, ok := aggregatedMissingSignature[wantK]; !ok {
			t.Errorf("AddMissingSignature() missingSignature NOT FOUND want %v ", wantK)
		} else {
			if wantV != gotV {
				t.Errorf("AddMissingSignature() number of missingSignature got = %+v, want = %v ", gotV, wantK)
			}
		}
	}
}

func TestSignatureCounter_GetAllSlashingPenalty(t *testing.T) {
	type fields struct {
		missingSignature map[string]uint
		penalties        []Penalty
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]Penalty
	}{
		{
			name: "no penalty",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 799,
					committeePublicKeys[1]: 0,
					committeePublicKeys[2]: 798,
					committeePublicKeys[3]: 1,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{},
		},
		{
			name: "penalty range 800",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 799,
					committeePublicKeys[1]: 800,
					committeePublicKeys[2]: 798,
					committeePublicKeys[3]: 801,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[1]: samplePenaltyRule[0],
				committeePublicKeys[3]: samplePenaltyRule[0],
			},
		},
		{
			name: "penalty range 800, 1500",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 1500,
					committeePublicKeys[1]: 800,
					committeePublicKeys[2]: 1499,
					committeePublicKeys[3]: 1501,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[1],
				committeePublicKeys[1]: samplePenaltyRule[0],
				committeePublicKeys[2]: samplePenaltyRule[0],
				committeePublicKeys[3]: samplePenaltyRule[1],
			},
		},
		{
			name: "penalty range 1500, 3000",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 1500,
					committeePublicKeys[1]: 3000,
					committeePublicKeys[2]: 3001,
					committeePublicKeys[3]: 2999,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[1],
				committeePublicKeys[1]: samplePenaltyRule[1],
				committeePublicKeys[2]: samplePenaltyRule[2],
				committeePublicKeys[3]: samplePenaltyRule[1],
			},
		},
		{
			name: "penalty range 800, 1500, 3000",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 800,
					committeePublicKeys[1]: 3000,
					committeePublicKeys[2]: 3001,
					committeePublicKeys[3]: 2999,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[0],
				committeePublicKeys[1]: samplePenaltyRule[1],
				committeePublicKeys[2]: samplePenaltyRule[2],
				committeePublicKeys[3]: samplePenaltyRule[1],
			},
		},
		{
			name: "no penalty, penalty range 800, 1500, 3000",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 800,
					committeePublicKeys[1]: 799,
					committeePublicKeys[2]: 3001,
					committeePublicKeys[3]: 2999,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[0],
				committeePublicKeys[2]: samplePenaltyRule[2],
				committeePublicKeys[3]: samplePenaltyRule[1],
			},
		},
		{
			name: "no penalty, penalty range 1500, 3000",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 1501,
					committeePublicKeys[1]: 799,
					committeePublicKeys[2]: 3001,
					committeePublicKeys[3]: 2999,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[1],
				committeePublicKeys[2]: samplePenaltyRule[2],
				committeePublicKeys[3]: samplePenaltyRule[1],
			},
		},
		{
			name: "penalty range 3000",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 30000,
					committeePublicKeys[1]: 3002,
					committeePublicKeys[2]: 3001,
					committeePublicKeys[3]: 30003,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[2],
				committeePublicKeys[1]: samplePenaltyRule[2],
				committeePublicKeys[2]: samplePenaltyRule[2],
				committeePublicKeys[3]: samplePenaltyRule[2],
			},
		},
		{
			name: "penalty range 1500",
			fields: fields{
				missingSignature: map[string]uint{
					committeePublicKeys[0]: 1500,
					committeePublicKeys[1]: 1501,
					committeePublicKeys[2]: 2999,
					committeePublicKeys[3]: 3000,
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[1],
				committeePublicKeys[1]: samplePenaltyRule[1],
				committeePublicKeys[2]: samplePenaltyRule[1],
				committeePublicKeys[3]: samplePenaltyRule[1],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SignatureCounter{
				missingSignature: tt.fields.missingSignature,
				penalties:        tt.fields.penalties,
				lock:             new(sync.RWMutex),
			}
			if got := s.GetAllSlashingPenalty(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllSlashingPenalty() = %v, want %v", got, tt.want)
			}
		})
	}
}
