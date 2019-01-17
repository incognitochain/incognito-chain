package metadata

import (
	"crypto/ecdsa"
	"errors"
	"math"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type UpdatingOracleBoard struct {
	Action        int8
	OraclePubKeys [][]byte
	Signs         map[string][]byte // key: pub key string, value: signature
	MetadataBase
}

func NewUpdatingOracleBoard(
	action int8,
	oraclePubKeys [][]byte,
	signs map[string][]byte,
	metaType int,
) *UpdatingOracleBoard {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &UpdatingOracleBoard{
		Action:        action,
		OraclePubKeys: oraclePubKeys,
		Signs:         signs,
		MetadataBase:  metadataBase,
	}
}

func (uob UpdatingOracleBoard) GetType() int {
	return uob.Type
}

func (uob UpdatingOracleBoard) CheckTransactionFee(
	tr Transaction,
	minFeePerKbTx uint64,
) bool {
	txFee := tr.GetTxFee()
	fullFee := minFeePerKbTx * tr.GetTxActualSize()
	if txFee < fullFee {
		return common.FalseValue
	}
	return common.TrueValue
}

func (uob UpdatingOracleBoard) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	govBoardPubKeys := bcr.GetBoardPubKeys("gov")
	boardLen := len(govBoardPubKeys)
	if boardLen == 0 {
		return common.FalseValue, errors.New("There is no one in GOV board yet.")
	}
	// verify signs
	txBytes := txr.Hash()[:]
	signs := uob.Signs
	verifiedSignCount := 0
	for _, pubKey := range govBoardPubKeys {
		sign, existed := signs[string(pubKey)]
		if !existed {
			continue
		}
		verKey := new(ecdsa.PublicKey)
		point := new(privacy.EllipticPoint)
		point, _ = privacy.DecompressKey(pubKey)
		verKey.X, verKey.Y = point.X, point.Y
		verKey.Curve = privacy.Curve

		// convert signature from byte array to ECDSASign
		r, s := common.FromByteArrayToECDSASig(sign)

		// verify signature
		res := ecdsa.Verify(verKey, txBytes, r, s)
		if res {
			verifiedSignCount += 1
		}
	}
	if verifiedSignCount < int(math.Floor(float64(boardLen/2)))+1 {
		return common.FalseValue, errors.New("Number of signatures is not enough.")
	}
	return common.TrueValue, nil
}

func (uob UpdatingOracleBoard) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if uob.Action == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's action")
	}
	if len(uob.OraclePubKeys) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's OraclePubKeys")
	}
	for _, pk := range uob.OraclePubKeys {
		if len(pk) == 0 {
			return common.FalseValue, common.FalseValue, errors.New("Wrong request info's OraclePubKey")
		}
	}
	if len(uob.Signs) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's Signs")
	}
	for pkStr, sign := range uob.Signs {
		if len(pkStr) == 0 || len(sign) == 0 {
			return common.FalseValue, common.FalseValue, errors.New("Wrong request info's Signs")
		}
	}
	return common.TrueValue, common.TrueValue, nil
}

func (uob UpdatingOracleBoard) ValidateMetadataByItself() bool {
	if uob.Type != UpdatingOracleBoardMeta {
		return common.FalseValue
	}
	if uob.Action != Add && uob.Action != Remove {
		return common.FalseValue
	}
	return common.TrueValue
}

func (uob UpdatingOracleBoard) Hash() *common.Hash {
	record := string(uob.Action)
	for _, pk := range uob.OraclePubKeys {
		record += string(pk)
	}
	// record += string(common.ToBytes(uob.Signs))
	record += uob.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
