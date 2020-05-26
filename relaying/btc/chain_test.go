package btcrelaying

import (
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// Block100000 defines block 100,000 of the block chain.  It is used to
// test Block operations.
var Block100000 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version: 1,
		PrevBlock: chainhash.Hash([32]byte{ // Make go vet happy.
			0x50, 0x12, 0x01, 0x19, 0x17, 0x2a, 0x61, 0x04,
			0x21, 0xa6, 0xc3, 0x01, 0x1d, 0xd3, 0x30, 0xd9,
			0xdf, 0x07, 0xb6, 0x36, 0x16, 0xc2, 0xcc, 0x1f,
			0x1c, 0xd0, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00,
		}), // 000000000002d01c1fccc21636b607dfd930d31d01c3a62104612a1719011250
		MerkleRoot: chainhash.Hash([32]byte{ // Make go vet happy.
			0x66, 0x57, 0xa9, 0x25, 0x2a, 0xac, 0xd5, 0xc0,
			0xb2, 0x94, 0x09, 0x96, 0xec, 0xff, 0x95, 0x22,
			0x28, 0xc3, 0x06, 0x7c, 0xc3, 0x8d, 0x48, 0x85,
			0xef, 0xb5, 0xa4, 0xac, 0x42, 0x47, 0xe9, 0xf3,
		}), // f3e94742aca4b5ef85488dc37c06c3282295ffec960994b2c0d5ac2a25a95766
		Timestamp: time.Unix(1293623863, 0), // 2010-12-29 11:57:43 +0000 UTC
		Bits:      0x1b04864c,               // 453281356
		Nonce:     0x10572b0f,               // 274148111
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  chainhash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x04, 0x4c, 0x86, 0x04, 0x1b, 0x02, 0x06, 0x02,
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x12a05f200, // 5000000000
					PkScript: []byte{
						0x41, // OP_DATA_65
						0x04, 0x1b, 0x0e, 0x8c, 0x25, 0x67, 0xc1, 0x25,
						0x36, 0xaa, 0x13, 0x35, 0x7b, 0x79, 0xa0, 0x73,
						0xdc, 0x44, 0x44, 0xac, 0xb8, 0x3c, 0x4e, 0xc7,
						0xa0, 0xe2, 0xf9, 0x9d, 0xd7, 0x45, 0x75, 0x16,
						0xc5, 0x81, 0x72, 0x42, 0xda, 0x79, 0x69, 0x24,
						0xca, 0x4e, 0x99, 0x94, 0x7d, 0x08, 0x7f, 0xed,
						0xf9, 0xce, 0x46, 0x7c, 0xb9, 0xf7, 0xc6, 0x28,
						0x70, 0x78, 0xf8, 0x01, 0xdf, 0x27, 0x6f, 0xdf,
						0x84, // 65-byte signature
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0x03, 0x2e, 0x38, 0xe9, 0xc0, 0xa8, 0x4c, 0x60,
							0x46, 0xd6, 0x87, 0xd1, 0x05, 0x56, 0xdc, 0xac,
							0xc4, 0x1d, 0x27, 0x5e, 0xc5, 0x5f, 0xc0, 0x07,
							0x79, 0xac, 0x88, 0xfd, 0xf3, 0x57, 0xa1, 0x87,
						}), // 87a157f3fd88ac7907c05fc55e271dc4acdc5605d187d646604ca8c0e9382e03
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xc3, 0x52, 0xd3,
						0xdd, 0x99, 0x3a, 0x98, 0x1b, 0xeb, 0xa4, 0xa6,
						0x3a, 0xd1, 0x5c, 0x20, 0x92, 0x75, 0xca, 0x94,
						0x70, 0xab, 0xfc, 0xd5, 0x7d, 0xa9, 0x3b, 0x58,
						0xe4, 0xeb, 0x5d, 0xce, 0x82, 0x02, 0x21, 0x00,
						0x84, 0x07, 0x92, 0xbc, 0x1f, 0x45, 0x60, 0x62,
						0x81, 0x9f, 0x15, 0xd3, 0x3e, 0xe7, 0x05, 0x5c,
						0xf7, 0xb5, 0xee, 0x1a, 0xf1, 0xeb, 0xcc, 0x60,
						0x28, 0xd9, 0xcd, 0xb1, 0xc3, 0xaf, 0x77, 0x48,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xf4, 0x6d, 0xb5, 0xe9, 0xd6, 0x1a, 0x9d,
						0xc2, 0x7b, 0x8d, 0x64, 0xad, 0x23, 0xe7, 0x38,
						0x3a, 0x4e, 0x6c, 0xa1, 0x64, 0x59, 0x3c, 0x25,
						0x27, 0xc0, 0x38, 0xc0, 0x85, 0x7e, 0xb6, 0x7e,
						0xe8, 0xe8, 0x25, 0xdc, 0xa6, 0x50, 0x46, 0xb8,
						0x2c, 0x93, 0x31, 0x58, 0x6c, 0x82, 0xe0, 0xfd,
						0x1f, 0x63, 0x3f, 0x25, 0xf8, 0x7c, 0x16, 0x1b,
						0xc6, 0xf8, 0xa6, 0x30, 0x12, 0x1d, 0xf2, 0xb3,
						0xd3, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x2123e300, // 556000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xc3, 0x98, 0xef, 0xa9, 0xc3, 0x92, 0xba, 0x60,
						0x13, 0xc5, 0xe0, 0x4e, 0xe7, 0x29, 0x75, 0x5e,
						0xf7, 0xf5, 0x8b, 0x32,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x108e20f00, // 4444000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x94, 0x8c, 0x76, 0x5a, 0x69, 0x14, 0xd4, 0x3f,
						0x2a, 0x7a, 0xc1, 0x77, 0xda, 0x2c, 0x2f, 0x6b,
						0x52, 0xde, 0x3d, 0x7c,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0xc3, 0x3e, 0xbf, 0xf2, 0xa7, 0x09, 0xf1, 0x3d,
							0x9f, 0x9a, 0x75, 0x69, 0xab, 0x16, 0xa3, 0x27,
							0x86, 0xaf, 0x7d, 0x7e, 0x2d, 0xe0, 0x92, 0x65,
							0xe4, 0x1c, 0x61, 0xd0, 0x78, 0x29, 0x4e, 0xcf,
						}), // cf4e2978d0611ce46592e02d7e7daf8627a316ab69759a9f3df109a7f2bf3ec3
						Index: 1,
					},
					SignatureScript: []byte{
						0x47, // OP_DATA_71
						0x30, 0x44, 0x02, 0x20, 0x03, 0x2d, 0x30, 0xdf,
						0x5e, 0xe6, 0xf5, 0x7f, 0xa4, 0x6c, 0xdd, 0xb5,
						0xeb, 0x8d, 0x0d, 0x9f, 0xe8, 0xde, 0x6b, 0x34,
						0x2d, 0x27, 0x94, 0x2a, 0xe9, 0x0a, 0x32, 0x31,
						0xe0, 0xba, 0x33, 0x3e, 0x02, 0x20, 0x3d, 0xee,
						0xe8, 0x06, 0x0f, 0xdc, 0x70, 0x23, 0x0a, 0x7f,
						0x5b, 0x4a, 0xd7, 0xd7, 0xbc, 0x3e, 0x62, 0x8c,
						0xbe, 0x21, 0x9a, 0x88, 0x6b, 0x84, 0x26, 0x9e,
						0xae, 0xb8, 0x1e, 0x26, 0xb4, 0xfe, 0x01,
						0x41, // OP_DATA_65
						0x04, 0xae, 0x31, 0xc3, 0x1b, 0xf9, 0x12, 0x78,
						0xd9, 0x9b, 0x83, 0x77, 0xa3, 0x5b, 0xbc, 0xe5,
						0xb2, 0x7d, 0x9f, 0xff, 0x15, 0x45, 0x68, 0x39,
						0xe9, 0x19, 0x45, 0x3f, 0xc7, 0xb3, 0xf7, 0x21,
						0xf0, 0xba, 0x40, 0x3f, 0xf9, 0x6c, 0x9d, 0xee,
						0xb6, 0x80, 0xe5, 0xfd, 0x34, 0x1c, 0x0f, 0xc3,
						0xa7, 0xb9, 0x0d, 0xa4, 0x63, 0x1e, 0xe3, 0x95,
						0x60, 0x63, 0x9d, 0xb4, 0x62, 0xe9, 0xcb, 0x85,
						0x0f, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0xf4240, // 1000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xb0, 0xdc, 0xbf, 0x97, 0xea, 0xbf, 0x44, 0x04,
						0xe3, 0x1d, 0x95, 0x24, 0x77, 0xce, 0x82, 0x2d,
						0xad, 0xbe, 0x7e, 0x10,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x11d260c0, // 299000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x6b, 0x12, 0x81, 0xee, 0xc2, 0x5a, 0xb4, 0xe1,
						0xe0, 0x79, 0x3f, 0xf4, 0xe0, 0x8a, 0xb1, 0xab,
						0xb3, 0x40, 0x9c, 0xd9,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0x0b, 0x60, 0x72, 0xb3, 0x86, 0xd4, 0xa7, 0x73,
							0x23, 0x52, 0x37, 0xf6, 0x4c, 0x11, 0x26, 0xac,
							0x3b, 0x24, 0x0c, 0x84, 0xb9, 0x17, 0xa3, 0x90,
							0x9b, 0xa1, 0xc4, 0x3d, 0xed, 0x5f, 0x51, 0xf4,
						}), // f4515fed3dc4a19b90a317b9840c243bac26114cf637522373a7d486b372600b
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xbb, 0x1a, 0xd2,
						0x6d, 0xf9, 0x30, 0xa5, 0x1c, 0xce, 0x11, 0x0c,
						0xf4, 0x4f, 0x7a, 0x48, 0xc3, 0xc5, 0x61, 0xfd,
						0x97, 0x75, 0x00, 0xb1, 0xae, 0x5d, 0x6b, 0x6f,
						0xd1, 0x3d, 0x0b, 0x3f, 0x4a, 0x02, 0x21, 0x00,
						0xc5, 0xb4, 0x29, 0x51, 0xac, 0xed, 0xff, 0x14,
						0xab, 0xba, 0x27, 0x36, 0xfd, 0x57, 0x4b, 0xdb,
						0x46, 0x5f, 0x3e, 0x6f, 0x8d, 0xa1, 0x2e, 0x2c,
						0x53, 0x03, 0x95, 0x4a, 0xca, 0x7f, 0x78, 0xf3,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xa7, 0x13, 0x5b, 0xfe, 0x82, 0x4c, 0x97,
						0xec, 0xc0, 0x1e, 0xc7, 0xd7, 0xe3, 0x36, 0x18,
						0x5c, 0x81, 0xe2, 0xaa, 0x2c, 0x41, 0xab, 0x17,
						0x54, 0x07, 0xc0, 0x94, 0x84, 0xce, 0x96, 0x94,
						0xb4, 0x49, 0x53, 0xfc, 0xb7, 0x51, 0x20, 0x65,
						0x64, 0xa9, 0xc2, 0x4d, 0xd0, 0x94, 0xd4, 0x2f,
						0xdb, 0xfd, 0xd5, 0xaa, 0xd3, 0xe0, 0x63, 0xce,
						0x6a, 0xf4, 0xcf, 0xaa, 0xea, 0x4e, 0xa1, 0x4f,
						0xbb, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0xf4240, // 1000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x39, 0xaa, 0x3d, 0x56, 0x9e, 0x06, 0xa1, 0xd7,
						0x92, 0x6d, 0xc4, 0xbe, 0x11, 0x93, 0xc9, 0x9b,
						0xf2, 0xeb, 0x9e, 0xe0,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
	},
}


