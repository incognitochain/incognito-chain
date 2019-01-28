package lvdb

import (
	"encoding/binary"
	"github.com/ninjadotorg/constant/privacy"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func iPlusPlus(x *int) int {
	*x += 1
	return *x - 1
}

func (db *db) AddVoteBoard(
	boardType string,
	boardIndex uint32,
	paymentAddress []byte,
	VoterPaymentAddress privacy.PaymentAddress,
	CandidatePaymentAddress privacy.PaymentAddress,
	amount uint64,
) error {
	//add to sum amount of vote token to this candidate
	key := GetKeyVoteBoardSum(boardType, boardIndex, &CandidatePaymentAddress)

	currentVoteInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		currentVoteInBytes = make([]byte, 8)
		binary.LittleEndian.PutUint64(currentVoteInBytes, uint64(0))
	}

	currentVote := binary.LittleEndian.Uint64(currentVoteInBytes)
	newVote := currentVote + amount

	newVoteInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newVoteInBytes, newVote)
	err = db.Put(key, newVoteInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to count amount of vote to this candidate
	key = GetKeyVoteBoardCount(boardType, boardIndex, CandidatePaymentAddress)
	currentCountInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		currentCountInBytes = make([]byte, 4)
		binary.LittleEndian.PutUint32(currentCountInBytes, uint32(0))
	}
	currentCount := binary.LittleEndian.Uint32(currentCountInBytes)
	newCount := currentCount + 1
	newCountInByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(newCountInByte, newCount)
	err = db.Put(key, newCountInByte)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to list voter new voter base on count as index
	key = GetKeyVoteBoardList(boardType, boardIndex, &CandidatePaymentAddress, &VoterPaymentAddress)
	oldAmountInByte, err := db.Get(key)
	oldAmount := uint64(0)
	if err == nil {
		oldAmount = ParseValueVoteBoardList(oldAmountInByte)
	}
	newAmount := oldAmount + amount
	newAmountInByte := GetValueVoteBoardList(newAmount)
	err = db.Put(key, newAmountInByte)
	return err
}

func GetNumberOfGovernor(boardType string) int {
	numberOfGovernors := common.NumberOfDCBGovernors
	if boardType == "gov" {
		numberOfGovernors = common.NumberOfGOVGovernors
	}
	return numberOfGovernors
}

func (db *db) GetTopMostVoteGovernor(boardType string, currentBoardIndex uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	newBoardIndex := currentBoardIndex + 1
	prefix := GetKeyVoteBoardSum(boardType, newBoardIndex, nil)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		_, _, paymentAddress, err := ParseKeyVoteBoardSum(iter.Key())
		countKey := GetKeyVoteBoardCount(boardType, newBoardIndex, *paymentAddress)
		if err != nil {
			return nil, err
		}
		countValue, err := db.Get(countKey)
		if err != nil {
			return nil, err
		}
		value := binary.LittleEndian.Uint64(iter.Value())
		candidateList = append(candidateList, database.CandidateElement{
			PaymentAddress: *paymentAddress,
			VoteAmount:     value,
			NumberOfVote:   common.BytesToUint32(countValue),
		})
	}
	sort.Sort(candidateList)
	numberOfGovernors := GetNumberOfGovernor(boardType)
	if len(candidateList) < numberOfGovernors {
		return nil, database.NewDatabaseError(database.NotEnoughCandidate, errors.Errorf("not enough Candidate"))
	}

	return candidateList[len(candidateList)-numberOfGovernors:], nil
}

func (db *db) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return db.lvdb.NewIterator(slice, ro)
}

func (db *db) AddVoteLv3Proposal(boardType string, constitutionIndex uint32, txID *common.Hash) error {
	//init sealer
	keySealer := GetKeyThreePhraseCryptoSealer(boardType, constitutionIndex, txID)
	ok, err := db.HasValue(keySealer)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate txid")
	}
	zeroInBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(zeroInBytes, 0)
	db.Put(keySealer, zeroInBytes)

	// init owner
	keyOwner := GetKeyThreePhraseCryptoOwner(boardType, constitutionIndex, txID)
	ok, err = db.HasValue(keyOwner)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate txid")
	}
	db.Put(keyOwner, zeroInBytes)

	return nil
}

