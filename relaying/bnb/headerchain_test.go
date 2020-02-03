package relaying

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/types"
	"testing"
	"time"
)

func TestHeaderChain_ReceiveNewHeader(t *testing.T) {
	timeBlock1, _ := time.Parse(time.RFC3339Nano, "2019-04-18T05:59:26.228734998Z")
	blockIDHash1, _ := hex.DecodeString("494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8")
	blockIDPartHash1, _ := hex.DecodeString("AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11")
	blockID1 := types.BlockID{
		Hash: blockIDHash1,
		PartsHeader: types.PartSetHeader{
			Total: 1,
			Hash:  blockIDPartHash1,
		},
	}
	lastCommitHash, _ := hex.DecodeString("")
	dataHash, _ := hex.DecodeString("")
	validatorsHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	nextValidatorsHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	consensusHash, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash, _ := hex.DecodeString("")
	lastResultsHash, _ := hex.DecodeString("")
	evidenceHash, _ := hex.DecodeString("")
	proposerAddress, _ := hex.DecodeString("14CFCE69B645F3F88BAF08EA5B77FA521E4480F9")

	header1 := &BNBHeader{
		Version:            Version{App: 0, Block: 10},
		ChainID:            "Binance-Chain-Tigris",
		Height:             1,
		Time:               timeBlock1,
		NumTxs:             0,
		TotalTxs:           0,
		LastBlockID:        types.BlockID{},
		LastCommitHash:     lastCommitHash,
		DataHash:           dataHash,
		ValidatorsHash:     validatorsHash,
		NextValidatorsHash: nextValidatorsHash,
		ConsensusHash:      consensusHash,
		AppHash:            appHash,
		LastResultsHash:    lastResultsHash,
		EvidenceHash:       evidenceHash,
		ProposerAddress:    proposerAddress,
	}

	lastCommit1 := &types.Commit{
		Height:     1,
		Round:      1,
		BlockID:    blockID1,
		Signatures: []types.CommitSig{},
	}

	var sigs = []struct {
		ValidatorAddress string
		Signature        string
		Time             string
	}{
		//{},
		{"14CFCE69B645F3F88BAF08EA5B77FA521E4480F9", "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw==", "2019-04-18T06:07:02.233083179Z"},
		{"17B42E8F284D3CA0E420262F89CD76C749BB12C9", "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw==", "2019-04-18T06:07:02.153180042Z"},
		{"3CD4AABABDDEB7ABFEA9618732E331077A861D2B", "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg==", "2019-04-18T06:07:02.23279118Z"},
		{"414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F", "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA==", "2019-04-18T06:07:02.154340159Z"},
		{"71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C", "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg==", "2019-04-18T06:07:02.23639005Z"},
		//{},
		{"A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E", "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw==", "2019-04-18T06:07:02.232155249Z"},
		{"A9157B3FA6EB4C1E396B9B746E95327A07DC42E5", "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ==", "2019-04-18T06:07:02.153428422Z"},
		{"B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D", "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw==", "2019-04-18T06:07:02.153867648Z"},
		//{},
	}

	lastCommit1.Signatures = make([]types.CommitSig, len(sigs))
	for i := 0; i < len(sigs); i++ {
		validatorAddressBytes := []byte{}
		signatureBytes := []byte{}
		if sigs[i].ValidatorAddress != "" {
			validatorAddressBytes, _ = hex.DecodeString(sigs[i].ValidatorAddress)
			signatureBytes, _ = base64.StdEncoding.DecodeString(sigs[i].Signature)
			time, _ := time.Parse(time.RFC3339Nano, sigs[i].Time)
			lastCommit1.Signatures[i] = types.CommitSig{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: validatorAddressBytes,
				Timestamp:        time,
				Signature:        signatureBytes,
			}
		}
	}


	//======= header 2 =======//

	timeBlock2, _ := time.Parse(time.RFC3339Nano, "2019-04-18T06:07:02.154340159Z")
	blockIDHash2, _ := hex.DecodeString("B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF")
	blockIDPartHash2, _ := hex.DecodeString("F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D")
	blockID2 := types.BlockID{
		Hash: blockIDHash2,
		PartsHeader: types.PartSetHeader{
			Total: 1,
			Hash:  blockIDPartHash2,
		},
	}
	lastCommitHash2, _ := hex.DecodeString("DBDCD4FB32B47D560A7FB045D254C997EDB456FB98F49495B9C1808981DB57AE")
	dataHash2, _ := hex.DecodeString("")
	validatorsHash2, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	nextValidatorsHash2, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	consensusHash2, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash2, _ := hex.DecodeString("B6E4D322EDA04F783EBDCD3D59AA908D211675F890EBC05A893C2093D101D009")
	lastResultsHash2, _ := hex.DecodeString("")
	evidenceHash2, _ := hex.DecodeString("")
	proposerAddress2, _ := hex.DecodeString("14CFCE69B645F3F88BAF08EA5B77FA521E4480F9")

	header2 := &BNBHeader{
		Version:            Version{App: 0, Block: 10},
		ChainID:            "Binance-Chain-Tigris",
		Height:             2,
		Time:               timeBlock2,
		NumTxs:             0,
		TotalTxs:           0,
		LastBlockID:        blockID1,
		LastCommitHash:     lastCommitHash2,
		DataHash:           dataHash2,
		ValidatorsHash:     validatorsHash2,
		NextValidatorsHash: nextValidatorsHash2,
		ConsensusHash:      consensusHash2,
		AppHash:            appHash2,
		LastResultsHash:    lastResultsHash2,
		EvidenceHash:       evidenceHash2,
		ProposerAddress:    proposerAddress2,
	}

	lastCommit2 := &types.Commit{
		Height:     2,
		Round:      1,
		BlockID:    blockID2,
		Signatures: []types.CommitSig{},
	}

	var sigs2 = []struct {
		ValidatorAddress string
		Signature        string
		Time             string
	}{
		//{},
		{"14CFCE69B645F3F88BAF08EA5B77FA521E4480F9", "VyHgok5hmgzjUJGPUPwAquXtmuLaJu4p2B5kzCvY/oI06LC2bbUPLUrkNFAtmz7XkRhVqUHMwgODSCsxpqpeAA==", "2019-04-18T06:07:07.81740164Z"},
		{"17B42E8F284D3CA0E420262F89CD76C749BB12C9", "kv6bsey/YxsRHuxLo2juzDHPrD/ZkxY432ii1gHDnuFXhbqwZM0xrKb+BU7dz7DLhCYJ8wuw/yo43vYP9/WCAQ==", "2019-04-18T06:07:07.7393153Z"},
		{"3CD4AABABDDEB7ABFEA9618732E331077A861D2B",  "vS7ASlEZsg+5PFAJ8YYDozCBswsbG0VrdQKBh6R6JhhJLt84KSqSaVkAGQtzFH397kAxDJpDEtULNkzd6v6NAw==", "2019-04-18T06:07:07.817334959Z"},
		{"414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F", "RUxRSxOA1pqqh8LvG2V7BRBanhcD51rC36gQNdztASlx4DVON7UhUPMaDYXKTjm/DQsMMPz8363HHflWz4RPAg==", "2019-04-18T06:07:07.735960102Z"},
		{"71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C", "zASakyX56qEkiAE+rXFXCjVch2Fr8CcOmUAkoQNvIu4efkK9wAkdTlfwSmDtxN+fa398MkadRGxkLWUHNS0BBg==", "2019-04-18T06:07:07.818207384Z"},
		//{},
		{"A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E", "0tW24yXhTOo5eW8hAIwhFZUumdsHB9fdMtJgAKbKcGP79nOoCBYAoEObYITVL/XpfXjHglmDcTHUmTnpqQDCAw==", "2019-04-18T06:07:07.816888459Z"},
		{"A9157B3FA6EB4C1E396B9B746E95327A07DC42E5", "V+dwIBkgrvZrDP659JP6sW3YejrStULTf2MROly+M8EZdRYfQLt56GcwcvricEl9VcHAHAJILgkENTmZG6WuDw==", "2019-04-18T06:07:07.739441221Z"},
		{"B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D", "zGrjAbOy20MQXk00aj/w5AkG7Bkbl11PQlZjKmEwKlerUXmLXDWBQ8Yez/bVvYJdfZLUek6WGPwvq7bXu4o5AA==", "2019-04-18T06:07:07.738348307Z"},
		//{},
	}

	lastCommit2.Signatures = make([]types.CommitSig, len(sigs2))
	for i := 0; i < len(sigs2); i++ {
		validatorAddressBytes := []byte{}
		signatureBytes := []byte{}
		if sigs2[i].ValidatorAddress != "" {
			validatorAddressBytes, _ = hex.DecodeString(sigs2[i].ValidatorAddress)
			signatureBytes, _ = base64.StdEncoding.DecodeString(sigs2[i].Signature)
			time, _ := time.Parse(time.RFC3339Nano, sigs2[i].Time)
			lastCommit2.Signatures[i] = types.CommitSig{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: validatorAddressBytes,
				Timestamp:        time,
				Signature:        signatureBytes,
			}
		}
	}

	//======= header 3 =======//

	timeBlock3, _ := time.Parse(time.RFC3339Nano, "2019-04-18T06:07:07.739441221Z")
	//blockIDHash3, _ := hex.DecodeString("494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8")
	//blockIDPartHash3, _ := hex.DecodeString("AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11")
	//blockID3 := types.BlockID{
	//	Hash: blockIDHash3,
	//	PartsHeader: types.PartSetHeader{
	//		Total: 1,
	//		Hash:  blockIDPartHash3,
	//	},
	//}
	lastCommitHash3, _ := hex.DecodeString("E7BDF3448318E45BA69BAE0BC3B94F5005279B4FFA1FEDF51AAF0F4CF1F6D529")
	dataHash3, _ := hex.DecodeString("")
	validatorsHash3, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	nextValidatorsHash3, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	consensusHash3, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash3, _ := hex.DecodeString("04C4F444F7101F26AAFC4F80507BEB61C066354BBB4B94502E6AF7E7DBE69CA1")
	lastResultsHash3, _ := hex.DecodeString("")
	evidenceHash3, _ := hex.DecodeString("")
	proposerAddress3, _ := hex.DecodeString("14CFCE69B645F3F88BAF08EA5B77FA521E4480F9")

	header3 := &BNBHeader{
		Version:            Version{App: 0, Block: 10},
		ChainID:            "Binance-Chain-Tigris",
		Height:             3,
		Time:               timeBlock3,
		NumTxs:             0,
		TotalTxs:           0,
		LastBlockID:        blockID2,
		LastCommitHash:     lastCommitHash3,
		DataHash:           dataHash3,
		ValidatorsHash:     validatorsHash3,
		NextValidatorsHash: nextValidatorsHash3,
		ConsensusHash:      consensusHash3,
		AppHash:            appHash3,
		LastResultsHash:    lastResultsHash3,
		EvidenceHash:       evidenceHash3,
		ProposerAddress:    proposerAddress3,
	}


	// header chain
	headerChain := new(HeaderChain)
	headerChain.HeaderChain = []*BNBHeader{}
	headerChain.prevHeader = []*BNBHeader{}

	// receive header 1
	isResult, err := headerChain.ReceiveNewHeader(header1, nil)
	assert.Equal(t, true, isResult)
	assert.Equal(t,  (*BNBRelayingError)(nil), err)

	fmt.Printf("Header chain after receive header 1: %v\n", headerChain.HeaderChain)
	fmt.Printf("Prev header after receive header 1: %v\n", headerChain.prevHeader)

	// receive header 2
	isResult2, err := headerChain.ReceiveNewHeader(header2, lastCommit1)
	assert.Equal(t, true, isResult2)
	assert.Equal(t, (*BNBRelayingError)(nil), err)

	fmt.Printf("Header chain after receive header 2: %v\n", headerChain.HeaderChain)
	fmt.Printf("Prev header after receive header 2: %v\n", headerChain.prevHeader)

	// receive header 3
	isResult3, err := headerChain.ReceiveNewHeader(header3, lastCommit2)
	assert.Equal(t, true, isResult3)
	assert.Equal(t, (*BNBRelayingError)(nil), err)

	fmt.Printf("Header chain after receive header 3: %v\n", headerChain.HeaderChain)
	fmt.Printf("Prev header after receive header 3: %v\n", headerChain.prevHeader)
	//_ = header3
}