func TestReorganizeChainV2(t *testing.T) {
	// Load up blocks such that there is a side chain.
	// (genesis block) -> 1 -> 2 -> 3
	//                     \-> 2a -> 3a -> 4a
	testFiles := []string{
		"blk_0_to_4.dat.bz2",
		"blk_3A.dat.bz2",
		"blk_4A.dat.bz2",
		"blk_5A.dat.bz2",
	}

	var blocks []*btcutil.Block
	var genBlk *wire.MsgBlock
	for _, file := range testFiles {
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Errorf("Error loading file: %v\n", err)
			return
		}
		blocks = append(blocks, blockTmp...)
	}

	for _, block := range blocks {
		if block.MsgBlock().BlockHash().String() == "00000000ebe5ec3e94d8dfe18100e5c0f3b1955bc6107fbe24d95732b814551b" {
			block.MsgBlock().ClearTransactions()
			genBlk = block.MsgBlock()
			break
		}
	}

	fmt.Println(genBlk)

	chain, err := GetChainV2("haveblock",
		&chaincfg.MainNetParams, 0)
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	// defer teardownFunc()

	// Since we're not dealing with the real block chain, set the coinbase
	// maturity to 1.
	chain.TstSetCoinbaseMaturity(1)

	for i := 1; i < len(blocks); i++ {
		blocks[i].MsgBlock().ClearTransactions()
		_, isOrphan, err := chain.ProcessBlockV2(blocks[i], BFNone)

		if err != nil {
			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
			return
		}
		if isOrphan {
			t.Errorf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
			return
		}
		fmt.Println("===== FINISHED PROCESS BLOCK ========")
	}

	// Insert an orphan block.
	_, isOrphan, err := chain.ProcessBlockV2(btcutil.NewBlock(&Block100000),
		BFNone)
	if err != nil {
		t.Errorf("Unable to process block: %v", err)
		return
	}
	if !isOrphan {
		t.Errorf("ProcessBlock indicated block is an not orphan when " +
			"it should be\n")
		return
	}

	tests := []struct {
		hash string
		want bool
	}{
		// Genesis block should be present (in the main chain).
		{hash: chaincfg.MainNetParams.GenesisHash.String(), want: true},

		// genesis block
		{hash: "00000000ebe5ec3e94d8dfe18100e5c0f3b1955bc6107fbe24d95732b814551b", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000952ccb1bf9b799fcd0cc654dd48363f76781f8b1c61dbf1696c39f97", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000bc3589303953766cc9364130cb97bc3749bae170f476d45f1e23f850", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000002f264d6504013e73b9c913de9098d4d771c1bb219af475d2a01b128e", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000474284d20067a4d33f6a02284e6ef70764a3a26d6a5b9df52ef663dd", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000551dc04c148242d1f648802577df8cf7d4e1b469211016280204a2bf", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000195f85184e77c18914bd0febd11278d950f5e4731a38f71ed79f044e", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506", want: true},
	}

	for i, test := range tests {
		hash, err := chainhash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewHashFromStr: %v", err)
			continue
		}

		result, err := chain.HaveBlock(hash)
		if err != nil {
			t.Errorf("HaveBlock #%d unexpected error: %v", i, err)
			return
		}
		if result != test.want {
			t.Errorf("HaveBlock #%d got %v want %v", i, result,
				test.want)
			continue
		}

		fmt.Println("=== check mainchain: ", test.hash, chain.MainChainHasBlock(hash))

		blk, err := chain.BlockByHash(hash)
		if err != nil {
			fmt.Println("Get block by hash error: ", err)
			continue
		}
		fmt.Println("===== Prev block hash : ", blk.MsgBlock().Header.PrevBlock.String())
		fmt.Println("===== Merkle root : ", blk.MsgBlock().Header.MerkleRoot.String())
		fmt.Println("===== Block hash : ", blk.MsgBlock().BlockHash())
		fmt.Println("===== Bits : ", blk.MsgBlock().Header.Bits)
		fmt.Println("===== Nonce : ", blk.MsgBlock().Header.Nonce)
	}

	fmt.Println("Finished test have block!!!!", chain.BestSnapshot().Hash, chain.BestSnapshot().Height)

	var initialized, hasBlockIndex bool
	chain.db.View(func(dbTx database.Tx) error {
		fmt.Println("chainStateKeyName: ", string(chainStateKeyName))
		fmt.Println("blockIndexBucketName: ", string(blockIndexBucketName))
		initialized = dbTx.Metadata().Get(chainStateKeyName) != nil
		hasBlockIndex = dbTx.Metadata().Bucket(blockIndexBucketName) != nil
		return nil
	})
	fmt.Println("initialized, hasBlockIndex: ", initialized, hasBlockIndex)

	chain.db.Close() // must have this statement because db is obtaining write lock

	// newChain, err := GetChain("haveblock",
	// 	&chaincfg.MainNetParams)
	// if err != nil {
	// 	t.Errorf("Failed to get chain instance: %v", err)
	// 	return
	// }
	// fmt.Println("info on new chain: ", newChain.BestSnapshot().Hash, newChain.BestSnapshot().Height)
	// newChain.db.Close()
}

