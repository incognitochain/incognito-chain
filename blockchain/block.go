package blockchain

import (
	"strconv"

	"encoding/json"
	"errors"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

const (
	// Default length of list tx in block
	defaultTransactionAlloc = 2048
)

type AgentDataPoint struct {
	AgentID          string
	AgentSig         string
	NumOfCoins       float64
	NumOfBonds       float64
	Tax              float64
	EligibleAgentIDs []string
	LockTime         int64
}

type Block struct {
	Header          BlockHeader
	Transactions    []transaction.Transaction
	AgentDataPoints map[string]*AgentDataPoint
	ValidatorSig    string // PoS only

	Height    int32
	blockHash *common.Hash
}

/**
Customer UnmarshalJSON to parse list Tx
*/
func (self *Block) UnmarshalJSON(data []byte) error {
	type Alias Block
	temp := &struct {
		Transactions []map[string]interface{}
		*Alias
	}{
		Alias: (*Alias)(self),
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	// process tx from tx interface of temp
	for _, txTemp := range temp.Transactions {
		if txTemp["Type"].(string) == common.TxNormalType {
			// init a tx
			txNormal := &transaction.Tx{
				Version:  int8(txTemp["Version"].(float64)),
				Type:     txTemp["Type"].(string),
				LockTime: int64(txTemp["LockTime"].(float64)),
				Fee:      uint64(txTemp["Fee"].(float64)),
			}
			jSPubKey, ok := txTemp["JSPubKey"]
			if ok && jSPubKey != nil {
				txNormal.JSPubKey = common.JsonUnmarshallByteArray(jSPubKey.(string))
			}
			jSSig, ok := txTemp["JSSig"]
			if ok && jSSig != nil {
				txNormal.JSSig = common.JsonUnmarshallByteArray(jSSig.(string))
			}
			desc, ok := txTemp["Descs"]
			if ok && desc != nil {
				descTemps := desc.([]interface{})
				for _, descTemp := range descTemps {
					item := descTemp.(map[string]interface{})
					desc := &transaction.JoinSplitDesc{
						Anchor:          common.JsonUnmarshallByteArray(item["Anchor"].(string)),
						Type:            item["Type"].(string),
						Reward:          uint64(item["Reward"].(float64)),
						EphemeralPubKey: common.JsonUnmarshallByteArray(item["EphemeralPubKey"].(string)),
						HSigSeed:        common.JsonUnmarshallByteArray(item["HSigSeed"].(string)),
					}
					// proof
					if ok := item["Proof"] != nil; ok {
						proofTemp := item["Proof"].(map[string]interface{})
						proof := &zksnark.PHGRProof{
							G_A:      common.JsonUnmarshallByteArray(proofTemp["g_A"].(string)),
							G_APrime: common.JsonUnmarshallByteArray(proofTemp["g_A_prime"].(string)),
							G_B:      common.JsonUnmarshallByteArray(proofTemp["g_B"].(string)),
							G_BPrime: common.JsonUnmarshallByteArray(proofTemp["g_B_prime"].(string)),
							G_C:      common.JsonUnmarshallByteArray(proofTemp["g_C"].(string)),
							G_CPrime: common.JsonUnmarshallByteArray(proofTemp["g_C_prime"].(string)),
							G_K:      common.JsonUnmarshallByteArray(proofTemp["g_K"].(string)),
							G_H:      common.JsonUnmarshallByteArray(proofTemp["g_H"].(string)),
						}
						desc.Proof = proof
					}

					// nullifier
					if ok := item["Nullifiers"] != nil; ok {
						nullifiersTemp := item["Nullifiers"].([]interface{})
						nullifiers := make([][]byte, 0)
						for _, n := range nullifiersTemp {
							nullifiers = append(nullifiers, common.JsonUnmarshallByteArray(n.(string)))
						}
						desc.Nullifiers = nullifiers
					}

					// commitment
					if ok := item["Commitments"] != nil; ok {
						commitmentsTemp := item["Commitments"].([]interface{})
						commitments := make([][]byte, 0)
						for _, n := range commitmentsTemp {
							commitments = append(commitments, common.JsonUnmarshallByteArray(n.(string)))
						}
						desc.Commitments = commitments
					}

					// encrypt data
					if ok := item["EncryptedData"] != nil; ok {
						datasTemp := item["EncryptedData"].([]interface{})
						datas := make([][]byte, 0)
						for _, n := range datasTemp {
							datas = append(datas, common.JsonUnmarshallByteArray(n.(string)))
						}
						desc.EncryptedData = datas
					}

					// vmac
					if ok := item["Vmacs"] != nil; ok {
						vmacsTemp := item["Vmacs"].([]interface{})
						vmacs := make([][]byte, 0)
						for _, n := range vmacsTemp {
							vmacs = append(vmacs, common.JsonUnmarshallByteArray(n.(string)))
						}
						desc.Vmacs = vmacs
					}
					txNormal.Descs = append(txNormal.Descs, desc)
				}
			}
			self.Transactions = append(self.Transactions, txNormal)
		} else if txTemp["Type"].(string) == common.TxActionParamsType {
			// init a tx
			param := transaction.Param{
				Tax:              txTemp["Tax"].(float64),
				AgentID:          txTemp["AgentID"].(string),
				AgentSig:         txTemp["AgentSig"].(string),
				NumOfBonds:       txTemp["NumOfBonds"].(float64),
				NumOfCoins:       txTemp["NumOfCoins"].(float64),
				EligibleAgentIDs: txTemp["EligibleAgentIDs"].([]string),
			}
			txAction := transaction.ActionParamTx{
				LockTime: int64(txTemp["LockTime"].(float64)),
				Type:     txTemp["Type"].(string),
				Version:  int8(txTemp["Version"].(float64)),
				Param:    &param,
			}
			self.Transactions = append(self.Transactions, &txAction)
		} else {
			return errors.New("Can not parse a wrong tx")
		}
	}

	self.Header = temp.Alias.Header
	return nil
}

func (self *Block) AddTransaction(tx transaction.Transaction) error {
	self.Transactions = append(self.Transactions, tx)
	return nil
}

func (self *Block) ClearTransactions() {
	self.Transactions = make([]transaction.Transaction, 0, defaultTransactionAlloc)
}

func (self Block) Hash() *common.Hash {
	//if self.blockHash != nil {
	//	return self.blockHash
	//}
	record := strconv.Itoa(self.Header.Version) +
		self.Header.MerkleRoot.String() +
		self.Header.MerkleRootCommitments.String() +
		strconv.FormatInt(self.Header.Timestamp, 10) +
		self.Header.PrevBlockHash.String() +
		strconv.Itoa(self.Header.Nonce) +
		strconv.Itoa(len(self.Transactions))
	hash := common.DoubleHashH([]byte(record))
	//self.blockHash = &hash
	//return self.blockHash
	return &hash
}