func TestCommitHash(t*testing.T){
	//timeBlock1, _ := time.Parse(time.RFC3339Nano, "2019-04-18T05:59:26.228734998Z")
	blockIDHash1, _ := hex.DecodeString("494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8")
	blockIDPartHash1, _ := hex.DecodeString("AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11")
	blockID1 := types.BlockID{
		Hash: blockIDHash1,
		PartsHeader: types.PartSetHeader{
			Total: 1,
			Hash:  blockIDPartHash1,
		},
	}

	lastCommit1 := &types.Commit{
		Height:     1,
		Round:      1,
		BlockID:    blockID1,
		Signatures: []types.CommitSig{},
	}

	var sigs = []struct {
		ValidatorAddress string
		Signature        string
		Time             string
	}{
		{"1175946A48EAA473868A0A6F52E6C66CCAF472EA", "", "" },
		{"14CFCE69B645F3F88BAF08EA5B77FA521E4480F9", "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw==", "2019-04-18T06:07:02.233083179Z"},
		{"17B42E8F284D3CA0E420262F89CD76C749BB12C9", "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw==", "2019-04-18T06:07:02.153180042Z"},
		{"3CD4AABABDDEB7ABFEA9618732E331077A861D2B", "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg==", "2019-04-18T06:07:02.23279118Z"},
		{"414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F", "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA==", "2019-04-18T06:07:02.154340159Z"},
		{"71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C", "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg==", "2019-04-18T06:07:02.23639005Z"},
		{"7235EF143D20FC0ABC427615D83014BB02D7C06C", "", ""},
		{"A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E", "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw==", "2019-04-18T06:07:02.232155249Z"},
		{"A9157B3FA6EB4C1E396B9B746E95327A07DC42E5", "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ==", "2019-04-18T06:07:02.153428422Z"},
		{"B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D", "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw==", "2019-04-18T06:07:02.153867648Z"},
		{"B7707D9F593C62E85BB9E1A2366D12A97CD5DFF2", "", ""},
	}

	lastCommit1.Signatures = make([]types.CommitSig, 0)
	for i := 0; i < len(sigs); i++ {
		validatorAddressBytes := []byte{}
		signatureBytes := []byte{}

		validatorAddressBytes, _ = hex.DecodeString(sigs[i].ValidatorAddress)
		if sigs[i].Signature != "" {
			signatureBytes, _ = base64.StdEncoding.DecodeString(sigs[i].Signature)
			time, _ := time.Parse(time.RFC3339Nano, sigs[i].Time)
			//fmt.Printf("validatorAddressBytes : %v\n", len(validatorAddressBytes))
			lastCommit1.Signatures = append(lastCommit1.Signatures, types.CommitSig{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: validatorAddressBytes,
				Timestamp:        time,
				Signature:        signatureBytes,
			})
			//}
			//} else{
			//	//lastCommit1.Signatures[i] = types.CommitSig{
			//	//	BlockIDFlag:      types.BlockIDFlagAbsent,
			//	//	//ValidatorAddress: validatorAddressBytes,
			//	//	//Timestamp:        ,
			//	//	//Signature:        signatureBytes,
			//	//}
			//	lastCommit1.Signatures[i] = types.NewCommitSigAbsent()
			//}
		}
	}

	fmt.Printf("commit: %v\n", lastCommit1.Signatures[0].Signature)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[0].ValidatorAddress)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[0].BlockIDFlag)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[0].Timestamp)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[1].Signature)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[1].ValidatorAddress)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[1].BlockIDFlag)
	fmt.Printf("commit: %v\n", lastCommit1.Signatures[1].Timestamp)

	lastCommitHash, _ := hex.DecodeString("DBDCD4FB32B47D560A7FB045D254C997EDB456FB98F49495B9C1808981DB57AE")
	lastCommitHash1 := lastCommit1.Hash().Bytes()

	fmt.Printf("lastCommitHash %v\n", lastCommitHash)
	fmt.Printf("lastCommitHash1 %v\n", lastCommitHash1)

	isEqual := bytes.Equal(lastCommitHash, lastCommitHash1)
	fmt.Println(isEqual)
}
