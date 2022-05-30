package coin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/wallet"
	. "github.com/stretchr/testify/assert"
)

func TestOTAReceiverMarshal(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := randomOTAReceiver()
		True(t, r.IsValid())
		jsonBytes, err := json.Marshal(r)
		NoError(t, err)
		rawBytes, err := r.Bytes()
		NoError(t, err)
		Len(t, rawBytes, 33+TxRandomGroupSize)
		var rAgain OTAReceiver
		err = json.Unmarshal(jsonBytes, &rAgain)
		NoError(t, err)
		jsonBytesAgain, err := json.Marshal(rAgain)
		NoError(t, err)
		True(t, bytes.Equal(jsonBytesAgain, jsonBytes))
	}
}

func TestOTAReceiverValid(t *testing.T) {
	type Testcase struct {
		name     string
		data     string
		expected bool
	}

	testcases := []Testcase{
		Testcase{"Valid1", "15sXoyo8kCZCHjurNC69b8WV2jMCvf5tVrcQ5mT1eH9Nm351XRjE1BH4WHHLGYPZy9dxTSLiKQd6KdfoGq4yb4gP1AU2oaJTeoGymsEzonyi1XSW2J2U7LeAVjS1S2gjbNDk1t3f9QUg2gk4", true},
		Testcase{"Valid2", "15ujixNQY1Qc5wyX9UYQW3s6cbcecFPNhrWjWiFCggeN5HukPVdjbKyRE3goUpFgZhawtBtRUK3ZSZb5LtH7bevhGzz3UTh1muzLHG3pvsE6RNB81y8xNGhyHdpHZfjwmSWDdwDe74Tg2CUP", true},
		Testcase{"Valid3", "16Q5kgtmmue79rCYHgnJpuTxwTbbXZQXZUgC65Tngkd6brUByi2ZQG1tz3xbu8Lq7H35szf2MALE6Hn4vFJqUBqeYWtLFonFe81P5hWgR47zP6xxbdC76G7Fh2fqyXHcdJwJRTgtc4Jj8SUT", true},
		Testcase{"Invalid - extra byte", "1R1sARaqtyxsZaxbyKUYGxJADoYnLrDewtkhvCr2aYTmPSgEc68T2LbDKjH1YuGY9FRZKkJQFZSfrXqexrQ92naQ8PdXumDKc9cyqCFEUFfPvYJYrUjEhhNfcdHkeiRsPUDEQZC3VQGwpWnjU", false},
		Testcase{"Invalid - extra point", "12ckAL4NW2XGdQFDRJXY6g2kCwp8QNzCqCGaP3jpzZ4wq1kntGWbZExc1MVHvXAaw8abNPaNSzJWBfMG9PdA2nvWq9wBv2PWyEqzKbb3trw7jMDQZV6KfWKkN9UWGv1TrQYAypmVtzbTrHoe3N6gcNg3rFMpa3YSm99a2Whm38BMATPxw2wnVeLJsca6", false},
		Testcase{"Invalid - key type", "12w5UVxaMedfp1W5rKaCJE8Bh15VxJQjyDRV6Dh7DMhUM5wGF9ZSGwevhy2YFSGAA4chLWJm4jj9qjXdXVAThaCyTNMQD7w9rWLvFCSDB2gb9fpb7hkKvLFCJCDNV6iYDGjpuejjc4dXG9qgH", false},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			var r OTAReceiver
			err := r.FromString(testcase.data)
			Equal(t, err == nil, testcase.expected)
			// data that fail IsValid() will not unmarshal successfully
			// Equal(t, r.IsValid(), testcase.expected)
		})
	}
}

func TestOTAReceiverFromAddress(t *testing.T) {
	type Testcase struct {
		name     string
		data     string
		expected bool
	}

	testcases := []Testcase{
		Testcase{"Valid", "12sxXUjkMJZHz6diDB6yYnSjyYcDYiT5QygUYFsUbGUqK8PH8uhxf4LePiAE8UYoDcNkHAdJJtT1J6T8hcvpZoWLHAp8g6h1BQEfp4h5LQgEPuhMpnVMquvr1xXZZueLhTNCXc8fkebLV8nDoJ17", true},
		Testcase{"Invalid - missing OTA public key", "12S6m2LpzN17jorYnLb2ApNKaV2EVeZtd6unvrPT1GH8yHGCyjYzKbywweQDZ7aAkhD31gutYAgfQizb2JhJTgBb3AJ8jW6mbusUm4j", false},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			kw, err := wallet.Base58CheckDeserialize(testcase.data)
			NoError(t, err)
			var r OTAReceiver
			err = r.FromAddress(kw.KeySet.PaymentAddress)
			Equal(t, err == nil, testcase.expected)
		})
	}
}

func showExampleOTAReceivers() {
	fmt.Println("Valid")
	for i := 0; i < 3; i++ {
		r := randomOTAReceiver()
		s, _ := r.String()
		fmt.Println(s)
	}
	fmt.Println("Invalid")
	r := randomOTAReceiver()
	tempBytes, _ := r.Bytes()
	tempBytes = append(tempBytes, byte(rand.Int31()))
	invalidOTAReceiver := base58.Base58Check{}.NewEncode(tempBytes, common.ZeroByte)
	fmt.Println(invalidOTAReceiver)

	tempBytes, _ = r.Bytes()
	tempBytes = append(tempBytes, operation.RandomPoint().ToBytesS()...)
	invalidOTAReceiver = base58.Base58Check{}.NewEncode(tempBytes, common.ZeroByte)
	fmt.Println(invalidOTAReceiver)

	tempBytes, _ = r.Bytes()
	tempBytes[0] = 99
	invalidOTAReceiver = base58.Base58Check{}.NewEncode(tempBytes, common.ZeroByte)
	fmt.Println(invalidOTAReceiver)
}

func randomOTAReceiver() OTAReceiver {
	txr := NewTxRandom()
	txr.SetTxOTARandomPoint(operation.RandomPoint())
	txr.SetTxConcealRandomPoint(operation.RandomPoint())
	txr.SetIndex(rand.Uint32())
	return OTAReceiver{
		PublicKey: operation.RandomPoint(),
		TxRandom:  *txr,
	}
}
