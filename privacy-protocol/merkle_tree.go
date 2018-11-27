package privacy

import (
	"fmt"

	"encoding/base64"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol/crypto/sha256"
)

// MerkleHash represents a hash in merkle tree
type MerkleHash []byte

// MarshalJSON converts hash to JSON
func (hash MerkleHash) MarshalJSON() ([]byte, error) {
	hashString := base64.StdEncoding.EncodeToString(hash[:])
	return json.Marshal(hashString)
}

// UnmarshalJSON converts JSON to hash
func (hash *MerkleHash) UnmarshalJSON(data []byte) error {
	hashString := ""
	_ = json.Unmarshal(data, &hashString)
	data, _ = base64.StdEncoding.DecodeString(hashString)
	copy(*hash, data)
	return nil
}

// String
func (h MerkleHash) String() string {
	return fmt.Sprintf("%s", []byte(h[:2]))
}

func combineAndHash(left, right []byte) MerkleHash {
	data := append(left, right...)
	hash := sha256.Sum256NoPad(data)
	return hash[:]
}

type MerklePath struct {
	AuthPath []MerkleHash
	Index    []bool
}

// CreateDummyPath creates a dummy MerklePath with all 0s to act as a placeholder for merkle tree check
func (mp *MerklePath) CreateDummyPath() *MerklePath {
	mp.Index = make([]bool, common.IncMerkleTreeHeight)
	for i := 0; i < common.IncMerkleTreeHeight; i++ {
		hash := make(MerkleHash, 32)
		mp.AuthPath = append(mp.AuthPath, hash)
	}
	return mp
}

var uncommittedNodeHash MerkleHash = make([]byte, 32)
var arrayPadding []MerkleHash

// IncMerkleTree compactly represents a fixed height merkle tree
type IncMerkleTree struct {
	nodes       []MerkleHash // One hash for each height, hash is hash value of all nodes have same height
	left, right MerkleHash   // Leaf nodes, they may not be appended to the tree
}

// MakeCopy creates a new merkle tree and copies data from the old one to it
func (tree *IncMerkleTree) MakeCopy() *IncMerkleTree {
	newTree := &IncMerkleTree{}

	if tree.left != nil {
		newTree.left = make([]byte, 32)
		copy(newTree.left, tree.left)
	}

	if tree.right != nil {
		newTree.right = make([]byte, 32)
		copy(newTree.right, tree.right)
	}

	for _, node := range tree.nodes {
		var nodeCopy MerkleHash
		if node != nil {
			nodeCopy = make([]byte, 32)
			copy(nodeCopy, node)
		}
		newTree.nodes = append(newTree.nodes, nodeCopy)
	}

	return newTree
}

// IncMerkleWitness stores the data for the process of building the merkle tree path
type IncMerkleWitness struct {
	snapshot, tmpTree *IncMerkleTree
	uncles            []MerkleHash
	uID, nextDepth    int
}

// TakeSnapshot takes a snapshot of a merkle tree to start building merkle path for the right most leaf
func (tree *IncMerkleTree) TakeSnapshot() *IncMerkleWitness {
	witness := &IncMerkleWitness{
		snapshot:  tree.MakeCopy(),
		tmpTree:   nil,
		uncles:    nil,
		nextDepth: 0,
		uID:       0,
	}
	return witness
}

func getPaddingAtDepth(d int) MerkleHash {
	if len(arrayPadding) >= common.IncMerkleTreeHeight {
		return arrayPadding[d]
	}
	var paddings []MerkleHash
	paddings = append(paddings, uncommittedNodeHash[:])
	lastHash := uncommittedNodeHash
	for i := 1; i <= common.IncMerkleTreeHeight; i++ {
		paddings = append(paddings, combineAndHash(lastHash[:], lastHash[:]))
		lastHash = paddings[len(paddings)-1]
	}
	arrayPadding = paddings
	return arrayPadding[d]
}

// AddNewNode incrementally adds a new leaf node data to the merkle tree
func (tree *IncMerkleTree) AddNewNode(hash MerkleHash) {

	hashCopy := make([]byte, 32) // Make a copy to make sure the hash is not changed while using IncMerkleTree
	copy(hashCopy, hash[:])
	if tree.left == nil {
		tree.left = hashCopy
	} else if tree.right == nil {
		tree.right = hashCopy
	} else {
		// Combine previous 2 leaves' hash to get parent hash
		prevHash := combineAndHash(tree.left, tree.right)

		// Save the new hash to new leafs
		tree.left = hashCopy
		tree.right = nil

		leftOverHash := true // Found new hash for parent node, push to array if no parents found
		for d, node := range tree.nodes {
			if node == nil {
				tree.nodes[d] = prevHash // Save hash for current depth
				leftOverHash = false     // Hash already saved in array
				break
			}

			prevHash = combineAndHash(node, prevHash)
			tree.nodes[d] = nil // Clear data at current depth for next hashes
		}
		if leftOverHash {
			if len(tree.nodes) >= common.IncMerkleTreeHeight-1 {
				panic("Too many hashes for IncMerkleTree")
			}
			tree.nodes = append(tree.nodes, prevHash)
		}
	}
	fmt.Printf("Number of nodes: %v\n", len(tree.nodes))
}

