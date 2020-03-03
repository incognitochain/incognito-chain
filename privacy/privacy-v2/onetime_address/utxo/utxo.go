package utxo

import "github.com/incognitochain/incognito-chain/privacy/operation"

type Utxo struct {
	index      uint8
	mask       *operation.Scalar
	amount     *operation.Scalar
	txData     *operation.Point
	addressee  *operation.Point // K^o = H_n(r * K_B^v )G + K_B^s
	commitment *operation.Point
}

func (this *Utxo) GetIndex() uint8                 { return this.index }
func (this *Utxo) GetMask() *operation.Scalar      { return this.mask }
func (this *Utxo) GetAmount() *operation.Scalar    { return this.amount }
func (this *Utxo) GetTxData() *operation.Point     { return this.txData }
func (this *Utxo) GetAddressee() *operation.Point  { return this.addressee }
func (this *Utxo) GetCommitment() *operation.Point { return this.commitment }

func (this *Utxo) SetMask(mask *operation.Scalar)          { this.mask = mask }
func (this *Utxo) SetAmount(amount *operation.Scalar)      { this.amount = amount }
func (this *Utxo) SetTxData(txData *operation.Point)       { this.txData = txData }
func (this *Utxo) SetAddressee(addressee *operation.Point) { this.addressee = addressee }

func NewUtxo(index uint8, mask *operation.Scalar, amount *operation.Scalar, txData *operation.Point, addressee *operation.Point, commitment *operation.Point) *Utxo {
	return &Utxo{
		index,
		mask,
		amount,
		txData,
		addressee,
		commitment,
	}
}
