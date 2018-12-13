package transaction

import (
	"encoding/hex"
	"testing"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"github.com/ninjadotorg/constant/wallet"
)

func TestSpendGenesisTx(t *testing.T) {
	key, _ := wallet.Base58CheckDeserialize("11111119q5P6bukedopEFUh7HDuiobEhcXb8VxdygNTzNoyDyXPzmAN13UXRKnwuXPEehA6AfD9UyGbsfKsg1aKvnf8AfX6nnfSQVr9bHio")
	key.KeyPair.ImportFromPrivateKey(&key.KeyPair.PrivateKey)

	receiverKey, _ := wallet.Base58CheckDeserialize("12Ryhwu68NruAQGC3Su34avHahqcFg6d86yJHFVBbL7xF6HkNtRRFpCqTLqQ3f45QxFB92KoJSys6Jb1YYsXDAynXCvDxy8ts65LTiY")
	paymentInfo := []*client.PaymentInfo{&client.PaymentInfo{
		PaymentAddress: receiverKey.KeyPair.PublicKey,
		Amount:         100000,
	}}

	rt, _ := hex.DecodeString("a13fb6e54c5f96cbd10eec571f98162c32649d3b55243568e5864cf15e70da2a")
	anchor := common.Hash{}
	copy(anchor[:], rt)

	r := [32]byte{1}
	rho, _ := hex.DecodeString("6cdf29a91e53b19f5ca49cd37e7b59d3ade30120fab21fc00e6d82e28b9133fa")
	note := []*client.Note{&client.Note{
		Value:          1000000000,
		PaymentAddress: key.KeyPair.PublicKey.Apk,
		Rho:            rho,
		R:              r[:],
	}}
	desc := []*JoinSplitDesc{&JoinSplitDesc{note: note}}
	usableTx := []*TxNormal{&TxNormal{Descs: desc}}

	cm1, _ := hex.DecodeString("d26356e6f726dfb4c0a395f3af134851139ce1c64cfed3becc3530c8c8ad5660")
	cm2, _ := hex.DecodeString("5aaf71f995db014006d630dedf7ffcbfa8854055e6a8cc9ef153629e3045b7e1")
	commitments := [][]byte{cm1, cm2}

	var nullifiers [][]byte
	_, err := CreateTx(&key.KeyPair.PrivateKey, paymentInfo, &anchor, usableTx, nullifiers, commitments)
	if err != nil {
		panic(err)
	}
}
