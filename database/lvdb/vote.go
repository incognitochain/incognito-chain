package lvdb

import (
	"encoding/binary"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) AddVoteDCBBoard(boardIndex uint32, paymentAddress []byte, VoterPubKey []byte, CandidatePubKey []byte, amount uint64) error {
	//add to sum amount of vote token to this candidate
	key := GetKeyVoteDCBBoardSum(boardIndex, CandidatePubKey)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(zeroInBytes, uint64(0))
		db.Put(key, zeroInBytes)
	}

	currentVoteInBytes, err := db.lvdb.Get(key, nil)
	currentVote := binary.LittleEndian.Uint64(currentVoteInBytes)
	newVote := currentVote + amount

	newVoteInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newVoteInBytes, newVote)
	err = db.Put(key, newVoteInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to count amount of vote to this candidate
	key = GetKeyVoteDCBBoardCount(boardIndex, CandidatePubKey)
	currentCountInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		return err
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
	key = GetKeyVoteDCBBoardList(boardIndex, CandidatePubKey, VoterPubKey)
	oldAmountInByte, _ := db.Get(key)
	oldAmount := ParseValueVoteDCBBoardList(oldAmountInByte)
	newAmount := oldAmount + amount
	newAmountInByte := GetValueVoteDCBBoardList(newAmount)
	err = db.Put(key, newAmountInByte)

	//add database to get paymentAddress from pubKey
	key = GetPubKeyToPaymentAddressKey(VoterPubKey)
	db.Put(key, paymentAddress)

	return nil
}

func (db *db) AddVoteGOVBoard(boardIndex uint32, paymentAddress []byte, VoterPubKey []byte, CandidatePubKey []byte, amount uint64) error {
	//add to sum amount of vote token to this candidate
	key := GetKeyVoteGOVBoardSum(boardIndex, CandidatePubKey)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(zeroInBytes, uint64(0))
		db.Put(key, zeroInBytes)
	}

	currentVoteInBytes, err := db.lvdb.Get(key, nil)
	currentVote := binary.LittleEndian.Uint64(currentVoteInBytes)
	newVote := currentVote + amount

	newVoteInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newVoteInBytes, newVote)
	err = db.Put(key, newVoteInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to count amount of vote to this candidate
	key = GetKeyVoteGOVBoardCount(boardIndex, CandidatePubKey)
	currentCountInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		return err
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
	key = GetKeyVoteGOVBoardList(boardIndex, CandidatePubKey, VoterPubKey)
	oldAmountInByte, _ := db.Get(key)
	oldAmount := ParseValueVoteGOVBoardList(oldAmountInByte)
	newAmount := oldAmount + amount
	newAmountInByte := GetValueVoteGOVBoardList(newAmount)
	err = db.Put(key, newAmountInByte)

	//add database to get paymentAddress from pubKey
	key = GetPubKeyToPaymentAddressKey(VoterPubKey)
	db.Put(key, paymentAddress)

	return nil
}

func (db *db) GetTopMostVoteDCBGovernor(currentBoardIndex uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	newBoardIndex := currentBoardIndex + 1
	prefix := GetKeyVoteDCBBoardSum(newBoardIndex, make([]byte, 0))
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		_, pubKey, err := ParseKeyVoteDCBBoardSum(iter.Key())
		countKey := GetKeyVoteDCBBoardCount(newBoardIndex, pubKey)
		if err != nil {
			return nil, err
		}
		countValue, err := db.Get(countKey)
		if err != nil {
			return nil, err
		}
		value := binary.LittleEndian.Uint64(iter.Value())
		candidateList = append(candidateList, database.CandidateElement{
			PubKey:       pubKey,
			VoteAmount:   value,
			NumberOfVote: common.BytesToUint32(countValue),
		})
	}
	sort.Sort(candidateList)
	if len(candidateList) < common.NumberOfDCBGovernors {
		return nil, database.NewDatabaseError(database.NotEnoughCandidateDCB, errors.Errorf("not enough DCB Candidate"))
	}

	return candidateList[len(candidateList)-common.NumberOfDCBGovernors:], nil
}

