package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

func GetPosFromLength(length []int) []int {
	pos := []int{0}
	for i := 0; i < len(length); i++ {
		pos = append(pos, pos[i]+length[i])
	}
	return pos
}

func CheckLength(key []byte, length []int) bool {
	return len(key) != length[len(length)-1]
}

func GetKeyFromVariadic(args ...[]byte) []byte {
	length := make([]int, 0)
	for i := 0; i < len(args); i++ {
		length = append(length, len(args[i]))
	}
	pos := GetPosFromLength(length)
	key := make([]byte, pos[len(pos)-1])
	for i := 0; i < len(pos)-1; i++ {
		copy(key[pos[i]:pos[i+1]], args[i])
	}
	return key
}

func ParseKeyToSlice(key []byte, length []int) ([][]byte, error) {
	pos := GetPosFromLength(length)
	if pos[len(pos)-1] != len(key) {
		return nil, errors.New("key and length of args not match")
	}
	res := make([][]byte, 0)
	for i := 0; i < len(pos)-1; i++ {
		res = append(res, key[pos[i]:pos[i+1]])
	}
	return res, nil
}

func GetKeyVoteBoardSum(boardType common.BoardType, boardIndex uint32, candidatePaymentAddress *privacy.PaymentAddress) []byte {
	var key []byte
	if candidatePaymentAddress == nil {
		key = GetKeyFromVariadic(voteBoardSumPrefix, boardType.Bytes(), common.Uint32ToBytes(boardIndex))
	} else {
		key = GetKeyFromVariadic(voteBoardSumPrefix, boardType.Bytes(), common.Uint32ToBytes(boardIndex), candidatePaymentAddress.Bytes())
	}
	return key
}

func ParseKeyVoteBoardSum(key []byte) (
	boardType common.BoardType,
	boardIndex uint32,
	paymentAddress *privacy.PaymentAddress,
	err error,
) {
	length := []int{len(voteBoardSumPrefix), 1, 4, common.PaymentAddressLength}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return 0, 0, nil, err
	}
	index := 1
	i := 1
	i++

	boardType = common.BoardType(elements[iPlusPlus(&index)][0])
	boardIndex = common.BytesToUint32(elements[iPlusPlus(&index)])
	paymentAddress = privacy.NewPaymentAddressFromByte(elements[iPlusPlus(&index)])
	return boardType, boardIndex, paymentAddress, nil
}

func GetKeyVoteBoardCount(boardType common.BoardType, boardIndex uint32, paymentAddress privacy.PaymentAddress) []byte {
	key := GetKeyFromVariadic(
		voteBoardCountPrefix,
		boardType.Bytes(),
		common.Uint32ToBytes(boardIndex),
		paymentAddress.Bytes(),
	)
	return key
}

func ParseKeyVoteBoardCount(key []byte) (boardType common.BoardType, boardIndex uint32, paymentAddress []byte, err error) {
	length := []int{len(voteBoardCountPrefix), 1, 4, common.PaymentAddressLength}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return 0, 0, nil, err
	}
	index := 1

	boardType = common.BoardType(elements[iPlusPlus(&index)][0])
	boardIndex = common.BytesToUint32(elements[iPlusPlus(&index)])
	paymentAddress = elements[iPlusPlus(&index)]
	return boardType, boardIndex, paymentAddress, nil
}

func GetKeyVoteBoardList(
	boardType common.BoardType,
	boardIndex uint32,
	candidatePaymentAddress *privacy.PaymentAddress,
	voterPaymentAddress *privacy.PaymentAddress,
) []byte {
	candidateBytes := make([]byte, 0)
	voterBytes := make([]byte, 0)
	if candidatePaymentAddress != nil {
		candidateBytes = candidatePaymentAddress.Bytes()
	}
	if voterPaymentAddress != nil {
		voterBytes = voterPaymentAddress.Bytes()
	}
	key := GetKeyFromVariadic(
		voteBoardListPrefix,
		boardType.Bytes(),
		common.Uint32ToBytes(boardIndex),
		candidateBytes,
		voterBytes,
	)
	return key
}

func ParseKeyVoteBoardList(key []byte) (boardType common.BoardType, boardIndex uint32, candidatePubKey []byte, voterPaymentAddress *privacy.PaymentAddress, err error) {
	length := []int{len(voteBoardListPrefix), 1, 4, common.PaymentAddressLength, common.PaymentAddressLength}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	index := 1

	boardType = common.BoardType(elements[iPlusPlus(&index)][0])
	boardIndex = common.BytesToUint32(elements[iPlusPlus(&index)])
	candidatePubKey = elements[iPlusPlus(&index)]
	voterPaymentAddress = privacy.NewPaymentAddressFromByte(elements[iPlusPlus(&index)])
	return boardType, boardIndex, candidatePubKey, voterPaymentAddress, nil
}

