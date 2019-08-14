package lvdb

import (
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type db struct {
	lvdb *leveldb.DB
}

// key prefix
var (
	prevShardPrefix          = []byte("prevShd-")
	prevBeaconPrefix         = []byte("prevBea-")
	beaconPrefix             = []byte("bea-")
	beaconBestBlockkeyPrefix = []byte("bea-bestBlock")
	committeePrefix          = []byte("com-")
	rewardReceiverPrefix     = []byte("rewardreceiver-")
	heightPrefix             = []byte("ep-") // TODO rename key value
	shardIDPrefix            = []byte("s-")
	blockKeyPrefix           = []byte("b-")
	blockHeaderKeyPrefix     = []byte("bh-")
	blockKeyIdxPrefix        = []byte("i-")
	crossShardKeyPrefix      = []byte("csh-")
	nextCrossShardKeyPrefix  = []byte("ncsh-")
	shardPrefix              = []byte("shd-")

	shardToBeaconKeyPrefix       = []byte("stb-")
	transactionKeyPrefix         = []byte("tx-")
	privateKeyPrefix             = []byte("prk-")
	serialNumbersPrefix          = []byte("serinalnumbers-")
	commitmentsPrefix            = []byte("commitments-")
	outcoinsPrefix               = []byte("outcoins-")
	snderivatorsPrefix           = []byte("snderivators-")
	bestBlockKeyPrefix           = []byte("bestBlock")
	feeEstimatorPrefix           = []byte("feeEstimator")
	tokenPrefix                  = []byte("token-")
	privacyTokenPrefix           = []byte("privacy-token-")
	privacyTokenCrossShardPrefix = []byte("privacy-cross-token-")
	tokenInitPrefix              = []byte("token-init-")
	privacyTokenInitPrefix       = []byte("privacy-token-init-")

	// multisigs
	multisigsPrefix = []byte("multisigs")

	// centralized bridge
	bridgePrefix              = []byte("bridge-")
	centralizedBridgePrefix   = []byte("centralizedbridge-")
	decentralizedBridgePrefix = []byte("decentralizedbridge-")
	ethTxHashIssuedPrefix     = []byte("ethtxhashissued-")

	// Incognito -> Ethereum relayer
	burnConfirmPrefix = []byte("burnConfirm-")

	//epoch reward
	shardRequestRewardPrefix = []byte("shardrequestreward-")
	committeeRewardPrefix    = []byte("committee-reward-")

	// public variable
	TokenPaymentAddressPrefix = []byte("token-paymentaddress-")
	Splitter                  = []byte("-[-]-")
)

// value
var (
	Spent   = []byte("spent")
	Unspent = []byte("unspent")
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

func (db *db) Batch(data []database.BatchData) leveldb.Batch {
	batch := new(leveldb.Batch)
	for _, v := range data {
		batch.Put(v.Key, v.Value)
	}
	return *batch
}

func (db *db) PutBatch(data []database.BatchData) error {
	batch := db.Batch(data)
	err := db.lvdb.Write(&batch, nil)
	return err
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
		return nil, database.NewDatabaseError(database.LvDbNotFound, errors.Wrap(err, "db.lvdb.Get"))
	}
	return value, nil
}
