# Client controller on command line for incognito-chain

## Backup and Restore Database
`$ go build -o [app-name]`

Run from root directory of project (incognito-chain/)
`$ ./[app-name] --cmd backupchain [flags]`

List of flags
```$xslt
 --beacon: backup beacon chain
 --shardids [1,2,3,...] or --shardids all: backup shard chain
 --chaindatadir "[...]/block": blockchain database to be backup
 --outdatadir "...." : directory where backup file store
 --filename "....": name of backup file
 --testnet: backup blockchain database is testnet or mainnet (only 2 option for now)  
```

Example:

    - Backup: `$ ./cmd/incognito --cmd backupchain --chaindatadir "data/shard0-0/testnet/block" --outdatadir "data/" --shardids "0" --testnetbackupchain`
    
    - Restore: `$ ./cmd/incognito --cmd restorechain --chaindatadir "data/shard0-0/testnet/block" --filename data/export-incognito-shard-0 --shardids 0 --testnet`

Note about blockchain database directory:
- By default block will be stored in .../testnet/block or .../mainnet/block