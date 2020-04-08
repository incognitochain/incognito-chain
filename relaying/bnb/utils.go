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

// getGenesisBNBBlockMainnet returns Block 75955500 from Binance testnet
func getGenesisBNBBlockTestnet() *tdmtypes.Block {
	lastBlockIDHash, _ := hex.DecodeString("D157C93E92BE2309EBD00FBED7DA9DDDACDA0CC8880D5378251F8253BE5A39B3")
	partsHeaderHash, _ := hex.DecodeString("38B304DC260622222FCB36C1A713005142956C4C9F984B05DD6AFAE9D546DB41")
	lastCommitHash, _ := hex.DecodeString("0AAC3C8D6A50792CA69361C4FFA6E03FBB8F2CAEC988E49EBE05A9F3061CA7E6")
	dataHash, _ := hex.DecodeString("EF38E0A2BA343259087FF49E0A709825F4271CFBC824B637D7BE37687E6F4AEA")
	validatorsHash, _ := hex.DecodeString("80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A")
	nextValidatorsHash, _ := hex.DecodeString("80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A")
	consensusHash, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash, _ := hex.DecodeString("DCB96B9F17459C78AAD15715DDEB9F3C7803C5CAFC0CF12393539C92F90FAEAB")
	//lastResultsHash, _ := hex.DecodeString("")
	//evidenceHash, _ := hex.DecodeString("")
	proposerAddress, _ := hex.DecodeString("37EF19AF29679B368D2B9E9DE3F8769B35786676")
	blkTime := time.Time{}
	blkTime.UnmarshalText([]byte("2020-04-08T11:00:38.832568688Z"))

	return &tdmtypes.Block{
		Header: tdmtypes.Header{
			Version: version.Consensus{
				Block: 10,
				App:   0,
			},
			ChainID:  TestnetBNBChainID,
			Height:   75955500,
			Time:     blkTime,
			NumTxs:   1,
			TotalTxs: 50470541,
			LastBlockID: tdmtypes.BlockID{
				Hash: lastBlockIDHash,
				PartsHeader: tdmtypes.PartSetHeader{
					Total: 1,
					Hash:  partsHeaderHash,
				},
			},
			LastCommitHash: lastCommitHash,
			DataHash:       dataHash,
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
