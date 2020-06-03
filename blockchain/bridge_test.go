package blockchain

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/pkg/errors"
)

func TestParseAndConcatPubkeys(t *testing.T) {
	testCases := []struct {
		desc string
		vals []string
		out  []byte
		err  bool
	}{
		{
			desc: "Valid validators",
			vals: getCommitteeKeys(),
			out:  getCommitteeAddresses(),
		},
		{
			desc: "Invalid validator keys",
			vals: func() []string {
				vals := getCommitteeKeys()
				vals[0] = vals[0] + "a"
				return vals
			}(),
			err: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			addrs, err := parseAndConcatPubkeys(tc.vals)
			isErr := err != nil
			if isErr != tc.err {
				t.Error(errors.Errorf("expect error = %t, got %v", tc.err, err))
			}
			if tc.err {
				return
			}
			if !bytes.Equal(addrs, tc.out) {
				t.Errorf("invalid committee addresses, expect %x, got %x", tc.out, addrs)
			}
		})
	}
}

func TestBuildSwapConfirmInstruction(t *testing.T) {
	testCases := []struct {
		desc string
		in   *swapInput
		err  bool
	}{
		{
			desc: "Valid swap confirm instruction",
			in:   getSwapBeaconConfirmInput(),
		},
		{
			desc: "Invalid committee",
			in: func() *swapInput {
				s := getSwapBeaconConfirmInput()
				s.vals[0] = s.vals[0] + "A"
				return s
			}(),
			err: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			inst, err := buildSwapConfirmInstruction(tc.in.meta, tc.in.vals, tc.in.startHeight)
			isErr := err != nil
			if isErr != tc.err {
				t.Error(errors.Errorf("expect error = %t, got %v", tc.err, err))
			}
			if tc.err {
				return
			}
			checkSwapConfirmInst(t, inst, tc.in)
		})
	}
}

func checkSwapConfirmInst(t *testing.T, inst []string, input *swapInput) {
	if meta, err := strconv.Atoi(inst[0]); err != nil || meta != input.meta {
		t.Errorf("invalid meta, expect %d, got %d, err = %v", input.meta, meta, err)
	}
	if bridgeID, err := strconv.Atoi(inst[1]); err != nil || bridgeID != 1 {
		t.Errorf("invalid bridgeID, expect %d, got %d, err = %v", 1, bridgeID, err)
	}
	if b, ok := checkDecodeB58(t, inst[2]); ok {
		if h := big.NewInt(0).SetBytes(b); h.Uint64() != input.startHeight {
			t.Errorf("invalid height, expect %d, got %d", input.startHeight, h)
		}
	}
	if b, ok := checkDecodeB58(t, inst[3]); ok {
		if numVals := big.NewInt(0).SetBytes(b); numVals.Uint64() != uint64(len(input.vals)) {
			t.Errorf("invalid #vals, expect %d, got %d", len(input.vals), numVals)
		}
	}
	if b, ok := checkDecodeB58(t, inst[4]); ok {
		if !bytes.Equal(b, getCommitteeAddresses()) {
			t.Errorf("invalid committee addresses, expect %x, got %x", input.vals, b)
		}
	}
}

func getSwapBeaconConfirmInput() *swapInput {
	return &swapInput{
		meta:        70,
		vals:        getCommitteeKeys(),
		startHeight: 123,
	}
}

func getSwapBridgeConfirmInput() *swapInput {
	return &swapInput{
		meta:        71,
		vals:        getCommitteeKeys(),
		startHeight: 123,
	}
}

type swapInput struct {
	meta        int
	vals        []string
	startHeight uint64
}

func getCommitteeAddresses() []byte {
	comm := []string{
		"9BC0faE7BB432828759B6e391e0cC99995057791",
		"6cbc2937FEe477bbda360A842EeEbF92c2FAb613",
		"cabF3DB93eB48a61d41486AcC9281B6240411403",
	}
	addrs := []byte{}
	for _, c := range comm {
		addr, _ := hex.DecodeString(c)
		addrs = append(addrs, addr...)
	}
	return addrs
}