// func TestReorganizeChain(t *testing.T) {
// 	// Load up blocks such that there is a side chain.
// 	// (genesis block) -> 1 -> 2 -> 3 -> 4
// 	//                          \-> 3a -> 4a -> 5a
// 	testFiles := []string{
// 		"blk_0_to_4.dat.bz2",
// 		"blk_3A.dat.bz2",
// 		"blk_4A.dat.bz2",
// 		"blk_5A.dat.bz2",
// 	}

// 	var blocks []*btcutil.Block
// 	for _, file := range testFiles {
// 		blockTmp, err := loadBlocks(file)
// 		if err != nil {
// 			t.Errorf("Error loading file: %v\n", err)
// 			return
// 		}
// 		blocks = append(blocks, blockTmp...)
// 	}

// 	// Create a new database and chain instance to run tests against.
// 	chain, _, err := chainSetup("haveblock",
// 		&chaincfg.MainNetParams)
// 	if err != nil {
// 		t.Errorf("Failed to setup chain instance: %v", err)
// 		return
// 	}
// 	// defer teardownFunc()

// 	// Since we're not dealing with the real block chain, set the coinbase
// 	// maturity to 1.
// 	chain.TstSetCoinbaseMaturity(1)

// 	for i := 1; i < len(blocks); i++ {
// 		blocks[i].MsgBlock().ClearTransactions()
// 		_, isOrphan, err := chain.ProcessBlockV2(blocks[i], BFNone)

// 		if err != nil {
// 			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
// 			return
// 		}
// 		if isOrphan {
// 			t.Errorf("ProcessBlock incorrectly returned block %v "+
// 				"is an orphan\n", i)
// 			return
// 		}
// 		fmt.Println("===== FINISHED PROCESS BLOCK ========")
// 	}

// 	// Insert an orphan block.
// 	_, isOrphan, err := chain.ProcessBlockV2(btcutil.NewBlock(&Block100000),
// 		BFNone)
// 	if err != nil {
// 		t.Errorf("Unable to process block: %v", err)
// 		return
// 	}
// 	if !isOrphan {
// 		t.Errorf("ProcessBlock indicated block is an not orphan when " +
// 			"it should be\n")
// 		return
// 	}

// 	tests := []struct {
// 		hash string
// 		want bool
// 	}{
// 		// Genesis block should be present (in the main chain).
// 		{hash: chaincfg.MainNetParams.GenesisHash.String(), want: true},

// 		// Block 3a should be present (on a side chain).
// 		{hash: "00000000ebe5ec3e94d8dfe18100e5c0f3b1955bc6107fbe24d95732b814551b", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "00000000952ccb1bf9b799fcd0cc654dd48363f76781f8b1c61dbf1696c39f97", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "00000000bc3589303953766cc9364130cb97bc3749bae170f476d45f1e23f850", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "000000002f264d6504013e73b9c913de9098d4d771c1bb219af475d2a01b128e", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "00000000474284d20067a4d33f6a02284e6ef70764a3a26d6a5b9df52ef663dd", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "00000000551dc04c148242d1f648802577df8cf7d4e1b469211016280204a2bf", want: true},

// 		// Random hashes should not be available.
// 		{hash: "123", want: false},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "00000000195f85184e77c18914bd0febd11278d950f5e4731a38f71ed79f044e", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506", want: true},
// 	}

// 	for i, test := range tests {
// 		hash, err := chainhash.NewHashFromStr(test.hash)
// 		if err != nil {
// 			t.Errorf("NewHashFromStr: %v", err)
// 			continue
// 		}

// 		result, err := chain.HaveBlock(hash)
// 		if err != nil {
// 			t.Errorf("HaveBlock #%d unexpected error: %v", i, err)
// 			return
// 		}
// 		if result != test.want {
// 			t.Errorf("HaveBlock #%d got %v want %v", i, result,
// 				test.want)
// 			continue
// 		}

// 		fmt.Println("=== check mainchain: ", test.hash, chain.MainChainHasBlock(hash))

// 		blk, err := chain.BlockByHash(hash)
// 		if err != nil {
// 			fmt.Println("Get block by hash error: ", err)
// 			continue
// 		}
// 		fmt.Println("===== Prev block hash : ", blk.MsgBlock().Header.PrevBlock.String())
// 		fmt.Println("===== Merkle root : ", blk.MsgBlock().Header.MerkleRoot.String())
// 		fmt.Println("===== Block hash : ", blk.MsgBlock().BlockHash())
// 		fmt.Println("===== Bits : ", blk.MsgBlock().Header.Bits)
// 		fmt.Println("===== Nonce : ", blk.MsgBlock().Header.Nonce)
// 	}

// 	fmt.Println("Finished test have block!!!!", chain.BestSnapshot().Hash, chain.BestSnapshot().Height)

