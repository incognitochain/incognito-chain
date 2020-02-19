package onetime_address

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

type UTXO struct {
	index      uint8
	mask       *privacy.Scalar
	amount     *privacy.Scalar
	txData     *privacy.Point
	addressee  *privacy.Point // K^o = H_n(r * K_B^v )G + K_B^s
	commitment *privacy.Point
}

func (this *UTXO) GetIndex() uint8               { return this.index }
func (this *UTXO) GetMask() *privacy.Scalar      { return this.mask }
func (this *UTXO) GetAmount() *privacy.Scalar    { return this.amount }
func (this *UTXO) GetTxData() *privacy.Point     { return this.txData }
func (this *UTXO) GetAddressee() *privacy.Point  { return this.addressee }
func (this *UTXO) GetCommitment() *privacy.Point { return this.commitment }

func (this *UTXO) SetMask(mask *privacy.Scalar)          { this.mask = mask }
func (this *UTXO) SetAmount(amount *privacy.Scalar)      { this.amount = amount }
func (this *UTXO) SetTxData(txData *privacy.Point)       { this.txData = txData }
func (this *UTXO) SetAddressee(addressee *privacy.Point) { this.addressee = addressee }
