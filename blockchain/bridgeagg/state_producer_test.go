package bridgeagg

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func Test_stateProducer_unshield(t *testing.T) {
	initLog()
	type args struct {
		contentStr        string
		unifiedTokenInfos map[common.Hash]map[uint]*Vault
		beaconHeight      uint64
		shardID           byte
		stateDB           *statedb.StateDB
	}
	tests := []struct {
		name                     string
		sp                       *stateProducer
		args                     args
		wantResInsts             [][]string
		wantResUnifiedTokenInfos map[common.Hash]map[uint]*Vault
		wantErr                  bool
	}{
		{
			name: "[Rejected] Not find unifiedTokenID",
			sp:   &stateProducer{},
			args: args{
				contentStr:        "eyJNZXRhIjp7IlRva2VuSUQiOiJiMzU3NTY0NTJkYzFmYTEyNjA1MTNmYTEyMWMyMGMyYjUxNmE4NjQ1ZjhkNDk2ZmE0MjM1Mjc0ZGFjMGIxYjUyIiwiRGF0YSI6W3siQnVybmluZ0Ftb3VudCI6MzAwLCJSZW1vdGVBZGRyZXNzIjoiY0U0MGNFNTExQTVEMDg0MDE3REJlZTdlM2ZGM2U0NTVlYTMyRDg1YyIsIklzRGVwb3NpdFRvU0MiOmZhbHNlLCJOZXR3b3JrSUQiOjMsIkV4cGVjdGVkQW1vdW50Ijo4fV0sIlJlY2VpdmVyIjoiMTVnbTg4Y1RaYnB6V0R0anc4ZEdCbWJ5U0JRWW5lM3pKTWd0U0ZBSkJpaEs1MXBQTkhwUWNNYWI0ZHZpUEpmemFVTDM4Q1NYSmhhcFRDVzZOc1RBTVZOWTNpMjFXejNDdDQ4REtSdUpyb0M2cUd3S1ZzVll4Mk1kOGd5amNrdU1uUXlROTFkZ3ZwRDY1Q2paIiwiVHlwZSI6MzQ1fSwiVHhSZXFJRCI6ImUyMzkyOGEzZWY1MTIxZDU1Nzg2ZGIxMTdiY2FkMDVhNmRlMmEzYTgyMDhiNmM3OTQ5OTBjMDIxNWNmZjM2ODQifQ==",
				unifiedTokenInfos: map[common.Hash]map[uint]*Vault{},
				beaconHeight:      10,
				shardID:           1,
				stateDB:           nil,
			},
			wantResInsts: [][]string{
				{
					"346", "1", "rejected", "eyJUeFJlcUlEIjoiZTIzOTI4YTNlZjUxMjFkNTU3ODZkYjExN2JjYWQwNWE2ZGUyYTNhODIwOGI2Yzc5NDk5MGMwMjE1Y2ZmMzY4NCIsIkVycm9yQ29kZSI6MTAwMCwiRGF0YSI6ImV5SlViMnRsYmtsRUlqb2lZak0xTnpVMk5EVXlaR014Wm1FeE1qWXdOVEV6Wm1FeE1qRmpNakJqTW1JMU1UWmhPRFkwTldZNFpEUTVObVpoTkRJek5USTNOR1JoWXpCaU1XSTFNaUlzSWtGdGIzVnVkQ0k2TXpBd0xDSlNaV05sYVhabGNpSTZJakUxWjIwNE9HTlVXbUp3ZWxkRWRHcDNPR1JIUW0xaWVWTkNVVmx1WlRONlNrMW5kRk5HUVVwQ2FXaExOVEZ3VUU1SWNGRmpUV0ZpTkdSMmFWQktabnBoVlV3ek9FTlRXRXBvWVhCVVExYzJUbk5VUVUxV1Rsa3phVEl4VjNvelEzUTBPRVJMVW5WS2NtOURObkZIZDB0V2MxWlplREpOWkRobmVXcGphM1ZOYmxGNVVUa3haR2QyY0VRMk5VTnFXaUo5In0=",
				},
			},
			wantResUnifiedTokenInfos: map[common.Hash]map[uint]*Vault{},
			wantErr:                  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducer{}
			gotResInsts, gotResUnifiedTokenInfos, err := sp.unshield(tt.args.contentStr, tt.args.unifiedTokenInfos, tt.args.beaconHeight, tt.args.shardID, tt.args.stateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducer.unshield() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResInsts, tt.wantResInsts) {
				t.Errorf("stateProducer.unshield() gotResInsts = %v, want %v", gotResInsts, tt.wantResInsts)
			}
			if !reflect.DeepEqual(gotResUnifiedTokenInfos, tt.wantResUnifiedTokenInfos) {
				t.Errorf("stateProducer.unshield() gotResUnifiedTokenInfos = %v, want %v", gotResUnifiedTokenInfos, tt.wantResUnifiedTokenInfos)
			}
		})
	}
}