func getCommitteeKeys() []string {
	return []string{
		"121VhftSAygpEJZ6i9jGk9depfiEJfPCUqMoeDS3QJgAURzB7XZFeoaQtPuXYTAd46CNDt5FS1fNgKkEdKcX4PbwxDoL8hACe1bdNoRaGnwvU4wHHY2TxY3kxpTe7w6GxMzGBLwb9GEoiRCh1r2RdxWNvAHMhHMPNzBBfRtJ45iXXtJYgbbB1rUqbGiCV4TDgt5QV3v4KZFYoTiXmURyqXbQeVJkkABRu1BR16HDrfGcNi5LL3s8Z8iemeTm8F1FAvrXdWBeqsTEQeqHuUrY6s5cPVCnTfuCDRRSJFDhLx33CmTiWux8vYdWfNKFuX1E8hJU2vaSgFzjypTWsZb814FMxsztHoq1ibnAXKfbXgZxj9RwjecXWe7285WWEHZsLcWZ3ncW1x6Bga5ZDVQX1zeQh88kSnsebxmfGwQzV8HWikRM",
		"121VhftSAygpEJZ6i9jGk9dvuAMKafpQ1EiTVzFiUVLDYBAfmjkidFCjAJD7UpJQkeakutbs4MfGx1AizjdQ49WY2TWDw2q5sMNsoeHSPSE3Qaqxd45HRAdHH2A7cWseo4sMAVWchFuRaoUJrTB36cqjXVKet1aK8sQJQbPwmnrmHnztsaEw6Soi6vg7TkoG96HJwxQVZaUtWfPpWBZQje5SnLyB15VYqs7KBSK2Fqz4jk2L18idrxXojQQYRfigfdNrLsjwT7FMJhNkN31YWiCs47yZX9hzixqwj4DpsmHQqM1S7FmNApWGePXT86woSTL9yUqAYaA9xXkYDPsajjbxag7vqDyGtbanG7rzZSP3L93oiR4bFxmstYyghsezoXVUoJs9wy98JGH3MmDgZ8gK64sAAsgAu6Lk4AjvkreEyK4K",
		"121VhftSAygpEJZ6i9jGk9dPdVubogXXJe23BYZ1uBiJq4x6aLuEar5iRzsk1TfR995g4C18bPV8yi8frkoeJdPfK2a9CAfaroJdgmBHSUi1yVVAWttSDDAT5PEbr1XhnTGmP1Z82dPwKucctwLwRzDTkBXPfWXwMpYCLs21umN8zpuoR47xZhMqDEN2ZAuWcjZhnBDoxpnmhDgoRBe7QwL2KGGGyBVhXJHc4P15V8msCLxArxKX9U2TT2bQMpw18p25vkfDX7XB2ZyozZox46cKj8PTVf2BjAhMzk5dghb3ipX4kp4p8cpVSnSpsGB8UJwer4LxHhN2sRDm88M8PH3xxtAgs1RZBmPH6EojnbxxU5XZgGtouRda1tjEp5jFDgp2h87gY5VzEME9u5FEKyiAjR1Ye7429PGTmiSf48mtm1xW",
	}
}

func TestBuildBeaconSwapConfirmInstruction(t *testing.T) {
	testCases := []struct {
		desc string
		in   *swapInput
		err  bool
	}{
		{
			desc: "Valid swap confirm instruction",
			in:   getSwapBeaconConfirmInput(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			inst, err := buildBeaconSwapConfirmInstruction(tc.in.vals, tc.in.startHeight)
			isErr := err != nil
			if isErr != tc.err {
				t.Error(errors.Errorf("expect error = %t, got %v", tc.err, err))
			}
			if tc.err {
				return
			}
			tc.in.startHeight += 1 // new committee starts signing next block
			checkSwapConfirmInst(t, inst, tc.in)
		})
	}
}

func TestBuildBridgeSwapConfirmInstruction(t *testing.T) {
	testCases := []struct {
		desc string
		in   *swapInput
		err  bool
	}{
		{
			desc: "Valid swap confirm instruction",
			in:   getSwapBridgeConfirmInput(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			inst, err := buildBridgeSwapConfirmInstruction(tc.in.vals, tc.in.startHeight)
			isErr := err != nil
			if isErr != tc.err {
				t.Error(errors.Errorf("expect error = %t, got %v", tc.err, err))
			}
			if tc.err {
				return
			}
			tc.in.startHeight += 1 // new committee starts signing next block
			checkSwapConfirmInst(t, inst, tc.in)
		})
	}
}