// 	var initialized, hasBlockIndex bool
// 	chain.db.View(func(dbTx database.Tx) error {
// 		fmt.Println("chainStateKeyName: ", string(chainStateKeyName))
// 		fmt.Println("blockIndexBucketName: ", string(blockIndexBucketName))
// 		initialized = dbTx.Metadata().Get(chainStateKeyName) != nil
// 		hasBlockIndex = dbTx.Metadata().Bucket(blockIndexBucketName) != nil
// 		return nil
// 	})
// 	fmt.Println("initialized, hasBlockIndex: ", initialized, hasBlockIndex)

// 	chain.db.Close() // must have this statement because db is obtaining write lock

// 	// newChain, err := GetChain("haveblock",
// 	// 	&chaincfg.MainNetParams)
// 	// if err != nil {
// 	// 	t.Errorf("Failed to get chain instance: %v", err)
// 	// 	return
// 	// }
// 	// fmt.Println("info on new chain: ", newChain.BestSnapshot().Hash, newChain.BestSnapshot().Height)
// 	// newChain.db.Close()
// }

func TestGetChainInfo(t *testing.T) {
	testFiles := []string{
		"blk_0_to_4.dat.bz2",
		"blk_3A.dat.bz2",
		"blk_4A.dat.bz2",
		"blk_5A.dat.bz2",
	}

	var blocks []*btcutil.Block
	var genBlk *wire.MsgBlock
	for _, file := range testFiles {
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Errorf("Error loading file: %v\n", err)
			return
		}
		blocks = append(blocks, blockTmp...)
	}

	for _, block := range blocks {
		if block.MsgBlock().BlockHash().String() == "00000000ebe5ec3e94d8dfe18100e5c0f3b1955bc6107fbe24d95732b814551b" {
			block.MsgBlock().ClearTransactions()
			genBlk = block.MsgBlock()
			break
		}
	}
	fmt.Println(genBlk)

	chain, err := GetChainV2("haveblock",
		&chaincfg.MainNetParams, 0)
	if err != nil {
		t.Errorf("Failed to get chain instance: %v", err)
		return
	}
	fmt.Println("info on new new chain: ", chain.BestSnapshot().Hash, chain.BestSnapshot().Height)

	tests := []struct {
		hash string
		want bool
	}{
		// Genesis block should be present (in the main chain).
		{hash: chaincfg.MainNetParams.GenesisHash.String(), want: true},

		// genesis block
		{hash: "00000000ebe5ec3e94d8dfe18100e5c0f3b1955bc6107fbe24d95732b814551b", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000952ccb1bf9b799fcd0cc654dd48363f76781f8b1c61dbf1696c39f97", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000bc3589303953766cc9364130cb97bc3749bae170f476d45f1e23f850", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000002f264d6504013e73b9c913de9098d4d771c1bb219af475d2a01b128e", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000474284d20067a4d33f6a02284e6ef70764a3a26d6a5b9df52ef663dd", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000551dc04c148242d1f648802577df8cf7d4e1b469211016280204a2bf", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},

		// Block 100000 should be present (as an orphan).
		{hash: "00000000195f85184e77c18914bd0febd11278d950f5e4731a38f71ed79f044e", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506", want: false},
	}

	for i, test := range tests {
		hash, err := chainhash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewHashFromStr: %v", err)
			continue
		}

		result, err := chain.HaveBlock(hash)
		if err != nil {
			t.Errorf("HaveBlock #%d unexpected error: %v", i, err)
			return
		}
		if result != test.want {
			t.Errorf("HaveBlock #%d got %v want %v", i, result,
				test.want)
			continue
		}

		fmt.Println("=== check mainchain: ", test.hash, chain.MainChainHasBlock(hash))
	}
}

// // TestHaveBlock tests the HaveBlock API to ensure proper functionality.
// func TestHaveBlock(t *testing.T) {
// 	// Load up blocks such that there is a side chain.
// 	// (genesis block) -> 1 -> 2 -> 3 -> 4
// 	//                          \-> 3a
// 	testFiles := []string{
// 		"blk_0_to_4.dat.bz2",
// 		"blk_3A.dat.bz2",
// 	}

// 	var blocks []*btcutil.Block
// 	for _, file := range testFiles {
// 		blockTmp, err := loadBlocks(file)
// 		if err != nil {
// 			t.Errorf("Error loading file: %v\n", err)
// 			return
// 		}
// 		blocks = append(blocks, blockTmp...)
// 	}

// 	// Create a new database and chain instance to run tests against.
// 	chain, teardownFunc, err := chainSetup("haveblock",
// 		&chaincfg.MainNetParams)
// 	if err != nil {
// 		t.Errorf("Failed to setup chain instance: %v", err)
// 		return
// 	}
// 	defer teardownFunc()

// 	// Since we're not dealing with the real block chain, set the coinbase
// 	// maturity to 1.
// 	chain.TstSetCoinbaseMaturity(1)

// 	for i := 1; i < len(blocks); i++ {
// 		blocks[i].MsgBlock().ClearTransactions()
// 		isMainChain, isOrphan, err := chain.ProcessBlock(blocks[i], BFNone)

// 		// bb, _ := json.Marshal(blocks[i].MsgBlock())
// 		// fmt.Println("===== Block contain: ", string(bb))
// 		fmt.Println("===== isMainChain, isOrphan: ", isMainChain, isOrphan)
// 		fmt.Println("===== Prev block hash : ", blocks[i].MsgBlock().Header.PrevBlock.String())
// 		fmt.Println("===== Block hash : ", blocks[i].MsgBlock().BlockHash())
// 		fmt.Println("===== Bits : ", blocks[i].MsgBlock().Header.Bits)
// 		fmt.Println("===== FINISHED PROCESS BLOCK ========")

// 		if err != nil {
// 			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
// 			return
// 		}
// 		if isOrphan {
// 			t.Errorf("ProcessBlock incorrectly returned block %v "+
// 				"is an orphan\n", i)
// 			return
// 		}
// 	}

// 	// Insert an orphan block.
// 	_, isOrphan, err := chain.ProcessBlock(btcutil.NewBlock(&Block100000),
// 		BFNone)
// 	if err != nil {
// 		t.Errorf("Unable to process block: %v", err)
// 		return
// 	}
// 	if !isOrphan {
// 		t.Errorf("ProcessBlock indicated block is an not orphan when " +
// 			"it should be\n")
// 		return
// 	}

// 	tests := []struct {
// 		hash string
// 		want bool
// 	}{
// 		// Genesis block should be present (in the main chain).
// 		{hash: chaincfg.MainNetParams.GenesisHash.String(), want: true},

// 		// Block 3a should be present (on a side chain).
// 		{hash: "00000000474284d20067a4d33f6a02284e6ef70764a3a26d6a5b9df52ef663dd", want: true},

// 		// Block 100000 should be present (as an orphan).
// 		{hash: "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506", want: true},

// 		// Random hashes should not be available.
// 		{hash: "123", want: false},
// 	}

// 	for i, test := range tests {
// 		hash, err := chainhash.NewHashFromStr(test.hash)
// 		if err != nil {
// 			t.Errorf("NewHashFromStr: %v", err)
// 			continue
// 		}

