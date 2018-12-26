package lvdb

import (
	"encoding/binary"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/voting"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) AddVoteDCBBoard(StartedBlockInt uint32, VoterPubKey []byte, CandidatePubKey []byte, amount uint64) error {
	StartedBlock := uint32(StartedBlockInt)
	//add to sum amount of vote token to this candidate
	key := GetVoteDCBBoardSumKey(StartedBlock, CandidatePubKey)
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
	key = GetVoteDCBBoardCountKey(StartedBlock, CandidatePubKey)
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
	key = GetVoteDCBBoardListKey(currentCount, StartedBlock, CandidatePubKey)
	amountInByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountInByte, amount)
	valueInByte := append([]byte(VoterPubKey), amountInByte...)
	err = db.Put(key, valueInByte)

	return nil
}

func (db *db) AddVoteGOVBoard(StartedBlockInt uint32, VoterPubKey []byte, CandidatePubKey []byte, amount uint64) error {
	StartedBlock := uint32(StartedBlockInt)
	//add to sum amount of vote token to this candidate
	key := GetVoteGOVBoardSumKey(StartedBlock, CandidatePubKey)
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
	key = GetVoteGOVBoardCountKey(StartedBlock, CandidatePubKey)
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
	key = GetVoteGOVBoardListKey(currentCount, StartedBlock, CandidatePubKey)
	amountInByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountInByte, amount)
	valueInByte := append([]byte(VoterPubKey), amountInByte...)
	err = db.Put(key, valueInByte)

	return nil
}

