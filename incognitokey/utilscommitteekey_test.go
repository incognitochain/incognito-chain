package incognitokey_test

import (
	"encoding/json"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/key"
)

func TestConvertCommitteePublicKeyBetweenStringAndStruct(t *testing.T) {
	strs1 := "1Q2cBtAJZUrbFAgm31NMTk4faWjLNfkpX984gFsGsVe4R2Ub5wLbgWSg2av8XiyNjZn8cnG29SDXYqGJcZ2EWUTWPoSs1BKhotbrHAna3dxN8GsAqDMbi9x1CeyyFCwHY32C4vGhog7nrpB1Lz68N7yocJCToEhiod7GVpgBCSXEUSo67DP9R2kdKvH8Beq2AC32guQm5MnrdAimA6HXdfKVbpevG1TXTE4v9nwAWnEzXhzmWjSrb9zRhGmua7qEYnG5kj59z7veqBbStPSYNQb9kZ6mVXSqnjshsEtaEmZ3CJJFt2w7bH9iomcYRnPt5rVAaV6X7sfm5cUEQisT8zYAfnXVPkv2EVKiTSuZrCy8PQNa4mAp6iyyG7qwiXFreKb3iaUD7HikVtmPj4TaNvv6rEBPdThiqe4rm7e1amM7oeZz2m3uqpG5LRiGppbWjJtawtCw95NVVCLmsUZDcSmPuBMaQy2VxxkKPafoaMW1DdeUzCCMPeSaJrGuSNeiYRF1zwYyTAbPwYuBBr1rWeHhE3E13DEpjcsuXiw7e5s9bgZcuGkyaZx3TrQTd35MsoxcpN3JUG6SidTptVDSikUtH7BvQRZ9mNE2tLbY4b1ppZZ8ZPy8b8H4NHXirvFwLhxL7eWtBgUzGD9qZoHqQjdVtKq272w4iYeY4f9h9DjzLdZJ7jpJFcvWHt7WE4PEnn73N5x2NHG3K5JDQAtXMsRh7NXVZTcP6d3XvugDxYwZjbpHjb9oJuLypMaQCjNqc7E97tkbK5JMqaVreLufX3jorD46C7EhaM5gAKfGECiY3jirVq2HvutDazQPSbh92hKPWwMKK7qJTh7bTsu75S4kfqaWUheo7JPLRPB89BdM28gjRgucLeUsBqL1MvMELV5FZVtsTCFBBQRyGRxv3dgN235MXcAjniJfV88bMkMtCncvAMN9rYqX6SZ2Fectacx6m3Xu3rxjyFkGp48G5Jh7Nqg63bVFTNa1ZtVmYQiftKtwmQWjN"
	strs2 := "121VhftSAygpEJZ6i9jGk9eXD8TsLogbjJKUXNFRj6U5Jiumpy2Q9LfuoNVrXGUsr8zXYNQRSZzbVX2sUJKVFnYZwC31W4Rx21cpJ46hTTJQx9o2BQvQ5hDfnmj6tmQ9gUj1VatLzUPM7pLGcJ4XtJHjZnHNoqrihvesfcK8i69GnkbSeGmtxNbxYZw6ToQgLqwbporAGMvH5QxtaeRe9xPGW81qQy2qH4KPjm67GznbaVaGXR96j3XxZbMCgr78WKU262GgvSrWRcZcb7ADSfVAej9mU8pbpF4SfzHgL3DzMise7SS3DWKBe4pmMePAvRrgmMG2z196SGYVjRiTMJEoRG2oaovGBWRduWgEj5xbscTnKrNyxDTbxFkdyFefw9MtbyyzVUsk3LUzqnSuzxomtqRNuspvL9J31Z5nuGw7e69B"
	strs3 := "121VhftSAygpEJZ6i9jGk9efCedRUdQkZrUmi6qFGvN2fvdVCfYW9wGJPAejJ7EgCXNnHcKoXiNbbAX47srDwgfi9UU5qop5MbjR9sxCcuocWqN3TisyV7hmuu4En9yyQnDAbHvx6StNckrLxsk266yb7qDStg18Hecz32GsEoL3ntmJgdYVoX2Ni4e894hXbegnuVTwDYuGsKQbLaLgo4KDH43rUgd6yd8PCK4UQbCZvyR7bZtLZA2JsXz5uBzj5mGUnhSYkMBcSbeTM2LJKyyj9qemkZRYWYZkQrX8D44dmDNXgbftjLZscepq5vj2NTLFMGGH6Rm5QxmxHavnyVJ8tMwP7w71FNhz6ZzrJFnyxrkQTT3emx8fkLEFvksPFYyp3Ac9G6NpjggiYyEHgpu2SFnGxXwsLqunsdsBNou2KE9N"
	tempStruct, err := incognitokey.CommitteeBase58KeyListToStruct([]string{strs1})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Temp Struct %+v", tempStruct)
	tempStruct2, err := incognitokey.CommitteeBase58KeyListToStruct([]string{strs2})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Temp Struct %+v", tempStruct2)
	tempStruct3, err := incognitokey.CommitteeBase58KeyListToStruct([]string{strs3})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Temp Struct %+v", tempStruct3)
	tempStructToString, _ := json.Marshal(tempStruct)
	t.Logf("Temp Struct -> String %+v", string(tempStructToString))
	t.Logf("Temp Struct Incognito Publickey %+v", tempStruct[0].GetIncKeyBase58())
	tempString, err := incognitokey.CommitteeKeyListToString(tempStruct)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tempString)
	revertTempStruct, err := incognitokey.CommitteeBase58KeyListToStruct(tempString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(revertTempStruct)
	t.Log(revertTempStruct[0].GetIncKeyBase58())
	c1 := new(incognitokey.CommitteePublicKey)
	err1 := c1.FromString(strs1)
	if err1 != nil {
		t.Fatal(err1)
	}
	c2 := new(incognitokey.CommitteePublicKey)
	err2 := c2.FromString(tempString[0])
	if err2 != nil {
		t.Fatal(err2)
	}
	if incognitokey.IsInBase58ShortFormat([]string{strs1}) {
		t.Fatal()
	}
	shotFormatStrs, err := incognitokey.ConvertToBase58ShortFormat(append(tempString, strs1))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(shotFormatStrs)
	t.Log(shotFormatStrs[0] == shotFormatStrs[1])
}
func TestUnmarshallkey(t *testing.T) {
	strs1 := "1Q2cBtAJZUrbFAgm31NMTk4faWjLNfkpX984gFsGsVe4R2Ub5wLbgWSg2av8XiyNjZn8cnG29SDXYqGJcZ2EWUTWPoSs1BKhotbrHAna3dxN8GsAqDMbi9x1CeyyFCwHY32C4vGhog7nrpB1Lz68N7yocJCToEhiod7GVpgBCSXEUSo67DP9R2kdKvH8Beq2AC32guQm5MnrdAimA6HXdfKVbpevG1TXTE4v9nwAWnEzXhzmWjSrb9zRhGmua7qEYnG5kj59z7veqBbStPSYNQb9kZ6mVXSqnjshsEtaEmZ3CJJFt2w7bH9iomcYRnPt5rVAaV6X7sfm5cUEQisT8zYAfnXVPkv2EVKiTSuZrCy8PQNa4mAp6iyyG7qwiXFreKb3iaUD7HikVtmPj4TaNvv6rEBPdThiqe4rm7e1amM7oeZz2m3uqpG5LRiGppbWjJtawtCw95NVVCLmsUZDcSmPuBMaQy2VxxkKPafoaMW1DdeUzCCMPeSaJrGuSNeiYRF1zwYyTAbPwYuBBr1rWeHhE3E13DEpjcsuXiw7e5s9bgZcuGkyaZx3TrQTd35MsoxcpN3JUG6SidTptVDSikUtH7BvQRZ9mNE2tLbY4b1ppZZ8ZPy8b8H4NHXirvFwLhxL7eWtBgUzGD9qZoHqQjdVtKq272w4iYeY4f9h9DjzLdZJ7jpJFcvWHt7WE4PEnn73N5x2NHG3K5JDQAtXMsRh7NXVZTcP6d3XvugDxYwZjbpHjb9oJuLypMaQCjNqc7E97tkbK5JMqaVreLufX3jorD46C7EhaM5gAKfGECiY3jirVq2HvutDazQPSbh92hKPWwMKK7qJTh7bTsu75S4kfqaWUheo7JPLRPB89BdM28gjRgucLeUsBqL1MvMELV5FZVtsTCFBBQRyGRxv3dgN235MXcAjniJfV88bMkMtCncvAMN9rYqX6SZ2Fectacx6m3Xu3rxjyFkGp48G5Jh7Nqg63bVFTNa1ZtVmYQiftKtwmQWjN"
	keyBytes1, ver, err := base58.Base58Check{}.Decode(strs1)
	if (ver != common.ZeroByte) || (err != nil) {
		t.Fatal()
	}
	interface1 := make(map[string]interface{})
	err = json.Unmarshal(keyBytes1, &interface1)
	if err != nil {
		t.Fatal(err)
	}
	tempIncKey1Bytes := interface1["IncPubKey"].([]interface{})
	incKey1Bytes := []byte{}
	for _, v := range tempIncKey1Bytes {
		incKey1Bytes = append(incKey1Bytes, byte(v.(float64)))
	}
	t.Log(incKey1Bytes)

	t.Log(interface1)
	strs2 := "121VhftSAygpEJZ6i9jGk9eXD8TsLogbjJKUXNFRj6U5Jiumpy2Q9LfuoNVrXGUsr8zXYNQRSZzbVX2sUJKVFnYZwC31W4Rx21cpJ46hTTJQx9o2BQvQ5hDfnmj6tmQ9gUj1VatLzUPM7pLGcJ4XtJHjZnHNoqrihvesfcK8i69GnkbSeGmtxNbxYZw6ToQgLqwbporAGMvH5QxtaeRe9xPGW81qQy2qH4KPjm67GznbaVaGXR96j3XxZbMCgr78WKU262GgvSrWRcZcb7ADSfVAej9mU8pbpF4SfzHgL3DzMise7SS3DWKBe4pmMePAvRrgmMG2z196SGYVjRiTMJEoRG2oaovGBWRduWgEj5xbscTnKrNyxDTbxFkdyFefw9MtbyyzVUsk3LUzqnSuzxomtqRNuspvL9J31Z5nuGw7e69B"
	keyBytes2, ver, err := base58.Base58Check{}.Decode(strs2)
	if (ver != common.ZeroByte) || (err != nil) {
		t.Fatal()
	}
	interface2 := make(map[string]interface{})
	err = json.Unmarshal(keyBytes2, &interface2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(interface2)
	incKey := interface2["IncPubKey"].(string)
	t.Log([]byte(incKey))
	incKeyBytes, err := json.Marshal(incKey)
	if err != nil {
		t.Fatal()
	}
	t.Log(incKeyBytes)
}

func TestCommitteePublicKey_IsValid(t *testing.T) {
	type fields struct {
		IncPubKey    key.PublicKey
		MiningPubKey map[string][]byte
	}
	type args struct {
		target incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Valid case",
			fields: fields{
				IncPubKey: []byte{0, 1, 2, 3, 4, 5},
				MiningPubKey: map[string][]byte{
					"BLS": []byte{0, 2, 4, 6, 8},
					"DSA": []byte{0, 2, 4, 6, 8},
				},
			},
			args: args{
				target: incognitokey.CommitteePublicKey{
					IncPubKey: []byte{0, 1, 2, 3, 4, 6},
					MiningPubKey: map[string][]byte{
						"BLS": []byte{0, 2, 4, 6, 9},
						"DSA": []byte{0, 2, 4, 6, 9},
					},
				},
			},
			want: true,
		},
		{
			name: "Invalid case",
			fields: fields{
				IncPubKey: []byte{0, 1, 2, 3, 4, 5},
				MiningPubKey: map[string][]byte{
					"BLS": []byte{0, 2, 4, 6, 8},
					"DSA": []byte{0, 2, 4, 6, 8},
				},
			},
			args: args{
				target: incognitokey.CommitteePublicKey{
					IncPubKey: []byte{0, 1, 2, 3, 4, 5},
					MiningPubKey: map[string][]byte{
						"BLS": []byte{0, 2, 4, 6, 9},
						"DSA": []byte{0, 2, 4, 6, 9},
					},
				},
			},
			want: false,
		},
		{
			name: "Invalid case",
			fields: fields{
				IncPubKey: []byte{0, 1, 2, 3, 4, 5},
				MiningPubKey: map[string][]byte{
					"BLS": []byte{0, 2, 4, 6, 8},
					"DSA": []byte{0, 2, 4, 6, 8},
				},
			},
			args: args{
				target: incognitokey.CommitteePublicKey{
					IncPubKey: []byte{0, 1, 2, 3, 4, 6},
					MiningPubKey: map[string][]byte{
						"BLS": []byte{0, 2, 4, 6, 8},
						"DSA": []byte{0, 2, 4, 6, 9},
					},
				},
			},
			want: false,
		},
		{
			name: "Invalid case",
			fields: fields{
				IncPubKey: []byte{0, 1, 2, 3, 4, 5},
				MiningPubKey: map[string][]byte{
					"BLS": []byte{0, 2, 4, 6, 8},
					"DSA": []byte{0, 2, 4, 6, 8},
				},
			},
			args: args{
				target: incognitokey.CommitteePublicKey{
					IncPubKey: []byte{0, 1, 2, 3, 4, 6},
					MiningPubKey: map[string][]byte{
						"BLS": []byte{0, 2, 4, 6, 6},
						"DSA": []byte{0, 2, 4, 6, 8},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			committeePublicKey := &incognitokey.CommitteePublicKey{
				IncPubKey:    tt.fields.IncPubKey,
				MiningPubKey: tt.fields.MiningPubKey,
			}
			if got := committeePublicKey.IsValid(tt.args.target); got != tt.want {
				t.Errorf("CommitteePublicKey.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