func (db *db) AddVoteLv1or2Proposal(boardType string, constitutionIndex uint32, txID *common.Hash) error {
	keySealer := GetKeyThreePhraseCryptoSealer(boardType, constitutionIndex, txID)
	ok, err := db.HasValue(keySealer)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate txid")
	}
	valueInBytes, err := db.Get(keySealer)
	if err != nil {
		return err
	}
	value := binary.LittleEndian.Uint32(valueInBytes)
	newValue := value + 1
	newValueInByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(newValueInByte, newValue)
	db.Put(keySealer, newValueInByte)
	return nil
}

func (db *db) AddVoteNormalProposalFromSealer(boardType string, constitutionIndex uint32, txID *common.Hash, voteValue []byte) error {
	err := db.AddVoteLv1or2Proposal(boardType, constitutionIndex, txID)
	if err != nil {
		return err
	}
	key := GetKeyThreePhraseVoteValue(boardType, constitutionIndex, txID)

	db.Put(key, voteValue)

	return nil
}

func (db *db) AddVoteNormalProposalFromOwner(boardType string, constitutionIndex uint32, txID *common.Hash, voteValue []byte) error {
	keyOwner := GetKeyThreePhraseCryptoOwner(boardType, constitutionIndex, txID)
	ok, err := db.HasValue(keyOwner)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate txid")
	}
	if err != nil {
		return err
	}
	newValueInByte := common.Uint32ToBytes(1)
	db.Put(keyOwner, newValueInByte)

	key := GetKeyThreePhraseVoteValue(boardType, constitutionIndex, txID)
	db.Put(key, voteValue)

	return nil
}

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

func GetKeyVoteBoardSum(boardType string, boardIndex uint32, candidatePaymentAddress *privacy.PaymentAddress) []byte {
	var key []byte
	if candidatePaymentAddress == nil {
		key = GetKeyFromVariadic(voteBoardSumPrefix, []byte(boardType), common.Uint32ToBytes(boardIndex))
	} else {
		key = GetKeyFromVariadic(voteBoardSumPrefix, []byte(boardType), common.Uint32ToBytes(boardIndex), candidatePaymentAddress.Bytes())
	}
	return key
}

func ParseKeyVoteBoardSum(key []byte) (boardType string, boardIndex uint32, paymentAddress *privacy.PaymentAddress, err error) {
	length := []int{len(voteBoardSumPrefix), 3, 4, common.PaymentAddressLength}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return "", 0, nil, err
	}
	index := 1

	boardType = string(elements[iPlusPlus(&index)])
	boardIndex = common.BytesToUint32(elements[iPlusPlus(&index)])
	paymentAddress = privacy.NewPaymentAddressFromByte(elements[iPlusPlus(&index)])
	return boardType, boardIndex, paymentAddress, nil
}

func GetKeyVoteBoardCount(boardType string, boardIndex uint32, paymentAddress privacy.PaymentAddress) []byte {
	key := GetKeyFromVariadic(voteBoardCountPrefix, []byte(boardType), common.Uint32ToBytes(boardIndex), paymentAddress.Bytes())
	return key
}

func ParseKeyVoteBoardCount(key []byte) (boardType string, boardIndex uint32, candidatePubKey []byte, err error) {
	length := []int{len(voteBoardCountPrefix), 3, 4, common.PubKeyLength}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return "", 0, nil, err
	}
	index := 1

	boardType = string(elements[iPlusPlus(&index)])
	boardIndex = common.BytesToUint32(elements[iPlusPlus(&index)])
	candidatePubKey = elements[iPlusPlus(&index)]
	return boardType, boardIndex, candidatePubKey, nil
}

func GetKeyVoteBoardList(
	boardType string,
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
		[]byte(boardType),
		common.Uint32ToBytes(boardIndex),
		candidateBytes,
		voterBytes,
	)
	return key
}

