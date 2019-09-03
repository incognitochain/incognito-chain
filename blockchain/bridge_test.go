package blockchain

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strconv"
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

func TestBuildSwapConfirmInstruction(t *testing.T) {
}

func TestBuildBeaconSwapConfirmInstruction(t *testing.T) {
}

func TestBuildBridgeSwapConfirmInstruction(t *testing.T) {
}

func TestPickBridgeSwapConfirmInst(t *testing.T) {
}

func TestParseAndPadAddress(t *testing.T) {
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
			decoded := decodeSwapConfirmInst(tc.inst)
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
		base58.Base58Check{}.Encode(big.NewInt(int64(height)).Bytes(), 0x00),
		base58.Base58Check{}.Encode(big.NewInt(int64(len(addrs))).Bytes(), 0x00),
		base58.Base58Check{}.Encode(a, 0x00),
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
