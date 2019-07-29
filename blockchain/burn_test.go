package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/mocks"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

const (
	remoteAddress = "e722D8b71DCC0152D47D2438556a45D3357d631f"
	txid          = "16d4660274a74fe286ba3060b500f0fc698d1e3a144779149e9c7fbc512f6cd9"
)

func TestBuildBridgeInst(t *testing.T) {
	height := uint64(123)
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
			out: [][]string{[]string{
				"72",
				"1",
				encode58(getExternalID(2)),
				remoteAddress,
				encode58(big.NewInt(2000).Bytes()),
				txid,
				encode58(big.NewInt(int64(height + 1)).Bytes()),
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			bc := BlockChain{}
			insts, err := bc.buildBridgeInstructions(
				0,
				tc.insts,
				&BestStateBeacon{BeaconHeight: height},
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

			expected := []string{
				"72",
				"1",
				encode58(getExternalID(tc.id)),
				remoteAddress,
				encode58(big.NewInt(tc.amount).Bytes()),
				txid,
				encode58(big.NewInt(height).Bytes()),
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
			t.Error(errors.Errorf("expected inst[%d] = %s, got %s", i, exp[i], s))
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
			db := &mocks.DatabaseInterface{}
			db.On("GetAllBridgeTokens").Return(tc.info, tc.err)

			_, err := findExternalTokenID(tokenID, db)
			if err == nil {
				t.Errorf("unexpected result for token info %+v, err == nil", tc.info)
			}
		})
	}
}

func setupDB(t *testing.T) *mocks.DatabaseInterface {
	tokenInfo := setupTokenInfos(t)
	db := &mocks.DatabaseInterface{}
	db.On("GetAllBridgeTokens").Return(tokenInfo, nil)
	return db
}

func setupTokenInfos(t *testing.T) []byte {
	tokens := []*lvdb.BridgeTokenInfo{
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

func newToken(b byte) *lvdb.BridgeTokenInfo {
	return &lvdb.BridgeTokenInfo{
		TokenID:         &common.Hash{b},
		ExternalTokenID: getExternalID(b),
	}
}

func getExternalID(b byte) []byte {
	var externalID []byte
	if b != 0 {
		externalID = make([]byte, 3)
		externalID[0] = b
	} else {
		externalID = make([]byte, 20)
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
