# Wallet

HD wallets generate a hierarchical tree-like structure of keys which start from the seed master key based on BIP 32. When you restore an HD wallet using the seed key, the wallet goes ahead and drives all the private keys of the tree using BIP 32.

Advantages of HD Wallets

- You need to backup only one key (i.e. “seed key”). It is the only backup you will ever need.
- You can generate many receiving addresses every time you receive bitcoins.
- You can protect your financial privacy.
- Confuse new users, as your receiving address changes every time.