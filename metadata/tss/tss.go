package tss

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
)

// VerifyTSS verifies ECDSA sig with:
// pubKey is TSS pubkey on S256 curve (with prefix)
// msg, sig are base64 encoded strings
func VerifyTSSSig(pubKeyStr string, msg string, sig string) (bool, error) {

	pubKeyBytes, err := ConvertStrToPubKeyBytes(pubKeyStr)
	if err != nil {
		return false, fmt.Errorf("[tss] Invalid pubKeyStr %v", err)
	}

	pub, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		return false, fmt.Errorf("[tss] Invalid pubKey %v", err)
	}

	msgBytes, err1 := base64.StdEncoding.DecodeString(msg)
	sigBytes, err2 := base64.StdEncoding.DecodeString(sig)
	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("[tss] Error decoding msg %v or sig %v", err1, err2)
	}
	if len(sigBytes) != ECDSASigLen {
		return false, fmt.Errorf("[tss] Invalid sig bytes %v", sigBytes)
	}

	// TODO 0xkraken: review format signature
	r := sigBytes[0:32]
	s := sigBytes[32:]
	isValid := ecdsa.Verify(pub.ToECDSA(), msgBytes, new(big.Int).SetBytes(r), new(big.Int).SetBytes(s))
	return isValid, nil
}
