package main

import (
	"log"

	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func shortKey(src string) string {
	return src[len(src)-5:]
}

func (v *Validator) shouldInBeaconWaiting(cs *jsonresult.CommiteeState) {
	k := shortKey(v.MiningPublicKey)
	for _, c := range cs.BeaconWaiting {
		if c == k {
			return
		}
	}
	log.Printf("key %s is not in beacon waiting list\n", k)
	panic("Stop")
}
