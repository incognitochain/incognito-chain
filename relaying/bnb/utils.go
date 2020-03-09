package relaying

import (
	"errors"
	"github.com/binance-chain/go-sdk/common/bech32"
	"github.com/binance-chain/go-sdk/common/types"
)

func GetAccAddressString(accAddress *types.AccAddress, networkType byte) (string, error) {
	switch networkType {
	case TestnetType: {
		bech32Addr, err := bech32.ConvertAndEncode(types.TestNetwork.Bech32Prefixes(), accAddress.Bytes())
		if err != nil {
			return "", err
		}
		return bech32Addr, nil
	}
	case MainnetType: {
		bech32Addr, err := bech32.ConvertAndEncode(types.ProdNetwork.Bech32Prefixes(), accAddress.Bytes())
		if err != nil {
			return "", err
		}
		return bech32Addr, nil
	}
	default:
		return "", errors.New("Invalid network type")
	}
}
