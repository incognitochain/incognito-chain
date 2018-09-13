package client

import (
	"bytes"
	"fmt"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client/crypto/sha256"
)

type MerkleHash []byte

type MerklePath struct {
	AuthPath []*MerkleHash
	Index    []bool
}

// CreateDummyPath creates a dummy MerklePath with all 0s to act as a placeholder for merkle tree check
func (mp *MerklePath) CreateDummyPath() *MerklePath {
	mp.Index = make([]bool, common.IncMerkleTreeHeight)
	for i := 0; i < common.IncMerkleTreeHeight; i++ {
		hash := make(MerkleHash, 32)
		mp.AuthPath = append(mp.AuthPath, &hash)
	}
	return mp
}

var uncommittedNodeHash MerkleHash = make([]byte, 32)
var arrayPadding []MerkleHash

// IncMerkleTree compactly represents a fixed height merkle tree
type IncMerkleTree struct {
	nodes       []MerkleHash // One hash for each height
	left, right MerkleHash   // Leaf nodes
}

// IncMerkleWitness stores the data for the process of building the merkle tree path
type IncMerkleWitness struct {
	snapshot, tmpTree *IncMerkleTree
	uncles            []MerkleHash
	nextDepth         int
}

// TakeSnapshot takes a snapshot of a merkle tree to start building merkle path for the right most leaf
func (tree *IncMerkleTree) TakeSnapshot() *IncMerkleWitness {
	treeCopy := &IncMerkleTree{}
	if tree.left != nil {
		treeCopy.left = make([]byte, 32)
		copy(treeCopy.left, tree.left)
	}
	if tree.right != nil {
		treeCopy.right = make([]byte, 32)
		copy(treeCopy.right, tree.right)
	}
	for _, node := range tree.nodes {
		var nodeCopy MerkleHash
		if node != nil {
			nodeCopy = make([]byte, 32)
			copy(nodeCopy, node)
		}
		treeCopy.nodes = append(treeCopy.nodes, nodeCopy)
	}

	witness := &IncMerkleWitness{
		snapshot:  treeCopy,
		tmpTree:   nil,
		uncles:    nil,
		nextDepth: 0,
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
	for i := 1; i < common.IncMerkleTreeHeight; i++ {
		data := append(lastHash[:], lastHash[:]...)
		lastHash := sha256.Sum256NoPad(data)
		paddings = append(paddings, lastHash[:])
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
		data := append(tree.left, tree.right...)
		merged := sha256.Sum256NoPad(data)
		prevHash := merged[:]

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

			data := append(node, prevHash...)
			tree.nodes[d] = nil // Clear data at current depth for next hashes
			newHash := sha256.Sum256NoPad(data)
			prevHash = newHash[:]

		}
		if leftOverHash {
			if len(tree.nodes) >= common.IncMerkleTreeHeight-1 {
				panic("Too many hashes for IncMerkleTree")
			}
			tree.nodes = append(tree.nodes, prevHash)
		}
	}
}

// GetRoot returns the merkle root of an unfinished tree; empty nodes are padded with default values
func (tree *IncMerkleTree) GetRoot(d int) MerkleHash {
	left := uncommittedNodeHash
	right := uncommittedNodeHash
	if tree.right != nil {
		if d == 0 {
			return tree.right
		}
		right = tree.right
	}
	if tree.left != nil {
		if d == 0 {
			return tree.left
		}
		left = tree.left[:]
	}
	combined := append(left, right...)
	prevHash := sha256.Sum256NoPad(combined)
	for i := 0; i < d-1; i++ {
		var data []byte
		if i >= len(tree.nodes) || tree.nodes[i] == nil {
			data = append(prevHash[:], getPaddingAtDepth(i)[:]...) // Previous hash is the left child, right child doesn't exist
		} else {
			data = append(tree.nodes[i], prevHash[:]...) // Previous hash is the right child
		}
		prevHash = sha256.Sum256NoPad(data)
	}
	return MerkleHash(prevHash[:])
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
	}

	w.tmpTree.AddNewNode(hash)
	if w.tmpTree.finishedDepth(w.nextDepth) {
		rt := w.tmpTree.GetRoot(w.nextDepth)
		w.uncles = append(w.uncles, rt)
		w.tmpTree = nil
	}
}

// getWitnessPath builds the path from the right most leaf of an IncMerkleWitness to the newest root
func (w *IncMerkleWitness) getWitnessPath() *MerklePath {
	return nil
}

// BuildWitnessPath builds commitments merkle path for all given notes
func BuildWitnessPath(notes []*JSInput, commitments [][]byte) error {
	tree := IncMerkleTree{}
	witnesses := make([]*IncMerkleWitness, len(notes))

	for _, cm := range commitments {
		tree.AddNewNode(cm)

		for i, note := range notes {
			if bytes.Equal(cm, note.InputNote.Cm) {
				if witnesses[i] != nil {
					return fmt.Errorf("Duplicate commitments for input notes")
				}
				witnesses[i] = tree.TakeSnapshot()
			}
		}

		for _, witness := range witnesses {
			if witness != nil {
				witness.addNewNode(cm)
			}
		}
	}

	// Check if all notes have witnesses
	for i, witness := range witnesses {
		if witness == nil {
			return fmt.Errorf("Input note with commitment %x not existed in commitment list", notes[i].InputNote.Cm)
		}
	}

	// Get the path to the merkle tree root of each witness's tree
	for i, note := range notes {
		note.WitnessPath = witnesses[i].getWitnessPath()
	}

	return nil
}
