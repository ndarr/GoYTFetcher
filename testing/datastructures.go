package main

import (
	"sync"
	"fmt"
	"strconv"
)

type YoutubeSong struct {
	id    string
	count int
}

func (youtubeSong YoutubeSong) GetUniqueId() string {
	return youtubeSong.id
}

func (youtubeSong YoutubeSong) GetCount() int {
	return youtubeSong.count
}

type rankingInterface interface {
	GetUniqueId() string
	GetCount() int
}

type rankingTree struct {
	start *node
	size  int

	lock sync.RWMutex
}

func (tree *rankingTree) insert(rankingItem rankingInterface) {
	tree.lock.Lock()
	defer tree.lock.Unlock()

	tree.size++
	if tree.start == nil {
		tree.start = &node{rankingItem: rankingItem}
		return
	}
	tree.start.insert(rankingItem)
}

func (tree *rankingTree) delete(rankingItem rankingInterface) bool {
	tree.lock.Lock()
	defer tree.lock.Unlock()

	if tree.start == nil {
		return false
	}
	if tree.start.rankingItem.GetUniqueId() == rankingItem.GetUniqueId() {
		tree.size--
		tree.start = createReplaceNode(tree.start)
		return true
	}
	if tree.start.delete(rankingItem) {
		tree.size--
		return true
	}
	return false
}

func (tree *rankingTree) getLowest() rankingInterface {
	tree.lock.RLock()
	defer tree.lock.RUnlock()

	if tree.start == nil {
		return nil
	}
	return tree.start.getLowest()
}

func (tree *rankingTree) getSize() int {
	tree.lock.RLock()
	defer tree.lock.RUnlock()

	return tree.size
}

type node struct {
	rankingItem rankingInterface
	left, right *node
}

func (nodeLeaf *node) insert(rankingItem rankingInterface) {
	if rankingItem.GetCount() < nodeLeaf.rankingItem.GetCount() {
		if nodeLeaf.left == nil {
			nodeLeaf.left = &node{rankingItem: rankingItem}
		} else {
			nodeLeaf.left.insert(rankingItem)
		}
	} else {
		if nodeLeaf.right == nil {
			nodeLeaf.right = &node{rankingItem: rankingItem}
		} else {
			nodeLeaf.right.insert(rankingItem)
		}
	}
}

func (nodeLeaf *node) delete(rankingItem rankingInterface) bool {
	if nodeLeaf.left != nil &&
		nodeLeaf.left.rankingItem.GetUniqueId() == rankingItem.GetUniqueId() {
		nodeLeaf.left = createReplaceNode(nodeLeaf.left)
		return true
	} else if nodeLeaf.right != nil &&
		nodeLeaf.right.rankingItem.GetUniqueId() == rankingItem.GetUniqueId() {
		nodeLeaf.right = createReplaceNode(nodeLeaf.right)
		return true
	}

	if rankingItem.GetCount() < nodeLeaf.rankingItem.GetCount() {
		if nodeLeaf.left != nil {
			return nodeLeaf.left.delete(rankingItem)
		}
	} else if nodeLeaf.right != nil {
		return nodeLeaf.right.delete(rankingItem)
	}

	return false
}

func (nodeLeaf *node) getLowest() rankingInterface {
	if nodeLeaf.left == nil {
		return nodeLeaf.rankingItem
	}
	return nodeLeaf.left.getLowest()
}

func createReplaceNode(replacedNode *node) *node {
	newNode := replacedNode.right
	if newNode == nil {
		return replacedNode.left
	}
	if replacedNode.left == nil {
		return newNode
	}

	if newNode.left == nil {
		newNode.left = replacedNode.left
		return newNode
	}
	lastLeftNode := newNode.left
	for lastLeftNode.left != nil {
		lastLeftNode = lastLeftNode.left
	}
	lastLeftNode.left = replacedNode.left
	return newNode
}

func (nodeLeaf *node) print(prefix string, isTail bool, position string) {
	if nodeLeaf == nil {
		return
	}
	message := "├── "
	if isTail {
		message = "└── "
	}

	fmt.Println(prefix + message + position + ": " + nodeLeaf.rankingItem.GetUniqueId() +
		" " + strconv.Itoa(nodeLeaf.rankingItem.GetCount()))
	message = "│   "
	if isTail {
		message = "    "
	}
	if nodeLeaf.left != nil {
		nodeLeaf.left.print(prefix+message, nodeLeaf.right == nil, "left")
	}
	if nodeLeaf.right != nil {
		nodeLeaf.right.print(prefix+message, true, "right")
	}
}
