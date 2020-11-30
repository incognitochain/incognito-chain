package data

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/transaction"
	"math/big"
)

type TxInfo struct {
	Hash     	 string `json:"Hash"`
	LockTime 	 int64  `json:"LockTime"`
	HexData  	 string `json:"HexData"`
}

type OutputCoin struct {
	CoinDetails          Coin
}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	CoinDetails Coin
}


type Coin struct {
	PublicKey      privacy.Point
	CoinCommitment privacy.Point
	SNDerivator    privacy.Scalar
	SerialNumber   privacy.Point
	Randomness     privacy.Scalar
	Value          uint64
	Info           []byte //256 bytes
	Index          uint64
	TokenID    	   string
	ShardID    	   byte
}


type TxPrivacyTokenData struct {
	Tx			   Transaction          // used for privacy functionality
	PropertyID     string// = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
}

type Transaction struct {
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant
	Info     []byte // 512 bytes
	// Sign and Privacy proof, required
	SigPubKey            []byte `json:"SigPubKey,omitempty"` // 33 bytes
	Sig                  []byte `json:"Sig,omitempty"`       //
	InputCoins  []InputCoin
	OutputCoins []OutputCoin
	PubKeyLastByteSender byte
	// Metadata, optional
	Metadata metadata.Metadata
	Proof                *zkp.PaymentProof `json:"Proof,omitempty"`
	TxPrivacy 	*TxPrivacyTokenData `json:"TxPrivacy,omitempty"` //NormalTX this field null don't serialize it
	Hash   string
}



type TokenState struct {
	TokenID        string
	PropertyName   string
	PropertySymbol string
	TokenType      int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
	Info           []byte
	InitTx         string
	Txs            []string
}

type CommitmentState struct {
	TokenID    string
	ShardID    byte
	Commitment []byte
	Index      uint64
}