// 		result, err := chain.HaveBlock(hash)
// 		if err != nil {
// 			t.Errorf("HaveBlock #%d unexpected error: %v", i, err)
// 			return
// 		}
// 		if result != test.want {
// 			t.Errorf("HaveBlock #%d got %v want %v", i, result,
// 				test.want)
// 			continue
// 		}
// 	}
// 	fmt.Println("Finished test have block!!!!")
// }

// // TestCalcSequenceLock tests the LockTimeToSequence function, and the
// // CalcSequenceLock method of a Chain instance. The tests exercise several
// // combinations of inputs to the CalcSequenceLock function in order to ensure
// // the returned SequenceLocks are correct for each test instance.
// func TestCalcSequenceLock(t *testing.T) {
// 	netParams := &chaincfg.SimNetParams

// 	// We need to activate CSV in order to test the processing logic, so
// 	// manually craft the block version that's used to signal the soft-fork
// 	// activation.
// 	csvBit := netParams.Deployments[chaincfg.DeploymentCSV].BitNumber
// 	blockVersion := int32(0x20000000 | (uint32(1) << csvBit))

// 	// Generate enough synthetic blocks to activate CSV.
// 	chain := newFakeChain(netParams)
// 	node := chain.bestChain.Tip()
// 	blockTime := node.Header().Timestamp
// 	numBlocksToActivate := (netParams.MinerConfirmationWindow * 3)
// 	for i := uint32(0); i < numBlocksToActivate; i++ {
// 		blockTime = blockTime.Add(time.Second)
// 		node = newFakeNode(node, blockVersion, 0, blockTime)
// 		chain.index.AddNode(node)
// 		chain.bestChain.SetTip(node)
// 	}

// 	// Create a utxo view with a fake utxo for the inputs used in the
// 	// transactions created below.  This utxo is added such that it has an
// 	// age of 4 blocks.
// 	targetTx := btcutil.NewTx(&wire.MsgTx{
// 		TxOut: []*wire.TxOut{{
// 			PkScript: nil,
// 			Value:    10,
// 		}},
// 	})
// 	utxoView := NewUtxoViewpoint()
// 	utxoView.AddTxOuts(targetTx, int32(numBlocksToActivate)-4)
// 	utxoView.SetBestHash(&node.hash)

// 	// Create a utxo that spends the fake utxo created above for use in the
// 	// transactions created in the tests.  It has an age of 4 blocks.  Note
// 	// that the sequence lock heights are always calculated from the same
// 	// point of view that they were originally calculated from for a given
// 	// utxo.  That is to say, the height prior to it.
// 	utxo := wire.OutPoint{
// 		Hash:  *targetTx.Hash(),
// 		Index: 0,
// 	}
// 	prevUtxoHeight := int32(numBlocksToActivate) - 4

// 	// Obtain the median time past from the PoV of the input created above.
// 	// The MTP for the input is the MTP from the PoV of the block *prior*
// 	// to the one that included it.
// 	medianTime := node.RelativeAncestor(5).CalcPastMedianTime().Unix()

// 	// The median time calculated from the PoV of the best block in the
// 	// test chain.  For unconfirmed inputs, this value will be used since
// 	// the MTP will be calculated from the PoV of the yet-to-be-mined
// 	// block.
// 	nextMedianTime := node.CalcPastMedianTime().Unix()
// 	nextBlockHeight := int32(numBlocksToActivate) + 1

// 	// Add an additional transaction which will serve as our unconfirmed
// 	// output.
// 	unConfTx := &wire.MsgTx{
// 		TxOut: []*wire.TxOut{{
// 			PkScript: nil,
// 			Value:    5,
// 		}},
// 	}
// 	unConfUtxo := wire.OutPoint{
// 		Hash:  unConfTx.TxHash(),
// 		Index: 0,
// 	}

// 	// Adding a utxo with a height of 0x7fffffff indicates that the output
// 	// is currently unmined.
// 	utxoView.AddTxOuts(btcutil.NewTx(unConfTx), 0x7fffffff)

// 	tests := []struct {
// 		tx      *wire.MsgTx
// 		view    *UtxoViewpoint
// 		mempool bool
// 		want    *SequenceLock
// 	}{
// 		// A transaction of version one should disable sequence locks
// 		// as the new sequence number semantics only apply to
// 		// transactions version 2 or higher.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 1,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 3),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     -1,
// 				BlockHeight: -1,
// 			},
// 		},
// 		// A transaction with a single input with max sequence number.
// 		// This sequence number has the high bit set, so sequence locks
// 		// should be disabled.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         wire.MaxTxInSequenceNum,
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     -1,
// 				BlockHeight: -1,
// 			},
// 		},
// 		// A transaction with a single input whose lock time is
// 		// expressed in seconds.  However, the specified lock time is
// 		// below the required floor for time based lock times since
// 		// they have time granularity of 512 seconds.  As a result, the
// 		// seconds lock-time should be just before the median time of
// 		// the targeted block.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 2),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     medianTime - 1,
// 				BlockHeight: -1,
// 			},
// 		},
// 		// A transaction with a single input whose lock time is
// 		// expressed in seconds.  The number of seconds should be 1023
// 		// seconds after the median past time of the last block in the
// 		// chain.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 1024),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     medianTime + 1023,
// 				BlockHeight: -1,
// 			},
// 		},
// 		// A transaction with multiple inputs.  The first input has a
// 		// lock time expressed in seconds.  The second input has a
// 		// sequence lock in blocks with a value of 4.  The last input
// 		// has a sequence number with a value of 5, but has the disable
// 		// bit set.  So the first lock should be selected as it's the
// 		// latest lock that isn't disabled.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 2560),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 4),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence: LockTimeToSequence(false, 5) |
// 						wire.SequenceLockTimeDisabled,
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     medianTime + (5 << wire.SequenceLockTimeGranularity) - 1,
// 				BlockHeight: prevUtxoHeight + 3,
// 			},
// 		},
// 		// Transaction with a single input.  The input's sequence number
// 		// encodes a relative lock-time in blocks (3 blocks).  The
// 		// sequence lock should  have a value of -1 for seconds, but a
// 		// height of 2 meaning it can be included at height 3.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 3),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     -1,
// 				BlockHeight: prevUtxoHeight + 2,
// 			},
// 		},
// 		// A transaction with two inputs with lock times expressed in
// 		// seconds.  The selected sequence lock value for seconds should
// 		// be the time further in the future.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 5120),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 2560),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     medianTime + (10 << wire.SequenceLockTimeGranularity) - 1,
// 				BlockHeight: -1,
// 			},
// 		},
// 		// A transaction with two inputs with lock times expressed in
// 		// blocks.  The selected sequence lock value for blocks should
// 		// be the height further in the future, so a height of 10
// 		// indicating it can be included at height 11.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 1),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 11),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     -1,
// 				BlockHeight: prevUtxoHeight + 10,
// 			},
// 		},
// 		// A transaction with multiple inputs.  Two inputs are time
// 		// based, and the other two are block based. The lock lying
// 		// further into the future for both inputs should be chosen.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 2560),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(true, 6656),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 3),
// 				}, {
// 					PreviousOutPoint: utxo,
// 					Sequence:         LockTimeToSequence(false, 9),
// 				}},
// 			},
// 			view: utxoView,
// 			want: &SequenceLock{
// 				Seconds:     medianTime + (13 << wire.SequenceLockTimeGranularity) - 1,
// 				BlockHeight: prevUtxoHeight + 8,
// 			},
// 		},
// 		// A transaction with a single unconfirmed input.  As the input
// 		// is confirmed, the height of the input should be interpreted
// 		// as the height of the *next* block.  So, a 2 block relative
// 		// lock means the sequence lock should be for 1 block after the
// 		// *next* block height, indicating it can be included 2 blocks
// 		// after that.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: unConfUtxo,
// 					Sequence:         LockTimeToSequence(false, 2),
// 				}},
// 			},
// 			view:    utxoView,
// 			mempool: true,
// 			want: &SequenceLock{
// 				Seconds:     -1,
// 				BlockHeight: nextBlockHeight + 1,
// 			},
// 		},
// 		// A transaction with a single unconfirmed input.  The input has
// 		// a time based lock, so the lock time should be based off the
// 		// MTP of the *next* block.
// 		{
// 			tx: &wire.MsgTx{
// 				Version: 2,
// 				TxIn: []*wire.TxIn{{
// 					PreviousOutPoint: unConfUtxo,
// 					Sequence:         LockTimeToSequence(true, 1024),
// 				}},
// 			},
// 			view:    utxoView,
// 			mempool: true,
// 			want: &SequenceLock{
// 				Seconds:     nextMedianTime + 1023,
// 				BlockHeight: -1,
// 			},
// 		},
// 	}

