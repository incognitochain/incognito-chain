# Client controller on command line for incognito-chain

## Backup and Restore Database
### Command
`$ go build -o [app-name]`

Run from root directory of project (incognito-chain/)
`$ ./[app-name] --cmd backupchain [flags]`

List of flags
```$xslt
 --beacon: backup beacon chain
 --shardids [string params can be splited with ","] or --shardids "all"
 --chaindatadir "[string params]/block": blockchain database to be backup
 --outdatadir [string params] : directory where backup file store
 --filename [string params]: name of backup file
 --testnet: backup blockchain database is testnet or mainnet (only 2 option for now)  
```

Example:
- Backup:
    - Beacon: 
    `$ ./cmd/incognito --cmd backupchain --chaindatadir "data/fullnode/testnet/block" --outdatadir "data/" --beacon --testnet`
    - Shard:
    `$ ./cmd/incognito --cmd backupchain --chaindatadir "data/fullnode/testnet/block" --outdatadir "data/" --shardids 0,1,2,3 --testnet`
    - All:
    `$ ./cmd/incognito --cmd backupchain --chaindatadir "data/fullnode/testnet/block" --outdatadir "data/" --shardids all --beacon --testnet`
  
- Restore: 
    - Beacon: Restore only Beacon Chain
    
    `$ ./cmd/incognito --cmd restorechain --chaindatadir "data/fullnode/testnet/block" --filename data/export-incognito-beacon --beacon --testnet`
    - Shard: Restore only Shard Chain (support multi shard at a time)
    
    `$ ./cmd/incognito --cmd restorechain --chaindatadir "data/fullnode/testnet/block" --filename "data/export-incognito-shard-0,data/export-incognito-shard-1" --testnet`

### Notice
- You SHOULD Restore Beacon Chain Database BEFORE Shard Chain Database
- By default block will be stored in .../testnet/block or .../mainnet/block
