# Memory pool for transactions
This package will process txs which are received from other peers in a broadcast tx message.
We need to make some policy and validation to check tx before add it into mempool.

Which valid txs in mempool, mining processing will get them and make consensus to create a new block

@Note: this is only one type of tx resource for mining