package btcrelaying

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestMerkleVerification1(t *testing.T) {
	contents := []string{"1", "2", "3", "4", "5"}
	hashes := make([]*chainhash.Hash, len(contents))
	for i, content := range contents {
		hashes[i], _ = chainhash.NewHashFromStr(content)
		fmt.Println(hashes[i].String())
	}

	merkleTree := buildMerkleTreeStoreFromTxHashes(hashes)
	mklRoot := merkleTree[len(merkleTree)-1]
	targetedTxHash, _ := chainhash.NewHashFromStr("3")
	mklProof := buildMerkleProof(hashes, targetedTxHash)

	for _, proof := range mklProof {
		fmt.Println("proof str: ", proof.ProofHash.String())
		fmt.Println("is left: ", proof.IsLeft)
	}

	result := verify(mklRoot, mklProof, targetedTxHash)
	fmt.Println("result: ", result)
}

func TestMerkleVerification2(t *testing.T) {
	contents := []string{"1", "2", "3", "4"}
	hashes := make([]*chainhash.Hash, len(contents))
	for i, content := range contents {
		hashes[i], _ = chainhash.NewHashFromStr(content)
		fmt.Println(hashes[i].String())
	}

	merkleTree := buildMerkleTreeStoreFromTxHashes(hashes)
	mklRoot := merkleTree[len(merkleTree)-1]
	targetedTxHash, _ := chainhash.NewHashFromStr("5")
	mklProof := buildMerkleProof(hashes, targetedTxHash)

	for _, proof := range mklProof {
		fmt.Println("proof str: ", proof.ProofHash.String())
		fmt.Println("is left: ", proof.IsLeft)
	}

	result := verify(mklRoot, mklProof, targetedTxHash)
	fmt.Println("result: ", result)

}

func TestMerkleVerification3(t *testing.T) {
	contents := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20"}
	hashes := make([]*chainhash.Hash, len(contents))
	for i, content := range contents {
		hashes[i], _ = chainhash.NewHashFromStr(content)
		fmt.Println(hashes[i].String())
	}

	merkleTree := buildMerkleTreeStoreFromTxHashes(hashes)
	mklRoot := merkleTree[len(merkleTree)-1]
	targetedTxHash, _ := chainhash.NewHashFromStr("18")
	mklProof := buildMerkleProof(hashes, targetedTxHash)

	for _, proof := range mklProof {
		fmt.Println("proof str: ", proof.ProofHash.String())
		fmt.Println("is left: ", proof.IsLeft)
	}

	result := verify(mklRoot, mklProof, targetedTxHash)
	fmt.Println("result: ", result)
}

func TestBuildTxMsg(t *testing.T) {
	txID := "8bae12b5f4c088d940733dcd1455efc6a3a69cf9340e17a981286d3778615684"
	msgTx := buildMsgTxFromCypher(txID, "main")
	if msgTx.TxHash().String() != txID {
		t.Errorf("Want tx hash %s but got %s", txID, msgTx.TxHash())
	}
}
