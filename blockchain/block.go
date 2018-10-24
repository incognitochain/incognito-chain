package blockchain

import (
	"strconv"
	"strings"

	"encoding/json"
	"errors"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/privacy/proto/zksnark"
	"github.com/ninjadotorg/cash/transaction"
)

/*
Block is struct present every block in blockchain
block contains many types of transaction
- normal tx:
- action tx:

*/
type Block struct {
	Header         BlockHeader
	Transactions   []transaction.Transaction
	ChainLeader    string
	ChainLeaderSig string

	Height    int32
	blockHash *common.Hash
}

/*
Customize UnmarshalJSON to parse list Tx
because we have many types of block, so we can need to customize data from marshal from json string to build a block
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
				Version:         int8(txTemp["Version"].(float64)),
				Type:            txTemp["Type"].(string),
				LockTime:        int64(txTemp["LockTime"].(float64)),
				Fee:             uint64(txTemp["Fee"].(float64)),
				AddressLastByte: byte(txTemp["AddressLastByte"].(float64)),
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

					// anchor
					if ok := item["Anchor"] != nil; ok {
						anchorsTemp := item["Anchor"].([]interface{})
						anchors := make([][]byte, 0)
						for _, n := range anchorsTemp {
							anchors = append(anchors, common.JsonUnmarshallByteArray(n.(string)))
						}
						desc.Anchor = anchors
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
		} else if txTemp["Type"].(string) == common.TxVotingType {
			// init a tx
			txNormal := &transaction.TxVoting{
				Tx: transaction.Tx{
					Version:         int8(txTemp["Version"].(float64)),
					Type:            txTemp["Type"].(string),
					LockTime:        int64(txTemp["LockTime"].(float64)),
					Fee:             uint64(txTemp["Fee"].(float64)),
					AddressLastByte: byte(txTemp["AddressLastByte"].(float64)),
				},
				NodeAddr: txTemp["NodeAddr"].(string),
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

					// anchor
					if ok := item["Anchor"] != nil; ok {
						anchorsTemp := item["Anchor"].([]interface{})
						anchors := make([][]byte, 0)
						for _, n := range anchorsTemp {
							anchors = append(anchors, common.JsonUnmarshallByteArray(n.(string)))
						}
						desc.Anchor = anchors
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
		} else {
			return errors.New("Can not parse a wrong tx")
		}
	}

	self.Header = temp.Alias.Header
	return nil
}

/*
AddTransaction adds a new transaction into block
*/
// #1 - tx
func (self *Block) AddTransaction(tx transaction.Transaction) error {
	self.Transactions = append(self.Transactions, tx)
	return nil
}

/*
Hash creates a hash from block data
*/

func (self Block) Hash() *common.Hash {
	if self.blockHash != nil {
		return self.blockHash
	}
	record := strconv.Itoa(self.Header.Version) +
		string(self.Header.ChainID) +
		self.Header.MerkleRoot.String() +
		self.Header.MerkleRootCommitments.String() +
		strconv.FormatInt(self.Header.Timestamp, 10) +
		self.Header.PrevBlockHash.String() +
		strconv.Itoa(len(self.Transactions)) +
		self.ChainLeader +
		strconv.Itoa(int(self.Height)) +
		strconv.Itoa(int(self.Header.SalaryFund)) +
		strings.Join(self.Header.Committee, ",")
	hash := common.DoubleHashH([]byte(record))
	self.blockHash = &hash
	return self.blockHash
}
