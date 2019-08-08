# Incognito Key structure
Incognito key include 3 value;
- Private Key: use for spending UTXO
- Payment Address: use for receiving token or PRV
- ReadonlyKey: use for decrypt data in output of tx if Payment Address output com from owner of readonly key

## Payment Address
Payment address includes two keys:
- Public key: It is derived from private key
- Transmission key: It is derived from Receiving key and it used to encrypt output's details when creating transactions

## Readonly Key 
Readonly key includes two keys:
- Public key: It is derived from private key
- Receiving key: It is random generated from private key and it used to decrypt output's details when receiving transactions