// 	t.Logf("Running %v SequenceLock tests", len(tests))
// 	for i, test := range tests {
// 		utilTx := btcutil.NewTx(test.tx)
// 		seqLock, err := chain.CalcSequenceLock(utilTx, test.view, test.mempool)
// 		if err != nil {
// 			t.Fatalf("test #%d, unable to calc sequence lock: %v", i, err)
// 		}

// 		if seqLock.Seconds != test.want.Seconds {
// 			t.Fatalf("test #%d got %v seconds want %v seconds",
// 				i, seqLock.Seconds, test.want.Seconds)
// 		}
// 		if seqLock.BlockHeight != test.want.BlockHeight {
// 			t.Fatalf("test #%d got height of %v want height of %v ",
// 				i, seqLock.BlockHeight, test.want.BlockHeight)
// 		}
// 	}
// }

// // nodeHashes is a convenience function that returns the hashes for all of the
// // passed indexes of the provided nodes.  It is used to construct expected hash
// // slices in the tests.
// func nodeHashes(nodes []*blockNode, indexes ...int) []chainhash.Hash {
// 	hashes := make([]chainhash.Hash, 0, len(indexes))
// 	for _, idx := range indexes {
// 		hashes = append(hashes, nodes[idx].hash)
// 	}
// 	return hashes
// }

// // nodeHeaders is a convenience function that returns the headers for all of
// // the passed indexes of the provided nodes.  It is used to construct expected
// // located headers in the tests.
// func nodeHeaders(nodes []*blockNode, indexes ...int) []wire.BlockHeader {
// 	headers := make([]wire.BlockHeader, 0, len(indexes))
// 	for _, idx := range indexes {
// 		headers = append(headers, nodes[idx].Header())
// 	}
// 	return headers
// }

// // TestLocateInventory ensures that locating inventory via the LocateHeaders and
// // LocateBlocks functions behaves as expected.
// func TestLocateInventory(t *testing.T) {
// 	// Construct a synthetic block chain with a block index consisting of
// 	// the following structure.
// 	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
// 	// 	                              \-> 16a -> 17a
// 	tip := tstTip
// 	chain := newFakeChain(&chaincfg.MainNetParams)
// 	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 18)
// 	branch1Nodes := chainedNodes(branch0Nodes[14], 2)
// 	for _, node := range branch0Nodes {
// 		chain.index.AddNode(node)
// 	}
// 	for _, node := range branch1Nodes {
// 		chain.index.AddNode(node)
// 	}
// 	chain.bestChain.SetTip(tip(branch0Nodes))

// 	// Create chain views for different branches of the overall chain to
// 	// simulate a local and remote node on different parts of the chain.
// 	localView := newChainView(tip(branch0Nodes))
// 	remoteView := newChainView(tip(branch1Nodes))

// 	// Create a chain view for a completely unrelated block chain to
// 	// simulate a remote node on a totally different chain.
// 	unrelatedBranchNodes := chainedNodes(nil, 5)
// 	unrelatedView := newChainView(tip(unrelatedBranchNodes))

