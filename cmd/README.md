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
    `$ ./cmd/incognito --cmd backupchain --chaindatadir "../testnet/fullnode/testnet/block" --outdatadir "../testnet/" --beacon --testnet`
    - Shard:
    `$ ./cmd/incognito --cmd backupchain --chaindatadir "../testnet/fullnode/testnet/block" --outdatadir "../testnet/" --shardids 0,1,2,3 --testnet`
    - All:
    `$ ./cmd/incognito-cmd --cmd backupchain --chaindatadir "../testnet/fullnode/testnet/block" --outdatadir "../testnet/" --shardids all --beacon --testnet`
  
- Restore: 
    - Beacon: Restore only Beacon Chain
    
    `$ ./cmd/incognito-cmd --cmd restorechain --chaindatadir /home/testnet1/fullnode/testnet/block" --filename ../testnet/export-incognito-beacon --beacon --testnet`
    - Shard: Restore only Shard Chain (support multi shard at a time)
    
    `$ ./cmd/incognito-cmd --cmd restorechain --chaindatadir "/home/testnet1/fullnode/testnet/block" --filename "../testnet/export-incognito-shard-0,../testnet/export-incognito-shard-1,../testnet/export-incognito-shard-2,../testnet/export-incognito-shard-3,../testnet/export-incognito-shard-4,../testnet/export-incognito-shard-5,../testnet/export-incognito-shard-6,../testnet/export-incognito-shard-7" --testnet`

### Notice
- You SHOULD Restore Beacon Chain Database BEFORE Shard Chain Database
- By default block will be stored in .../testnet/block or .../mainnet/block
