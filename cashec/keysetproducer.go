package cashec

/*import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"golang.org/x/crypto/ed25519"
)

// TODO - remove this object after apply ecliptic curves for keysetprivacy

type KeySetProducer struct {
	SprivateKey     []byte
	SpublicKey      []byte
	SpendingAddress [client.SpendingAddressLength]byte
	TransmissionKey [client.TransmissionKeyLength]byte
	ReceivingKey    [client.ReceivingKeyLength]byte
}

func (self *KeySetProducer) GenerateKey(seed []byte) (*KeySetProducer, error) {
	var err error
	self.SpublicKey, self.SprivateKey, err = ed25519.GenerateKey(bytes.NewBuffer(seed))
	if err != nil {
		return self, err
	}
	return self, nil
}

func (self *KeySetProducer) Import(privateKey string) (*KeySetProducer, error) {
	key := ed25519.PrivateKey{}
	base58C := base58.Base58Check{}
	key, _, err := base58C.Decode(privateKey)
	if err != nil {
		return self, err
	}
	self.SpublicKey = key.Public().(ed25519.PublicKey)
	self.SprivateKey = key
	return self, nil
}

func (self *KeySetProducer) Verify(data, signature []byte) (bool, error) {
	isValid := false
	isValid = ed25519.Verify(self.SpublicKey, data, signature)
	return isValid, nil
}

func (self *KeySetProducer) Sign(data []byte) ([]byte, error) {
	result := ed25519.Sign(self.SprivateKey, data)
	return result, nil
}

func (self *KeySetProducer) EncodeToString() string {
	val, _ := json.Marshal(self)
	result := base58.Base58Check{}.Encode(val, byte(0x00))
	return result
}

func (self *KeySetProducer) DecodeToKeySet(keystring string) (*KeySetProducer, error) {
	base58C := base58.Base58Check{}
	keyBytes, _, _ := base58C.Decode(keystring)
	json.Unmarshal(keyBytes, self)
	return self, nil
}

func (self *KeySetProducer) GetPaymentAddress() (client.PaymentAddress, error) {
	var paymentAddr client.PaymentAddress
	paymentAddr.Apk = self.SpendingAddress
	paymentAddr.Pkenc = self.TransmissionKey
	return paymentAddr, nil
}

func (self *KeySetProducer) GetViewingKey() (client.ViewingKey, error) {
	var viewingKey client.ViewingKey
	viewingKey.Apk = self.SpendingAddress
	viewingKey.Skenc = self.ReceivingKey
	return viewingKey, nil
}

func ValidateDataB58(pubkey string, sig string, data []byte) error {
	decPubkey, _, err := base58.Base58Check{}.Decode(pubkey)
	if err != nil {
		return errors.New("can't decode public key:" + err.Error())
	}

	validatorKp := KeySetProducer{
		SpublicKey: decPubkey,
	}
	decSig, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return errors.New("can't decode signature: " + err.Error())
	}

	isValid, err := validatorKp.Verify(data, decSig)
	if err != nil {
		return errors.New("error when verify data: " + err.Error())
	}
	if !isValid {
		return errors.New("Invalid signature")
	}
	return nil
}*/
