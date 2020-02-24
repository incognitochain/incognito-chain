package utxo

import "github.com/incognitochain/incognito-chain/privacy"

type Utxo struct {
	index      uint8
	mask       *privacy.Scalar
	amount     *privacy.Scalar
	txData     *privacy.Point
	addressee  *privacy.Point // K^o = H_n(r * K_B^v )G + K_B^s
	commitment *privacy.Point
}

func (this *Utxo) GetIndex() uint8               { return this.index }
func (this *Utxo) GetMask() *privacy.Scalar      { return this.mask }
func (this *Utxo) GetAmount() *privacy.Scalar    { return this.amount }
func (this *Utxo) GetTxData() *privacy.Point     { return this.txData }
func (this *Utxo) GetAddressee() *privacy.Point  { return this.addressee }
func (this *Utxo) GetCommitment() *privacy.Point { return this.commitment }

func (this *Utxo) SetMask(mask *privacy.Scalar)          { this.mask = mask }
func (this *Utxo) SetAmount(amount *privacy.Scalar)      { this.amount = amount }
func (this *Utxo) SetTxData(txData *privacy.Point)       { this.txData = txData }
func (this *Utxo) SetAddressee(addressee *privacy.Point) { this.addressee = addressee }

func NewUtxo(index uint8, mask *privacy.Scalar, amount *privacy.Scalar, txData *privacy.Point, addressee *privacy.Point, commitment *privacy.Point) *Utxo {
	return &Utxo{
		index,
		mask,
		amount,
		txData,
		addressee,
		commitment,
	}
}
