package lvdb

import (
	"encoding/binary"
	"sort"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) AddVoteDCBBoard(StartedBlockInt uint32, VoterPubKey []byte, CandidatePubKey []byte, amount uint64) error {
	StartedBlock := uint32(StartedBlockInt)
	//add to sum amount of vote token to this candidate
	key := db.GetKey(string(voteDCBBoardSumPrefix), string(StartedBlock)+string(CandidatePubKey))
	ok, err := db.hasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(zeroInBytes, uint64(0))
		db.put(key, zeroInBytes)
	}

	currentVoteInBytes, err := db.lvdb.Get(key, nil)
	currentVote := binary.LittleEndian.Uint64(currentVoteInBytes)
	newVote := currentVote + amount

	newVoteInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newVoteInBytes, newVote)
	err = db.put(key, newVoteInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to count amount of vote to this candidate
	key = db.GetKey(string(voteDCBBoardCountPrefix), string(StartedBlock)+string(CandidatePubKey))
	currentCountInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		return err
	}
	currentCount := binary.LittleEndian.Uint32(currentCountInBytes)
	newCount := currentCount + 1
	newCountInByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(newCountInByte, newCount)
	err = db.put(key, newCountInByte)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to list voter new voter base on count as index
	key = db.GetKey(string(VoteDCBBoardListPrefix), string(currentCount)+string(StartedBlock)+string(CandidatePubKey))
	amountInByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountInByte, amount)
	valueInByte := append([]byte(VoterPubKey), amountInByte...)
	err = db.put(key, valueInByte)

	return nil
}

func (db *db) AddVoteGOVBoard(StartedBlockInt uint32, VoterPubKey []byte, CandidatePubKey []byte, amount uint64) error {
	StartedBlock := uint32(StartedBlockInt)
	//add to sum amount of vote token to this candidate
	key := db.GetKey(string(voteGOVBoardSumPrefix), string(StartedBlock)+string(CandidatePubKey))
	ok, err := db.hasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(zeroInBytes, uint64(0))
		db.put(key, zeroInBytes)
	}

	currentVoteInBytes, err := db.lvdb.Get(key, nil)
	currentVote := binary.LittleEndian.Uint64(currentVoteInBytes)
	newVote := currentVote + amount

	newVoteInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newVoteInBytes, newVote)
	err = db.put(key, newVoteInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to count amount of vote to this candidate
	key = db.GetKey(string(voteGOVBoardCountPrefix), string(StartedBlock)+string(CandidatePubKey))
	currentCountInBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		return err
	}
	currentCount := binary.LittleEndian.Uint32(currentCountInBytes)
	newCount := currentCount + 1
	newCountInByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(newCountInByte, newCount)
	err = db.put(key, newCountInByte)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	// add to list voter new voter base on count as index
	key = db.GetKey(string(VoteGOVBoardListPrefix), string(currentCount)+string(StartedBlock)+string(CandidatePubKey))
	amountInByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountInByte, amount)
	valueInByte := append([]byte(VoterPubKey), amountInByte...)
	err = db.put(key, valueInByte)

	return nil
}

func (db *db) GetTopMostVoteDCBGovernor(StartedBlock uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	prefix := db.GetKey(string(voteDCBBoardSumPrefix), string(StartedBlock))
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		keyI, _ := db.ReverseGetKey(string(voteDCBBoardSumPrefix), iter.Key())
		key := keyI.([]byte)
		pubKey := key[len(string(StartedBlock)+"#"):]
		value := binary.LittleEndian.Uint64(iter.Value())
		candidateList = append(candidateList, database.CandidateElement{pubKey, value})
	}
	sort.Sort(candidateList)
	if len(candidateList) < blockchain.NumberOfDCBGovernors {
		return nil, database.NewDatabaseError(database.NotEnoughCandidateDCB, errors.Errorf("not enough DCB Candidate"))
	}

	return candidateList[len(candidateList)-blockchain.NumberOfDCBGovernors:], nil
}

func (db *db) GetTopMostVoteGOVGovernor(StartedBlock uint32) (database.CandidateList, error) {
	var candidateList database.CandidateList
	//use prefix  as in file lvdb/block.go FetchChain
	prefix := db.GetKey(string(voteGOVBoardSumPrefix), string(StartedBlock))
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		keyI, _ := db.ReverseGetKey(string(voteGOVBoardSumPrefix), iter.Key())
		key := keyI.([]byte)
		pubKey := key[len(string(StartedBlock)+"#"):]
		value := binary.LittleEndian.Uint64(iter.Value())
		candidateList = append(candidateList, database.CandidateElement{pubKey, value})
	}
	sort.Sort(candidateList)
	if len(candidateList) < blockchain.NumberOfGOVGovernors {
		return nil, database.NewDatabaseError(database.NotEnoughCandidateGOV, errors.Errorf("not enough GOV Candidate"))
	}

	return candidateList[len(candidateList)-blockchain.NumberOfGOVGovernors:], nil
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