type Shard struct {
	ShardID           byte               `json:"ShardID"`
	BlockHash         string             `json:"Hash"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	Height            uint64             `json:"Height"`
	Version           int                `json:"Version"`
	TxRoot            string             `json:"TxRoot"`
	Time              int64              `json:"Time"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []TxInfo			 `json:"Txs"`
	BlockProducer     string             `json:"BlockProducer"`
	ValidationData    string             `json:"ValidationData"`
	ConsensusType     string             `json:"ConsensusType"`
	Data              string             `json:"Data"`
	BeaconHeight      uint64             `json:"BeaconHeight"`
	BeaconBlockHash   string             `json:"BeaconBlockHash"`
	Round             int                `json:"Round"`
	Epoch             uint64             `json:"Epoch"`
	Reward            uint64             `json:"Reward"`
	RewardBeacon      uint64             `json:"RewardBeacon"`
	Fee               uint64             `json:"Fee"`
	Size              uint64             `json:"Size"`
	Instruction       		[][]string         `json:"Instruction"`
	CrossShardBitMap  		[]int              `json:"CrossShardBitMap"`
	NumTxns                uint64                            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64                            `json:"TotalTxns"`              // The total number of txns in the chain.
	NumTxnsExcludeSalary   uint64                            `json:"NumTxnsExcludeSalary"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary uint64                            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int                               `json:"ActiveShards"`
	ConsensusAlgorithm     string                            `json:"ConsensusAlgorithm"`
	NumOfBlocksByProducers map[string]uint64 			`json:"NumOfBlocksByProducers"`
	ShardCommittee         []incognitokey.CommitteeKeyString `json:"ShardCommittee"`
	ShardPendingValidator  []incognitokey.CommitteeKeyString `json:"ShardPendingValidator"`

	CommitteeRewardState map[string]map[common.Hash]uint64 `json:"CommitteeRewardState"`
	TokenState 				[]TokenState  `json:"TokenState"`
	Transactions			[]Transaction  `json:"Transaction"`
}


func NewShardFromShardState (shardState *blockchain.ShardBestState) *Shard {
	shard := &Shard{
		ShardID: shardState.ShardID,
		BlockHash: shardState.BestBlockHash.String(),
		PreviousBlockHash: shardState.GetPreviousHash().String(),
		Height: shardState.GetHeight(),
		Version: shardState.BestBlock.GetVersion(),
		TxRoot: shardState.BestBlock.Header.TxRoot.String(),
		Time: shardState.GetBlockTime(),
		TxHashes: []string{},
		Txs: getBlockTx(shardState.BestBlock.Body.Transactions),
		BlockProducer: shardState.BestBlock.GetProducer(),
		ValidationData: shardState.BestBlock.GetValidationField(),
		ConsensusType: shardState.BestBlock.GetConsensusType(),
		Data: "",
		BeaconHeight: shardState.GetBeaconHeight(),
		BeaconBlockHash: shardState.BestBeaconHash.String(),
		Round: shardState.BestBlock.GetRound(),
		Epoch: shardState.BestBlock.GetCurrentEpoch(),
		Reward: 0, //fill by below code
		RewardBeacon: 0, //fill by below code
		Fee: 0, // fill by below code.
		Size: getBlockSize(shardState.BestBlock),
		Instruction: getInstruction(shardState.BestBlock.GetInstructions()),
		CrossShardBitMap: getCrossShardBitMap(shardState.BestBlock.Header.CrossShardBitMap),
		NumTxns: shardState.NumTxns,
		TotalTxns: shardState.TotalTxns,
		NumTxnsExcludeSalary: shardState.NumTxnsExcludeSalary,
		TotalTxnsExcludeSalary: shardState.TotalTxnsExcludeSalary,
		ActiveShards: shardState.ActiveShards,
		ConsensusAlgorithm: shardState.ConsensusAlgorithm,
		NumOfBlocksByProducers: getNumOfBlocksByProducers(shardState.NumOfBlocksByProducers),
		ShardCommittee: incognitokey.CommitteeKeyListToStringList(shardState.ShardCommittee),
		ShardPendingValidator: incognitokey.CommitteeKeyListToStringList(shardState.ShardPendingValidator),
		CommitteeRewardState: getRewardCommittee(shardState.GetShardRewardStateDB()),
		TokenState: getPrivacyToken(shardState.GetCopiedTransactionStateDB()),
		Transactions: parseTransactions(shardState),
	}
	if len(shardState.BestBlock.Body.Transactions) > 0 {
		for _, tx := range shardState.BestBlock.Body.Transactions {
			if tx.GetMetadataType() == metadata.ShardBlockReward {
				shard.Reward += tx.GetProof().GetOutputCoins()[0].CoinDetails.GetValue()
			} else if tx.GetMetadataType() == metadata.BeaconSalaryResponseMeta {
				shard.RewardBeacon += tx.GetProof().GetOutputCoins()[0].CoinDetails.GetValue()
			}
		}
	}
	for _, tx := range shardState.BestBlock.Body.Transactions {
		shard.TxHashes = append(shard.TxHashes, tx.Hash().String())
		shard.Fee += tx.GetTxFee()
	}
	return shard
}

func getCrossShardBitMap(srcCrossShardBitMap []byte ) []int {
	var CrossShardBitMap []int
	if len(srcCrossShardBitMap) > 0 {
		for _, shardID := range srcCrossShardBitMap {
			CrossShardBitMap = append(CrossShardBitMap, int(shardID))
		}
	}
	return CrossShardBitMap
}

func getBlockSize(block *blockchain.ShardBlock) uint64 {
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return 0
	}
	return uint64(len(blockBytes))
}

func getInstruction(insts [][]string) [][]string{
	dest := make ([][]string, 0, len(insts))
	for _, inst := range insts {
		i := make([]string, len(inst))
		copy(i, inst)
		dest = append(dest, i)
	}
	return dest
}

func getNumOfBlocksByProducers (NumOfBlocksByProducers map[string]uint64) map[string]uint64 {
	blockByProducer := make (map[string]uint64)

	for key, val := range NumOfBlocksByProducers {
		blockByProducer[key] = val
	}
	return blockByProducer
}

func getBlockTx (Transactions []metadata.Transaction) []TxInfo {
	Txs := make([]TxInfo, 0)
	for _, tx := range Transactions {
		transactionResult := TxInfo{}
		transactionResult.Hash = tx.Hash().String()
		switch tx.GetType() {
			case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
				txN := tx.(*transaction.Tx)
				data, err := json.Marshal(txN)
				if err != nil {
					continue
				}
				transactionResult.HexData = hex.EncodeToString(data)
				transactionResult.LockTime = txN.LockTime

			case common.TxCustomTokenPrivacyType:
				txP := tx.(*transaction.TxCustomTokenPrivacy)
				data, err := json.Marshal(txP)
				if err != nil {
					continue
				}
				transactionResult.HexData = hex.EncodeToString(data)
				transactionResult.LockTime = txP.LockTime
		}
		Txs = append(Txs, transactionResult)
	}
	return Txs
}

func getRewardCommittee(db *statedb.StateDB) map[string]map[common.Hash]uint64 {
	return statedb.ListCommitteeReward(db)
}

func getPrivacyToken(transactionDB *statedb.StateDB) []TokenState{
	privacyToken := statedb.ListPrivacyToken(transactionDB)

	tokenState := []TokenState{}

	for _, token := range privacyToken {
		tokenState = append(tokenState, TokenState{
			TokenID:        token.TokenID().String(),
			PropertyName:   token.PropertyName(),
			PropertySymbol: token.PropertySymbol(),
			TokenType:      token.TokenType(),
			Mintable:       token.Mintable(),
			Amount:         token.Amount(),
			Info:           token.Info(),
			InitTx:         token.InitTx().String(),
			Txs:            convertHashSliceToHashString(token.Txs()),
		})
	}

	return tokenState
}

func convertHashSliceToHashString(hashes []common.Hash ) []string {
	result := make ([]string, 0, len(hashes))
	for _, hash := range  hashes {
		result = append(result, hash.String())
	}
	return result
}

func convertPrivacyInputCoinToInputCoin (privacyCoin *privacy.InputCoin) InputCoin {
	return InputCoin {CoinDetails: Coin{
		PublicKey:      *privacyCoin.CoinDetails.GetPublicKey(),
		CoinCommitment: *privacyCoin.CoinDetails.GetCoinCommitment(),
		SNDerivator:    *privacyCoin.CoinDetails.GetSNDerivator(),
		SerialNumber:   *privacyCoin.CoinDetails.GetSerialNumber(),
		Randomness:     *privacyCoin.CoinDetails.GetRandomness(),
		Value:          privacyCoin.CoinDetails.GetValue(),
		Info:           privacyCoin.CoinDetails.GetInfo(),
		Index:          0,
	}}
}

func convertSlicePrivacyInputCoinToSliceInputCoin(privacyCoins []*privacy.InputCoin) []InputCoin{
	coins := make([]InputCoin, 0, len(privacyCoins))
	for _, privacyCoin := range privacyCoins {
		coins = append(coins, convertPrivacyInputCoinToInputCoin(privacyCoin))
	}
	return coins
}

func convertPrivacyOutputCoinToOutputCoin (coin *privacy.OutputCoin) OutputCoin {
	return OutputCoin {CoinDetails: Coin{
		PublicKey:      *coin.CoinDetails.GetPublicKey(),
		CoinCommitment: *coin.CoinDetails.GetCoinCommitment(),
		SNDerivator:    *coin.CoinDetails.GetSNDerivator(),
		SerialNumber:   *coin.CoinDetails.GetSerialNumber(),
		Randomness:     *coin.CoinDetails.GetRandomness(),
		Value:          coin.CoinDetails.GetValue(),
		Info:           coin.CoinDetails.GetInfo(),
		Index:          0,
	}}
}

func convertSlicePrivacyOutputCoinToSliceOutputCoin(privacyCoins []*privacy.OutputCoin) []OutputCoin{
	coins := make([]OutputCoin, 0, len(privacyCoins))
	for _, privacyCoin := range privacyCoins {
		coins = append(coins, convertPrivacyOutputCoinToOutputCoin(privacyCoin))
	}
	return coins
}

func convertNormalInternalTransactionToTransaction (normalTx *transaction.Tx) Transaction {
	return Transaction{
		Version:              normalTx.Version,
		Type:                 normalTx.Type,
		LockTime:             normalTx.LockTime,
		Fee:                  normalTx.Fee,
		Info:                 normalTx.Info,
		SigPubKey:            normalTx.SigPubKey,
		Sig:                  normalTx.Sig,
		InputCoins:           convertSlicePrivacyInputCoinToSliceInputCoin(normalTx.Proof.GetInputCoins()),
		OutputCoins:          convertSlicePrivacyOutputCoinToSliceOutputCoin(normalTx.Proof.GetOutputCoins()),
		PubKeyLastByteSender: normalTx.PubKeyLastByteSender,
		Metadata:             normalTx.Metadata,
		Proof:                normalTx.Proof,
		Hash:                 normalTx.Hash().String(),
	}
}

func convertInternalTokenPrivacyToTokenPrivacy(data *transaction.TxCustomTokenPrivacy) *TxPrivacyTokenData {
	return &TxPrivacyTokenData{
		Tx:    convertNormalInternalTransactionToTransaction(&data.TxPrivacyTokenData.TxNormal),
		PropertyID:     data.TxPrivacyTokenData.PropertyID.String(),
		PropertyName:   data.TxPrivacyTokenData.PropertyName,
		PropertySymbol: data.TxPrivacyTokenData.PropertySymbol,
		Type:           data.TxPrivacyTokenData.Type,
		Mintable:       data.TxPrivacyTokenData.Mintable,
		Amount:         data.TxPrivacyTokenData.Amount,
	}
}

func getTransaction(transactions []metadata.Transaction) []Transaction {
	newTransactions := make([]Transaction, 0, len(transactions))
	for _, tx := range transactions {
		switch tx.GetType() {
		case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
			{
				normalTx := tx.(*transaction.Tx)
				newTransactions = append(newTransactions, convertNormalInternalTransactionToTransaction(normalTx))
			}
		case common.TxCustomTokenPrivacyType:
			{
				tx := tx.(*transaction.TxCustomTokenPrivacy)
				newTransactions = append(newTransactions, Transaction{
					Version:              tx.Version,
					Type:                 tx.Type,
					LockTime:             tx.LockTime,
					Fee:                  tx.Fee,
					Info:                 tx.Info,
					SigPubKey:            tx.SigPubKey,
					Sig:                  tx.Sig,
					InputCoins:           convertSlicePrivacyInputCoinToSliceInputCoin(tx.Proof.GetInputCoins()),
					OutputCoins:          convertSlicePrivacyOutputCoinToSliceOutputCoin(tx.Proof.GetOutputCoins()),
					PubKeyLastByteSender: tx.PubKeyLastByteSender,
					Metadata:             tx.Metadata,
					Proof:                tx.Proof,
					Hash:                 tx.Hash().String(),
					TxPrivacy:			  convertInternalTokenPrivacyToTokenPrivacy(tx),
				})

			}
		}
	}
	return newTransactions
}

func parseTransactions(shardBestState *blockchain.ShardBestState)  []Transaction {
	prevTransactionDB := shardBestState.GetPrevTransactionStateDB()
	currentTransactionDB := shardBestState.GetCopiedTransactionStateDB()
	tokenState := statedb.ListPrivacyToken(currentTransactionDB)  //TODO: Should enhance to get only token in this block
	//Parse Transaction
	transactions := getTransaction(shardBestState.BestBlock.Body.Transactions)

	commitmentStates := make([]CommitmentState,0)
	//Get Commitment
	for tokenID := range tokenState {
		//Get Commitment length
		previousLength, _ := statedb.GetCommitmentLength(prevTransactionDB, tokenID, shardBestState.ShardID)
		if previousLength == nil {
			continue
		}
		if previousLength.Uint64() != 0 {
			temp := previousLength.Uint64() + 1
			previousLength = new(big.Int).SetUint64(temp)
		}
		currentLength, _ := statedb.GetCommitmentLength(currentTransactionDB, tokenID, shardBestState.ShardID)
		if currentLength == nil {
			continue
		}
		for previousLength.Cmp(currentLength) <= 0  {
			if commitmentState, err := statedb.GetCommitmentStateByIndex(currentTransactionDB, tokenID, previousLength.Uint64(), shardBestState.ShardID); err != nil {
				commitmentStates = append(commitmentStates, CommitmentState{
					TokenID:    commitmentState.TokenID().String(),
					ShardID:    commitmentState.ShardID(),
					Commitment: commitmentState.Commitment(),
					Index:      commitmentState.Index().Uint64()})
			}
			temp2 := previousLength.Uint64() + 1
			previousLength = new(big.Int).SetUint64(temp2)
		}
	}
	for _ , commitmentState := range commitmentStates {
		for _, transaction := range transactions {
			updateIndexForTransactionOutputCoin(transaction,commitmentState)
		}
	}
	return transactions
}

func updateIndexForTransactionOutputCoin(transaction Transaction, commitment CommitmentState) {
	for _, outputCoin := range transaction.OutputCoins {
		if bytes.Compare(commitment.Commitment, outputCoin.CoinDetails.CoinCommitment.ToBytesS()) == 0 {
			outputCoin.CoinDetails.ShardID = commitment.ShardID
			outputCoin.CoinDetails.TokenID = commitment.TokenID
			outputCoin.CoinDetails.Index = commitment.Index
		}
	}

}