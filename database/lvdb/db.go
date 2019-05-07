package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type db struct {
	lvdb *leveldb.DB
}

type hasher interface {
	Hash() *common.Hash
}

var (
	beaconPrefix            = []byte("bea-")
	beaconBestBlockkey      = []byte("bea-bestBlock")
	stabilityPrefix         = []byte("sta-")
	committeePrefix         = []byte("com-")
	epochPrefix             = []byte("ep-")
	shardIDPrefix           = []byte("s")
	blockKeyPrefix          = []byte("b-")
	blockHeaderKeyPrefix    = []byte("bh-")
	blockKeyIdxPrefix       = []byte("i-")
	crossShardKeyPrefix     = []byte("csh-")
	nextCrossShardKeyPrefix = []byte("ncsh-")

	shardToBeaconKeyPrefix       = []byte("stb-")
	transactionKeyPrefix         = []byte("tx-")
	privateKeyPrefix             = []byte("prk-")
	serialNumbersPrefix          = []byte("serinalnumbers-")
	commitmentsPrefix            = []byte("commitments-")
	outcoinsPrefix               = []byte("outcoins-")
	snderivatorsPrefix           = []byte("snderivators-")
	bestBlockKey                 = []byte("bestBlock")
	feeEstimator                 = []byte("feeEstimator")
	Splitter                     = []byte("-[-]-")
	TokenPrefix                  = []byte("token-")
	PrivacyTokenPrefix           = []byte("privacy-token-")
	PrivacyTokenCrossShardPrefix = []byte("privacy-cross-token-")
	TokenPaymentAddressPrefix    = []byte("token-paymentaddress-")
	tokenInitPrefix              = []byte("token-init-")
	privacyTokenInitPrefix       = []byte("privacy-token-init-")
	rewared                      = []byte("reward")

	//vote prefix
	voteBoardSumPrefix           = []byte("votesumboard-")
	voteBoardCountPrefix         = []byte("votecountboard-")
	VoteBoardListPrefix          = []byte("votelistboard-")
	threePhraseCryptoOwnerPrefix = []byte("threephrasecryptoownerprefix-")
	VoteProposalPrefix           = []byte("voteproposalprefix-")
	BoardFundPrefix              = []byte("boardfundprefix-")
	ConstantPricePrefix          = []byte("constantspriceprefix-")
	listVoterOfProposalPrefix    = []byte("listvoterofproposalprefix-")
	SubmitProposalPrefix         = []byte("submitproposalprefix-")
	threePhraseVoteValuePrefix   = []byte("threephrasevotevalueprefix-")
	winningVoterPrefix           = []byte("winningvoterprefix-")
	//pubKeyToPaymentAddress          = []byte("pubkeytopaymentaddress")
	encryptFlagPrefix               = []byte("encryptFlagPrefix")
	encryptionLastBlockHeightPrefix = []byte("encryptionlastblockheightprefix")

	// multisigs
	multisigsPrefix = []byte("multisigs")

	// bond type
	bondTypePrefix = []byte("bondtype-")

	// crowdsale
	crowdsalePrefix       = []byte("crowdsale-")
	dcbBondInfoPrefix     = []byte("dcbBondInfo-")
	tradeActivationPrefix = []byte("tradeAct-")

	Unreward   = []byte("unreward")
	Spent      = []byte("spent")
	Unspent    = []byte("unspent")
	Mintable   = []byte("mintable")
	UnMintable = []byte("unmintable")

	// reserve
	reserveIssuingInfoPrefix     = []byte("rsIss")
	reserveContractingInfoPrefix = []byte("rsCon")
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
		return nil, database.NewDatabaseError(database.LvDbNotFound, errors.Wrap(err, "db.lvdb.Get"))
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
		dbkey = append(serialNumbersPrefix, []byte(key.(*common.Hash).String())...)
	case string(commitmentsPrefix):
		dbkey = append(commitmentsPrefix, []byte(key.(*common.Hash).String())...)
	case string(outcoinsPrefix):
		dbkey = append(outcoinsPrefix, []byte(key.(*common.Hash).String())...)
	case string(snderivatorsPrefix):
		dbkey = append(snderivatorsPrefix, []byte(key.(*common.Hash).String())...)
	case string(TokenPrefix):
		dbkey = append(TokenPrefix, []byte(key.(*common.Hash).String())...)
	case string(PrivacyTokenPrefix):
		dbkey = append(PrivacyTokenPrefix, []byte(key.(*common.Hash).String())...)
	case string(PrivacyTokenCrossShardPrefix):
		dbkey = append(PrivacyTokenCrossShardPrefix, []byte(key.(*common.Hash).String())...)
	case string(tokenInitPrefix):
		dbkey = append(tokenInitPrefix, []byte(key.(*common.Hash).String())...)
	case string(privacyTokenInitPrefix):
		dbkey = append(privacyTokenInitPrefix, []byte(key.(*common.Hash).String())...)
	}
	return dbkey
}