func TestPickBridgeSwapConfirmInst(t *testing.T) {
	testCases := []struct {
		desc  string
		insts [][]string
		out   [][]string
	}{
		{
			desc:  "No swap inst",
			insts: [][]string{[]string{"1", "2"}, []string{"3", "4"}},
		},
		{
			desc:  "Check metaType",
			insts: [][]string{[]string{"70", "2"}, []string{"71", "1", "2", "3", "4"}},
			out:   [][]string{[]string{"71", "1", "2", "3", "4"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			insts := pickBridgeSwapConfirmInst(tc.insts)
			if len(tc.out) != len(insts) {
				t.Errorf("incorrect number of insts, expect %d, got %d", len(tc.out), len(insts))
			}
			for i, inst := range insts {
				if strings.Join(inst, "") != strings.Join(tc.out[i], "") {
					t.Errorf("incorrect bridge swap inst, expect %s, got %s", tc.out[i], inst)
				}
			}
		})
	}
}

func TestParseAndPadAddress(t *testing.T) {
	testCases := []struct {
		desc string
		inst string
		err  bool
	}{
		{
			desc: "Valid instruction",
			inst: base58.EncodeCheck(getCommitteeAddresses()),
		},
		{
			desc: "Decode fail",
			inst: func() string {
				inst := base58.EncodeCheck(getCommitteeAddresses())
				inst = inst + "a"
				return inst
			}(),
			err: true,
		},
		{
			desc: "Invalid address length",
			inst: func() string {
				addrs := getCommitteeAddresses()
				addrs = append(addrs, byte(123))
				return base58.EncodeCheck(addrs)
			}(),
			err: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			addrs, err := parseAndPadAddress(tc.inst)
			isErr := err != nil
			if isErr != tc.err {
				t.Error(errors.Errorf("expect error = %t, got %v", tc.err, err))
			}
			if tc.err {
				return
			}

			checkPaddedAddresses(t, addrs, tc.inst)
		})
	}
}

func checkPaddedAddresses(t *testing.T, addrs []byte, inst string) {
	b, _, _ := base58.DecodeCheck(inst)
	numAddrs := len(b) / 20
	if len(addrs) != numAddrs*32 {
		t.Fatalf("incorrect padded length, expect %d, got %d", numAddrs*32, len(addrs))
	}
	zero := make([]byte, 12)
	for i := 0; i < numAddrs; i++ {
		prefix := addrs[i*32 : i*32+12]
		if !bytes.Equal(zero, prefix) {
			t.Errorf("address must start with 12 bytes of 0, expect %x, got %x", zero, prefix)
		}

		addr := addrs[i*32+12 : (i+1)*32]
		exp := b[i*20 : (i+1)*20]
		if !bytes.Equal(exp, addr) {
			t.Errorf("wrong address of committee member, expect %x, got %x", exp, addr)
		}
	}
}

func TestDecodeSwapConfirm(t *testing.T) {
	addrs := []string{
		"834f98e1b7324450b798359c9febba74fb1fd888",
		"1250ba2c592ac5d883a0b20112022f541898e65b",
		"2464c00eab37be5a679d6e5f7c8f87864b03bfce",
		"6d4850ab610be9849566c09da24b37c5cfa93e50",
	}
	testCases := []struct {
		desc string
		inst []string
		out  []byte
		err  bool
	}{
		{
			desc: "Swap beacon instruction",
			inst: buildEncodedSwapConfirmInst(70, 1, 123, addrs),
			out:  buildDecodedSwapConfirmInst(70, 1, 123, addrs),
		},
		{
			desc: "Swap bridge instruction",
			inst: buildEncodedSwapConfirmInst(71, 1, 19827312, []string{}),
			out:  buildDecodedSwapConfirmInst(71, 1, 19827312, []string{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			decoded, err := decodeSwapConfirmInst(tc.inst)
			isErr := err != nil
			if isErr != tc.err {
				t.Error(errors.Errorf("expect error = %t, got %v", tc.err, err))
			}
			if tc.err {
				return
			}

			if !bytes.Equal(decoded, tc.out) {
				t.Errorf("invalid decoded swap inst, expect\n%v, got\n%v", tc.out, decoded)
			}
		})
	}
}

func buildEncodedSwapConfirmInst(meta, shard, height int, addrs []string) []string {
	a := []byte{}
	for _, addr := range addrs {
		d, _ := hex.DecodeString(addr)
		a = append(a, d...)
	}
	inst := []string{
		strconv.Itoa(meta),
		strconv.Itoa(shard),
		base58.EncodeCheck(big.NewInt(int64(height)).Bytes()),
		base58.EncodeCheck(big.NewInt(int64(len(addrs))).Bytes()),
		base58.EncodeCheck(a),
	}
	return inst
}

func buildDecodedSwapConfirmInst(meta, shard, height int, addrs []string) []byte {
	a := []byte{}
	for _, addr := range addrs {
		d, _ := hex.DecodeString(addr)
		a = append(a, toBytes32BigEndian(d)...)
	}
	decoded := []byte{byte(meta)}
	decoded = append(decoded, byte(shard))
	decoded = append(decoded, toBytes32BigEndian(big.NewInt(int64(height)).Bytes())...)
	decoded = append(decoded, toBytes32BigEndian(big.NewInt(int64(len(addrs))).Bytes())...)
	decoded = append(decoded, a...)
	return decoded
}

func checkDecodeB58(t *testing.T, e string) ([]byte, bool) {
	b, _, err := base58.DecodeCheck(e)
	if err != nil {
		t.Error(errors.WithStack(err))
		return nil, false
	}
	return b, true
}
