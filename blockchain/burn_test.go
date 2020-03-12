package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"math/big"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/mocks"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

const (
	remoteAddress = "e722D8b71DCC0152D47D2438556a45D3357d631f"
	txid          = "16d4660274a74fe286ba3060b500f0fc698d1e3a144779149e9c7fbc512f6cd9"
)

var incToken, _ = common.Hash{}.NewHashFromStr("1234ac4a9b9e0707957e21725381e39866ae247000000000000000000000000")

func TestPickBurningConfirm(t *testing.T) {
	testCases := []struct {
		desc   string
		insts  [][]string
		num    int
		height int64
	}{
		{
			desc:  "No burning inst",
			insts: [][]string{[]string{"1", "2"}, []string{"3", "4"}},
		},
		{
			desc:   "Check height",
			insts:  [][]string{setupBurningConfirmInst(123, incToken[:])},
			height: 456,
			num:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			b := []*BeaconBlock{&BeaconBlock{Body: BeaconBody{Instructions: tc.insts}}}
			insts := pickBurningConfirmInstruction(b, uint64(tc.height))

			if tc.num != len(insts) {
				t.Errorf("incorrect number of insts, expect %d, got %d", tc.num, len(insts))
			}
			e := encode58(big.NewInt(tc.height).Bytes())
			for _, inst := range insts {
				o := inst[len(inst)-1]
				if e != o {
					t.Errorf("incorrect shard height of inst, expect %s, got %s", e, o)
				}
			}
		})
	}
}

func setupBurningConfirmInst(height int64, token []byte) []string {
	return []string{
		"72",
		"1",
		encode58(getExternalID(2)),
		remoteAddress,
		encode58(big.NewInt(2000).Bytes()),
		txid,
		encode58(token),
		encode58(big.NewInt(int64(height)).Bytes()),
	}
}

func TestFlattenInst(t *testing.T) {
	testCases := []struct {
		desc  string
		insts [][]string
		err   bool
	}{
		{
			desc:  "Generic inst",
			insts: [][]string{[]string{"1", "2"}, []string{"3", "4"}},
			err:   false,
		},
		{
			desc:  "Corrupted BurningConfirm inst",
			insts: [][]string{[]string{"72", "1", "2", "3", "4", "5", "6"}},
			err:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			d, err := FlattenAndConvertStringInst(tc.insts)
			ok := err != nil
			if ok != tc.err {
				t.Error(err)
			}

			if !tc.err && len(tc.insts) != len(d) {
				t.Errorf("incorrect number of insts, expect %d, got %d", len(tc.insts), len(d))
			}
		})
	}
}

func TestDecodeInstruction(t *testing.T) {
	testCases := []struct {
		desc string
		inst []string
		err  bool
		leng int
	}{
		{
			desc: "Generic inst",
			inst: []string{"1", "2", "3"},
			err:  false,
			leng: 3,
		},
		{
			desc: "Corrupted BurningConfirm inst",
			inst: []string{"72", "1", "2", "3", "4", "5", "6"},
			err:  true,
		},
		{
			desc: "Valid BurningConfirm inst",
			inst: setupBurningConfirmInst(123, incToken[:]),
			err:  false,
			leng: 194,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			d, err := DecodeInstruction(tc.inst)
			ok := err != nil
			if ok != tc.err {
				t.Error(err)
			}

			if !tc.err && tc.leng != len(d) {
				t.Errorf("incorrect length of decoded inst, expect %d, got %d", tc.leng, len(d))
			}
		})
	}
}