func ParseKeyVoteBoardList(key []byte) (boardType string, boardIndex uint32, candidatePubKey []byte, voterPaymentAddress *privacy.PaymentAddress, err error) {
	length := []int{len(voteBoardListPrefix), 3, 4, common.PubKeyLength, common.PubKeyLength}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return "", 0, nil, nil, err
	}
	index := 1

	boardType = string(elements[iPlusPlus(&index)])
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

func GetKeyVoteTokenAmount(boardType string, boardIndex uint32, paymentAddress privacy.PaymentAddress) []byte {
	key := GetKeyFromVariadic(VoteTokenAmountPrefix, []byte(boardType), common.Uint32ToBytes(boardIndex), paymentAddress.Bytes())
	return key
}

func (db *db) GetVoteTokenAmount(boardType string, boardIndex uint32, paymentAddress privacy.PaymentAddress) (uint32, error) {
	key := GetKeyVoteTokenAmount(boardType, boardIndex, paymentAddress)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func (db *db) SetVoteTokenAmount(boardType string, boardIndex uint32, paymentAddress privacy.PaymentAddress, newAmount uint32) error {
	key := GetKeyVoteTokenAmount(boardType, boardIndex, paymentAddress)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := common.Uint32ToBytes(uint32(0))
		db.Put(key, zeroInBytes)
	}

	newAmountInBytes := common.Uint32ToBytes(newAmount)
	err = db.Put(key, newAmountInBytes)
	if err != nil {
		return err
	}
	return nil
}

func GetKeyThreePhraseCryptoOwner(boardType string, constitutionIndex uint32, txId *common.Hash) []byte {
	txIdByte := make([]byte, 0)
	if txId != nil {
		txIdByte = txId.GetBytes()
	}
	key := GetKeyFromVariadic(threePhraseCryptoOwnerPrefix, []byte(boardType), common.Uint32ToBytes(constitutionIndex), txIdByte)
	return key
}

func ParseKeyThreePhraseCryptoOwner(key []byte) (boardType string, constitutionIndex uint32, txId *common.Hash, err error) {
	length := []int{len(threePhraseCryptoOwnerPrefix), 3, 4, common.PubKeyLength}
	if CheckLength(key, length) {
		length[len(length)-1] = 0
	}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return "", 0, nil, err
	}
	index := 1

	boardType = string(elements[iPlusPlus(&index)])
	constitutionIndex = common.BytesToUint32(elements[iPlusPlus(&index)])

	txId = nil
	txIdData := elements[iPlusPlus(&index)]
	if len(txIdData) != 0 {
		newHash, err1 := common.NewHash(txIdData)
		if err1 != nil {
			return "", 0, nil, err1
		}
		txId = newHash
	}

	return boardType, constitutionIndex, txId, nil
}

func ParseValueThreePhraseCryptoOwner(value []byte) (uint32, error) {
	i := common.BytesToUint32(value)
	return i, nil
}

func GetKeyThreePhraseCryptoSealer(boardType string, constitutionIndex uint32, txId *common.Hash) []byte {
	txIdByte := make([]byte, 0)
	if txId != nil {
		txIdByte = txId.GetBytes()
	}
	key := GetKeyFromVariadic(threePhraseCryptoSealerPrefix, []byte(boardType), common.Uint32ToBytes(constitutionIndex), txIdByte)
	return key
}

func ParseKeyThreePhraseCryptoSealer(key []byte) (boardType string, constitutionIndex uint32, txId *common.Hash, err error) {
	length := []int{len(threePhraseCryptoSealerPrefix), 3, 4, common.PubKeyLength}
	if CheckLength(key, length) {
		length[len(length)-1] = 0
	}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return "", 0, nil, err
	}
	index := 1

	boardType = string(elements[iPlusPlus(&index)])
	constitutionIndex = common.BytesToUint32(elements[iPlusPlus(&index)])

	txId = nil
	txIdData := elements[iPlusPlus(&index)]
	if len(txIdData) != 0 {
		newHash, err1 := common.NewHash(txIdData)
		if err1 != nil {
			return "", 0, nil, err1
		}
		txId = newHash
	}
	return boardType, constitutionIndex, txId, nil
}