func (db *db) GetTopMostVoteGOVGovernor(currentBoardIndex uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	newBoardIndex := currentBoardIndex + 1
	prefix := GetKeyVoteGOVBoardSum(newBoardIndex, make([]byte, 0))
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		_, pubKey, err := ParseKeyVoteGOVBoardSum(iter.Key())
		countKey := GetKeyVoteGOVBoardCount(newBoardIndex, pubKey)
		if err != nil {
			return nil, err
		}
		countValue, err := db.Get(countKey)
		if err != nil {
			return nil, err
		}
		value := binary.LittleEndian.Uint64(iter.Value())
		candidateList = append(candidateList, database.CandidateElement{
			PubKey:       pubKey,
			VoteAmount:   value,
			NumberOfVote: common.BytesToUint32(countValue),
		})
	}
	sort.Sort(candidateList)
	if len(candidateList) < common.NumberOfGOVGovernors {
		return nil, database.NewDatabaseError(database.NotEnoughCandidateGOV, errors.Errorf("not enough GOV Candidate"))
	}

	return candidateList[len(candidateList)-common.NumberOfGOVGovernors:], nil
}

func (db *db) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return db.lvdb.NewIterator(slice, ro)
}

func (db *db) AddVoteLv3Proposal(boardType string, constitutionIndex uint32, txID *common.Hash) error {
	//init sealer
	keySealer := GetThreePhraseCryptoSealerKey(boardType, constitutionIndex, txID)
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
	keyOwner := GetThreePhraseCryptoOwnerKey(boardType, constitutionIndex, txID)
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
	keySealer := GetThreePhraseCryptoSealerKey(boardType, constitutionIndex, txID)
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
	key := GetThreePhraseVoteValueKey(boardType, constitutionIndex, txID)

	db.Put(key, voteValue)

	return nil
}

func (db *db) AddVoteNormalProposalFromOwner(boardType string, constitutionIndex uint32, txID *common.Hash, voteValue []byte) error {
	keyOwner := GetThreePhraseCryptoOwnerKey(boardType, constitutionIndex, txID)
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

	key := GetThreePhraseVoteValueKey(boardType, constitutionIndex, txID)
	db.Put(key, voteValue)

	return nil
}

func GetPubKeyToPaymentAddressKey(pubKey []byte) []byte {
	key := append(pubKeyToPaymentAddress, pubKey...)
	return key
}

func GetKeyVoteDCBBoardSum(boardIndex uint32, candidatePubKey []byte) []byte {
	key := append(voteDCBBoardSumPrefix, common.Uint32ToBytes(boardIndex)...)
	key = append(key, candidatePubKey...)
	return key
}

func ParseKeyVoteDCBBoardSum(key []byte) (uint32, []byte, error) {
	realKey := key[len(voteDCBBoardSumPrefix):]
	boardIndex := common.BytesToUint32(realKey[:4])
	candidatePubKey := realKey[4:]
	return boardIndex, candidatePubKey, nil
}

func GetKeyVoteDCBBoardCount(boardIndex uint32, candidatePubKey []byte) []byte {
	key := append(voteDCBBoardCountPrefix, common.Uint32ToBytes(boardIndex)...)
	key = append(key, candidatePubKey...)
	return key
}

func ParseKeyVoteGOVBoardSum(key []byte) (uint32, []byte, error) {
	realKey := key[len(voteGOVBoardSumPrefix):]
	boardIndex := common.BytesToUint32(realKey[:4])
	candidatePubKey := realKey[4:]
	return boardIndex, candidatePubKey, nil
}

func GetPosFromLength(length []int) []int {
	pos := []int{0}
	for i := 0; i < len(length); i++ {
		pos = append(pos, pos[i]+length[i])
	}
	return pos
}

func GetKeyVoteDCBBoardList(boardIndex uint32, candidatePubKey []byte, voterPubKey []byte) []byte {
	length := []int{len(VoteDCBBoardListPrefix), 4, common.PubKeyLength, common.PubKeyLength}
	pos := GetPosFromLength(length)

	key := make([]byte, pos[len(pos)-1])
	copy(key[pos[0]:pos[1]], VoteDCBBoardListPrefix)
	copy(key[pos[1]:pos[2]], common.Uint32ToBytes(boardIndex))
	copy(key[pos[2]:pos[3]], candidatePubKey)
	copy(key[pos[3]:pos[4]], voterPubKey)
	return key
}

func ParseKeyVoteDCBBoardList(key []byte) (uint32, []byte, []byte, error) {
	length := []int{len(VoteDCBBoardListPrefix), 4, common.PubKeyLength, common.PubKeyLength}
	pos := GetPosFromLength(length)

	_ = key[pos[0]:pos[1]]
	boardIndex := common.BytesToUint32(key[pos[1]:pos[2]])
	candidatePubKey := key[pos[2]:pos[3]]
	voterPubKey := key[pos[3]:pos[4]]
	return boardIndex, candidatePubKey, voterPubKey, nil
}