func TestDecodeBurningConfirmInst(t *testing.T) {
	// Precompute result
	token := make([]byte, 32)
	token[31] = 2

	ra, _ := hex.DecodeString(remoteAddress)
	addr := make([]byte, 32)
	copy(addr[12:], ra)

	tx, _ := common.Hash{}.NewHashFromStr(txid)

	testCases := []struct {
		desc string
		inst []string
		err  bool
		out  *confirm
	}{
		{
			desc: "Invalid inst",
			inst: []string{"72", "-1", "", "", "", "", ""},
			err:  true,
			out:  nil,
		},
		{
			desc: "Invalid length inst",
			inst: []string{"72", "1"},
			err:  true,
			out:  nil,
		},
		{
			desc: "Valid inst",
			inst: setupBurningConfirmInst(123, incToken[:]),
			err:  false,
			out: &confirm{
				meta:   72,
				shard:  1,
				token:  token,
				addr:   addr,
				amount: 2000,
				txid:   tx[:],
				inc:    incToken[:],
				height: 123,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			d, err := decodeBurningConfirmInst(tc.inst)
			ok := err != nil
			if ok != tc.err {
				t.Error(err)
			}

			if !tc.err {
				out := parseBurningConfirmInst(d)
				checkDecodedBurningConfirmInst(t, out, tc.out)
			}
		})
	}
}

type confirm struct {
	meta   byte
	shard  byte
	token  []byte
	addr   []byte
	amount uint64
	txid   []byte
	inc    []byte
	height uint64
}

func checkDecodedBurningConfirmInst(t *testing.T, a, e *confirm) {
	if a.meta != e.meta {
		t.Errorf("%+v", errors.Errorf("expect meta = %v, got %v", e.meta, a.meta))
	}
	if a.shard != e.shard {
		t.Errorf("%+v", errors.Errorf("expect shard = %v, got %v", e.shard, a.shard))
	}
	if !bytes.Equal(a.token, e.token) {
		t.Errorf("%+v", errors.Errorf("expect token = %v, got %v", e.token, a.token))
	}
	if !bytes.Equal(a.addr, e.addr) {
		t.Errorf("%+v", errors.Errorf("expect addr = %x, got %x", e.addr, a.addr))
	}
	if a.amount != e.amount {
		t.Errorf("%+v", errors.Errorf("expect amount = %d, got %d", e.amount, a.amount))
	}
	if !bytes.Equal(a.txid, e.txid) {
		t.Errorf("%+v", errors.Errorf("expect txid = %x, got %x", e.txid, a.txid))
	}
	if !bytes.Equal(a.inc, e.inc) {
		t.Errorf("%+v", errors.Errorf("expect inc = %x, got %x", e.inc, a.inc))
	}
	if a.height != e.height {
		t.Errorf("%+v", errors.Errorf("expect height = %d, got %d", e.height, a.height))
	}
}

func parseBurningConfirmInst(inst []byte) *confirm {
	return &confirm{
		meta:   inst[0],
		shard:  inst[1],
		token:  inst[2:34],
		addr:   inst[34:66],
		amount: big.NewInt(0).SetBytes(inst[66:98]).Uint64(),
		txid:   inst[98:130],
		inc:    inst[130:162],
		height: big.NewInt(0).SetBytes(inst[162:194]).Uint64(),
	}
}

func TestBuildBridgeInst(t *testing.T) {
	height := uint64(123)
	token := common.Hash{2}
	testCases := []struct {
		desc  string
		insts [][]string
		out   [][]string
	}{
		{
			desc:  "Action from consensus",
			insts: [][]string{[]string{"set"}, []string{"swap"}, []string{"random"}, []string{"stake"}, []string{"assign"}},
			out:   [][]string{},
		},
		{
			desc:  "Corrupted metaType",
			insts: [][]string{[]string{"27a"}},
			out:   [][]string{},
		},
		{
			desc:  "ERC20",
			insts: [][]string{setupBurningRequest(2)},
			out:   [][]string{setupBurningConfirmInst(int64(height), token[:])},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			bc := BlockChain{}
			insts, err := bc.buildBridgeInstructions(
				0,
				tc.insts,
				height,
				setupDB(t),
			)
			if err != nil {
				t.Error(err)
			}
			if len(insts) != len(tc.out) {
				t.Errorf("expected %d BurningConfirm inst, got %+v", len(tc.out), insts)
			}
			for i, e := range tc.out {
				checkBurningConfirmInst(t, insts[i], e)
			}
		})
	}
}