// GetRoot returns the merkle root of an unfinished tree; empty nodes are padded with default values
func (tree *IncMerkleTree) GetRoot(d int) MerkleHash {
	fmt.Printf("get root at depth %d\n", d)
	if d > common.IncMerkleTreeHeight {
		d = common.IncMerkleTreeHeight // Maximum height of the tree to get root
	}

	left := uncommittedNodeHash
	right := uncommittedNodeHash
	if tree.right != nil {
		fmt.Printf("have tree.right != nil\n")
		if d == 0 {
			fmt.Printf("d == 0, return tree.right\n")
			return tree.right
		}
		right = tree.right
	}
	if tree.left != nil {
		fmt.Printf("have tree.left != nil\n")
		if d == 0 {
			fmt.Printf("d == 0, return tree.left\n")
			return tree.left
		}
		left = tree.left[:]
	}
	prevHash := combineAndHash(left, right)
	for i := 0; i < d-1; i++ {
		if i >= len(tree.nodes) || tree.nodes[i] == nil {
			// fmt.Printf("getroot depth %d: getPadding\n", i)
			// Previous hash is the left child, right child doesn't exist
			prevHash = combineAndHash(prevHash[:], getPaddingAtDepth(i+1))
		} else {
			// Previous hash is the right child
			prevHash = combineAndHash(tree.nodes[i], prevHash)
			fmt.Printf("getroot depth %d: combine hash %x\n", i, prevHash)
		}
	}
	return prevHash
}

// finishedDepth checks if a merkle tree of depth d is exactly full (cannot be higher or lower)
func (tree *IncMerkleTree) finishedDepth(d int) bool {
	if tree.left == nil { // Empty merkle tree
		return false
	} else if tree.right == nil {
		if d == 0 { // Require merkle tree with only 1 node
			return true
		}
		return false // Merkle tree cannot be full if the rightmost leaf is nil
	}
	for i := 0; i < d-1; i++ {
		if i >= len(tree.nodes) || tree.nodes[i] == nil {
			return false
		}
	}
	return true
}

func (tree *IncMerkleTree) getNextUncleDepth(numUnclesBuilt int) int {
	if tree.left == nil { // Empty merkle tree
		if numUnclesBuilt <= 0 {
			return 0
		}
		numUnclesBuilt--
	} else if tree.right == nil {
		if numUnclesBuilt <= 0 {
			return 0
		}
		numUnclesBuilt--
	}

	nextDepth := 1
	for i := 0; i < common.IncMerkleTreeHeight; i++ {
		if i >= len(tree.nodes) || tree.nodes[i] == nil {
			if numUnclesBuilt <= 0 {
				return nextDepth
			}
			numUnclesBuilt--
		}
		nextDepth++
	}
	return nextDepth
}

func (w *IncMerkleWitness) addNewNode(hash MerkleHash) {
	if w.tmpTree == nil {
		w.nextDepth = w.snapshot.getNextUncleDepth(len(w.uncles))
		w.tmpTree = new(IncMerkleTree)
		fmt.Printf("tmpTree is nil, nextDepth: %d\n", w.nextDepth)
	}

	w.tmpTree.AddNewNode(hash)
	if w.tmpTree.finishedDepth(w.nextDepth) {
		fmt.Printf("tmpTree finished\n")
		rt := w.tmpTree.GetRoot(w.nextDepth)
		w.uncles = append(w.uncles, rt)
		w.tmpTree = nil
	}
}

func (w *IncMerkleWitness) getUncle(d int) MerkleHash {
	if w.uID >= len(w.uncles) {
		fmt.Printf("no more uncles to take, get empty rt at depth %d\n", d)
		return (&IncMerkleTree{}).GetRoot(d)
	}
	fmt.Printf("return uncle with id %d at depth %d\n", w.uID, d)
	hash := w.uncles[w.uID]
	w.uID++
	return hash
}

