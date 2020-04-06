package bnb

import (
	"encoding/hex"
	"errors"
	"github.com/binance-chain/go-sdk/common/bech32"
	"github.com/binance-chain/go-sdk/common/types"
	tdmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
	"time"
)

func GetAccAddressString(accAddress *types.AccAddress, chainID string) (string, error) {
	switch chainID {
	case TestnetBNBChainID:
		{
			bech32Addr, err := bech32.ConvertAndEncode(types.TestNetwork.Bech32Prefixes(), accAddress.Bytes())
			if err != nil {
				return "", err
			}
			return bech32Addr, nil
		}
	case MainnetBNBChainID:
		{
			bech32Addr, err := bech32.ConvertAndEncode(types.ProdNetwork.Bech32Prefixes(), accAddress.Bytes())
			if err != nil {
				return "", err
			}
			return bech32Addr, nil
		}
	default:
		return "", errors.New("Invalid network chainID")
	}
}

func GetGenesisBNBHeaderBlockHeight(chainID string) (int64, error) {
	switch chainID {
	case TestnetBNBChainID:
		{
			return TestnetGenesisBlockHeight, nil
		}
	case MainnetBNBChainID:
		{
			return MainnetGenesisBlockHeight, nil
		}
	default:
		return int64(0), errors.New("Invalid network chainID")
	}
}

func createGenesisHeaderChain(chainID string) (*tdmtypes.Block, error) {
	if chainID == MainnetBNBChainID {
		return getGenesisBNBBlockMainnet(), nil
	} else if chainID == TestnetBNBChainID {
		return getGenesisBNBBlockTestnet(), nil
	}

	return nil, errors.New("Invalid network chainID")
}

// getGenesisBNBBlockMainnet returns Block 79394120 from Binance mainnet
func getGenesisBNBBlockMainnet() *tdmtypes.Block {
	lastBlockIDHash, _ := hex.DecodeString("B0930A48A06AB07BBDDA67CD8AFBFA9F9D1659561585D763AB71D01CDF5039DF")
	partsHeaderHash, _ := hex.DecodeString("3CF1C5B9DE88425D37DF9DD7FCF0F6203BCAE5D13B1B47758DE1076B8BAFC5D8")
	lastCommitHash, _ := hex.DecodeString("E42713455A1D9A30F1F3507394656A9C864AADF56EF62535F7B35ACC59BB40A0")
	//dataHash, _ := hex.DecodeString("")
	validatorsHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	nextValidatorsHash, _ := hex.DecodeString("43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49")
	consensusHash, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash, _ := hex.DecodeString("1F0659D2683CEA03D7B74426D6565CE2BFC2FEA7FD8F99A8CEB27B4FFAAB62DF")
	lastResultsHash, _ := hex.DecodeString("6E340B9CFFB37A989CA544E6BB780A2C78901D3FB33738768511A30617AFA01D")
	//evidenceHash, _ := hex.DecodeString("")
	proposerAddress, _ := hex.DecodeString("14CFCE69B645F3F88BAF08EA5B77FA521E4480F9")
	blkTime := time.Time{}
	blkTime.UnmarshalText([]byte("2020-04-06T07:01:28.00123013Z"))

	return &tdmtypes.Block{
		Header: tdmtypes.Header{
			Version: version.Consensus{
				Block: 10,
				App:   0,
			},
			ChainID:  MainnetBNBChainID,
			Height:   79394120,
			Time:     blkTime,
			NumTxs:   0,
			TotalTxs: 76030277,
			LastBlockID: tdmtypes.BlockID{
				Hash: lastBlockIDHash,
				PartsHeader: tdmtypes.PartSetHeader{
					Total: 1,
					Hash:  partsHeaderHash,
				},
			},
			LastCommitHash: lastCommitHash,
			DataHash:       []byte{},
			ValidatorsHash: validatorsHash,

			NextValidatorsHash: nextValidatorsHash,
			ConsensusHash:      consensusHash,
			AppHash:            appHash,
			LastResultsHash:    lastResultsHash,
			EvidenceHash:       []byte{},
			ProposerAddress:    proposerAddress,
		},
		Data:       tdmtypes.Data{},
		Evidence:   tdmtypes.EvidenceData{},
		LastCommit: &tdmtypes.Commit{},
	}
}

// getGenesisBNBBlockMainnet returns Block 75424400 from Binance testnet
func getGenesisBNBBlockTestnet() *tdmtypes.Block {
	lastBlockIDHash, _ := hex.DecodeString("39473A9AA0F62BA90F412B2E386C952C6F3BAD8ECEA16FAD21E28D89E4DDB50E")
	partsHeaderHash, _ := hex.DecodeString("C20ECAC5DBDC1D2B436554E285743C7851FDEE5E6F3CC6E794A0EB8FB767257A")
	lastCommitHash, _ := hex.DecodeString("1015B7870199A22AB6CA85FA946AECA3948DF8DA089280936EE73C384BA3EC06")
	//dataHash, _ := hex.DecodeString("")
	validatorsHash, _ := hex.DecodeString("80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A")
	nextValidatorsHash, _ := hex.DecodeString("80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A")
	consensusHash, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash, _ := hex.DecodeString("8E5EA59E99A47CD8588A559C33DB20FCE6B55A48746DF2B8A512AC8C2D11335D")
	//lastResultsHash, _ := hex.DecodeString("")
	//evidenceHash, _ := hex.DecodeString("")
	proposerAddress, _ := hex.DecodeString("7B343E041CA130000A8BC00C35152BD7E7740037")
	blkTime := time.Time{}
	blkTime.UnmarshalText([]byte("2020-04-05T12:02:18.622604487Z"))

	return &tdmtypes.Block{
		Header: tdmtypes.Header{
			Version: version.Consensus{
				Block: 10,
				App:   0,
			},
			ChainID:  TestnetBNBChainID,
			Height:   75424400,
			Time:     blkTime,
			NumTxs:   0,
			TotalTxs: 50438403,
			LastBlockID: tdmtypes.BlockID{
				Hash: lastBlockIDHash,
				PartsHeader: tdmtypes.PartSetHeader{
					Total: 1,
					Hash:  partsHeaderHash,
				},
			},
			LastCommitHash: lastCommitHash,
			DataHash:       []byte{},
			ValidatorsHash: validatorsHash,

			NextValidatorsHash: nextValidatorsHash,
			ConsensusHash:      consensusHash,
			AppHash:            appHash,
			LastResultsHash:    []byte{},
			EvidenceHash:       []byte{},
			ProposerAddress:    proposerAddress,
		},
		Data:       tdmtypes.Data{},
		Evidence:   tdmtypes.EvidenceData{},
		LastCommit: &tdmtypes.Commit{},
	}
}
