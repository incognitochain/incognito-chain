package incognitokey_test

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestCommitteeBase58KeyListToStruct(t *testing.T) {
	strs1 := "1Q2cBtAJZUrbFAgm31NMTk4faWjLNfkpX984gFsGsVe4R2Ub5wLbgWSg2av8XiyNjZn8cnG29SDXYqGJcZ2EWUTWPoSs1BKhotbrHAna3dxN8GsAqDMbi9x1CeyyFCwHY32C4vGhog7nrpB1Lz68N7yocJCToEhiod7GVpgBCSXEUSo67DP9R2kdKvH8Beq2AC32guQm5MnrdAimA6HXdfKVbpevG1TXTE4v9nwAWnEzXhzmWjSrb9zRhGmua7qEYnG5kj59z7veqBbStPSYNQb9kZ6mVXSqnjshsEtaEmZ3CJJFt2w7bH9iomcYRnPt5rVAaV6X7sfm5cUEQisT8zYAfnXVPkv2EVKiTSuZrCy8PQNa4mAp6iyyG7qwiXFreKb3iaUD7HikVtmPj4TaNvv6rEBPdThiqe4rm7e1amM7oeZz2m3uqpG5LRiGppbWjJtawtCw95NVVCLmsUZDcSmPuBMaQy2VxxkKPafoaMW1DdeUzCCMPeSaJrGuSNeiYRF1zwYyTAbPwYuBBr1rWeHhE3E13DEpjcsuXiw7e5s9bgZcuGkyaZx3TrQTd35MsoxcpN3JUG6SidTptVDSikUtH7BvQRZ9mNE2tLbY4b1ppZZ8ZPy8b8H4NHXirvFwLhxL7eWtBgUzGD9qZoHqQjdVtKq272w4iYeY4f9h9DjzLdZJ7jpJFcvWHt7WE4PEnn73N5x2NHG3K5JDQAtXMsRh7NXVZTcP6d3XvugDxYwZjbpHjb9oJuLypMaQCjNqc7E97tkbK5JMqaVreLufX3jorD46C7EhaM5gAKfGECiY3jirVq2HvutDazQPSbh92hKPWwMKK7qJTh7bTsu75S4kfqaWUheo7JPLRPB89BdM28gjRgucLeUsBqL1MvMELV5FZVtsTCFBBQRyGRxv3dgN235MXcAjniJfV88bMkMtCncvAMN9rYqX6SZ2Fectacx6m3Xu3rxjyFkGp48G5Jh7Nqg63bVFTNa1ZtVmYQiftKtwmQWjN"
	tempStruct, err := incognitokey.CommitteeBase58KeyListToStruct([]string{strs1})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Temp Struct %+v", tempStruct)
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

}