// getWitnessPath builds the path from the right most leaf of an IncMerkleWitness to the newest root
// Note: this should only be called once for a IncMerkleWitness; after the newest commitment
// had already been added to it
func (w *IncMerkleWitness) getWitnessPath() *MerklePath {
	path := &MerklePath{}
	if w.snapshot.left == nil { // Empty merkle tree, return dummy path
		return (&MerklePath{}).CreateDummyPath()
	}

	// Build the last uncle if the temporary tree is not finished
	if w.tmpTree != nil {
		fmt.Printf("tmpTree != nil\n")
		rt := w.tmpTree.GetRoot(w.nextDepth)
		w.uncles = append(w.uncles, rt)
		w.tmpTree = nil
	}
	w.uID = 0

	if w.snapshot.right != nil { // Right most leaf (witnessing commitment) is the right child
		fmt.Printf("witness is the right child\n")
		hash := make([]byte, 32)
		copy(hash, w.snapshot.left[:])
		path.AuthPath = append(path.AuthPath, w.snapshot.left)
		path.Index = append(path.Index, true)
	} else { // Right most leaf is the left child
		fmt.Printf("witness is the left child\n")
		hash := w.getUncle(0)
		path.AuthPath = append(path.AuthPath, hash)
		path.Index = append(path.Index, false)
	}

	for d := 0; d < common.IncMerkleTreeHeight-1; d++ {
		if d >= len(w.snapshot.nodes) || w.snapshot.nodes[d] == nil {
			fmt.Printf("getWitnessPath at depth %d: need newer nodes\n", d)
			// Need a node newer than the witnessing commitment
			path.AuthPath = append(path.AuthPath, w.getUncle(d+1))
			path.Index = append(path.Index, false)
		} else {
			fmt.Printf("getWitnessPath at depth %d: already have uncle in snapshot\n", d)
			// Need a node older than the witnessing commitment
			// The node is stored in the snapshot tree at height d (from the bottom, ignore leaves)
			path.AuthPath = append(path.AuthPath, w.snapshot.nodes[d])
			path.Index = append(path.Index, true)
		}
	}
	return path
}

// // BuildWitnessPath builds commitments merkle path for all given notes
// func BuildWitnessPath(notes []*JSInput, commitments [][]byte) ([]byte, error) {
// 	tree := IncMerkleTree{}
// 	witnesses := make([]*IncMerkleWitness, len(notes))

// 	for _, cm := range commitments {
// 		tree.AddNewNode(cm)

// 		// Add "newer" nodes to witness' merkle tree to build merkle path of the latest merkle root
// 		for _, witness := range witnesses {
// 			if witness != nil {
// 				fmt.Printf("add new node to witness: %x\n", cm)
// 				witness.addNewNode(cm)
// 			}
// 		}

// 		// If we find a needed commitment, take a snapshot of the tree whose right most leaf is the commitment
// 		for i, note := range notes {
// 			if bytes.Equal(cm, note.InputNote.Cm) {
// 				if witnesses[i] != nil {
// 					return nil, fmt.Errorf("Duplicate commitments for input notes")
// 				}
// 				witnesses[i] = tree.TakeSnapshot()
// 			}
// 		}
// 	}

// 	// Check if all notes have witnesses
// 	for i, witness := range witnesses {
// 		if witness == nil {
// 			return nil, fmt.Errorf("Input note with commitment %x not existed in commitment list", notes[i].InputNote.Cm)
// 		}
// 	}

// 	fmt.Printf("\ngetWitnessPath\n")
// 	// Get the path to the merkle tree root of each witness's tree
// 	for i, note := range notes {
// 		note.WitnessPath = witnesses[i].getWitnessPath()
// 	}

// 	fmt.Printf("\ngetting tree root of newest tree\n")
// 	fmt.Printf("newest tree root: %x\n", tree.GetRoot(common.IncMerkleTreeHeight))
// 	fmt.Printf("new@28: %x\n", tree.GetRoot(28))
// 	fmt.Printf("witness@29: %x\n", notes[0].WitnessPath.AuthPath[28])
// 	fmt.Printf("new@28+witness@29: %x\n", combineAndHash(tree.GetRoot(28), notes[0].WitnessPath.AuthPath[28]))
// 	fmt.Printf("new@29: %x\n", tree.GetRoot(29))
// 	// fmt.Printf("anchor: %x\n", notes[0].WitnessPath.AuthPath[len(notes[0].WitnessPath.AuthPath)-1])
// 	newRt := tree.GetRoot(common.IncMerkleTreeHeight)
// 	return newRt, nil
// }

// // BuildWitnessPathMultiChain builds witness path for multiple input notes from different chains
// func BuildWitnessPathMultiChain(inputs map[byte][]*JSInput, commitments map[byte][][]byte) (map[byte][]byte, error) {
// 	mapRt := make(map[byte][]byte)
// 	for chainID, inputList := range inputs {
// 		rt, err := BuildWitnessPath(inputList, commitments[chainID])
// 		if err != nil {
// 			return nil, err
// 		}
// 		mapRt[chainID] = rt
// 	}
// 	return mapRt, nil
// }
