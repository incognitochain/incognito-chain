package bridgehub

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	sdk "github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
	"github.com/stretchr/testify/assert"
	"gitlab.com/thorchain/tss/go-tss/conversion"
)

func getSignature(r, s string) ([]byte, error) {
	rBytes, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		return nil, err
	}
	sBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	R := new(big.Int).SetBytes(rBytes)
	S := new(big.Int).SetBytes(sBytes)
	N := btcec.S256().N
	halfOrder := new(big.Int).Rsh(N, 1)
	// see: https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if S.Cmp(halfOrder) == 1 {
		S.Sub(N, S)
	}

	// Serialize signature to R || S.
	// R, S are padded to 32 bytes respectively.
	rBytes = R.Bytes()
	sBytes = S.Bytes()

	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes, nil
}

func TestVerifyTSSSig(t *testing.T) {
	// set up
	conversion.SetupBech32Prefix()

	poolPubKeySytr := "thorpub1addwnpepqf4rwyg6zvkyq9h0dcektq6rgajvamxqj68mw7um0d9ydgdl9gals4urvz8"
	encodedMsg := "k2oYXKqiZrucvpgengXLeM1zKwsygOuURBK7b4+PB68="
	R := "NHm5q6QsmDseGeLURg2hCNzwz1Em0TAMLxsJ0CtVHYw="
	S := "K9cteNj5QNCYqOGUsx+Q8aMPh4/ieh4sLC1bhA2Y+HY="
	// RecoveryID := "AA=="

	sigBytes, _ := getSignature(R, S)
	sig := base64.StdEncoding.EncodeToString(sigBytes)

	pubKey, err := sdk.UnmarshalPubKey(sdk.AccPK, poolPubKeySytr)
	if err != nil {
		fmt.Printf("Error unmarshal pubkey: %v\n", err)
		return
	}

	isValid, err := VerifyTSSSig(pubKey, encodedMsg, sig)

	assert.Equal(t, true, isValid)
	assert.Equal(t, nil, err)
}