func GetValueVoteDCBBoardList(amount uint64) []byte {
	return common.Uint64ToBytes(amount)
}

func ParseValueVoteDCBBoardList(value []byte) uint64 {
	return common.BytesToUint64(value)
}

func GetKeyDCBVoteTokenAmount(boardIndex uint32, pubKey []byte) []byte {
	key := append(DCBVoteTokenAmountPrefix, common.Uint32ToBytes(boardIndex)...)
	key = append(key, pubKey...)
	return key
}

func (db *db) GetDCBVoteTokenAmount(boardIndex uint32, pubKey []byte) (uint32, error) {
	key := GetKeyDCBVoteTokenAmount(boardIndex, pubKey)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}
func (db *db) GetGOVVoteTokenAmount(boardIndex uint32, pubKey []byte) (uint32, error) {
	key := GetGOVVoteTokenAmountKey(boardIndex, pubKey)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func GetKeyVoteGOVBoardSum(boardIndex uint32, candidatePubKey []byte) []byte {
	key := append(voteGOVBoardSumPrefix, common.Uint32ToBytes(boardIndex)...)
	key = append(key, candidatePubKey...)
	return key
}

func GetKeyVoteGOVBoardCount(boardIndex uint32, candidatePubKey []byte) []byte {
	key := append(voteGOVBoardCountPrefix, common.Uint32ToBytes(boardIndex)...)
	key = append(key, candidatePubKey...)
	return key
}

func GetKeyVoteGOVBoardList(boardIndex uint32, candidatePubKey []byte, voterPubKey []byte) []byte {
	length := []int{len(VoteGOVBoardListPrefix), 4, common.PubKeyLength, common.PubKeyLength}
	pos := GetPosFromLength(length)

	key := make([]byte, pos[len(pos)-1])
	copy(key[pos[0]:pos[1]], VoteGOVBoardListPrefix)
	copy(key[pos[1]:pos[2]], common.Uint32ToBytes(boardIndex))
	copy(key[pos[2]:pos[3]], candidatePubKey)
	copy(key[pos[3]:pos[4]], voterPubKey)
	return key
}

func ParseKeyVoteGOVBoardList(key []byte) (uint32, []byte, []byte, error) {
	length := []int{len(VoteGOVBoardListPrefix), 4, common.PubKeyLength, common.PubKeyLength}
	pos := GetPosFromLength(length)

	_ = key[pos[0]:pos[1]]
	boardIndex := common.BytesToUint32(key[pos[1]:pos[2]])
	candidatePubKey := key[pos[2]:pos[3]]
	voterPubKey := key[pos[3]:pos[4]]
	return boardIndex, candidatePubKey, voterPubKey, nil
}

func GetValueVoteGOVBoardList(amount uint64) []byte {
	return common.Uint64ToBytes(amount)
}

func ParseValueVoteGOVBoardList(value []byte) uint64 {
	return common.BytesToUint64(value)
}

func GetGOVVoteTokenAmountKey(boardIndex uint32, pubKey []byte) []byte {
	key := append(GOVVoteTokenAmountPrefix, common.Uint32ToBytes(boardIndex)...)
	key = append(key, pubKey...)
	return key
}

func GetThreePhraseCryptoOwnerKey(boardType string, constitutionIndex uint32, txId *common.Hash) []byte {
	key := append(threePhraseCryptoOwnerPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(constitutionIndex)...)
	if txId != nil {
		key = append(key, txId.GetBytes()...)
	}
	return key
}

func ParseKeyThreePhraseCryptoOwner(key []byte) (string, uint32, *common.Hash, error) {
	realKey := key[len(threePhraseCryptoOwnerPrefix):]
	boardType := realKey[:3]
	constitutionIndex := common.BytesToUint32(realKey[3 : 3+4])
	hash := common.NewHash([]byte(realKey[3+4:]))
	return string(boardType), uint32(constitutionIndex), &hash, nil
}

func ParseValueThreePhraseCryptoOwner(value []byte) (uint32, error) {
	i := common.BytesToUint32(value)
	return i, nil
}

func GetThreePhraseCryptoSealerKey(boardType string, constitutionIndex uint32, txId *common.Hash) []byte {
	key := append(threePhraseCryptoSealerPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(constitutionIndex)...)
	if txId != nil {
		key = append(key, txId.GetBytes()...)
	}
	return key
}

func ParseKeyThreePhraseCryptoSealer(key []byte) (string, uint32, *common.Hash, error) {
	realKey := key[len(threePhraseCryptoSealerPrefix):]
	boardType := realKey[:3]
	constitutionIndex := common.BytesToUint32(realKey[3 : 3+4])
	hash := common.NewHash([]byte(realKey[3+4:]))
	return string(boardType), uint32(constitutionIndex), &hash, nil
}

func GetThreePhraseVoteValueKey(boardType string, constitutionIndex uint32, txId *common.Hash) []byte {
	key := append(threePhraseCryptoSealerPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(constitutionIndex)...)
	if txId != nil {
		key = append(key, txId.GetBytes()...)
	}
	return key
}

func GetKeyWinningVoter(boardType string, constitutionIndex uint32) []byte {
	key := append(winningVoterPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(constitutionIndex)...)
	return key
}

func ParseKeyThreePhraseVoteValue(key []byte) (string, uint32, *common.Hash, error) {
	realKey := key[len(threePhraseVoteValuePrefix):]
	boardType := realKey[:3]
	constitutionIndex := common.BytesToUint32(realKey[3 : 3+4])
	hash := common.NewHash([]byte(realKey[3+4:]))
	return string(boardType), uint32(constitutionIndex), &hash, nil
}

func ParseValueThreePhraseVoteValue(value []byte) (*common.Hash, int32, error) {
	txId := common.NewHash(value[:common.HashSize])
	amount := common.BytesToInt32(value[common.HashSize:])
	return &txId, amount, nil
}

func (db *db) GetAmountVoteToken(boardType string, boardIndex uint32, pubKey []byte) (uint32, error) {
	key := make([]byte, 0)
	if boardType == "dcb" {
		key = GetKeyDCBVoteTokenAmount(boardIndex, pubKey)
	} else if boardType == "gov" {
		key = GetGOVVoteTokenAmountKey(boardIndex, pubKey)
	} else {
		return 0, errors.New("wrong board type")
	}
	currentAmountInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		return 0, err
	}
	currentAmount := common.BytesToUint32(currentAmountInBytes)
	return currentAmount, nil
}

