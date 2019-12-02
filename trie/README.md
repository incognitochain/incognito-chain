#Trie package
It's patricia merkle tree based on ethereum design and implementation (geth)

The core of the trie, and its sole requirement in terms of the protocol specification is 
to provide a single value that identifies a given set of key-value pairs,
which may be either a 32 byte sequence or the empty byte sequence
## encoding
compact hex data to save storage, this kind of encoding only effect data from node to storage and vice versa
## node
implemented tree nodes in patricia merkle tree, node is purely place to hold data and form a trie

### node and raw node
- node: collapsed trie node (node in trie)
- raw node: already encoded rlp binary object (not in trie but already encoded as rlp)

## hasher
hash node and return hashNode (hash of that node)

## intermediate Writer
- manage node and raw full node
- get raw node from db
- commit all nodes in trie from one root node (traverse all nodes from root node then put to batch then write batch to db)


## trie and secure trie
- get, update, insert, hash by trie rule from intermediate writer