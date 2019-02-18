package lvdb

import (
	"encoding/binary"
	"sort"

	"github.com/ninjadotorg/constant/privacy"

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
	boardType byte,
	boardIndex uint32,
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

func GetNumberOfGovernor(boardType byte) int {
	numberOfGovernors := common.NumberOfDCBGovernors
	if boardType == common.GOVBoard {
		numberOfGovernors = common.NumberOfGOVGovernors
	}
	return numberOfGovernors
}

func (db *db) GetTopMostVoteGovernor(boardType byte, boardIndex uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	prefix := GetKeyVoteBoardSum(boardType, boardIndex, nil)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		_, _, paymentAddress, err := ParseKeyVoteBoardSum(iter.Key())
		countKey := GetKeyVoteBoardCount(boardType, boardIndex, *paymentAddress)
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

func (db *db) AddVoteLv3Proposal(boardType byte, constitutionIndex uint32, txID *common.Hash) error {
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

func (db *db) AddVoteLv1or2Proposal(boardType byte, constitutionIndex uint32, txID *common.Hash) error {
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

func (db *db) AddVoteNormalProposalFromSealer(boardType byte, constitutionIndex uint32, txID *common.Hash, voteValue []byte) error {
	err := db.AddVoteLv1or2Proposal(boardType, constitutionIndex, txID)
	if err != nil {
		return err
	}
	key := GetKeyThreePhraseVoteValue(boardType, constitutionIndex, txID)

	db.Put(key, voteValue)

	return nil
}

func (db *db) AddVoteNormalProposalFromOwner(boardType byte, constitutionIndex uint32, txID *common.Hash, voteValue []byte) error {
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

func (db *db) GetVoteTokenAmount(boardType byte, boardIndex uint32, paymentAddress privacy.PaymentAddress) (uint32, error) {
	key := GetKeyVoteTokenAmount(boardType, boardIndex, paymentAddress)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func (db *db) SetVoteTokenAmount(boardType byte, boardIndex uint32, paymentAddress privacy.PaymentAddress, newAmount uint32) error {
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

func (db *db) GetEncryptFlag(boardType byte) (uint32, error) {
	key := GetKeyEncryptFlag(boardType)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func (db *db) SetEncryptFlag(boardType byte, flag uint32) {
	key := GetKeyEncryptFlag(boardType)
	value := common.Uint32ToBytes(flag)
	db.Put(key, value)
}

func (db *db) GetEncryptionLastBlockHeight(boardType byte) (uint32, error) {
	key := GetKeyEncryptionLastBlockHeight(boardType)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint32(value), nil
}

func (db *db) SetEncryptionLastBlockHeight(boardType byte, height uint32) {
	key := GetKeyEncryptionLastBlockHeight(boardType)
	value := common.Uint32ToBytes(height)
	db.Put(key, value)
}

func (db *db) TakeVoteTokenFromWinner(boardType byte, boardIndex uint32, voterPaymentAddress privacy.PaymentAddress, amountOfVote int32) error {
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

func (db *db) SetNewProposalWinningVoter(boardType byte, constitutionIndex uint32, voterPaymentAddress privacy.PaymentAddress) error {
	key := GetKeyWinningVoter(boardType, constitutionIndex)
	db.Put(key, voterPaymentAddress.Bytes())
	return nil
}

func (db *db) GetBoardVoterList(boardType byte, candidatePaymentAddress privacy.PaymentAddress, boardIndex uint32) []privacy.PaymentAddress {
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