// 	tests := []struct {
// 		name       string
// 		locator    BlockLocator       // locator for requested inventory
// 		hashStop   chainhash.Hash     // stop hash for locator
// 		maxAllowed uint32             // max to locate, 0 = wire const
// 		headers    []wire.BlockHeader // expected located headers
// 		hashes     []chainhash.Hash   // expected located hashes
// 	}{
// 		{
// 			// Empty block locators and unknown stop hash.  No
// 			// inventory should be located.
// 			name:     "no locators, no stop",
// 			locator:  nil,
// 			hashStop: chainhash.Hash{},
// 			headers:  nil,
// 			hashes:   nil,
// 		},
// 		{
// 			// Empty block locators and stop hash in side chain.
// 			// The expected result is the requested block.
// 			name:     "no locators, stop in side",
// 			locator:  nil,
// 			hashStop: tip(branch1Nodes).hash,
// 			headers:  nodeHeaders(branch1Nodes, 1),
// 			hashes:   nodeHashes(branch1Nodes, 1),
// 		},
// 		{
// 			// Empty block locators and stop hash in main chain.
// 			// The expected result is the requested block.
// 			name:     "no locators, stop in main",
// 			locator:  nil,
// 			hashStop: branch0Nodes[12].hash,
// 			headers:  nodeHeaders(branch0Nodes, 12),
// 			hashes:   nodeHashes(branch0Nodes, 12),
// 		},
// 		{
// 			// Locators based on remote being on side chain and a
// 			// stop hash local node doesn't know about.  The
// 			// expected result is the blocks after the fork point in
// 			// the main chain and the stop hash has no effect.
// 			name:     "remote side chain, unknown stop",
// 			locator:  remoteView.BlockLocator(nil),
// 			hashStop: chainhash.Hash{0x01},
// 			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on side chain and a
// 			// stop hash in side chain.  The expected result is the
// 			// blocks after the fork point in the main chain and the
// 			// stop hash has no effect.
// 			name:     "remote side chain, stop in side",
// 			locator:  remoteView.BlockLocator(nil),
// 			hashStop: tip(branch1Nodes).hash,
// 			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on side chain and a
// 			// stop hash in main chain, but before fork point.  The
// 			// expected result is the blocks after the fork point in
// 			// the main chain and the stop hash has no effect.
// 			name:     "remote side chain, stop in main before",
// 			locator:  remoteView.BlockLocator(nil),
// 			hashStop: branch0Nodes[13].hash,
// 			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on side chain and a
// 			// stop hash in main chain, but exactly at the fork
// 			// point.  The expected result is the blocks after the
// 			// fork point in the main chain and the stop hash has no
// 			// effect.
// 			name:     "remote side chain, stop in main exact",
// 			locator:  remoteView.BlockLocator(nil),
// 			hashStop: branch0Nodes[14].hash,
// 			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on side chain and a
// 			// stop hash in main chain just after the fork point.
// 			// The expected result is the blocks after the fork
// 			// point in the main chain up to and including the stop
// 			// hash.
// 			name:     "remote side chain, stop in main after",
// 			locator:  remoteView.BlockLocator(nil),
// 			hashStop: branch0Nodes[15].hash,
// 			headers:  nodeHeaders(branch0Nodes, 15),
// 			hashes:   nodeHashes(branch0Nodes, 15),
// 		},
// 		{
// 			// Locators based on remote being on side chain and a
// 			// stop hash in main chain some time after the fork
// 			// point.  The expected result is the blocks after the
// 			// fork point in the main chain up to and including the
// 			// stop hash.
// 			name:     "remote side chain, stop in main after more",
// 			locator:  remoteView.BlockLocator(nil),
// 			hashStop: branch0Nodes[16].hash,
// 			headers:  nodeHeaders(branch0Nodes, 15, 16),
// 			hashes:   nodeHashes(branch0Nodes, 15, 16),
// 		},
// 		{
// 			// Locators based on remote being on main chain in the
// 			// past and a stop hash local node doesn't know about.
// 			// The expected result is the blocks after the known
// 			// point in the main chain and the stop hash has no
// 			// effect.
// 			name:     "remote main chain past, unknown stop",
// 			locator:  localView.BlockLocator(branch0Nodes[12]),
// 			hashStop: chainhash.Hash{0x01},
// 			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on main chain in the
// 			// past and a stop hash in a side chain.  The expected
// 			// result is the blocks after the known point in the
// 			// main chain and the stop hash has no effect.
// 			name:     "remote main chain past, stop in side",
// 			locator:  localView.BlockLocator(branch0Nodes[12]),
// 			hashStop: tip(branch1Nodes).hash,
// 			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on main chain in the
// 			// past and a stop hash in the main chain before that
// 			// point.  The expected result is the blocks after the
// 			// known point in the main chain and the stop hash has
// 			// no effect.
// 			name:     "remote main chain past, stop in main before",
// 			locator:  localView.BlockLocator(branch0Nodes[12]),
// 			hashStop: branch0Nodes[11].hash,
// 			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on main chain in the
// 			// past and a stop hash in the main chain exactly at that
// 			// point.  The expected result is the blocks after the
// 			// known point in the main chain and the stop hash has
// 			// no effect.
// 			name:     "remote main chain past, stop in main exact",
// 			locator:  localView.BlockLocator(branch0Nodes[12]),
// 			hashStop: branch0Nodes[12].hash,
// 			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
// 			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Locators based on remote being on main chain in the
// 			// past and a stop hash in the main chain just after
// 			// that point.  The expected result is the blocks after
// 			// the known point in the main chain and the stop hash
// 			// has no effect.
// 			name:     "remote main chain past, stop in main after",
// 			locator:  localView.BlockLocator(branch0Nodes[12]),
// 			hashStop: branch0Nodes[13].hash,
// 			headers:  nodeHeaders(branch0Nodes, 13),
// 			hashes:   nodeHashes(branch0Nodes, 13),
// 		},
// 		{
// 			// Locators based on remote being on main chain in the
// 			// past and a stop hash in the main chain some time
// 			// after that point.  The expected result is the blocks
// 			// after the known point in the main chain and the stop
// 			// hash has no effect.
// 			name:     "remote main chain past, stop in main after more",
// 			locator:  localView.BlockLocator(branch0Nodes[12]),
// 			hashStop: branch0Nodes[15].hash,
// 			headers:  nodeHeaders(branch0Nodes, 13, 14, 15),
// 			hashes:   nodeHashes(branch0Nodes, 13, 14, 15),
// 		},
// 		{
// 			// Locators based on remote being at exactly the same
// 			// point in the main chain and a stop hash local node
// 			// doesn't know about.  The expected result is no
// 			// located inventory.
// 			name:     "remote main chain same, unknown stop",
// 			locator:  localView.BlockLocator(nil),
// 			hashStop: chainhash.Hash{0x01},
// 			headers:  nil,
// 			hashes:   nil,
// 		},
// 		{
// 			// Locators based on remote being at exactly the same
// 			// point in the main chain and a stop hash at exactly
// 			// the same point.  The expected result is no located
// 			// inventory.
// 			name:     "remote main chain same, stop same point",
// 			locator:  localView.BlockLocator(nil),
// 			hashStop: tip(branch0Nodes).hash,
// 			headers:  nil,
// 			hashes:   nil,
// 		},
// 		{
// 			// Locators from remote that don't include any blocks
// 			// the local node knows.  This would happen if the
// 			// remote node is on a completely separate chain that
// 			// isn't rooted with the same genesis block.  The
// 			// expected result is the blocks after the genesis
// 			// block.
// 			name:     "remote unrelated chain",
// 			locator:  unrelatedView.BlockLocator(nil),
// 			hashStop: chainhash.Hash{},
// 			headers: nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
// 				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
// 			hashes: nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
// 				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Locators from remote for second block in main chain
// 			// and no stop hash, but with an overridden max limit.
// 			// The expected result is the blocks after the second
// 			// block limited by the max.
// 			name:       "remote genesis",
// 			locator:    locatorHashes(branch0Nodes, 0),
// 			hashStop:   chainhash.Hash{},
// 			maxAllowed: 3,
// 			headers:    nodeHeaders(branch0Nodes, 1, 2, 3),
// 			hashes:     nodeHashes(branch0Nodes, 1, 2, 3),
// 		},
// 		{
// 			// Poorly formed locator.
// 			//
// 			// Locator from remote that only includes a single
// 			// block on a side chain the local node knows.  The
// 			// expected result is the blocks after the genesis
// 			// block since even though the block is known, it is on
// 			// a side chain and there are no more locators to find
// 			// the fork point.
// 			name:     "weak locator, single known side block",
// 			locator:  locatorHashes(branch1Nodes, 1),
// 			hashStop: chainhash.Hash{},
// 			headers: nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
// 				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
// 			hashes: nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
// 				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Poorly formed locator.
// 			//
// 			// Locator from remote that only includes multiple
// 			// blocks on a side chain the local node knows however
// 			// none in the main chain.  The expected result is the
// 			// blocks after the genesis block since even though the
// 			// blocks are known, they are all on a side chain and
// 			// there are no more locators to find the fork point.
// 			name:     "weak locator, multiple known side blocks",
// 			locator:  locatorHashes(branch1Nodes, 1),
// 			hashStop: chainhash.Hash{},
// 			headers: nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
// 				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
// 			hashes: nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
// 				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
// 		},
// 		{
// 			// Poorly formed locator.
// 			//
// 			// Locator from remote that only includes multiple
// 			// blocks on a side chain the local node knows however
// 			// none in the main chain but includes a stop hash in
// 			// the main chain.  The expected result is the blocks
// 			// after the genesis block up to the stop hash since
// 			// even though the blocks are known, they are all on a
// 			// side chain and there are no more locators to find the
// 			// fork point.
// 			name:     "weak locator, multiple known side blocks, stop in main",
// 			locator:  locatorHashes(branch1Nodes, 1),
// 			hashStop: branch0Nodes[5].hash,
// 			headers:  nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5),
// 			hashes:   nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5),
// 		},
// 	}
// 	for _, test := range tests {
// 		// Ensure the expected headers are located.
// 		var headers []wire.BlockHeader
// 		if test.maxAllowed != 0 {
// 			// Need to use the unexported function to override the
// 			// max allowed for headers.
// 			chain.chainLock.RLock()
// 			headers = chain.locateHeaders(test.locator,
// 				&test.hashStop, test.maxAllowed)
// 			chain.chainLock.RUnlock()
// 		} else {
// 			headers = chain.LocateHeaders(test.locator,
// 				&test.hashStop)
// 		}
// 		if !reflect.DeepEqual(headers, test.headers) {
// 			t.Errorf("%s: unxpected headers -- got %v, want %v",
// 				test.name, headers, test.headers)
// 			continue
// 		}

