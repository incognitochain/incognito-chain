package tss

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestVerifySig(t *testing.T) {
	// set up
	Bech32Prefix = "thorpub"

	// poolPubKeySytr := "thorpub1addwnpepqf4rwyg6zvkyq9h0dcektq6rgajvamxqj68mw7um0d9ydgdl9gals4urvz8"
	poolPubKeySytr := "thorpub1qf4rwyg6zvkyq9h0dcektq6rgajvamxqj68mw7um0d9ydgdl9galsqpzcat"
	encodedMsg := "k2oYXKqiZrucvpgengXLeM1zKwsygOuURBK7b4+PB68="
	R := "NHm5q6QsmDseGeLURg2hCNzwz1Em0TAMLxsJ0CtVHYw="
	S := "K9cteNj5QNCYqOGUsx+Q8aMPh4/ieh4sLC1bhA2Y+HY="

	// RecoveryID := "AA=="

	// msg, err := base64.StdEncoding.DecodeString(encodedMsg)
	r, err := base64.StdEncoding.DecodeString(R)
	s, err := base64.StdEncoding.DecodeString(S)

	sig := append(r, s...)
	sigStr := base64.StdEncoding.EncodeToString(sig)

	isValid, err := VerifyTSSSig(poolPubKeySytr, encodedMsg, sigStr)
	// recovery, err := base64.StdEncoding.DecodeString(RecoveryID)
	// sig := tsslibcommon.ECSignature{
	// 	R:                 r,
	// 	S:                 s,
	// 	SignatureRecovery: recovery,
	// }

	// sig.Signature

	// signature := keysign.NewSignature(encodedMsg, R, S, RecoveryID)

	fmt.Printf("Result: %v - %v\n", isValid, err)

}
