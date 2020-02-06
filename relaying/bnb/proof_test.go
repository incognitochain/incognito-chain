package relaying

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTxByHash(t *testing.T){
	getProofFromTxHash("24B93E8B6C5817B159870E5C617597EBD0BDAE100430DB8242BFBA5DA37D70CE")
}

func TestGetTxsInBlockHeight(t *testing.T){
	txs, _ := getTxsInBlockHeight(66239775)
	fmt.Printf("Len txs: %v\n", txs)
}

func TestGetBlock(t *testing.T){
	getBlock(66239775)
}

func TestBuildProof1_VerifyProof(t *testing.T){
	//txIndex := 0
	txHash := "24B93E8B6C5817B159870E5C617597EBD0BDAE100430DB8242BFBA5DA37D70CE"
	blockHeight := int64(60479432)
	dataHash, _ := hex.DecodeString("D81AD27D7C1D8114EB339158897C02337820BC17E10AB6405143EFE8E52AB526")
	proof, err := BuildProof1(txHash)
	assert.Nil(t, err)

	isValid, err := VerifyProof(proof, blockHeight, dataHash)
	fmt.Printf("err: %+v\n", err)
	assert.Nil(t, err)
	assert.Equal(t, true, isValid)
}

func TestBuildProof2_VerifyProof(t *testing.T){
	txIndex := 0
	//txHash := "24B93E8B6C5817B159870E5C617597EBD0BDAE100430DB8242BFBA5DA37D70CE"
	blockHeight := int64(60479432)
	dataHash, _ := hex.DecodeString("D81AD27D7C1D8114EB339158897C02337820BC17E10AB6405143EFE8E52AB526")
	proof, err := BuildProof2(txIndex, blockHeight)
	assert.Nil(t, err)
	//fmt.Printf("Proof1: %+v\n", proof)

	isValid, err := VerifyProof(proof, blockHeight, dataHash)
	assert.Nil(t, err)
	assert.Equal(t, true, isValid)
}

func TestBuildProof1_BuildProof2(t *testing.T){
	txIndex := 0
	txHash := "24B93E8B6C5817B159870E5C617597EBD0BDAE100430DB8242BFBA5DA37D70CE"
	blockHeight := int64(60479432)
	dataHash, _ := hex.DecodeString("D81AD27D7C1D8114EB339158897C02337820BC17E10AB6405143EFE8E52AB526")
	proof1, err := BuildProof1(txHash)
	assert.Nil(t, err)

	proof2, err := BuildProof2(txIndex, blockHeight)
	assert.Nil(t, err)

	assert.Equal(t, proof1, proof2)

	isValid1, err := VerifyProof(proof1, blockHeight, dataHash)
	assert.Nil(t, err)
	assert.Equal(t, true, isValid1)

	isValid2, err := VerifyProof(proof2, blockHeight, dataHash)
	assert.Nil(t, err)
	assert.Equal(t, true, isValid2)
}

func TestParseDataToTx(t *testing.T) {
	data := `
		pwLwYl3uCkgqLIf6CiAKFBxGk+JFWp2mPF2PEkC+PYRmzQ5GEggKA0JOQhCQThIgChQcRpPiRVqdpjxdjxJAvj2EZs0ORhIICgNCTkIQkE4ScQom61rphyEDeYW1MIWu9puLSBtb81veeyDb+Y25cJCQSPclhJQS46wSQIEg/andMybUQMjQWIUa+1gwvG4S74lrUT7gCV1HJz3/CmQDgqhI6bYo9rJGclFcM3dqvjjqpfoYjchGfaGmPDYYolwg0OwDGmR7InByb3RvY29sIjoiYWxlcGgtb2ZmY2hhaW4iLCJ2ZXJzaW9uIjoxLCJjb250ZW50IjoiUW1jZExaaFNzbUc2THJEWkQ5c2RhamRwd2FiZjRZeDVWUEg5UDdhUjk3ZTJNUiJ9
	`
	dataBytes, _ := base64.StdEncoding.DecodeString(data)
	fmt.Printf("dataBytes: %v\n", dataBytes)
	fmt.Printf("dataBytes: %v\n", len(dataBytes))

	data2, _ := hex.DecodeString("A702F0625DEE0A482A2C87FA0A200A141C4693E2455A9DA63C5D8F1240BE3D8466CD0E4612080A03424E4210904E12200A141C4693E2455A9DA63C5D8F1240BE3D8466CD0E4612080A03424E4210904E12710A26EB5AE98721037985B53085AEF69B8B481B5BF35BDE7B20DBF98DB970909048F725849412E3AC12408120FDA9DD3326D440C8D058851AFB5830BC6E12EF896B513EE0095D47273DFF0A640382A848E9B628F6B24672515C33776ABE38EAA5FA188DC8467DA1A63C3618A25C20D0EC031A647B2270726F746F636F6C223A22616C6570682D6F6666636861696E222C2276657273696F6E223A312C22636F6E74656E74223A22516D63644C5A6853736D47364C72445A44397364616A647077616266345978355650483950376152393765324D52227D")

	tx, err := ParseDataToTx(data2)
	fmt.Printf("stdTx: %+v\n", tx)
	fmt.Printf("stdTx.Memo: %v\n", tx.Memo)

	senderAddress := "bnb1r3rf8cj9t2w6v0za3ufyp03as3nv6rjxteku6g"
	amount := int64(10000)
	receiverAddress := "bnb1r3rf8cj9t2w6v0za3ufyp03as3nv6rjxteku6g"
	assert.Nil(t, err)
	assert.Equal(t, senderAddress, tx.Msgs[0].(msg.SendMsg).Inputs[0].Address.String())
	assert.Equal(t, receiverAddress, tx.Msgs[0].(msg.SendMsg).Outputs[0].Address.String())
	outputCoins := tx.Msgs[0].(msg.SendMsg).Outputs[0].Coins
	actualAmount := int64(0)
	for _, coin := range outputCoins{
		actualAmount += coin.Amount
	}
	assert.Equal(t, amount, actualAmount)
}