func (db *db) GetTopMostVoteDCBGovernor(StartedBlock uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	prefix := GetVoteDCBBoardSumKey(StartedBlock, make([]byte, 0))
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		_, pubKey, err := ParseKeyVoteDCBBoardSum(iter.Key())
		countKey := GetVoteDCBBoardCountKey(StartedBlock, pubKey)
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

func (db *db) GetTopMostVoteGOVGovernor(StartedBlock uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	prefix := GetVoteGOVBoardSumKey(StartedBlock, make([]byte, 0))
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		_, pubKey, err := ParseKeyVoteGOVBoardSum(iter.Key())
		countKey := GetVoteGOVBoardCountKey(StartedBlock, pubKey)
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

func (db *db) GetVoteDCBBoardListPrefix() []byte {
	return VoteDCBBoardListPrefix
}

func (db *db) GetVoteGOVBoardListPrefix() []byte {
	return VoteGOVBoardListPrefix
}

func (db *db) GetThreePhraseSealerPrefix() []byte {
	return threePhraseCryptoSealerPrefix
}

func (db *db) GetThreePhraseOwnerPrefix() []byte {
	return threePhraseCryptoOwnerPrefix
}

func (db *db) GetThreePhraseVoteValuePrefix() []byte {
	return threePhraseVoteValuePrefix
}

func (db *db) AddVoteLv3Proposal(boardType string, startedBlock uint32, txID *common.Hash) error {
	//init sealer
	keySealer := GetThreePhraseCryptoSealerKey(boardType, startedBlock, txID)
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
	keyOwner := GetThreePhraseCryptoOwnerKey(boardType, startedBlock, txID)
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

func (db *db) AddVoteLv1or2Proposal(boardType string, startedBlock uint32, txID *common.Hash) error {
	keySealer := GetThreePhraseCryptoSealerKey(boardType, startedBlock, txID)
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

func (db *db) AddVoteNormalProposalFromSealer(boardType string, startedBlock uint32, txID *common.Hash, voteValue []byte) error {
	err := db.AddVoteLv1or2Proposal(boardType, startedBlock, txID)
	if err != nil {
		return err
	}
	key := GetThreePhraseVoteValueKey(boardType, startedBlock, txID)

	db.Put(key, voteValue)

	return nil
}

func (db *db) AddVoteNormalProposalFromOwner(boardType string, startedBlock uint32, txID *common.Hash, voteValue []byte) error {
	keyOwner := GetThreePhraseCryptoOwnerKey(boardType, startedBlock, txID)
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

	key := GetThreePhraseVoteValueKey(boardType, startedBlock, txID)
	db.Put(key, voteValue)

	return nil
}

func GetVoteDCBBoardSumKey(startedBlock uint32, candidatePubKey []byte) []byte {
	key := append(voteDCBBoardSumPrefix, common.Uint32ToBytes(startedBlock)...)
	key = append(key, candidatePubKey...)
	return key
}

func ParseKeyVoteDCBBoardSum(key []byte) (uint32, []byte, error) {
	realKey := key[len(voteDCBBoardSumPrefix):]
	startedBlock := common.BytesToUint32(realKey[:4])
	candidatePubKey := realKey[4:]
	return startedBlock, candidatePubKey, nil
}

func GetVoteDCBBoardCountKey(startedBlock uint32, candidatePubKey []byte) []byte {
	key := append(voteDCBBoardCountPrefix, common.Uint32ToBytes(startedBlock)...)
	key = append(key, candidatePubKey...)
	return key
}

func ParseKeyVoteGOVBoardSum(key []byte) (uint32, []byte, error) {
	realKey := key[len(voteGOVBoardSumPrefix):]
	startedBlock := common.BytesToUint32(realKey[:4])
	candidatePubKey := realKey[4:]
	return startedBlock, candidatePubKey, nil
}

func GetVoteDCBBoardListKey(currentCount uint32, startedBlock uint32, candidatePubKey []byte) []byte {
	key := append(VoteDCBBoardListPrefix, common.Uint32ToBytes(currentCount)...)
	key = append(key, common.Uint32ToBytes(startedBlock)...)
	key = append(key, candidatePubKey...)
	return key
}

func ParseKeyVoteDCBBoardList(key []byte) (uint32, []byte, uint32, error) {
	realKey := key[len(VoteDCBBoardListPrefix):]
	startedBlock := common.BytesToUint32(realKey[:4])
	pubKey := realKey[4 : 4+common.PubKeyLength]
	currentIndex := common.BytesToUint32(realKey[4+common.PubKeyLength:])
	return startedBlock, pubKey, currentIndex, nil
}

func GetDCBVoteTokenAmountKey(startedBlock uint32, pubKey []byte) []byte {
	key := append(DCBVoteTokenAmountPrefix, common.Uint32ToBytes(startedBlock)...)
	key = append(key, pubKey...)
	return key
}

func (db *db) GetDCBVoteTokenAmount(startedBlock uint32, pubKey []byte) uint32 {
	key := GetDCBVoteTokenAmountKey(startedBlock, pubKey)
	value, _ := db.Get(key)
	return common.BytesToUint32(value)
}
func (db *db) GetGOVVoteTokenAmount(startedBlock uint32, pubKey []byte) uint32 {
	key := GetGOVVoteTokenAmountKey(startedBlock, pubKey)
	value, _ := db.Get(key)
	return common.BytesToUint32(value)
}

func GetVoteGOVBoardSumKey(startedBlock uint32, candidatePubKey []byte) []byte {
	key := append(voteGOVBoardSumPrefix, common.Uint32ToBytes(startedBlock)...)
	key = append(key, candidatePubKey...)
	return key
}

func GetVoteGOVBoardCountKey(startedBlock uint32, candidatePubKey []byte) []byte {
	key := append(voteGOVBoardCountPrefix, common.Uint32ToBytes(startedBlock)...)
	key = append(key, candidatePubKey...)
	return key
}

func GetVoteGOVBoardListKey(currentCount uint32, startedBlock uint32, candidatePubKey []byte) []byte {
	key := append(VoteGOVBoardListPrefix, common.Uint32ToBytes(currentCount)...)
	key = append(key, common.Uint32ToBytes(startedBlock)...)
	key = append(key, candidatePubKey...)
	return key
}

func ParseKeyVoteGOVBoardList(key []byte) (uint32, []byte, uint32, error) {
	realKey := key[len(VoteGOVBoardListPrefix):]
	startedBlock := common.BytesToUint32(realKey[:4])
	pubKey := realKey[4 : 4+common.PubKeyLength]
	currentIndex := common.BytesToUint32(realKey[4+common.PubKeyLength:])
	return startedBlock, pubKey, currentIndex, nil
}

func GetGOVVoteTokenAmountKey(startedBlock uint32, pubKey []byte) []byte {
	key := append(GOVVoteTokenAmountPrefix, common.Uint32ToBytes(startedBlock)...)
	key = append(key, pubKey...)
	return key
}

func GetThreePhraseCryptoOwnerKey(boardType string, startedBlock uint32, txId *common.Hash) []byte {
	key := append(threePhraseCryptoOwnerPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(startedBlock)...)
	if txId != nil {
		key = append(key, txId.GetBytes()...)
	}
	return key
}

func ParseKeyThreePhraseCryptoOwner(key []byte) (string, uint32, *common.Hash, error) {
	realKey := key[len(threePhraseCryptoOwnerPrefix):]
	boardType := realKey[:3]
	startedBlock := common.BytesToUint32(realKey[3 : 3+4])
	hash := common.NewHash([]byte(realKey[3+4:]))
	return string(boardType), uint32(startedBlock), &hash, nil
}

func ParseValueThreePhraseCryptoOwner(value []byte) (uint32, error) {
	i := common.BytesToUint32(value)
	return i, nil
}

func GetThreePhraseCryptoSealerKey(boardType string, startedBlock uint32, txId *common.Hash) []byte {
	key := append(threePhraseCryptoSealerPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(startedBlock)...)
	if txId != nil {
		key = append(key, txId.GetBytes()...)
	}
	return key
}

func ParseKeyThreePhraseCryptoSealer(key []byte) (string, uint32, *common.Hash, error) {
	realKey := key[len(threePhraseCryptoSealerPrefix):]
	boardType := realKey[:3]
	startedBlock := common.BytesToUint32(realKey[3 : 3+4])
	hash := common.NewHash([]byte(realKey[3+4:]))
	return string(boardType), uint32(startedBlock), &hash, nil
}

func GetThreePhraseVoteValueKey(boardType string, startedBlock uint32, txId *common.Hash) []byte {
	key := append(threePhraseCryptoSealerPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(startedBlock)...)
	if txId != nil {
		key = append(key, txId.GetBytes()...)
	}
	return key
}

func GetWinningVoterKey(boardType string, startedBlock uint32) []byte {
	key := append(winningVoterPrefix, []byte(boardType)...)
	key = append(key, common.Uint32ToBytes(startedBlock)...)
	return key
}

func ParseKeyThreePhraseVoteValue(key []byte) (string, uint32, *common.Hash, error) {
	realKey := key[len(threePhraseVoteValuePrefix):]
	boardType := realKey[:3]
	startedBlock := common.BytesToUint32(realKey[3 : 3+4])
	hash := common.NewHash([]byte(realKey[3+4:]))
	return string(boardType), uint32(startedBlock), &hash, nil
}

func ParseValueThreePhraseVoteValue(value []byte) (*common.Hash, int32, error) {
	txId := common.NewHash(value[:common.HashSize])
	amount := common.BytesToInt32(value[common.HashSize:])
	return &txId, amount, nil
}

func (db *db) GetAmountVoteToken(boardType string, startedBlock uint32, pubKey []byte) (uint32, error) {
	key := make([]byte, 0)
	if boardType == "dcb" {
		key = GetDCBVoteTokenAmountKey(startedBlock, pubKey)
	} else if boardType == "gov" {
		key = GetGOVVoteTokenAmountKey(startedBlock, pubKey)
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

func (db *db) TakeVoteTokenFromWinner(boardType string, startedBlock uint32, voter voting.Voter) error {
	key := make([]byte, 0)
	if boardType == "dcb" {
		key = GetDCBVoteTokenAmountKey(startedBlock, voter.PubKey)
	} else if boardType == "gov" {
		key = GetGOVVoteTokenAmountKey(startedBlock, voter.PubKey)
	} else {
		return errors.New("wrong board type")
	}
	currentAmountInByte, err := db.Get(key)
	if err != nil {
		return err
	}
	currentAmount := common.BytesToUint32(currentAmountInByte)
	newAmount := currentAmount - uint32(voter.AmountOfVote)
	db.Put(key, common.Uint32ToBytes(newAmount))
	return nil
}

func (db *db) SetNewWinningVoter(boardType string, startedBlock uint32, voterPubKey []byte) error {
	key := make([]byte, 0)
	if boardType == "dcb" {
		key = GetWinningVoterKey(boardType, startedBlock)
	} else if boardType == "gov" {
		key = GetWinningVoterKey(boardType, startedBlock)
	} else {
		return errors.New("wrong board type")
	}
	db.Put(key, voterPubKey)
	return nil
}
