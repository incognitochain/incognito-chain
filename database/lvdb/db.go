package lvdb

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"log"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type db struct {
	lvdb *leveldb.DB
}

type hasher interface {
	Hash() *common.Hash
}

var (
	chainIDPrefix             = []byte("c")
	blockKeyPrefix            = []byte("b-")
	blockHeaderKeyPrefix      = []byte("bh-")
	blockKeyIdxPrefix         = []byte("i-")
	transactionKeyPrefix      = []byte("tx-")
	privateKeyPrefix          = []byte("prk-")
	serialNumbersPrefix       = []byte("serinalnumbers-")
	commitmentsPrefix         = []byte("commitments-")
	outcoinsPrefix            = []byte("outcoins-")
	snderivatorsPrefix        = []byte("snderivators-")
	bestBlockKey              = []byte("bestBlock")
	feeEstimator              = []byte("feeEstimator")
	Splitter                  = []byte("-[-]-")
	TokenPrefix               = []byte("token-")
	TokenPaymentAddressPrefix = []byte("token-paymentaddress-")
	tokenInitPrefix           = []byte("token-init-")
	loanIDKeyPrefix           = []byte("loanID-")
	loanTxKeyPrefix           = []byte("loanTx-")
	loanRequestPostfix        = []byte("-req")
	loanResponsePostfix       = []byte("-res")
	rewared                   = []byte("reward")

	//vote prefix
	voteDCBBoardSumPrefix         = []byte("votedcbsumboard-")
	voteDCBBoardCountPrefix       = []byte("votedcbcountboard-")
	VoteDCBBoardListPrefix        = []byte("votedcblistboard-")
	DCBVoteTokenAmountPrefix      = []byte("dcbvotetokenamount-")
	voteGOVBoardSumPrefix         = []byte("votegovsumboard-")
	voteGOVBoardCountPrefix       = []byte("votegovcountboard-")
	VoteGOVBoardListPrefix        = []byte("votegovlistboard-")
	GOVVoteTokenAmountPrefix      = []byte("govvotetokenamount-")
	threePhraseCryptoOwnerPrefix  = []byte("threephrasecryptoownerprefix-")
	threePhraseCryptoSealerPrefix = []byte("threephrasecryptosealerprefix-")
	threePhraseVoteValuePrefix    = []byte("threephrasevotevalueprefix-")

	Unreward = []byte("unreward")
	Spent    = []byte("spent")
	Unspent  = []byte("unspent")
)

func open(dbPath string) (database.DatabaseInterface, error) {
	lvdb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.OpenDbErr, errors.Wrapf(err, "levelvdb.OpenFile %s", dbPath))
	}
	return &db{lvdb: lvdb}, nil
}

func (db *db) Close() error {
	return errors.Wrap(db.lvdb.Close(), "db.lvdb.Close")
}

func (db *db) HasValue(key []byte) (bool, error) {
	ret, err := db.lvdb.Has(key, nil)
	if err != nil {
		return false, database.NewDatabaseError(database.NotExistValue, err)
	}
	return ret, nil
}

func (db *db) Put(key, value []byte) error {
	if err := db.lvdb.Put(key, value, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Put"))
	}
	return nil
}

func (db *db) Delete(key []byte) error {
	err := db.lvdb.Delete(key, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	return nil
}

func (db *db) Get(key []byte) ([]byte, error) {
	value, err := db.lvdb.Get(key, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return value, nil
}

func (db db) GetKey(keyType string, key interface{}) []byte {
	var dbkey []byte
	switch keyType {
	case string(blockKeyPrefix):
		dbkey = append(blockKeyPrefix, key.(*common.Hash)[:]...)
	case string(blockKeyIdxPrefix):
		dbkey = append(blockKeyIdxPrefix, key.(*common.Hash)[:]...)
	case string(serialNumbersPrefix):
		dbkey = append(serialNumbersPrefix, []byte(key.(string))...)
	case string(commitmentsPrefix):
		dbkey = append(commitmentsPrefix, []byte(key.(string))...)
	case string(outcoinsPrefix):
		dbkey = append(outcoinsPrefix, []byte(key.(string))...)
	case string(snderivatorsPrefix):
		dbkey = append(snderivatorsPrefix, []byte(key.(string))...)
	case string(TokenPrefix):
		dbkey = append(TokenPrefix, key.(*common.Hash)[:]...)
	case string(tokenInitPrefix):
		dbkey = append(tokenInitPrefix, key.(*common.Hash)[:]...)
	}
	return dbkey
}
