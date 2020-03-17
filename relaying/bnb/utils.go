package relaying

import (
	"errors"
	"github.com/binance-chain/go-sdk/common/bech32"
	"github.com/binance-chain/go-sdk/common/types"
)

func GetAccAddressString(accAddress *types.AccAddress, chainID string) (string, error) {
	switch chainID {
	case TestnetBNBChainID: {
		bech32Addr, err := bech32.ConvertAndEncode(types.TestNetwork.Bech32Prefixes(), accAddress.Bytes())
		if err != nil {
			return "", err
		}
		return bech32Addr, nil
	}
	case MainnetBNBChainID: {
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
