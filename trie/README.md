#Trie package
It's patricia merkle tree based on ethereum design and implementation (geth)

## encoding
compact hex data to save storage, this kind of encoding only effect data from node to storage and vice versa
## node
implemented tree nodes in patricia merkle tree, node is purely place to hold data and form a trie

### node and raw node
- node: collapsed trie node (node in trie)
- raw node: already encoded rlp binary object (not in trie but already encoded as rlp)

## hasher
hash node and return hashNode (hash of that node)

## Next is Intermediate Writer
structure?
function?
what is do?
how it do things?
why?