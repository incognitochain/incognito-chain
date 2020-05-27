package bnb

import (
	"encoding/hex"
	"errors"
	client2 "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/common/bech32"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
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

func IsValidBNBAddress(bnbAddr string, chainID string) bool {
	prefix, keyBytes, err := bech32.DecodeAndConvert(bnbAddr)
	if err != nil || len(keyBytes) != types.AddrLen {
		return false

	}
	if chainID == TestnetBNBChainID && prefix != types.TestNetwork.Bech32Prefixes() {
		return false
	}
	if chainID == MainnetBNBChainID && prefix != types.ProdNetwork.Bech32Prefixes() {
		return false
	}

	return true
}

func generateBNBAddress(network types.ChainNetwork) string {
	km, _ := keys.NewKeyManager()
	client, _ := client2.NewDexClient("dex.binance.org", network, km) // api string can be "https://testnet-dex.binance.org" for testnet
	accn := client.GetKeyManager().GetAddr().String()
	return accn
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

func getGenesisBlock(chainID string) (*tdmtypes.Block, error) {
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

// getGenesisBNBBlockTestnet returns Block 79473100 from Binance testnet
func getGenesisBNBBlockTestnet() *tdmtypes.Block {
	lastBlockIDHash, _ := hex.DecodeString("C709E30E6F3A7D2C8C0E38D38D7402014E3463C64057E80BAAE3316886266145")
	partsHeaderHash, _ := hex.DecodeString("BF1F250C37E9F77349230764C447EFACF692411B168292206573ADDCFCADD9C0")
	lastCommitHash, _ := hex.DecodeString("D572594AFDC07168765DE24BFBEE231747C08C631965C655CBBA3160B1DC4A7F")
	//dataHash, _ := hex.DecodeString("")
	validatorsHash, _ := hex.DecodeString("80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A")
	nextValidatorsHash, _ := hex.DecodeString("80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A")
	consensusHash, _ := hex.DecodeString("294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93")
	appHash, _ := hex.DecodeString("D118233FA9F24A9F63268F3D9A99932B454BEC148A68A66C9B506289AFAC92ED")
	//lastResultsHash, _ := hex.DecodeString("")
	//evidenceHash, _ := hex.DecodeString("")
	proposerAddress, _ := hex.DecodeString("18E69CC672973992BB5F76D049A5B2C5DDF77436")
	blkTime := time.Time{}
	blkTime.UnmarshalText([]byte("2020-04-28T07:38:47.455006399Z"))

	return &tdmtypes.Block{
		Header: tdmtypes.Header{
			Version: version.Consensus{
				Block: 10,
				App:   0,
			},
			ChainID:  TestnetBNBChainID,
			Height:   79473100,
			Time:     blkTime,
			NumTxs:   0,
			TotalTxs: 50579416,
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
