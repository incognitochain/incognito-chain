package bls

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"
)

func randomMessage() []byte {
	msg := make([]byte, 32)
	rand.Read(msg)
	return msg
}

func TestSignVerify(t *testing.T) {
	for i := 0; i < 100; i++ {
		msg := randomMessage()
		pub, priv, err := GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("error generating key: %s", err)
		}

		sig := Sign(priv, msg)
		ok := Verify(pub, msg, sig)
		if !ok {
			t.Fatalf("expected signature to verify: msg=%q pub=%s priv=%s sig=%#v", msg, pub.gx, priv.x, sig)
		}

		compressedSig := CompressG1(sig.s)
		ok = VerifyCompressSig(pub, msg, compressedSig)
		if !ok {
			t.Fatalf("expected compressed signature to verify: msg=%q pub=%s priv=%s sig=%#v", msg, pub.gx, priv.x, sig)
		}

		msg[0] = ^msg[0]
		ok = Verify(pub, msg, sig)
		if ok {
			t.Fatalf("expected signature to not verify")
		}
	}
}

func TestSignVerifyAgg(t *testing.T) {
	for index := 0; index < 100; index++ {

		members := CCommiteeSize
		pubLs := make([]*PublicKey, members)
		privLs := make([]*PrivateKey, members)
		sigLs := make([]*Signature, members)
		msg := randomMessage()
		for i := 0; i < members; i++ {
			pubLs[i], privLs[i], _ = GenerateKey(rand.Reader)
			sigLs[i] = Sign(privLs[i], msg)
		}

		var subset []int
		for i := 0; i < int(members*2/3+1); i++ {
			flag := false
			tmp := mrand.Intn(members)
			for j := 0; j < len(subset); j++ {
				if subset[j] == tmp {
					flag = true
					break
				}
			}
			if flag == false {
				subset = append(subset, tmp)
			}
		}

		subsetPub := make([]*PublicKey, len(subset))
		subsetSig := make([]*Signature, len(subset))

		for i := 0; i < len(subset); i++ {
			subsetPub[i] = pubLs[subset[i]]
			subsetSig[i] = sigLs[subset[i]]
		}

		aggSig := Aggregate(subsetSig, subsetPub)
		ok := VerifyAgg(subsetPub, msg, aggSig)
		if !ok {
			t.Fatalf("expected aggregated signature to verify")
		}

		compressedSig := CompressG1(aggSig.s)
		ok = VerifyAggCompressSig(subsetPub, msg, compressedSig)
		if !ok {
			t.Fatalf("expected aggregated compressed signature to verify")
		}

		msg[10] = ^msg[10]
		ok = VerifyAgg(subsetPub, msg, aggSig)
		if ok {
			t.Fatalf("expected aggregated signature to not verify")
		}

		subsetPub[0] = pubLs[1]
		ok = VerifyAgg(subsetPub, msg, aggSig)
		if ok {
			t.Fatalf("expected aggregated signature to not verify")
		}

		aggSig = Aggregate(sigLs, pubLs)
		ok = VerifyAgg(subsetPub, msg, aggSig)
		if ok {
			t.Fatalf("expected aggregated signature to not verify")
		}
	}
}

func TestBatchVerifyDistinct(t *testing.T) {
	for index := 0; index < 50; index++ {
		members := CCommiteeSize
		msgLs := make([][]byte, members)
		pubLs := make([]*PublicKey, members)
		privLs := make([]*PrivateKey, members)
		sigLs := make([]*Signature, members)

		for i := 0; i < members; i++ {
			msgLs[i] = randomMessage()
			pubLs[i], privLs[i], _ = GenerateKey(rand.Reader)
			sigLs[i] = Sign(privLs[i], msgLs[i])
		}

		ok := BatchVerifyDistinct(pubLs, msgLs, sigLs)
		if !ok {
			t.Fatalf("expected batching (pub, msg, sig) to verify")
		}

		// change message
		tmpMsg := msgLs[0][10]
		msgLs[0][10] = ^msgLs[0][10]
		ok = BatchVerifyDistinct(pubLs, msgLs, sigLs)
		if ok {
			t.Fatalf("expected batching (pub, msg, sig) to not verify")
		}
		msgLs[0][10] = tmpMsg

		tmpPub, tmpPriv, _ := GenerateKey(rand.Reader)
		tmpSig := Sign(tmpPriv, msgLs[0])

		// change public key
		tmp1 := pubLs[0]
		pubLs[0] = tmpPub
		ok = BatchVerifyDistinct(pubLs, msgLs, sigLs)
		if ok {
			t.Fatalf("expected batching (pub, msg, sig) to not verify")
		}
		pubLs[0] = tmp1

		// change signature
		tmp2 := sigLs[0]
		sigLs[0] = Sign(tmpPriv, msgLs[0])
		ok = BatchVerifyDistinct(pubLs, msgLs, sigLs)
		if ok {
			t.Fatalf("expected batching (pub, msg, sig) to not verify")
		}
		sigLs[0] = tmp2

		// change one tuple
		sigLs[0] = tmpSig
		pubLs[0] = tmpPub
		ok = BatchVerifyDistinct(pubLs, msgLs, sigLs)
		if !ok {
			t.Fatalf("expected batching (pub, msg, sig) to verify")
		}
	}

}

func BenchmarkSign(b *testing.B) {
	msg := randomMessage()
	_, priv, _ := GenerateKey(rand.Reader)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sign(priv, msg)
	}
}

func BenchmarkVerify(b *testing.B) {
	msg := randomMessage()
	pub, priv, _ := GenerateKey(rand.Reader)
	sig := Sign(priv, msg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Verify(pub, msg, sig)
	}
}

func BenchmarkAggregate(b *testing.B) {
	members := CCommiteeSize
	pubLs := make([]*PublicKey, members)
	privLs := make([]*PrivateKey, members)
	sigLs := make([]*Signature, members)
	msg := randomMessage()
	for i := 0; i < members; i++ {
		pubLs[i], privLs[i], _ = GenerateKey(rand.Reader)
		sigLs[i] = Sign(privLs[i], msg)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Aggregate(sigLs, pubLs)
	}
}

func BenchmarkVerifyAgg(b *testing.B) {
	members := CCommiteeSize
	pubLs := make([]*PublicKey, members)
	privLs := make([]*PrivateKey, members)
	sigLs := make([]*Signature, members)
	msg := randomMessage()
	for i := 0; i < members; i++ {
		pubLs[i], privLs[i], _ = GenerateKey(rand.Reader)
		sigLs[i] = Sign(privLs[i], msg)
	}
	sig := Aggregate(sigLs, pubLs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		VerifyAgg(pubLs, msg, sig)
	}
}

func BenchmarkBatchVerifyDistinct(b *testing.B) {
	members := 1000
	msgLs := make([][]byte, members)
	pubLs := make([]*PublicKey, members)
	privLs := make([]*PrivateKey, members)
	sigLs := make([]*Signature, members)

	for i := 0; i < members; i++ {
		msgLs[i] = randomMessage()
		pubLs[i], privLs[i], _ = GenerateKey(rand.Reader)
		sigLs[i] = Sign(privLs[i], msgLs[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		BatchVerifyDistinct(pubLs, msgLs, sigLs)
	}
}
