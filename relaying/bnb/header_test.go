package bnb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/types"
	"testing"
)

//
//import (
//	"bytes"
//	"crypto/rand"
//	"encoding/base64"
//	"encoding/hex"
//	"fmt"
//	"github.com/stretchr/testify/assert"
//	"github.com/tendermint/tendermint/crypto"
//	"github.com/tendermint/tendermint/crypto/ed25519"
//	"github.com/tendermint/tendermint/crypto/merkle"
//	"github.com/tendermint/tendermint/crypto/tmhash"
//	"github.com/tendermint/tendermint/types"
//	"math"
//	"testing"
//	"time"
//)
//
//func makeBlockIDRandom() types.BlockID {
//	var (
//		blockHash   = make([]byte, tmhash.Size)
//		partSetHash = make([]byte, tmhash.Size)
//	)
//	rand.Read(blockHash)   //nolint: gosec
//	rand.Read(partSetHash) //nolint: gosec
//	return types.BlockID{blockHash, types.PartSetHeader{123, partSetHash}}
//}
//
//// NOTE: privValidators are in order
//func randVoteSet(
//	height int64,
//	round int,
//	signedMsgType types.SignedMsgType,
//	numValidators int,
//	votingPower int64,
//) (*types.VoteSet, *types.ValidatorSet, []types.PrivValidator) {
//	valSet, privValidators := types.RandValidatorSet(numValidators, votingPower)
//	return types.NewVoteSet("test_chain_id", height, round, signedMsgType, valSet), valSet, privValidators
//}
//
//func randCommit() *types.Commit {
//	lastID := makeBlockIDRandom()
//	h := int64(3)
//	voteSet, _, vals := randVoteSet(h-1, 1, types.PrecommitType, 10, 1)
//	commit, err := types.MakeCommit(lastID, h-1, 1, voteSet, vals)
//	if err != nil {
//		panic(err)
//	}
//	return commit
//}
//
//func TestBNBSignedHeaderValidateBasic(t *testing.T) {
//	commit := randCommit()
//	chainID := "𠜎"
//	timestamp := time.Date(math.MaxInt64, 0, 0, 0, 0, 0, math.MaxInt64, time.UTC)
//	h := BNBHeader{
//		Version:            Version{Block: math.MaxInt64, App: math.MaxInt64},
//		ChainID:            chainID,
//		Height:             commit.Height,
//		Time:               timestamp,
//		LastBlockID:        commit.BlockID,
//		LastCommitHash:     commit.Hash(),
//		DataHash:           commit.Hash(),
//		ValidatorsHash:     commit.Hash(),
//		NextValidatorsHash: commit.Hash(),
//		ConsensusHash:      commit.Hash(),
//		AppHash:            commit.Hash(),
//		LastResultsHash:    commit.Hash(),
//		EvidenceHash:       commit.Hash(),
//		ProposerAddress:    crypto.AddressHash([]byte("proposer_address")),
//	}
//
//	validSignedHeader := BNBSignedHeader{Header: &h, Commit: commit}
//	validSignedHeader.Commit.BlockID.Hash = validSignedHeader.Header.Hash()
//	invalidSignedHeader := BNBSignedHeader{}
//
//	testCases := []struct {
//		testName  string
//		shHeader  *BNBHeader
//		shCommit  *types.Commit
//		expectErr bool
//	}{
//		{"Valid Signed Header", validSignedHeader.Header, validSignedHeader.Commit, false},
//		{"Invalid Signed Header", invalidSignedHeader.Header, validSignedHeader.Commit, true},
//		{"Invalid Signed Header", validSignedHeader.Header, invalidSignedHeader.Commit, true},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		t.Run(tc.testName, func(t *testing.T) {
//			sh := BNBSignedHeader{
//				Header: tc.shHeader,
//				Commit: tc.shCommit,
//			}
//			assert.Equal(
//				t,
//				tc.expectErr,
//				sh.ValidateBasic(validSignedHeader.Header.ChainID) != nil,
//				"Validate Basic had an unexpected result",
//			)
//		})
//	}
//}
//
//var header2 = `
//{ "header" : {
//	"version": {
//	"block": "10",
//	"app": "0"
//	},
//	"chain_id": "Binance-Chain-Tigris",
//	"height": "1",
//	"time": "2019-04-18T05:59:26.228734998Z",
//	"num_txs": "0",
//	"total_txs": "0",
//	"last_block_id": {
//	"hash": "",
//	"parts": {
//	"total": "0",
//	"hash": ""
//	}
//	},
//	"last_commit_hash": "",
//	"data_hash": "",
//	"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
//	"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
//	"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
//	"app_hash": "",
//	"last_results_hash": "",
//	"evidence_hash": "",
//	"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
//},
//"commit":
//        "precommits": [
//          null,
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.233083179Z",
//            "validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
//            "validator_index": "1",
//            "signature": "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw=="
//          },
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.153180042Z",
//            "validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
//            "validator_index": "2",
//            "signature": "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw=="
//          },
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.23279118Z",
//            "validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
//            "validator_index": "3",
//            "signature": "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg=="
//          },
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.154340159Z",
//            "validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
//            "validator_index": "4",
//            "signature": "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA=="
//          },
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.23639005Z",
//            "validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
//            "validator_index": "5",
//            "signature": "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg=="
//          },
//          null,
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.232155249Z",
//            "validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
//            "validator_index": "7",
//            "signature": "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw=="
//          },
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.153428422Z",
//            "validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
//            "validator_index": "8",
//            "signature": "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ=="
//          },
//          {
//            "type": 2,
//            "height": "1",
//            "round": "1",
//            "block_id": {
//              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
//              "parts": {
//                "total": "1",
//                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
//              }
//            },
//            "timestamp": "2019-04-18T06:07:02.153867648Z",
//            "validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
//            "validator_index": "9",
//            "signature": "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw=="
//          },
//          null
//        ]
//      }
//}`
//
//func TestBNBSignedHeader_Verify(t *testing.T) {
//	timeBlock, _ := time.Parse(time.RFC3339Nano, "2019-04-18T05:59:26.228734998Z")
//	blockIDHash, _ := hex.DecodeString("494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8")
//	blockIDPartHash, _ := hex.DecodeString("AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11")
//	blockID := types.BlockID{
//		Hash: blockIDHash,
//		PartsHeader: types.PartSetHeader{
//			Total: 1,
//			Hash:  blockIDPartHash,
//		},
//	}
//	lastCommitHash, _ := hex.DecodeString("")
//	dataHash, _ := hex.DecodeString("")
//	validatorsHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
//	nextValidatorsHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
//	consensusHash, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
//	appHash, _ := hex.DecodeString("")
//	lastResultsHash, _ := hex.DecodeString("")
//	evidenceHash, _ := hex.DecodeString("")
//	proposerAddress, _ := hex.DecodeString("14CFCE69B645F3F88BAF08EA5B77FA521E4480F9")
//
//	sh := new(BNBSignedHeader)
//	sh.Header = &BNBHeader{
//		Version:            Version{App: 0, Block: 10},
//		ChainID:            "Binance-Chain-Tigris",
//		Height:             1,
//		Time:               timeBlock,
//		NumTxs:             0,
//		TotalTxs:           0,
//		LastBlockID:        types.BlockID{},
//		LastCommitHash:     lastCommitHash,
//		DataHash:           dataHash,
//		ValidatorsHash:     validatorsHash,
//		NextValidatorsHash: nextValidatorsHash,
//		ConsensusHash:      consensusHash,
//		AppHash:            appHash,
//		LastResultsHash:    lastResultsHash,
//		EvidenceHash:       evidenceHash,
//		ProposerAddress:    proposerAddress,
//	}
//
//	sh.Commit = &types.Commit{
//		Height:     1,
//		Round:      1,
//		BlockID:    blockID,
//		Signatures: []types.CommitSig{},
//	}
//
//	var sigs = []struct {
//		ValidatorAddress string
//		Signature        string
//		Time             string
//	}{
//		//{},
//		{"14CFCE69B645F3F88BAF08EA5B77FA521E4480F9", "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw==", "2019-04-18T06:07:02.233083179Z"},
//		{"17B42E8F284D3CA0E420262F89CD76C749BB12C9", "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw==", "2019-04-18T06:07:02.153180042Z"},
//		{"3CD4AABABDDEB7ABFEA9618732E331077A861D2B", "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg==", "2019-04-18T06:07:02.23279118Z"},
//		{"414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F", "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA==", "2019-04-18T06:07:02.154340159Z"},
//		{"71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C", "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg==", "2019-04-18T06:07:02.23639005Z"},
//		//{},
//		{"A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E", "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw==", "2019-04-18T06:07:02.232155249Z"},
//		{"A9157B3FA6EB4C1E396B9B746E95327A07DC42E5", "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ==", "2019-04-18T06:07:02.153428422Z"},
//		{"B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D", "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw==", "2019-04-18T06:07:02.153867648Z"},
//		//{},
//	}
//
//	sh.Commit.Signatures = make([]types.CommitSig, len(sigs))
//	for i := 0; i < len(sigs); i++ {
//		validatorAddressBytes := []byte{}
//		signatureBytes := []byte{}
//		if sigs[i].ValidatorAddress != "" {
//			validatorAddressBytes, _ = hex.DecodeString(sigs[i].ValidatorAddress)
//			signatureBytes, _ = base64.StdEncoding.DecodeString(sigs[i].Signature)
//			time, _ := time.Parse(time.RFC3339Nano, sigs[i].Time)
//			sh.Commit.Signatures[i] = types.CommitSig{
//				BlockIDFlag:      types.BlockIDFlagCommit,
//				ValidatorAddress: validatorAddressBytes,
//				Timestamp:        time,
//				Signature:        signatureBytes,
//			}
//		}
//	}
//
//	isValid, err := sh.Verify()
//	assert.Equal(t, true, isValid)
//	assert.Equal(t, nil, err)
//}
//
//func TestVoteSignVerify(t *testing.T) {
//	privVal := types.NewMockPV()
//	pubkey := privVal.GetPubKey()
//
//	time, _ := time.Parse(time.RFC3339Nano, "2019-04-18T06:07:02.233083179Z")
//
//	blockIDHash, _ := hex.DecodeString("494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8")
//	blockIDPartHash, _ := hex.DecodeString("AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11")
//	blockID := types.BlockID{
//		Hash: blockIDHash,
//		PartsHeader: types.PartSetHeader{
//			Total: 1,
//			Hash:  blockIDPartHash,
//		},
//	}
//
//	//pubkey := validatorsMainnet["14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"].publicKey
//
//	vote := &types.Vote{
//		Type:             2,
//		Height:           1,
//		Round:            1,
//		Timestamp:        time,
//		BlockID:          blockID,
//		ValidatorAddress: pubkey.Bytes(),
//		ValidatorIndex:   1,
//	}
//	vote.ValidatorAddress = pubkey.Address()
//	//vote.Signature, _ = base64.StdEncoding.DecodeString("jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw==")
//
//	privVal.SignVote("Binance-Chain-Tigris", vote)
//
//	// verify with unmatching public key
//	err := vote.Verify("Binance-Chain-Tigris", ed25519.GenPrivKey().PubKey())
//	assert.NotEqual(t, nil, err)
//
//	// verify with matching public key
//	err = vote.Verify("Binance-Chain-Tigris", pubkey)
//	assert.Equal(t, nil, err)
//}
//
//
////block 2
////"header": {
////"version": {
////"block": "10",
////"app": "0"
////},
////"chain_id": "Binance-Chain-Tigris",
////"height": "2",
////"time": "2019-04-18T06:07:02.154340159Z",
////"num_txs": "0",
////"total_txs": "0",
////"last_block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"last_commit_hash": "DBDCD4FB32B47D560A7FB045D254C997EDB456FB98F49495B9C1808981DB57AE",
////"data_hash": "",
////"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
////"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
////"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
////"app_hash": "B6E4D322EDA04F783EBDCD3D59AA908D211675F890EBC05A893C2093D101D009",
////"last_results_hash": "",
////"evidence_hash": "",
////"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
////},
////"data": {
////"txs": null
////},
////"evidence": {
////"evidence": null
////},
////"last_commit": {
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"precommits": [
////null,
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.233083179Z",
////"validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
////"validator_index": "1",
////"signature": "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw=="
////},
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.153180042Z",
////"validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
////"validator_index": "2",
////"signature": "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw=="
////},
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.23279118Z",
////"validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
////"validator_index": "3",
////"signature": "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg=="
////},
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.154340159Z",
////"validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
////"validator_index": "4",
////"signature": "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA=="
////},
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.23639005Z",
////"validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
////"validator_index": "5",
////"signature": "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg=="
////},
////null,
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.232155249Z",
////"validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
////"validator_index": "7",
////"signature": "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw=="
////},
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.153428422Z",
////"validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
////"validator_index": "8",
////"signature": "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ=="
////},
////{
////"type": 2,
////"height": "1",
////"round": "1",
////"block_id": {
////"hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
////"parts": {
////"total": "1",
////"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
////}
////},
////"timestamp": "2019-04-18T06:07:02.153867648Z",
////"validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
////"validator_index": "9",
////"signature": "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw=="
////},
////null
////]
////
////
////
//////block 3
////"block": {
////"header": {
////"version": {
////"block": "10",
////"app": "0"
////},
////"chain_id": "Binance-Chain-Tigris",
////"height": "3",
////"time": "2019-04-18T06:07:07.739441221Z",
////"num_txs": "0",
////"total_txs": "0",
////"last_block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"last_commit_hash": "E7BDF3448318E45BA69BAE0BC3B94F5005279B4FFA1FEDF51AAF0F4CF1F6D529",
////"data_hash": "",
////"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
////"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
////"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
////"app_hash": "04C4F444F7101F26AAFC4F80507BEB61C066354BBB4B94502E6AF7E7DBE69CA1",
////"last_results_hash": "",
////"evidence_hash": "",
////"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
////},
////"data": {
////"txs": null
////},
////"evidence": {
////"evidence": null
////},
////"last_commit": {
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"precommits": [
////null,
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.81740164Z",
////"validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
////"validator_index": "1",
////"signature": "VyHgok5hmgzjUJGPUPwAquXtmuLaJu4p2B5kzCvY/oI06LC2bbUPLUrkNFAtmz7XkRhVqUHMwgODSCsxpqpeAA=="
////},
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.7393153Z",
////"validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
////"validator_index": "2",
////"signature": "kv6bsey/YxsRHuxLo2juzDHPrD/ZkxY432ii1gHDnuFXhbqwZM0xrKb+BU7dz7DLhCYJ8wuw/yo43vYP9/WCAQ=="
////},
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.817334959Z",
////"validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
////"validator_index": "3",
////"signature": "vS7ASlEZsg+5PFAJ8YYDozCBswsbG0VrdQKBh6R6JhhJLt84KSqSaVkAGQtzFH397kAxDJpDEtULNkzd6v6NAw=="
////},
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.735960102Z",
////"validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
////"validator_index": "4",
////"signature": "RUxRSxOA1pqqh8LvG2V7BRBanhcD51rC36gQNdztASlx4DVON7UhUPMaDYXKTjm/DQsMMPz8363HHflWz4RPAg=="
////},
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.818207384Z",
////"validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
////"validator_index": "5",
////"signature": "zASakyX56qEkiAE+rXFXCjVch2Fr8CcOmUAkoQNvIu4efkK9wAkdTlfwSmDtxN+fa398MkadRGxkLWUHNS0BBg=="
////},
////null,
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.816888459Z",
////"validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
////"validator_index": "7",
////"signature": "0tW24yXhTOo5eW8hAIwhFZUumdsHB9fdMtJgAKbKcGP79nOoCBYAoEObYITVL/XpfXjHglmDcTHUmTnpqQDCAw=="
////},
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.739441221Z",
////"validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
////"validator_index": "8",
////"signature": "V+dwIBkgrvZrDP659JP6sW3YejrStULTf2MROly+M8EZdRYfQLt56GcwcvricEl9VcHAHAJILgkENTmZG6WuDw=="
////},
////{
////"type": 2,
////"height": "2",
////"round": "1",
////"block_id": {
////"hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
////"parts": {
////"total": "1",
////"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
////}
////},
////"timestamp": "2019-04-18T06:07:07.738348307Z",
////"validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
////"validator_index": "9",
////"signature": "zGrjAbOy20MQXk00aj/w5AkG7Bkbl11PQlZjKmEwKlerUXmLXDWBQ8Yez/bVvYJdfZLUek6WGPwvq7bXu4o5AA=="
////},
////null
////]
////
//
//
//func TestValidatorHash(t *testing.T){
//	validatorHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
//
//	validatorBytes := make([][]byte, 11)
//	for i, v := range MainnetValidatorAddresses{
//		validatorBytes[i] = validatorsMainnet[v].Bytes()
//	}
//
//	hash := merkle.SimpleHashFromByteSlices(validatorBytes)
//	fmt.Printf("validatorHash: %v\n", validatorHash)
//	fmt.Printf("hash: %v\n", hash)
//	fmt.Printf("Result: %v\n", bytes.Equal(validatorHash, hash))
//
//	//sortedValidator
//}

func TestParseHeaderFromJson(t *testing.T) {
	json := `{
	"version": {
		"block": "10",
			"app": "0"
	},
	"chain_id": "Binance-Chain-Tigris",
		"height": "66038977",
		"time": "2020-02-04T07:16:30.773117407Z",
		"num_txs": "1",
		"total_txs": "55866604",
		"last_block_id": {
		"hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
			"parts": {
			"total": "1",
				"hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
		}
	},
	"last_commit_hash": "FBED868DE0C209B53A9D8016F07DD711BEBFF32206B0CF0388B4505277394C35",
		"data_hash": "E9BDA8FD942F6908A7398B7D5FA1C46553240BC919B0836083878F2C7BA91709",
		"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
		"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
		"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
		"app_hash": "478A8630FCEF61B1E6239D263BB149866A74E1ABAF586A32076C455B5800854B",
		"last_results_hash": "C0D03C905732518BC495C4D88E92D2C9833F40355847A0B2AD4D15537A40DC81",
		"evidence_hash": "",
		"proposer_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9"
}`
ParseHeaderFromJson(json)
}


func TestParseCommitFromJson(t *testing.T) {
	json := `
	{
        "block_id": {
          "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
          "parts": {
            "total": 1,
            "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
          }
        },
        "precommits": [
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.801216049Z",
            "validator_address": "1175946A48EAA473868A0A6F52E6C66CCAF472EA",
            "validator_index": 0,
            "signature": "bC71LRWXi43xsY76MPeBvolmblC4jkFvMlrOSb5lXF9WZA1t8qeA2FoLZ4xvnfvDqhV5gXLDsCx0T1lOIEOxCA=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.774335439Z",
            "validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
            "validator_index": 1,
            "signature": "W9B9ujKqDQ7i7sneE+4msH0QB/MeDKC9nnvJs0k4ZNf1zsBjTZn9w1CGzI7kKjcLugSAfVDVeCmMhl7LlaVOAA=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.685083047Z",
            "validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
            "validator_index": 2,
            "signature": "NJv3Y1CZByjQYgCqS0L5TOeYun4VFkrF+kGA9WVxsaQ5UHl6nCb0usHDdvHhd31OMKFIvQZx/Tj+sxDO7O7PCA=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.773117407Z",
            "validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
            "validator_index": 3,
            "signature": "yjSvfkcpuaLiElfmL+yZtgRd8SEQonuUcn88OqHMB0sfD1Yt3Hv1BxkXwX3VKOS5H5P3z6UhLJ/sGhAxg+XxBw=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.683656798Z",
            "validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
            "validator_index": 4,
            "signature": "C2N9G0QN0Pb2pw19nfdlCG9VvAOAAWzKbqDribQxm8jzwLbv6FJZSZTcpNJj3VcFGGbvhMKurFc0rkN9veJUBw=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.774747928Z",
            "validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
            "validator_index": 5,
            "signature": "WgCnCFBzAH6jzcRuKawBlyXWujI0n1swjtZmiaC8AtyGUWpXCsPHX5V6aYzJYP6QRKibqIyjTDWxDTf+psJ+Cw=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.795537037Z",
            "validator_address": "7235EF143D20FC0ABC427615D83014BB02D7C06C",
            "validator_index": 6,
            "signature": "lgFAGZsbiUW7MgwVeGAUB1oFpHHPuYdog5SBLOGTjyuCgCOSXrWOLaDgjaS0mLrV4G0fKEGGJXlu4/aaoWdtCw=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.772955892Z",
            "validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
            "validator_index": 7,
            "signature": "+dKs7gdou+cB63BAhSLwK13p/B+L4OtYPRbTcCcwKJzXDFEScliwjjszQB2ARD6BkReSCdyC0cd9y4JD7qZpDw=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.683610963Z",
            "validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
            "validator_index": 8,
            "signature": "XtJkkzCzzcXm3g7Lh6y+7/w+EdooB302BRFqEpzGrbW32KLqqInbKa9ITFaByaoZrbDYS1a38pn5fIOqorhrCg=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.677151461Z",
            "validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
            "validator_index": 9,
            "signature": "7EJJTTIvqadGt8pGW1/64veDth5IkWhCQrq9eUo64ylQkALDHyh2GO72CMAt78CchdWrm5bhPI6wmdagYz5pCg=="
          },
          {
            "type": 2,
            "height": 66038976,
            "round": 0,
            "block_id": {
              "hash": "BD0056A133E081FA6C1847EECE504A011F57BBA265AA13166AF055465B8E3D90",
              "parts": {
                "total": 1,
                "hash": "EFED1843385A46A995B140F3DBC41CFA124DF0EFB9B2EAC652911851A43841C3"
              }
            },
            "timestamp": "2020-02-04T07:16:30.801295534Z",
            "validator_address": "B7707D9F593C62E85BB9E1A2366D12A97CD5DFF2",
            "validator_index": 10,
            "signature": "//+Awm4ANTwOHZL+WuAIm7Ibaa54a6AbxE5Nm0fELizv/X2/zk3gaLnhbQNqYSMmpQ5XvFQcwJQWJMBcFvx5CQ=="
          }
        ]
      }
`

ParseCommitFromJson(json)
}




var header1 = `
{
	"version": {
	"block": 10,
	"app": 0
	},
	"chain_id": "Binance-Chain-Tigris",
	"height": 1,
	"time": "2019-04-18T05:59:26.228734998Z",
	"num_txs": 0,
	"total_txs": 0,
	"last_block_id": {
	"hash": "",
	"parts": {
	"total": 0,
	"hash": ""
	}
	},
	"last_commit_hash": "",
	"data_hash": "",
	"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
	"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
	"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
	"app_hash": "",
	"last_results_hash": "",
	"evidence_hash": "",
	"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
}`

type BNBHeader struct {
	Header *types.Header		`json:"Header"`
	LastCommit *types.Commit	`json:"LastCommit"`
}

func TestGenerateB64EncodeHeader(t *testing.T){
	h := new(BNBHeader)
	var err error
	h.Header, err = ParseHeaderFromJson(header1)
	if err != nil {
		fmt.Println("Error parse: ", err)
	}
	h.LastCommit = nil

	hBytes, err := json.Marshal(h)
	if err != nil {
		fmt.Println("Error marshal: ", err)
	}
	hStr := base64.StdEncoding.EncodeToString(hBytes)
	fmt.Printf("header base64 encode: %v\n", hStr)

	h2Bytes, err := base64.StdEncoding.DecodeString(hStr)
	if err != nil {
		fmt.Println("Error decode: ", err)
	}

	var h2 BNBHeader
	err = json.Unmarshal(h2Bytes, &h2)
	if err != nil {
		fmt.Println("Error unmarshal: ", err)
	}

	assert.Equal(t, h2.Header.Height, h.Header.Height)
}

var header2 = `
{
        "version": {
          "block": 10,
          "app": 0
        },
        "chain_id": "Binance-Chain-Tigris",
        "height": 2,
        "time": "2019-04-18T06:07:02.154340159Z",
        "num_txs": 0,
        "total_txs": 0,
        "last_block_id": {
          "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
          "parts": {
            "total": 1,
            "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
          }
        },
        "last_commit_hash": "DBDCD4FB32B47D560A7FB045D254C997EDB456FB98F49495B9C1808981DB57AE",
        "data_hash": "",
        "validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
        "next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
        "consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
        "app_hash": "B6E4D322EDA04F783EBDCD3D59AA908D211675F890EBC05A893C2093D101D009",
        "last_results_hash": "",
        "evidence_hash": "",
        "proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
      }
`
var commit1 = `
	{
        "block_id": {
          "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
          "parts": {
            "total": 1,
            "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
          }
        },
        "precommits": [
          null,
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.233083179Z",
            "validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
            "validator_index": 1,
            "signature": "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw=="
          },
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.153180042Z",
            "validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
            "validator_index": 2,
            "signature": "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw=="
          },
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.23279118Z",
            "validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
            "validator_index": 3,
            "signature": "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg=="
          },
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.154340159Z",
            "validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
            "validator_index": 4,
            "signature": "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA=="
          },
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.23639005Z",
            "validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
            "validator_index": 5,
            "signature": "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg=="
          },
          null,
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.232155249Z",
            "validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
            "validator_index": 7,
            "signature": "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw=="
          },
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.153428422Z",
            "validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
            "validator_index": 8,
            "signature": "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ=="
          },
          {
            "type": 2,
            "height": 1,
            "round": 1,
            "block_id": {
              "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
              "parts": {
                "total": 1,
                "hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
              }
            },
            "timestamp": "2019-04-18T06:07:02.153867648Z",
            "validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
            "validator_index": 9,
            "signature": "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw=="
          },
          null
        ]
      }
`
var header3 = `
	{
        "version": {
          "block": 10,
          "app": 0
        },
        "chain_id": "Binance-Chain-Tigris",
        "height": 3,
        "time": "2019-04-18T06:07:07.739441221Z",
        "num_txs": 0,
        "total_txs": 0,
        "last_block_id": {
          "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
          "parts": {
            "total": 1,
            "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
          }
        },
        "last_commit_hash": "E7BDF3448318E45BA69BAE0BC3B94F5005279B4FFA1FEDF51AAF0F4CF1F6D529",
        "data_hash": "",
        "validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
        "next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
        "consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
        "app_hash": "04C4F444F7101F26AAFC4F80507BEB61C066354BBB4B94502E6AF7E7DBE69CA1",
        "last_results_hash": "",
        "evidence_hash": "",
        "proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
      }
	`
var commit2 = `
	{
        "block_id": {
          "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
          "parts": {
            "total": 1,
            "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
          }
        },
        "precommits": [
          null,
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.81740164Z",
            "validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
            "validator_index": 1,
            "signature": "VyHgok5hmgzjUJGPUPwAquXtmuLaJu4p2B5kzCvY/oI06LC2bbUPLUrkNFAtmz7XkRhVqUHMwgODSCsxpqpeAA=="
          },
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.7393153Z",
            "validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
            "validator_index": 2,
            "signature": "kv6bsey/YxsRHuxLo2juzDHPrD/ZkxY432ii1gHDnuFXhbqwZM0xrKb+BU7dz7DLhCYJ8wuw/yo43vYP9/WCAQ=="
          },
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.817334959Z",
            "validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
            "validator_index": 3,
            "signature": "vS7ASlEZsg+5PFAJ8YYDozCBswsbG0VrdQKBh6R6JhhJLt84KSqSaVkAGQtzFH397kAxDJpDEtULNkzd6v6NAw=="
          },
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.735960102Z",
            "validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
            "validator_index": 4,
            "signature": "RUxRSxOA1pqqh8LvG2V7BRBanhcD51rC36gQNdztASlx4DVON7UhUPMaDYXKTjm/DQsMMPz8363HHflWz4RPAg=="
          },
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.818207384Z",
            "validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
            "validator_index": 5,
            "signature": "zASakyX56qEkiAE+rXFXCjVch2Fr8CcOmUAkoQNvIu4efkK9wAkdTlfwSmDtxN+fa398MkadRGxkLWUHNS0BBg=="
          },
          null,
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.816888459Z",
            "validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
            "validator_index": 7,
            "signature": "0tW24yXhTOo5eW8hAIwhFZUumdsHB9fdMtJgAKbKcGP79nOoCBYAoEObYITVL/XpfXjHglmDcTHUmTnpqQDCAw=="
          },
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.739441221Z",
            "validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
            "validator_index": 8,
            "signature": "V+dwIBkgrvZrDP659JP6sW3YejrStULTf2MROly+M8EZdRYfQLt56GcwcvricEl9VcHAHAJILgkENTmZG6WuDw=="
          },
          {
            "type": 2,
            "height": 2,
            "round": 1,
            "block_id": {
              "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
              "parts": {
                "total": 1,
                "hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
              }
            },
            "timestamp": "2019-04-18T06:07:07.738348307Z",
            "validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
            "validator_index": 9,
            "signature": "zGrjAbOy20MQXk00aj/w5AkG7Bkbl11PQlZjKmEwKlerUXmLXDWBQ8Yez/bVvYJdfZLUek6WGPwvq7bXu4o5AA=="
          },
          null
        ]
      }
`
func TestGenerateB64EncodeHeader2(t *testing.T) {
	h := new(BNBHeader)
	var err error
	h.Header, err = ParseHeaderFromJson(header3)
	if err != nil {
		fmt.Println("Error parse: ", err)
	}
	h.LastCommit, err = ParseCommitFromJson(commit2)
	if err != nil {
		fmt.Println("Error parse last commit: ", err)
	}

	hBytes, err := json.Marshal(h)
	if err != nil {
		fmt.Println("Error marshal: ", err)
	}
	hStr := base64.StdEncoding.EncodeToString(hBytes)
	fmt.Printf("header base64 encode: %v\n", hStr)
}


//"eyJIZWFkZXIiOnsidmVyc2lvbiI6eyJibG9jayI6MTAsImFwcCI6MH0sImNoYWluX2lkIjoiQmluYW5jZS1DaGFpbi1UaWdyaXMiLCJoZWlnaHQiOjEsInRpbWUiOiIyMDE5LTA0LTE4VDA1OjU5OjI2LjIyODczNDk5OFoiLCJudW1fdHhzIjowLCJ0b3RhbF90eHMiOjAsImxhc3RfYmxvY2tfaWQiOnsiaGFzaCI6IiIsInBhcnRzIjp7InRvdGFsIjowLCJoYXNoIjoiIn19LCJsYXN0X2NvbW1pdF9oYXNoIjoiIiwiZGF0YV9oYXNoIjoiIiwidmFsaWRhdG9yc19oYXNoIjoiNDNDNTNBNTBEODY1M0VGOENGMUU1NzE2REE2ODEyMEZCNTFCNjM2REM2RDExMUVDMzI3N0IwOThFQ0Q0MkQ0OSIsIm5leHRfdmFsaWRhdG9yc19oYXNoIjoiNDNDNTNBNTBEODY1M0VGOENGMUU1NzE2REE2ODEyMEZCNTFCNjM2REM2RDExMUVDMzI3N0IwOThFQ0Q0MkQ0OSIsImNvbnNlbnN1c19oYXNoIjoiMjk0RDhGQkQwQjk0Qjc2N0E3RUJBOTg0MEYyOTlBMzU4NkRBN0ZFNkI1REVBRDNCN0VFQ0JBMTkzQzQwMEY5MyIsImFwcF9oYXNoIjoiIiwibGFzdF9yZXN1bHRzX2hhc2giOiIiLCJldmlkZW5jZV9oYXNoIjoiIiwicHJvcG9zZXJfYWRkcmVzcyI6IjE0Q0ZDRTY5QjY0NUYzRjg4QkFGMDhFQTVCNzdGQTUyMUU0NDgwRjkifSwiTGFzdENvbW1pdCI6bnVsbH0="
//"eyJIZWFkZXIiOnsidmVyc2lvbiI6eyJibG9jayI6MTAsImFwcCI6MH0sImNoYWluX2lkIjoiQmluYW5jZS1DaGFpbi1UaWdyaXMiLCJoZWlnaHQiOjIsInRpbWUiOiIyMDE5LTA0LTE4VDA2OjA3OjAyLjE1NDM0MDE1OVoiLCJudW1fdHhzIjowLCJ0b3RhbF90eHMiOjAsImxhc3RfYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sImxhc3RfY29tbWl0X2hhc2giOiJEQkRDRDRGQjMyQjQ3RDU2MEE3RkIwNDVEMjU0Qzk5N0VEQjQ1NkZCOThGNDk0OTVCOUMxODA4OTgxREI1N0FFIiwiZGF0YV9oYXNoIjoiIiwidmFsaWRhdG9yc19oYXNoIjoiNDNDNTNBNTBEODY1M0VGOENGMUU1NzE2REE2ODEyMEZCNTFCNjM2REM2RDExMUVDMzI3N0IwOThFQ0Q0MkQ0OSIsIm5leHRfdmFsaWRhdG9yc19oYXNoIjoiNDNDNTNBNTBEODY1M0VGOENGMUU1NzE2REE2ODEyMEZCNTFCNjM2REM2RDExMUVDMzI3N0IwOThFQ0Q0MkQ0OSIsImNvbnNlbnN1c19oYXNoIjoiMjk0RDhGQkQwQjk0Qjc2N0E3RUJBOTg0MEYyOTlBMzU4NkRBN0ZFNkI1REVBRDNCN0VFQ0JBMTkzQzQwMEY5MyIsImFwcF9oYXNoIjoiQjZFNEQzMjJFREEwNEY3ODNFQkRDRDNENTlBQTkwOEQyMTE2NzVGODkwRUJDMDVBODkzQzIwOTNEMTAxRDAwOSIsImxhc3RfcmVzdWx0c19oYXNoIjoiIiwiZXZpZGVuY2VfaGFzaCI6IiIsInByb3Bvc2VyX2FkZHJlc3MiOiIxNENGQ0U2OUI2NDVGM0Y4OEJBRjA4RUE1Qjc3RkE1MjFFNDQ4MEY5In0sIkxhc3RDb21taXQiOnsiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInByZWNvbW1pdHMiOltudWxsLHsidHlwZSI6MiwiaGVpZ2h0IjoxLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDIuMjMzMDgzMTc5WiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiMTRDRkNFNjlCNjQ1RjNGODhCQUYwOEVBNUI3N0ZBNTIxRTQ0ODBGOSIsInZhbGlkYXRvcl9pbmRleCI6MSwic2lnbmF0dXJlIjoiamJ1SGFDZlQvdVpEZGJMY0VIR1daSHVxYzFwa1Y2aGxnYXFxN3dmclBmNU5ObkJBR3BxY0Q1UjlWOGRHcSthdkFjZzdHLzdmeHdYSzZQdCtRMmQvQ3c9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoxLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDIuMTUzMTgwMDQyWiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiMTdCNDJFOEYyODREM0NBMEU0MjAyNjJGODlDRDc2Qzc0OUJCMTJDOSIsInZhbGlkYXRvcl9pbmRleCI6Miwic2lnbmF0dXJlIjoidzROMkZUMWZIRVg4SHJIWTVqaXE4QXBOc3VMQktLOHI3S3h3d0xDY0l0eWRCTUU5N0lJSzBTRGNKT3VOQ3VSZDd6TWZncDVKUkdpSnVTSlNCOEw2QXc9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoxLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDIuMjMyNzkxMThaIiwidmFsaWRhdG9yX2FkZHJlc3MiOiIzQ0Q0QUFCQUJEREVCN0FCRkVBOTYxODczMkUzMzEwNzdBODYxRDJCIiwidmFsaWRhdG9yX2luZGV4IjozLCJzaWduYXR1cmUiOiJtelp1TUdNL3ErOTJHaVk0WkwyUXY0V3ZSY3k3Vmg5MVV3SzV1aFhiU0E1aitSMVJhYm1UalpwRkJ0aFFzT2MyVEdUeUN2c0U1ZW41RjFwd2ZpV3RCZz09In0seyJ0eXBlIjoyLCJoZWlnaHQiOjEsInJvdW5kIjoxLCJibG9ja19pZCI6eyJoYXNoIjoiNDk0QTY3NEZDREFEQjBGNjdCQzk5Q0JDMzNCNkY4MDYxRTFCNzZBNzg2NDg1QjY2MkVFNDIyQzY5MThDODZDOCIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiQUY0Q0VERkFCN0ZCN0U1MjYzRDlDNUE3MkY1RUNEM0YzQTRDMTg2RDhEM0MxMEQ4MEM5NUZDQjJCQ0VFMEUxMSJ9fSwidGltZXN0YW1wIjoiMjAxOS0wNC0xOFQwNjowNzowMi4xNTQzNDAxNTlaIiwidmFsaWRhdG9yX2FkZHJlc3MiOiI0MTRGQjNCQkEyMTZBRjg0QzQ3RTA3RDZFQkFBMkRDRkMzNTYzQTJGIiwidmFsaWRhdG9yX2luZGV4Ijo0LCJzaWduYXR1cmUiOiJKRVRzY0hyb0t1TUJrbnRRN0FMb1dmVHdwdHE1eXlPRG1pV3RBN0FjOXYxLy9Fc1FKLzBqVkUvWUhtVTFMWUZYdUo2OFUzSTdsNk5IN2pNRWd4dG5BQT09In0seyJ0eXBlIjoyLCJoZWlnaHQiOjEsInJvdW5kIjoxLCJibG9ja19pZCI6eyJoYXNoIjoiNDk0QTY3NEZDREFEQjBGNjdCQzk5Q0JDMzNCNkY4MDYxRTFCNzZBNzg2NDg1QjY2MkVFNDIyQzY5MThDODZDOCIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiQUY0Q0VERkFCN0ZCN0U1MjYzRDlDNUE3MkY1RUNEM0YzQTRDMTg2RDhEM0MxMEQ4MEM5NUZDQjJCQ0VFMEUxMSJ9fSwidGltZXN0YW1wIjoiMjAxOS0wNC0xOFQwNjowNzowMi4yMzYzOTAwNVoiLCJ2YWxpZGF0b3JfYWRkcmVzcyI6IjcxRjI1M0U2RkVBOUVERDRCNDc1M0Y1NDgzNTQ5RkU0RjBGM0EyMUMiLCJ2YWxpZGF0b3JfaW5kZXgiOjUsInNpZ25hdHVyZSI6Im44b1BOVFlMalBJR3pyeGFXa2FBWFpRYTZpcWNHWHNWZlhqcDI3OXNsWWdFcVZBY0d1SHdNdWJ6UFBGUUcxS1dUdXZwZndBaEtCUE5xNG5xczBnTEJnPT0ifSxudWxsLHsidHlwZSI6MiwiaGVpZ2h0IjoxLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDIuMjMyMTU1MjQ5WiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiQTcxRTVDRDA3OEI4QzVDN0IxQUY4OEJDRTg0REQ3MEIwNTU3RDkzRSIsInZhbGlkYXRvcl9pbmRleCI6Nywic2lnbmF0dXJlIjoiMG51aGZVWGNINlhmY3gzS0loQ0lwYkJaZmpFVTNGUTNPN3lHQkhxbnd3bGpDUi9GRzA2Kzd3VElPMmxmUGtBbmFrWHdzcGorUzM2ZEJUa25nUWVBRHc9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoxLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDIuMTUzNDI4NDIyWiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiQTkxNTdCM0ZBNkVCNEMxRTM5NkI5Qjc0NkU5NTMyN0EwN0RDNDJFNSIsInZhbGlkYXRvcl9pbmRleCI6OCwic2lnbmF0dXJlIjoicS9ZdzB5YXNKdHRJTk5wVUNRUmY5WXdocUxvVmNkRWlodVJLOHZmYWtoaEo2TUZvblgzWW50NVpHb0xseUliVnlhZW5WbVA3bGNlZmhlRmtuYzdOQVE9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoxLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IjQ5NEE2NzRGQ0RBREIwRjY3QkM5OUNCQzMzQjZGODA2MUUxQjc2QTc4NjQ4NUI2NjJFRTQyMkM2OTE4Qzg2QzgiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkFGNENFREZBQjdGQjdFNTI2M0Q5QzVBNzJGNUVDRDNGM0E0QzE4NkQ4RDNDMTBEODBDOTVGQ0IyQkNFRTBFMTEifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDIuMTUzODY3NjQ4WiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiQjBGQkI1MkZGN0VFOTNDQzQ3NkRGRTZCNzRGQTFGQzg4NTg0RjMwRCIsInZhbGlkYXRvcl9pbmRleCI6OSwic2lnbmF0dXJlIjoiVFYyVnpDcXh1QVdRYURGSG00ZFhwWHRFY1l3L0VkTXhNdEQ1d1NNb3hMZWYyU1B2blFsa0cvc1kyek9iM0tOdGdPbWplbGtka3pxZUEwa0FQSWxMQnc9PSJ9LG51bGxdfX0="
//"eyJIZWFkZXIiOnsidmVyc2lvbiI6eyJibG9jayI6MTAsImFwcCI6MH0sImNoYWluX2lkIjoiQmluYW5jZS1DaGFpbi1UaWdyaXMiLCJoZWlnaHQiOjMsInRpbWUiOiIyMDE5LTA0LTE4VDA2OjA3OjA3LjczOTQ0MTIyMVoiLCJudW1fdHhzIjowLCJ0b3RhbF90eHMiOjAsImxhc3RfYmxvY2tfaWQiOnsiaGFzaCI6IkI3RkIyMjI2MkVFNjM5MjFFQjI2MEUxMEEyODVBQzE0QTk2Q0M3OTRGNjI5OEVBQkI3MTU1REQxQjRFRjIyQ0YiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkY4MDE1MjBEQjg1NzU0RjE4MDAyNzczNzUyQzAwMjA2MDVCM0MyNUEzNTczMzk4MjZDMTJENzFBMTUyMjQ3NUQifX0sImxhc3RfY29tbWl0X2hhc2giOiJFN0JERjM0NDgzMThFNDVCQTY5QkFFMEJDM0I5NEY1MDA1Mjc5QjRGRkExRkVERjUxQUFGMEY0Q0YxRjZENTI5IiwiZGF0YV9oYXNoIjoiIiwidmFsaWRhdG9yc19oYXNoIjoiNDNDNTNBNTBEODY1M0VGOENGMUU1NzE2REE2ODEyMEZCNTFCNjM2REM2RDExMUVDMzI3N0IwOThFQ0Q0MkQ0OSIsIm5leHRfdmFsaWRhdG9yc19oYXNoIjoiNDNDNTNBNTBEODY1M0VGOENGMUU1NzE2REE2ODEyMEZCNTFCNjM2REM2RDExMUVDMzI3N0IwOThFQ0Q0MkQ0OSIsImNvbnNlbnN1c19oYXNoIjoiMjk0RDhGQkQwQjk0Qjc2N0E3RUJBOTg0MEYyOTlBMzU4NkRBN0ZFNkI1REVBRDNCN0VFQ0JBMTkzQzQwMEY5MyIsImFwcF9oYXNoIjoiMDRDNEY0NDRGNzEwMUYyNkFBRkM0RjgwNTA3QkVCNjFDMDY2MzU0QkJCNEI5NDUwMkU2QUY3RTdEQkU2OUNBMSIsImxhc3RfcmVzdWx0c19oYXNoIjoiIiwiZXZpZGVuY2VfaGFzaCI6IiIsInByb3Bvc2VyX2FkZHJlc3MiOiIxNENGQ0U2OUI2NDVGM0Y4OEJBRjA4RUE1Qjc3RkE1MjFFNDQ4MEY5In0sIkxhc3RDb21taXQiOnsiYmxvY2tfaWQiOnsiaGFzaCI6IkI3RkIyMjI2MkVFNjM5MjFFQjI2MEUxMEEyODVBQzE0QTk2Q0M3OTRGNjI5OEVBQkI3MTU1REQxQjRFRjIyQ0YiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkY4MDE1MjBEQjg1NzU0RjE4MDAyNzczNzUyQzAwMjA2MDVCM0MyNUEzNTczMzk4MjZDMTJENzFBMTUyMjQ3NUQifX0sInByZWNvbW1pdHMiOltudWxsLHsidHlwZSI6MiwiaGVpZ2h0IjoyLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IkI3RkIyMjI2MkVFNjM5MjFFQjI2MEUxMEEyODVBQzE0QTk2Q0M3OTRGNjI5OEVBQkI3MTU1REQxQjRFRjIyQ0YiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkY4MDE1MjBEQjg1NzU0RjE4MDAyNzczNzUyQzAwMjA2MDVCM0MyNUEzNTczMzk4MjZDMTJENzFBMTUyMjQ3NUQifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDcuODE3NDAxNjRaIiwidmFsaWRhdG9yX2FkZHJlc3MiOiIxNENGQ0U2OUI2NDVGM0Y4OEJBRjA4RUE1Qjc3RkE1MjFFNDQ4MEY5IiwidmFsaWRhdG9yX2luZGV4IjoxLCJzaWduYXR1cmUiOiJWeUhnb2s1aG1nempVSkdQVVB3QXF1WHRtdUxhSnU0cDJCNWt6Q3ZZL29JMDZMQzJiYlVQTFVya05GQXRtejdYa1JoVnFVSE13Z09EU0NzeHBxcGVBQT09In0seyJ0eXBlIjoyLCJoZWlnaHQiOjIsInJvdW5kIjoxLCJibG9ja19pZCI6eyJoYXNoIjoiQjdGQjIyMjYyRUU2MzkyMUVCMjYwRTEwQTI4NUFDMTRBOTZDQzc5NEY2Mjk4RUFCQjcxNTVERDFCNEVGMjJDRiIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiRjgwMTUyMERCODU3NTRGMTgwMDI3NzM3NTJDMDAyMDYwNUIzQzI1QTM1NzMzOTgyNkMxMkQ3MUExNTIyNDc1RCJ9fSwidGltZXN0YW1wIjoiMjAxOS0wNC0xOFQwNjowNzowNy43MzkzMTUzWiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiMTdCNDJFOEYyODREM0NBMEU0MjAyNjJGODlDRDc2Qzc0OUJCMTJDOSIsInZhbGlkYXRvcl9pbmRleCI6Miwic2lnbmF0dXJlIjoia3Y2YnNleS9ZeHNSSHV4TG8yanV6REhQckQvWmt4WTQzMmlpMWdIRG51RlhoYnF3Wk0weHJLYitCVTdkejdETGhDWUo4d3V3L3lvNDN2WVA5L1dDQVE9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoyLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IkI3RkIyMjI2MkVFNjM5MjFFQjI2MEUxMEEyODVBQzE0QTk2Q0M3OTRGNjI5OEVBQkI3MTU1REQxQjRFRjIyQ0YiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkY4MDE1MjBEQjg1NzU0RjE4MDAyNzczNzUyQzAwMjA2MDVCM0MyNUEzNTczMzk4MjZDMTJENzFBMTUyMjQ3NUQifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDcuODE3MzM0OTU5WiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiM0NENEFBQkFCRERFQjdBQkZFQTk2MTg3MzJFMzMxMDc3QTg2MUQyQiIsInZhbGlkYXRvcl9pbmRleCI6Mywic2lnbmF0dXJlIjoidlM3QVNsRVpzZys1UEZBSjhZWURvekNCc3dzYkcwVnJkUUtCaDZSNkpoaEpMdDg0S1NxU2FWa0FHUXR6RkgzOTdrQXhESnBERXRVTE5remQ2djZOQXc9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoyLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IkI3RkIyMjI2MkVFNjM5MjFFQjI2MEUxMEEyODVBQzE0QTk2Q0M3OTRGNjI5OEVBQkI3MTU1REQxQjRFRjIyQ0YiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkY4MDE1MjBEQjg1NzU0RjE4MDAyNzczNzUyQzAwMjA2MDVCM0MyNUEzNTczMzk4MjZDMTJENzFBMTUyMjQ3NUQifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDcuNzM1OTYwMTAyWiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiNDE0RkIzQkJBMjE2QUY4NEM0N0UwN0Q2RUJBQTJEQ0ZDMzU2M0EyRiIsInZhbGlkYXRvcl9pbmRleCI6NCwic2lnbmF0dXJlIjoiUlV4UlN4T0ExcHFxaDhMdkcyVjdCUkJhbmhjRDUxckMzNmdRTmR6dEFTbHg0RFZPTjdVaFVQTWFEWVhLVGptL0RRc01NUHo4MzYzSEhmbFd6NFJQQWc9PSJ9LHsidHlwZSI6MiwiaGVpZ2h0IjoyLCJyb3VuZCI6MSwiYmxvY2tfaWQiOnsiaGFzaCI6IkI3RkIyMjI2MkVFNjM5MjFFQjI2MEUxMEEyODVBQzE0QTk2Q0M3OTRGNjI5OEVBQkI3MTU1REQxQjRFRjIyQ0YiLCJwYXJ0cyI6eyJ0b3RhbCI6MSwiaGFzaCI6IkY4MDE1MjBEQjg1NzU0RjE4MDAyNzczNzUyQzAwMjA2MDVCM0MyNUEzNTczMzk4MjZDMTJENzFBMTUyMjQ3NUQifX0sInRpbWVzdGFtcCI6IjIwMTktMDQtMThUMDY6MDc6MDcuODE4MjA3Mzg0WiIsInZhbGlkYXRvcl9hZGRyZXNzIjoiNzFGMjUzRTZGRUE5RURENEI0NzUzRjU0ODM1NDlGRTRGMEYzQTIxQyIsInZhbGlkYXRvcl9pbmRleCI6NSwic2lnbmF0dXJlIjoiekFTYWt5WDU2cUVraUFFK3JYRlhDalZjaDJGcjhDY09tVUFrb1FOdkl1NGVma0s5d0FrZFRsZndTbUR0eE4rZmEzOThNa2FkUkd4a0xXVUhOUzBCQmc9PSJ9LG51bGwseyJ0eXBlIjoyLCJoZWlnaHQiOjIsInJvdW5kIjoxLCJibG9ja19pZCI6eyJoYXNoIjoiQjdGQjIyMjYyRUU2MzkyMUVCMjYwRTEwQTI4NUFDMTRBOTZDQzc5NEY2Mjk4RUFCQjcxNTVERDFCNEVGMjJDRiIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiRjgwMTUyMERCODU3NTRGMTgwMDI3NzM3NTJDMDAyMDYwNUIzQzI1QTM1NzMzOTgyNkMxMkQ3MUExNTIyNDc1RCJ9fSwidGltZXN0YW1wIjoiMjAxOS0wNC0xOFQwNjowNzowNy44MTY4ODg0NTlaIiwidmFsaWRhdG9yX2FkZHJlc3MiOiJBNzFFNUNEMDc4QjhDNUM3QjFBRjg4QkNFODRERDcwQjA1NTdEOTNFIiwidmFsaWRhdG9yX2luZGV4Ijo3LCJzaWduYXR1cmUiOiIwdFcyNHlYaFRPbzVlVzhoQUl3aEZaVXVtZHNIQjlmZE10SmdBS2JLY0dQNzluT29DQllBb0VPYllJVFZML1hwZlhqSGdsbURjVEhVbVRucHFRRENBdz09In0seyJ0eXBlIjoyLCJoZWlnaHQiOjIsInJvdW5kIjoxLCJibG9ja19pZCI6eyJoYXNoIjoiQjdGQjIyMjYyRUU2MzkyMUVCMjYwRTEwQTI4NUFDMTRBOTZDQzc5NEY2Mjk4RUFCQjcxNTVERDFCNEVGMjJDRiIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiRjgwMTUyMERCODU3NTRGMTgwMDI3NzM3NTJDMDAyMDYwNUIzQzI1QTM1NzMzOTgyNkMxMkQ3MUExNTIyNDc1RCJ9fSwidGltZXN0YW1wIjoiMjAxOS0wNC0xOFQwNjowNzowNy43Mzk0NDEyMjFaIiwidmFsaWRhdG9yX2FkZHJlc3MiOiJBOTE1N0IzRkE2RUI0QzFFMzk2QjlCNzQ2RTk1MzI3QTA3REM0MkU1IiwidmFsaWRhdG9yX2luZGV4Ijo4LCJzaWduYXR1cmUiOiJWK2R3SUJrZ3J2WnJEUDY1OUpQNnNXM1llanJTdFVMVGYyTVJPbHkrTThFWmRSWWZRTHQ1Nkdjd2N2cmljRWw5VmNIQUhBSklMZ2tFTlRtWkc2V3VEdz09In0seyJ0eXBlIjoyLCJoZWlnaHQiOjIsInJvdW5kIjoxLCJibG9ja19pZCI6eyJoYXNoIjoiQjdGQjIyMjYyRUU2MzkyMUVCMjYwRTEwQTI4NUFDMTRBOTZDQzc5NEY2Mjk4RUFCQjcxNTVERDFCNEVGMjJDRiIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiRjgwMTUyMERCODU3NTRGMTgwMDI3NzM3NTJDMDAyMDYwNUIzQzI1QTM1NzMzOTgyNkMxMkQ3MUExNTIyNDc1RCJ9fSwidGltZXN0YW1wIjoiMjAxOS0wNC0xOFQwNjowNzowNy43MzgzNDgzMDdaIiwidmFsaWRhdG9yX2FkZHJlc3MiOiJCMEZCQjUyRkY3RUU5M0NDNDc2REZFNkI3NEZBMUZDODg1ODRGMzBEIiwidmFsaWRhdG9yX2luZGV4Ijo5LCJzaWduYXR1cmUiOiJ6R3JqQWJPeTIwTVFYazAwYWovdzVBa0c3QmtibDExUFFsWmpLbUV3S2xlclVYbUxYRFdCUThZZXovYlZ2WUpkZlpMVWVrNldHUHd2cTdiWHU0bzVBQT09In0sbnVsbF19fQ=="