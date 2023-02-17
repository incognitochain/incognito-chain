package tss

import (
	"errors"
	"fmt"
)

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("bech32str is empty")
	}

	hrp, bz, err := DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}

func ConvertStrToPubKeyBytes(pubKeyStr string) ([]byte, error) {
	bz, err := GetFromBech32(pubKeyStr, Bech32Prefix)
	if err != nil {
		return nil, err
	}
	return bz, nil
}
