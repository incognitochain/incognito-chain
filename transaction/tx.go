package transaction

import (
	"strconv"
	"github.com/ninjadotorg/cash-prototype/common"
	//"encoding/json"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

/*type Tx struct {
	Version  int     `json:"Version"`
	Type     string  `json:"Type"` // NORMAL / ACTION_PARAMS
	TxIn     []TxIn  `json:"TxIn"`
	TxOut    []TxOut `json:"TxOut"`
	LockTime int     `json:"LockTime"`
}*/

// TODO(@0xbunyip): add randomSeed, MACs and epk
type JoinSplitDesc struct {
	Anchor        []byte             `json:Anchor`
	Nullifiers    [][]byte           `json:Nullifiers`
	Commitments   [][]byte           `json:Commitments`
	Proof         *zksnark.PHGRProof `json:Proof`
	EncryptedData []byte             `json:EncryptedData`
}

type Tx struct {
	Version  int    `json:"Version"`
	Type     string `json:"Type"` // NORMAL / ACTION_PARAMS
	LockTime int    `json:"LockTime"`
	Fee      uint64 `json:"Fee"`

	Desc     []*JoinSplitDesc `json:Desc`
	JSPubKey []byte           `json:JSPubKey` // 32 bytes
	JSSig    []byte           `json:JSSig`    // 64 bytes
}

func (desc *JoinSplitDesc) toString() string {
	s := string(desc.Anchor)
	for _, nf := range desc.Nullifiers {
		s += string(nf)
	}
	for _, cm := range desc.Commitments {
		s += string(cm)
	}
	s += desc.Proof.String()
	s += string(desc.EncryptedData)
	return s
}

// Hash returns the hash of all fields of the transaction
func (tx *Tx) Hash() *common.Hash {
	record := strconv.Itoa(tx.Version)
	record += tx.Type
	record += strconv.Itoa(tx.LockTime)
	record += strconv.Itoa(len(tx.Desc))
	for _, desc := range tx.Desc {
		record += desc.toString()
	}
	record += string(tx.JSPubKey)
	record += string(tx.JSSig)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction returns true if transaction is valid:
// - All data fields are well formed
// - JSDescriptions are valid (zk-snark proof satisfied)
// - Signature matches the signing public key
// Note: This method doesn't check for double spending
func (tx *Tx) ValidateTransaction() bool {
	// TODO(@0xbunyip): implement
	return true
}

// GetType returns the type of the transaction
func (tx *Tx) GetType() string {
	return tx.Type
}
