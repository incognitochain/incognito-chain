package syncker

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/patrickmn/go-cache"
)

// var MAX_RANGE =

type AVLSync struct {
	root     *AVLNode
	maxRange uint64
	cache    *cache.Cache
}

func NewAVLSync(cacher *cache.Cache) *AVLSync {
	return &AVLSync{
		root:     nil,
		maxRange: config.Param().EpochParam.NumberOfBlockInEpochV2,
		cache:    cacher,
	}
}

func (t *AVLSync) Add(key uint64, value uint64) {
	t.root = t.root.add(key, value)
}

func (t *AVLSync) Remove(key uint64) {
	t.root = t.root.remove(key)
}

func (t *AVLSync) Update(oldKey uint64, newKey uint64, newValue uint64) {
	t.root = t.root.remove(oldKey)
	t.root = t.root.add(newKey, newValue)
}

func (t *AVLSync) SearchMissingPair(left, right uint64) (newPair [2][]uint64) {
	missing := [2][]uint64{}
	missing[0] = append(missing[0], left)
	missing[1] = append(missing[1], right)
	return t.root.searchMissingPair(missing)
}

func (t *AVLSync) InsertPair(left, right uint64) (node *AVLNode) {
	t.root = t.root.insertMissingPair(left, right)
	return t.root
}

func (t *AVLSync) Find(key uint64) *AVLNode {
	return t.root.find(key)
}
func (t *AVLSync) SearchPair(left, right uint64) (node *AVLNode) {
	root := t.root
	for {
		if root == nil {
			root.add(left, right)
			return root
		}
		found, turnLeft, _ := root.searchPair(left, right)
		if found {
			root.update(left, right)
			return root
		}
		if turnLeft {
			root = root.left
		} else {
			root = root.right
		}
	}
}

func (t *AVLSync) DisplayInOrder() {
	fmt.Println("=================================")
	if t.root != nil {
		t.root.displayNodesInOrder()
	} else {
		Logger.Errorf("root is nil")
	}
	fmt.Println()
	fmt.Println("=================================")
}

// AVLNode structure
type AVLNode struct {
	key   uint64
	Value uint64

	// height counts nodes (not edges)
	height int
	left   *AVLNode
	right  *AVLNode
}

func (n *AVLNode) ContainPair(l, r uint64) bool {
	return ((l <= n.key) && (r >= n.key)) || ((l <= n.Value) && (r >= n.Value)) || ((l >= n.key) && (r <= n.Value))
}

// Adds a new node
func (n *AVLNode) add(key uint64, value uint64) *AVLNode {
	if n == nil {
		return &AVLNode{key, value, 1, nil, nil}
	}
	if key < n.key {
		n.left = n.left.add(key, value)
	} else {
		n.right = n.right.add(key, value)
	}

	return n.rebalanceTree()
}

// Removes a node
func (n *AVLNode) remove(key uint64) *AVLNode {
	if n == nil {
		return nil
	}
	if key < n.key {
		n.left = n.left.remove(key)
	} else if key > n.key {
		n.right = n.right.remove(key)
	} else {
		if n.left != nil && n.right != nil {
			// node to delete found with both children;
			// replace values with smallest node of the right sub-tree
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			n.Value = rightMinNode.Value
			// delete smallest node that we replaced
			n.right = n.right.remove(rightMinNode.key)
		} else if n.left != nil {
			// node only has left child
			n = n.left
		} else if n.right != nil {
			// node only has right child
			n = n.right
		} else {
			// node has no children
			n = nil
			return n
		}

	}
	return n.rebalanceTree()
}

func (n *AVLNode) update(key, value uint64) {
	if n.key > key {
		n.key = key
	}
	if n.Value < value {
		n.Value = value
	}
}

// Searches for a node
func (n *AVLNode) find(key uint64) *AVLNode {

	if (n == nil) || (n.key == key) {
		return n
	}
	if n.key < key {
		return n.right.find(key)
	}
	return n.left.find(key)
}

// Searches for a node
func (n *AVLNode) searchPair(l, r uint64) (
	contained bool,
	leftOfNode bool,
	rightOfNode bool,
) {
	if n.ContainPair(l, r) {
		return true, false, false
	}
	if n.Value < l {
		return false, false, true
	}
	return false, false, true
}

func (n *AVLNode) searchMissingPair(missingPair [2][]uint64) (newPair [2][]uint64) {
	if n == nil {
		return missingPair
	}
	leftMissing := [2][]uint64{}
	rightMissing := [2][]uint64{}
	for i := 0; i < len(missingPair[0]); i++ {
		from := missingPair[0][i]
		to := missingPair[1][i]
		if from < n.key {
			leftMissing[0] = append(leftMissing[0], from)
			if to >= n.key {
				leftMissing[1] = append(leftMissing[1], n.key-1)
			} else {
				leftMissing[1] = append(leftMissing[1], to)
			}
		}
		if to > n.Value {
			rightMissing[1] = append(rightMissing[1], to)
			if from <= n.Value {
				rightMissing[0] = append(rightMissing[0], n.Value+1)
			} else {
				rightMissing[0] = append(rightMissing[0], from)
			}
		}
	}
	leftMissing = n.left.searchMissingPair(leftMissing)
	rightMissing = n.right.searchMissingPair(rightMissing)
	newPair[0] = append(leftMissing[0], rightMissing[0]...)
	newPair[1] = append(leftMissing[1], rightMissing[1]...)
	return newPair
}

func (n *AVLNode) insertMissingPair(left, right uint64) *AVLNode {
	if n == nil {
		return &AVLNode{left, right, 1, nil, nil}
	}
	leftPair := [2]uint64{0, 0}
	rightPair := [2]uint64{0, 0}
	if left < n.key {
		if (n.Value-left+1 < 350) && (right+1 >= n.key) {
			n.key = left
		} else {
			leftPair[0] = left
			if right >= n.key {
				leftPair[1] = n.key - 1
			} else {
				leftPair[1] = right
			}
		}
	}
	if right > n.Value {
		if (right-n.key+1 < 350) && (left-1 <= n.Value) {
			n.Value = right
		} else {
			rightPair[1] = right
			if left <= n.Value {
				rightPair[0] = n.Value + 1
			} else {
				rightPair[0] = left
			}
		}
	}
	if leftPair[0] != 0 {
		n.left = n.left.insertMissingPair(leftPair[0], leftPair[1])
	}
	if rightPair[0] != 0 {
		n.right = n.right.insertMissingPair(rightPair[0], rightPair[1])
	}
	return n.rebalanceTree()
}

// Displays nodes left-depth first (used for debugging)
func (n *AVLNode) displayNodesInOrder() {
	if n.left != nil {
		n.left.displayNodesInOrder()
	}
	fmt.Printf("[%v-%v] ", n.key, n.Value)
	if n.right != nil {
		n.right.displayNodesInOrder()
	}
}

func (n *AVLNode) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *AVLNode) recalculateHeight() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// Checks if node is balanced and rebalance
func (n *AVLNode) rebalanceTree() *AVLNode {
	if n == nil {
		return n
	}
	n.recalculateHeight()

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor == -2 {
		// check if child is left-heavy and rotateRight first
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	} else if balanceFactor == 2 {
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	}
	return n
}

// Rotate nodes left to balance node
func (n *AVLNode) rotateLeft() *AVLNode {
	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Rotate nodes right to balance node
func (n *AVLNode) rotateRight() *AVLNode {
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Finds the smallest child (based on the key) for the current node
func (n *AVLNode) findSmallest() *AVLNode {
	if n.left != nil {
		return n.left.findSmallest()
	} else {
		return n
	}
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