func GetKeyWinningVoter(boardType string, constitutionIndex uint32) []byte {
	key := GetKeyFromVariadic(winningVoterPrefix, []byte(boardType), common.Uint32ToBytes(constitutionIndex))
	return key
}

func GetKeyThreePhraseVoteValue(boardType string, constitutionIndex uint32, txId *common.Hash) []byte {
	txIdByte := make([]byte, 0)
	if txId != nil {
		txIdByte = txId.GetBytes()
	}
	key := GetKeyFromVariadic(threePhraseVoteValuePrefix, []byte(boardType), common.Uint32ToBytes(constitutionIndex), txIdByte)
	return key
}

func ParseKeyThreePhraseVoteValue(key []byte) (boardType string, constitutionIndex uint32, txId *common.Hash, err error) {
	length := []int{len(threePhraseVoteValuePrefix), 3, 4, common.PubKeyLength}
	if CheckLength(key, length) {
		length[len(length)-1] = 0
	}
	elements, err := ParseKeyToSlice(key, length)
	if err != nil {
		return "", 0, nil, err
	}
	index := 1

	boardType = string(elements[iPlusPlus(&index)])
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
	return boardType, constitutionIndex, txId, err
}

func GetKeyEncryptFlag(boardType string) []byte {
	key := GetKeyFromVariadic(encryptFlagPrefix, []byte(boardType))
	return key
}

func (db *db) GetEncryptFlag(boardType string) (uint32, error) {
	key := GetKeyEncryptFlag(boardType)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func (db *db) SetEncryptFlag(boardType string, flag uint32) {
	key := GetKeyEncryptFlag(boardType)
	value := common.Uint32ToBytes(flag)
	db.Put(key, value)
}

func GetKeyEncryptionLastBlockHeight(boardType string) []byte {
	key := GetKeyFromVariadic(encryptionLastBlockHeightPrefix, []byte(boardType))
	return key
}

func (db *db) GetEncryptionLastBlockHeight(boardType string) (uint32, error) {
	key := GetKeyEncryptionLastBlockHeight(boardType)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func (db *db) SetEncryptionLastBlockHeight(boardType string, height uint32) {
	key := GetKeyEncryptionLastBlockHeight(boardType)
	value := common.Uint32ToBytes(height)
	db.Put(key, value)
}

func (db *db) TakeVoteTokenFromWinner(boardType string, boardIndex uint32, voterPaymentAddress privacy.PaymentAddress, amountOfVote int32) error {
	key := GetKeyVoteTokenAmount(boardType, boardIndex, voterPaymentAddress)
	currentAmountInByte, err := db.Get(key)
	if err != nil {
		return err
	}
	currentAmount := common.BytesToUint32(currentAmountInByte)
	newAmount := currentAmount - uint32(amountOfVote)
	db.Put(key, common.Uint32ToBytes(newAmount))
	return nil
}

func (db *db) SetNewProposalWinningVoter(boardType string, constitutionIndex uint32, voterPaymentAddress privacy.PaymentAddress) error {
	key := GetKeyWinningVoter(boardType, constitutionIndex)
	db.Put(key, voterPaymentAddress.Bytes())
	return nil
}

func (db *db) GetBoardVoterList(boardType string, candidatePaymentAddress privacy.PaymentAddress, boardIndex uint32) []privacy.PaymentAddress {
	begin := GetKeyVoteBoardList(boardType, boardIndex, &candidatePaymentAddress, nil)
	end := GetKeyVoteBoardList(boardType, boardIndex, &candidatePaymentAddress, nil)
	end = common.BytesPlusOne(end)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := db.NewIterator(&searchRange, nil)
	listVoter := make([]privacy.PaymentAddress, 0)
	for iter.Next() {
		key := iter.Key()
		_, _, _, candidatePaymentAddress, _ := ParseKeyVoteBoardList(key)
		listVoter = append(listVoter, *candidatePaymentAddress)
	}
	return listVoter
}
