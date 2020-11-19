package blockchain

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"math/big"
	"strconv"
)

type IncProofInterface interface {
	FindConfirmInst(insts [][]string, txID *common.Hash) ([]string, int)
	FindBeaconBlockWithConfirmInst(beaconBlocks []*BeaconBlock, inst []string) (*BeaconBlock, int)
	ConvertInstToBytes(inst []string) ([]byte, error)
}

type IncProof struct {
	metaType int
}

func NewIncProof(metaType int) IncProofInterface {
	switch metaType {
	case metadata.PortalCustodianWithdrawConfirmMetaV3, metadata.PortalRedeemFromLiquidationPoolConfirmMetaV3, metadata.PortalLiquidateRunAwayCustodianConfirmMetaV3:
		return PortalWithdrawCollateralProof{
			&IncProof{
				metaType: metaType,
			},
		}
	default:
		return nil
	}
}

type PortalWithdrawCollateralProof struct {
	*IncProof
}

// FindConfirmInst finds a specific instruction in a list, returns it along with its index
func (withdrawProof PortalWithdrawCollateralProof) FindConfirmInst(insts [][]string, txID *common.Hash) ([]string, int) {
	for i, inst := range insts {
		if inst[0] != strconv.Itoa(withdrawProof.metaType) || len(inst) < 8 {
			continue
		}

		h, err := common.Hash{}.NewHashFromStr(inst[6])
		if err != nil {
			continue
		}

		if h.IsEqual(txID) {
			return inst, i
		}
	}
	return nil, -1
}

// FindBeaconBlockWithConfirmInst finds a beacon block with a specific incognito instruction and the instruction's index; nil if not found
func (withdrawProof PortalWithdrawCollateralProof) FindBeaconBlockWithConfirmInst(beaconBlocks []*BeaconBlock, inst []string) (*BeaconBlock, int) {
	for _, b := range beaconBlocks {
		for k, blkInst := range b.Body.Instructions {
			diff := false
			// Ignore block height (last element)
			for i, part := range inst[:len(inst)-1] {
				if i >= len(blkInst) || part != blkInst[i] {
					diff = true
					break
				}
			}
			if !diff {
				return b, k
			}
		}
	}
	return nil, -1
}

func (withdrawProof PortalWithdrawCollateralProof) ConvertInstToBytes(inst []string) ([]byte, error) {
	if len(inst) < 8 {
		return nil, errors.New("invalid length of WithdrawCollateralConfirm inst")
	}

	m, _ := strconv.Atoi(inst[0])
	metaType := byte(m)
	s, _ := strconv.Atoi(inst[1])
	shardID := byte(s)
	l, _ := strconv.Atoi(inst[2])
	lenExternalCollateral := byte(l)
	cusPaymentAddress := []byte(inst[3])
	externalAddress, err := common.DecodeETHAddr(inst[4])
	if err != nil {
		Logger.log.Errorf("Decode external address error: ", err)
		return nil, err
	}
	exteralCollaterals, _, err := base58.Base58Check{}.Decode(inst[5])
	if err != nil {
		Logger.log.Errorf("Decode exteral collaterals error: ", err)
		return nil, err
	}

	txIDStr := inst[6]
	txID, _ := common.Hash{}.NewHashFromStr(txIDStr)

	beaconHeightStr := inst[7]
	bcHeightBN, _ := new(big.Int).SetString(beaconHeightStr, 10)
	bcHeightBytes := common.AddPaddingBigInt(bcHeightBN, 32)

	//Logger.log.Errorf("metaType: %v", metaType)
	//Logger.log.Errorf("shardID: %v", shardID)
	//Logger.log.Errorf("cusPaymentAddress: %v - %v", cusPaymentAddress, len(cusPaymentAddress))
	//Logger.log.Errorf("externalAddress: %v - %v", externalAddress, len(externalAddress))
	//Logger.log.Errorf("externalTokenID: %v - %v", externalTokenID, len(externalTokenID))
	//Logger.log.Errorf("amountBytes: %v - %v", amountBytes, len(amountBytes))
	//Logger.log.Errorf("txID: %v - %v", txID[:])

	//BLogger.log.Infof("Decoded WithdrawCollateralConfirm inst, amount: %d, remoteAddr: %x, externalTokenID: %x", amount, externalAddress, externalTokenID)
	flatten := []byte{}
	flatten = append(flatten, metaType)
	flatten = append(flatten, shardID)
	flatten = append(flatten, lenExternalCollateral)
	flatten = append(flatten, cusPaymentAddress...)
	flatten = append(flatten, externalAddress...)
	flatten = append(flatten, exteralCollaterals...)
	flatten = append(flatten, txID[:]...)
	flatten = append(flatten, bcHeightBytes...)
	Logger.log.Infof("flatten: %v - %v", flatten, len(flatten))
	return flatten, nil
}
