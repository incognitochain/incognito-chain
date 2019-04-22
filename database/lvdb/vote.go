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
	fmt.Println("[ndh] - [Add Vote Board] Enter add vote board", boardType, boardIndex)
	key := GetKeyVoteBoardSum(boardType, boardIndex, &CandidatePaymentAddress)

	currentVoteInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		currentVoteInBytes = make([]byte, 8)
		binary.LittleEndian.PutUint64(currentVoteInBytes, uint64(0))
	}
	fmt.Printf("[ndh] - [Add Vote Board] %+v - %+v\n", CandidatePaymentAddress, currentVoteInBytes)
	currentVote := binary.LittleEndian.Uint64(currentVoteInBytes)
	newVote := currentVote + amount

	newVoteInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newVoteInBytes, newVote)
	err = db.Put(key, newVoteInBytes)
	if err != nil {
		fmt.Println("[ndh] - [Add Vote Board] - Error1: ", err)
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
		fmt.Println("[ndh] - [Add Vote Board] - Error2: ", err)
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to list voter new voter base on count as index
	key = GetKeyVoteBoardList(boardType, boardIndex, &CandidatePaymentAddress, &VoterPaymentAddress)
	fmt.Printf("[ndh] Board Type: %+v, Board Index: %+v, key: %+v\n", boardType, boardIndex, key)
	amountInByte := GetValueVoteBoardList(amount)
	err = db.Put(key, amountInByte)
	if err != nil {
		fmt.Printf("[ndh] - - - Error when put key %+v\n", err)
	} else {
		vl, _ := db.Get(key)
		fmt.Printf("[ndh] - - - - - - - - Value %+v \n", vl)
	}

	return err
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

func (db *db) AddVoteProposalDB(boardType common.BoardType, constitutionIndex uint32, voterPayment []byte, proposalTxID []byte) error {
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

func (db *db) GetSubmitProposalDB(boardType common.BoardType, constitutionIndex uint32, proposalTxID []byte) ([]byte, error) {
	key := GetKeySubmitProposal(boardType, constitutionIndex, proposalTxID)
	value, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (db *db) AddSubmitProposalDB(boardType common.BoardType, constitutionIndex uint32, proposalTxID []byte, submitter []byte) error {
	key := GetKeySubmitProposal(boardType, constitutionIndex, proposalTxID) //privacy.NewPaymentAddressFromByte(submitter)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate proposal txid")
	}
	if err != nil {
		return err
	}
	err = db.Put(key, submitter)
	if err != nil {
		return err
	}

	return nil
}

func (db *db) AddListVoterOfProposalDB(boardType common.BoardType, constitutionIndex uint32, voterPayment []byte, proposalTxID []byte) error {
	key := GetKeyListVoterOfProposal(boardType, constitutionIndex, proposalTxID, privacy.NewPaymentAddressFromByte(voterPayment))
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("duplicate vote")
	}
	if err != nil {
		return err
	}
	err = db.Put(key, []byte{0})
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

func (db *db) SetNewProposalWinningVoter(boardType common.BoardType, constitutionIndex uint32, paymentAddresseses []privacy.PaymentAddress) error {
	key := GetKeyWinningVoter(boardType, constitutionIndex)
	fmt.Printf("[ndh] Set New Proposal WinningVoter: BoardType: %+v; constitutionIndex: %+v\n", boardType, constitutionIndex)
	value := concatListPaymentAddresses(paymentAddresseses)
	fmt.Printf("[ndh] Set New Proposal WinningVoter: Value: %+v\n", value)
	db.Put(key, value)
	return nil
}

func (db *db) GetCurrentProposalWinningVoter(boardType common.BoardType, constitutionIndex uint32) ([]privacy.PaymentAddress, error) {
	key := GetKeyWinningVoter(boardType, constitutionIndex)
	fmt.Printf("[ndh] Get New Proposal WinningVoter: BoardType: %+v; constitutionIndex: %+v\n", boardType, constitutionIndex)
	value, err := db.Get(key)
	fmt.Printf("[ndh] Get New Proposal WinningVoter: Value: %+v\n", value)
	if err != nil {
		return nil, err
	}
	res, err1 := decompListPaymentAddressesByte(value)
	return res, err1
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

func (db *db) AddBoardFundDB(boardType common.BoardType, constitutionIndex uint32, amountOfBoardFund uint64) error {
	fmt.Printf("[ndh]-[boardfund] - %+v %+v %+v \n", boardType, constitutionIndex, amountOfBoardFund)
	key := GetKeyBoardFund(boardType, constitutionIndex)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf("Can not update board fund")
	}
	err = db.Put(key, common.Uint64ToBytes(amountOfBoardFund))
	return err
}

func (db *db) GetBoardFundDB(boardType common.BoardType, constitutionIndex uint32) (uint64, error) {
	key := GetKeyBoardFund(boardType, constitutionIndex)
	ok, err := db.HasValue(key)
	if err != nil {
		return 0, err
	}
	if ok {
		byteTemp, _ := db.Get(key)
		return common.BytesToUint64(byteTemp), nil
	}
	return 0, errors.Errorf("Board Fund not found")
}

func (db *db) AddConstantsPriceDB(constitutionIndex uint32, price uint64) error {
	fmt.Printf("[ndh]-[constantprice] %+v %+v\n", constitutionIndex, price)
	key := GetKeyConstantsPrice(constitutionIndex)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if ok {
		valueBytes, _ := db.Get(key)
		value := common.BytesToUint64(valueBytes)
		value += price
		err = db.Put(key, common.Uint64ToBytes(value))
		return err
	}
	err = db.Put(key, common.Uint64ToBytes(price))
	return err
}

func (db *db) DeleteAnyProposalButThisDB(boardType common.BoardType, constitutionIndex uint32, proposalTxID []byte) error {
	begin := GetKeySubmitProposal(boardType, constitutionIndex, nil)
	end := GetKeySubmitProposal(boardType, constitutionIndex+1, nil)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchRange, nil)
	for iter.Next() {
		key := iter.Key()
		_, _, proposalTxIDTemp, err := ParseKeySubmitProposal(key)
		if err != nil {
			return err
		}
		if !common.ByteEqual(proposalTxID, proposalTxIDTemp) {
			err = db.Delete(key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *db) GetProposalSubmitterByConstitutionIndexDB(boardType common.BoardType, constitutionIndex uint32) ([]byte, error) {
	begin := GetKeySubmitProposal(boardType, constitutionIndex, nil)
	end := GetKeySubmitProposal(boardType, constitutionIndex+1, nil)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchRange, nil)
	for iter.Next() {
		return iter.Value(), nil
	}
	return nil, errors.Errorf("Proposal submitter not found")
}

func concatListPaymentAddresses(paymentAddresses []privacy.PaymentAddress) []byte {
	res := make([]byte, len(paymentAddresses)*common.PaymentAddressLength)
	i := 0
	for _, paymentAddress := range paymentAddresses {
		i += copy(res[i:], paymentAddress.Bytes())
	}
	return res
}

func decompListPaymentAddressesByte(paymentAddressesByte []byte) ([]privacy.PaymentAddress, error) {
	if len(paymentAddressesByte)%common.PaymentAddressLength != 0 {
		return nil, errors.New("Wrong payment address length")
	}
	res := make([]privacy.PaymentAddress, len(paymentAddressesByte)/common.PaymentAddressLength)
	for i, _ := range res {
		res[i].SetBytes(paymentAddressesByte[i*common.PaymentAddressLength : (i+1)*common.PaymentAddressLength])
	}
	return res, nil
}
