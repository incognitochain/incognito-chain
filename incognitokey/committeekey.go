package incognitokey

import "encoding/json"

type CommitteePubKey struct {
	IncPubKey    []byte
	MiningPubKey map[string][]byte
}

// // func (priKey *CommitteePriKey) LoadKey(consensusPriKey map[string]string) error {
// // 	priKey.PriKey = map[string][]byte{}
// // 	for key, value := range consensusPriKey {
// // 		keyBytes, ver, err := base58.Base58Check{}.Decode(value)
// // 		if (ver != common.ZeroByte) || (err != nil) {
// // 			return err
// // 		}
// // 		priKey.PriKey[key] = keyBytes
// // 	}
// // 	return nil
// // }

// // func (pubKey *CommitteePubKey) LoadKey(consensusPubKey map[string]string) error {

// // }

// func (keyset *MiningKey) GetPublicKeyBase58() string {
// 	return base58.Base58Check{}.Encode(keyset.Publickey, common.ZeroByte)
// }

func (pubKey *CommitteePubKey) FromBytes(keyBytes []byte) error {
	return json.Unmarshal(keyBytes, pubKey)
}

func (pubKey *CommitteePubKey) Bytes() ([]byte, error) {
	return json.Marshal(pubKey)
}

func (pubKey *CommitteePubKey) GetNormalKey() []byte {
	return pubKey.IncPubKey
}

func (pubKey *CommitteePubKey) GetMiningKey(schemeName string) []byte {
	return pubKey.MiningPubKey[schemeName]
}