func TestBurnConfirmScaleAmount(t *testing.T) {
	testCases := []struct {
		desc   string
		id     byte
		amount int64
	}{
		{
			desc:   "ERC20",
			id:     byte(2),
			amount: 2000,
		},
		{
			desc:   "ETH",
			id:     byte(0),
			amount: 2000000000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			inst := setupBurningRequest(tc.id)
			height := int64(123)
			db := setupDB(t)

			c, err := buildBurningConfirmInst(inst, uint64(height), db)
			if err != nil {
				t.Error(err)
			}

			token := common.Hash{tc.id}
			expected := []string{
				"72",
				"1",
				encode58(getExternalID(tc.id)),
				remoteAddress,
				encode58(big.NewInt(tc.amount).Bytes()),
				txid,
				encode58(token[:]),
				encode58(big.NewInt(int64(height)).Bytes()),
			}
			checkBurningConfirmInst(t, c, expected)
		})
	}
}

func setupBurningRequest(id byte) []string {
	meta := metadata.BurningRequest{
		BurnerAddress: privacy.PaymentAddress{},
		BurningAmount: 2000,
		TokenID:       common.Hash{id},
		TokenName:     "token",
		RemoteAddress: remoteAddress,
	}
	txHash, _ := common.Hash{}.NewHashFromStr(txid)
	actionContent, _ := json.Marshal(map[string]interface{}{
		"meta":          meta,
		"RequestedTxID": txHash,
	})
	action := base64.StdEncoding.EncodeToString(actionContent)
	return []string{"27", action}
}

func checkBurningConfirmInst(t *testing.T, got, exp []string) {
	for i, s := range got {
		if s != exp[i] {
			t.Errorf("%+v", errors.Errorf("expected inst[%d] = %s, got %s", i, exp[i], s))
		}
	}
}

func TestFindExternalTokenID(t *testing.T) {
	testCases := []struct {
		desc string
		id   byte
		err  bool
	}{
		{
			desc: "New tokenID",
			id:   byte(99),
			err:  true,
		},
		{
			desc: "Valid tokenID",
			id:   byte(2),
			err:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tokenID := &common.Hash{tc.id}
			db := setupDB(t)

			_, err := findExternalTokenID(tokenID, db)
			e := err != nil
			if e != tc.err {
				t.Errorf("unexpected result for tokenID %d: expected %t, got %+v", tc.id, tc.err, err)
			}
		})
	}
}

func TestInvalidTokenInfo(t *testing.T) {
	testCases := []struct {
		desc string
		info []byte
		err  error
	}{
		{
			desc: "Empty token info",
			info: nil,
			err:  fmt.Errorf("Empty"),
		},
		{
			desc: "Invalid token info",
			info: make([]byte, 10),
			err:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tokenID := &common.Hash{}
			db := &mocks.Database{}
			db.On("GetAllBridgeTokens").Return(tc.info, tc.err)

			_, err := findExternalTokenID(tokenID, db)
			if err == nil {
				t.Errorf("unexpected result for token info %+v, err == nil", tc.info)
			}
		})
	}
}

func setupDB(t *testing.T) *mocks.Database {
	tokenInfo := setupTokenInfos(t)
	db := &mocks.Database{}
	db.On("GetAllBridgeTokens").Return(tokenInfo, nil)
	return db
}

func setupTokenInfos(t *testing.T) []byte {
	tokens := []*rawdbv2.BridgeTokenInfo{
		newToken(1),
		newToken(2),
		newToken(3),
		newToken(0),
	}
	tokenInfo, err := json.Marshal(tokens)
	if err != nil {
		t.Error(err)
	}
	return tokenInfo
}

func newToken(b byte) *rawdbv2.BridgeTokenInfo {
	return &rawdbv2.BridgeTokenInfo{
		TokenID:         &common.Hash{b},
		ExternalTokenID: getExternalID(b),
	}
}

func getExternalID(b byte) []byte {
	externalID := make([]byte, 20)
	if b != 0 {
		externalID[19] = b
	}
	return externalID
}

func encode58(data []byte) string {
	return base58.Base58Check{}.Encode(data, 0x00)
}

func init() {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	BLogger.Init(common.NewBackend(nil).Logger("test", true))
}