func GetValueVoteBoardList(amount uint64) []byte {
	return common.Uint64ToBytes(amount)
}

func ParseValueVoteBoardList(value []byte) uint64 {
	return common.BytesToUint64(value)
}

func GetKeyThreePhraseCryptoOwner(boardType common.BoardType, constitutionIndex uint32) []byte {
	key := GetKeyFromVariadic(threePhraseCryptoOwnerPrefix,
		boardType.Bytes(),
		common.Uint32ToBytes(constitutionIndex),
	)
	return key
}

func ParseKeyThreePhraseCryptoOwner(key []byte) (boardType common.BoardType, constitutionIndex uint32, txId *common.Hash, err error) {
	length := []int{len(threePhraseCryptoOwnerPrefix), 1, 4, common.PaymentAddressLength}
	if CheckLength(key, length) {
		length[len(length)-1] = 0
	}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return 0, 0, nil, err
	}
	index := 1

	boardType = common.BoardType(elements[iPlusPlus(&index)][0])
	constitutionIndex = common.BytesToUint32(elements[iPlusPlus(&index)])

	txId = nil
	txIdData := elements[iPlusPlus(&index)]
	if len(txIdData) != 0 {
		newHash, err1 := common.NewHash(txIdData)
		if err1 != nil {
			return 0, 0, nil, err1
		}
		txId = newHash
	}

	return boardType, constitutionIndex, txId, nil
}

func ParseValueThreePhraseCryptoOwner(value []byte) (uint32, error) {
	i := common.BytesToUint32(value)
	return i, nil
}

func GetKeyThreePhraseCryptoSealer(boardType common.BoardType, constitutionIndex uint32, txId *common.Hash) []byte {
	txIdByte := make([]byte, 0)
	if txId != nil {
		txIdByte = txId.GetBytes()
	}
	key := GetKeyFromVariadic(threePhraseCryptoSealerPrefix, boardType.Bytes(), common.Uint32ToBytes(constitutionIndex), txIdByte)
	return key
}

func ParseKeyThreePhraseCryptoSealer(key []byte) (boardType common.BoardType, constitutionIndex uint32, txId *common.Hash, err error) {
	length := []int{len(threePhraseCryptoSealerPrefix), 1, 4, common.PaymentAddressLength}
	if CheckLength(key, length) {
		length[len(length)-1] = 0
	}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return 0, 0, nil, err
	}
	index := 1

	boardType = common.BoardType(elements[iPlusPlus(&index)][0])
	constitutionIndex = common.BytesToUint32(elements[iPlusPlus(&index)])

	txId = nil
	txIdData := elements[iPlusPlus(&index)]
	if len(txIdData) != 0 {
		newHash, err1 := common.NewHash(txIdData)
		if err1 != nil {
			return 0, 0, nil, err1
		}
		txId = newHash
	}
	return boardType, constitutionIndex, txId, nil
}

func GetKeyWinningVoter(boardType common.BoardType, constitutionIndex uint32) []byte {
	key := GetKeyFromVariadic(winningVoterPrefix, boardType.Bytes(), common.Uint32ToBytes(constitutionIndex))
	return key
}

func GetKeyThreePhraseVoteValue(boardType common.BoardType, constitutionIndex uint32) []byte {
	key := GetKeyFromVariadic(threePhraseVoteValuePrefix, boardType.Bytes(), common.Uint32ToBytes(constitutionIndex))
	return key
}

func ParseKeyThreePhraseVoteValue(key []byte) (boardType common.BoardType, constitutionIndex uint32, txId *common.Hash, err error) {
	length := []int{len(threePhraseVoteValuePrefix), 1, 4, common.PaymentAddressLength}
	if CheckLength(key, length) {
		length[len(length)-1] = 0
	}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return 0, 0, nil, err
	}
	index := 1

	boardType = common.BoardType(elements[iPlusPlus(&index)][0])
	constitutionIndex = common.BytesToUint32(elements[iPlusPlus(&index)])

	txId = nil
	txIdData := elements[iPlusPlus(&index)]
	if len(txIdData) != 0 {
		newHash, err1 := common.NewHash(txIdData)
		if err1 != nil {
			return boardType, constitutionIndex, txId, err1
		}
		txId = newHash
	}
	return boardType, constitutionIndex, txId, nil
}

func GetKeyEncryptFlag(boardType common.BoardType) []byte {
	key := GetKeyFromVariadic(encryptFlagPrefix, boardType.Bytes())
	return key
}

func GetKeyEncryptionLastBlockHeight(boardType common.BoardType) []byte {
	key := GetKeyFromVariadic(encryptionLastBlockHeightPrefix, boardType.Bytes())
	return key
}
