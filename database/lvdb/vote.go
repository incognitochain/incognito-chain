package lvdb

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
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
	boardType common.BoardType,
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

// GetNumberOfGovernor remove-soon
func GetNumberOfGovernor(boardType common.BoardType) int {
	numberOfGovernors := common.NumberOfDCBGovernors
	if boardType == common.GOVBoard {
		numberOfGovernors = common.NumberOfGOVGovernors
	}
	return numberOfGovernors
}

// GetNumberOfGovernorRange return
func GetNumberOfGovernorRange(boardType common.BoardType) (int, int) {
	if boardType == common.GOVBoard {
		return common.GOVGovernorsLowerBound, common.GOVGovernorsUpperBound
	}
	return common.DCBGovernorsLowerBound, common.DCBGovernorsUpperBound
}

func (db *db) GetTopMostVoteGovernor(boardType common.BoardType, boardIndex uint32) (database.CandidateList, error) {
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
	fmt.Println("\n\n\n\n\n\n\n\n\n")
	fmt.Println(candidateList.Len())
	for _, candidateElement := range candidateList {
		fmt.Println(candidateElement)
	}
	fmt.Println("\n\n\n\n\n\n\n\n\n")
	lenCandidateList := len(candidateList)
	lowerBound, upperBound := GetNumberOfGovernorRange(boardType)
	if lowerBound > lenCandidateList {
		return nil, database.NewDatabaseError(database.NotEnoughCandidate, errors.Errorf("not enough Candidate"))
	}
	if lenCandidateList > upperBound {
		return candidateList[lenCandidateList-upperBound:], nil
	}
	return candidateList, nil
}

func (db *db) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return db.lvdb.NewIterator(slice, ro)
}

func (db *db) AddVoteNormalProposalDB(boardType common.BoardType, constitutionIndex uint32, voterPayment []byte, proposalTxID []byte) error {
	key := GetKeyVoteProposal(boardType, constitutionIndex, privacy.NewPaymentAddressFromByte(voterPayment))
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate txid")
	}
	if err != nil {
		return err
	}
	err = db.Put(key, proposalTxID)
	if err != nil {
		return err
	}

	return nil
}

func (db *db) GetEncryptFlag(boardType common.BoardType) (byte, error) {
	key := GetKeyEncryptFlag(boardType)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	if len(value) != 1 {
		return 0, errors.New("wrong flag format")
	}
	return value[0], nil
}

func (db *db) SetEncryptFlag(boardType common.BoardType, flag byte) {
	key := GetKeyEncryptFlag(boardType)
	value := common.ByteToBytes(flag)
	db.Put(key, value)
}

func (db *db) GetEncryptionLastBlockHeight(boardType common.BoardType) (uint64, error) {
	key := GetKeyEncryptionLastBlockHeight(boardType)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return common.BytesToUint64(value), nil
}

func (db *db) SetEncryptionLastBlockHeight(boardType common.BoardType, height uint64) {
	key := GetKeyEncryptionLastBlockHeight(boardType)
	value := common.Uint64ToBytes(height)
	db.Put(key, value)
}

func (db *db) SetNewProposalWinningVoter(boardType common.BoardType, constitutionIndex uint32, voterPaymentAddress privacy.PaymentAddress) error {
	key := GetKeyWinningVoter(boardType, constitutionIndex)
	db.Put(key, voterPaymentAddress.Bytes())
	return nil
}

func (db *db) GetBoardVoterList(boardType common.BoardType, candidatePaymentAddress privacy.PaymentAddress, boardIndex uint32) []privacy.PaymentAddress {
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

func (db *db) GetListSupporters(boardType common.BoardType, candidateAddress privacy.PaymentAddress) ([]*privacy.PaymentAddress, error) {
	// todo @jackalope
	return nil, nil
}