// 		// Ensure the expected block hashes are located.
// 		maxAllowed := uint32(wire.MaxBlocksPerMsg)
// 		if test.maxAllowed != 0 {
// 			maxAllowed = test.maxAllowed
// 		}
// 		hashes := chain.LocateBlocks(test.locator, &test.hashStop,
// 			maxAllowed)
// 		if !reflect.DeepEqual(hashes, test.hashes) {
// 			t.Errorf("%s: unxpected hashes -- got %v, want %v",
// 				test.name, hashes, test.hashes)
// 			continue
// 		}
// 	}
// }

// // TestHeightToHashRange ensures that fetching a range of block hashes by start
// // height and end hash works as expected.
// func TestHeightToHashRange(t *testing.T) {
// 	// Construct a synthetic block chain with a block index consisting of
// 	// the following structure.
// 	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
// 	// 	                              \-> 16a -> 17a -> 18a (unvalidated)
// 	tip := tstTip
// 	chain := newFakeChain(&chaincfg.MainNetParams)
// 	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 18)
// 	branch1Nodes := chainedNodes(branch0Nodes[14], 3)
// 	for _, node := range branch0Nodes {
// 		chain.index.SetStatusFlags(node, statusValid)
// 		chain.index.AddNode(node)
// 	}
// 	for _, node := range branch1Nodes {
// 		if node.height < 18 {
// 			chain.index.SetStatusFlags(node, statusValid)
// 		}
// 		chain.index.AddNode(node)
// 	}
// 	chain.bestChain.SetTip(tip(branch0Nodes))

// 	tests := []struct {
// 		name        string
// 		startHeight int32            // locator for requested inventory
// 		endHash     chainhash.Hash   // stop hash for locator
// 		maxResults  int              // max to locate, 0 = wire const
// 		hashes      []chainhash.Hash // expected located hashes
// 		expectError bool
// 	}{
// 		{
// 			name:        "blocks below tip",
// 			startHeight: 11,
// 			endHash:     branch0Nodes[14].hash,
// 			maxResults:  10,
// 			hashes:      nodeHashes(branch0Nodes, 10, 11, 12, 13, 14),
// 		},
// 		{
// 			name:        "blocks on main chain",
// 			startHeight: 15,
// 			endHash:     branch0Nodes[17].hash,
// 			maxResults:  10,
// 			hashes:      nodeHashes(branch0Nodes, 14, 15, 16, 17),
// 		},
// 		{
// 			name:        "blocks on stale chain",
// 			startHeight: 15,
// 			endHash:     branch1Nodes[1].hash,
// 			maxResults:  10,
// 			hashes: append(nodeHashes(branch0Nodes, 14),
// 				nodeHashes(branch1Nodes, 0, 1)...),
// 		},
// 		{
// 			name:        "invalid start height",
// 			startHeight: 19,
// 			endHash:     branch0Nodes[17].hash,
// 			maxResults:  10,
// 			expectError: true,
// 		},
// 		{
// 			name:        "too many results",
// 			startHeight: 1,
// 			endHash:     branch0Nodes[17].hash,
// 			maxResults:  10,
// 			expectError: true,
// 		},
// 		{
// 			name:        "unvalidated block",
// 			startHeight: 15,
// 			endHash:     branch1Nodes[2].hash,
// 			maxResults:  10,
// 			expectError: true,
// 		},
// 	}
// 	for _, test := range tests {
// 		hashes, err := chain.HeightToHashRange(test.startHeight, &test.endHash,
// 			test.maxResults)
// 		if err != nil {
// 			if !test.expectError {
// 				t.Errorf("%s: unexpected error: %v", test.name, err)
// 			}
// 			continue
// 		}

// 		if !reflect.DeepEqual(hashes, test.hashes) {
// 			t.Errorf("%s: unxpected hashes -- got %v, want %v",
// 				test.name, hashes, test.hashes)
// 		}
// 	}
// }

// // TestIntervalBlockHashes ensures that fetching block hashes at specified
// // intervals by end hash works as expected.
// func TestIntervalBlockHashes(t *testing.T) {
// 	// Construct a synthetic block chain with a block index consisting of
// 	// the following structure.
// 	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
// 	// 	                              \-> 16a -> 17a -> 18a (unvalidated)
// 	tip := tstTip
// 	chain := newFakeChain(&chaincfg.MainNetParams)
// 	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 18)
// 	branch1Nodes := chainedNodes(branch0Nodes[14], 3)
// 	for _, node := range branch0Nodes {
// 		chain.index.SetStatusFlags(node, statusValid)
// 		chain.index.AddNode(node)
// 	}
// 	for _, node := range branch1Nodes {
// 		if node.height < 18 {
// 			chain.index.SetStatusFlags(node, statusValid)
// 		}
// 		chain.index.AddNode(node)
// 	}
// 	chain.bestChain.SetTip(tip(branch0Nodes))

// 	tests := []struct {
// 		name        string
// 		endHash     chainhash.Hash
// 		interval    int
// 		hashes      []chainhash.Hash
// 		expectError bool
// 	}{
// 		{
// 			name:     "blocks on main chain",
// 			endHash:  branch0Nodes[17].hash,
// 			interval: 8,
// 			hashes:   nodeHashes(branch0Nodes, 7, 15),
// 		},
// 		{
// 			name:     "blocks on stale chain",
// 			endHash:  branch1Nodes[1].hash,
// 			interval: 8,
// 			hashes: append(nodeHashes(branch0Nodes, 7),
// 				nodeHashes(branch1Nodes, 0)...),
// 		},
// 		{
// 			name:     "no results",
// 			endHash:  branch0Nodes[17].hash,
// 			interval: 20,
// 			hashes:   []chainhash.Hash{},
// 		},
// 		{
// 			name:        "unvalidated block",
// 			endHash:     branch1Nodes[2].hash,
// 			interval:    8,
// 			expectError: true,
// 		},
// 	}
// 	for _, test := range tests {
// 		hashes, err := chain.IntervalBlockHashes(&test.endHash, test.interval)
// 		if err != nil {
// 			if !test.expectError {
// 				t.Errorf("%s: unexpected error: %v", test.name, err)
// 			}
// 			continue
// 		}

// 		if !reflect.DeepEqual(hashes, test.hashes) {
// 			t.Errorf("%s: unxpected hashes -- got %v, want %v",
// 				test.name, hashes, test.hashes)
// 		}
// 	}
// }
