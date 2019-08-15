# Mempool Persistence Database (MPDB)
## Introduction
Mempool Persistence Database is a leveldb, which seperated from the main leveldb. 
This database only store transactions in mempool.

Transactions in mempool is stored in Memory (RAM), they will be erased when node is turned off or crashed.
So, MPDB is used to stored current transactions in mempool. If transaction is removed in mempool for any reason, it will be removed out of mpdb as well.

This feature can be turn on and off by using config.
## Feature
- Add Transaction: add transaction to database
- Remove Transaction: remove transaction out of database
- Has Transaction: check transaction existence
- Reset: delete all transactions in database
- Load: load all transaction from database into memory