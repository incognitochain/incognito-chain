package relaying

import (
	"encoding/hex"
	"fmt"
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