func (db *db) TakeVoteTokenFromWinner(boardType string, boardIndex uint32, voterPubKey []byte, amountOfVote int32) error {
	key := make([]byte, 0)
	if boardType == "dcb" {
		key = GetKeyDCBVoteTokenAmount(boardIndex, voterPubKey)
	} else if boardType == "gov" {
		key = GetGOVVoteTokenAmountKey(boardIndex, voterPubKey)
	} else {
		return errors.New("wrong board type")
	}
	currentAmountInByte, err := db.Get(key)
	if err != nil {
		return err
	}
	currentAmount := common.BytesToUint32(currentAmountInByte)
	newAmount := currentAmount - uint32(amountOfVote)
	db.Put(key, common.Uint32ToBytes(newAmount))
	return nil
}

func (db *db) SetNewProposalWinningVoter(boardType string, constitutionIndex uint32, voterPubKey []byte) error {
	key := make([]byte, 0)
	if boardType == "dcb" {
		key = GetKeyWinningVoter(boardType, constitutionIndex)
	} else if boardType == "gov" {
		key = GetKeyWinningVoter(boardType, constitutionIndex)
	} else {
		return errors.New("wrong board type")
	}
	db.Put(key, voterPubKey)
	return nil
}

func (db *db) GetPaymentAddressFromPubKey(pubKey []byte) []byte {
	key := GetPubKeyToPaymentAddressKey(pubKey)
	value, _ := db.Get(key)
	return value
}

func (db *db) GetBoardVoterList(candidatePubKey []byte, boardIndex uint32) [][]byte {
	begin := GetKeyVoteDCBBoardList(boardIndex, candidatePubKey, make([]byte, common.PubKeyLength))
	end := GetKeyVoteDCBBoardList(boardIndex, common.BytesPlusOne(candidatePubKey), make([]byte, common.PubKeyLength))
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := db.NewIterator(&searchRange, nil)
	listVoter := make([][]byte, 0)
	for iter.Next() {
		key := iter.Key()
		_, _, pubKey, _ := ParseKeyVoteDCBBoardList(key)
		listVoter = append(listVoter, pubKey)
	}
	return listVoter
}
