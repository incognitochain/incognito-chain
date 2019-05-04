package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/wallet"
)

func buildBuySellConfirmInst(inst []string, shardID byte) []string {
	contentBytes, _ := base64.StdEncoding.DecodeString(inst[3])
	var buySellReqAction BuySellReqAction
	json.Unmarshal(contentBytes, &buySellReqAction)
	if len(buySellReqAction.Meta.TradeID) == 0 {
		return []string{}
	}

	fmt.Printf("[db] build buy sell confirm inst\n")
	return []string{
		strconv.Itoa(component.ConfirmBuySellRequestMeta),
		strconv.Itoa(int(shardID)),
		inst[2],
		inst[3],
		inst[4],
	}
}

func buildBuyBackConfirmInst(inst []string, shardID byte) []string {
	var buyBackInfo BuyBackInfo
	json.Unmarshal([]byte(inst[3]), &buyBackInfo)
	if len(buyBackInfo.TradeID) == 0 {
		return []string{}
	}

	fmt.Printf("[db] build buy back confirm inst\n")
	return []string{
		strconv.Itoa(component.ConfirmBuyBackRequestMeta),
		strconv.Itoa(int(shardID)),
		inst[2],
		inst[3],
	}
}

func (blockgen *BlkTmplGenerator) buildTradeBondConfirmInsts(beaconBlocks []*BeaconBlock, shardID byte) ([][]string, error) {
	insts := [][]string{}
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	dcbShardID := common.GetShardIDFromLastByte(dcbPk[len(dcbPk)-1])

	if shardID == dcbShardID {
		for _, beaconBlock := range beaconBlocks {
			for _, l := range beaconBlock.Body.Instructions {
				if len(l) <= 2 {
					continue
				}
				switch l[0] {
				case strconv.Itoa(metadata.BuyFromGOVRequestMeta):
					buySellConfirmInst := buildBuySellConfirmInst(l, shardID)
					if len(buySellConfirmInst) > 0 {
						insts = append(insts, buySellConfirmInst)
					}

				case strconv.Itoa(metadata.BuyBackRequestMeta):
					buyBackConfirmInst := buildBuyBackConfirmInst(l, shardID)
					if len(buyBackConfirmInst) > 0 {
						insts = append(insts, buyBackConfirmInst)
					}
				}
			}
		}
	}
	return insts, nil
}
