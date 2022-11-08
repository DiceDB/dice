package core

import (
	"fmt"
)

type freqLinkedList struct {
	head, tail *freqNode
}

// node to store the list of key which are val times accessed
type freqNode struct {
	val        int16
	next, prev *freqNode
	head, tail *keyNode
}

// node to store the metadata of key and frequency count
type keyNode struct {
	key          string
	up, down     *keyNode
	freqHeadnode *freqNode
}

// keyMap is to store the key and corresponding key-node
var keyMap map[string]*keyNode

// freqList is doubly linked list to store the frequency nodes
var freqList *freqLinkedList

// function to initialize the keyMap and freqList
func initLFU() {
	keyMap = make(map[string]*keyNode)
	freqList = &freqLinkedList{head: nil, tail: nil}
}

// function to get freqNode if present inside the freqList of
func (freql *freqLinkedList) getOrCreateFreqNode(currFreqNode *freqNode, freq int16) *freqNode {
	// if head is null then create node at head position with given freq
	if freql.head == nil {
		newFreqNode := &freqNode{
			val: freq,
		}
		freql.head = newFreqNode
		freql.tail = newFreqNode
		return freql.head
	}

	currNode := freql.head

	// if freq of head node is greater then required freq node.
	// create new head of list
	if currNode.val > freq {
		newFreqNode := &freqNode{
			val: freq,
		}
		newFreqNode.next = freql.head
		freql.head.prev = newFreqNode
		freql.head = newFreqNode
		return freql.head
	}

	// if we need to insert immediate neighbour node (val=>val+1)
	// then we will create neighbour node from curent node
	// else find the position where to insert the new node
	// i.e node which has higher val then freq (upper bound)
	if currFreqNode != nil && freq == currFreqNode.val+1 {
		currNode = currFreqNode.next
	} else {
		for currNode != nil && currNode.val < freq {
			currNode = currNode.next
		}
	}

	// if node able to find upper bound of given node
	// add new node at tail of list
	if currNode == nil {
		newFreqNode := &freqNode{
			val: freq,
		}

		freql.tail.next = newFreqNode
		newFreqNode.prev = freql.tail
		freql.tail = newFreqNode
		return freql.tail
	}

	if currNode.val == freq {
		return currNode
	}

	newFreqNode := &freqNode{
		val: freq,
	}
	// insert a node between 2 nodes
	newFreqNode.next = currNode
	currNode.prev.next = newFreqNode
	currNode.prev = newFreqNode

	return newFreqNode
}

// function to insert key in doubly linkedlist
//  1 --> nil                    1 --> nil
//  |         ==>  remove y ==>  |
//  x                            x
//  							 |
//								 y
func (freql *freqLinkedList) Insert(key string) *keyNode {
	//create freq nod at freq 1
	currFreqNode := freqList.getOrCreateFreqNode(nil, 1)

	newKeyNode := &keyNode{
		key:          key,
		up:           currFreqNode.tail,
		freqHeadnode: currFreqNode,
	}

	// if no node is present at this freq
	if currFreqNode.head == nil {
		currFreqNode.head = newKeyNode
		currFreqNode.tail = newKeyNode
	} else {
		currFreqNode.tail.down = newKeyNode
		newKeyNode.up = currFreqNode.tail
		currFreqNode.tail = newKeyNode
	}

	//add key to keymap
	keyMap[key] = newKeyNode

	return newKeyNode
}

// function to remove frequency node
//  1 -->  2 --> 3 --> nil                    1 --> 3 --> nil
//  |            |         ==>  remove 2 ==>  |     |
//  x            y                            x     y
func (freql *freqLinkedList) removeFreqNode(fNode *freqNode) {
	// check if given node is head node then replace head node with next node
	// else if chek if node is tail node then replace tail with prev node
	// else connect neighbouring node
	if freql.head == fNode {
		freql.head = freql.head.next
		if freql.head != nil {
			freql.head.prev = nil
		}
	} else if freql.tail == fNode {
		freql.tail = freql.tail.prev
		freql.tail.next = nil
	} else {
		fNode.prev.next = fNode.next
		fNode.next.prev = fNode.prev
	}
}

// function to reallocation key when access freq is changed
//  1 --> 3 --> nil                             2 --> 3 --> nil
//  |     |         ==>  reallocate x at 2 ==>  |     |
//  x     y                                     x     y
func (freql *freqLinkedList) reallocateKeyNode(kNode *keyNode, newFreqNode *freqNode) *keyNode {
	// connect neighbouring node to remove from the list
	if kNode.up != nil {
		kNode.up.down = kNode.down
	}
	if kNode.down != nil {
		kNode.down.up = kNode.up
	}

	// if given node  is head node then change the head
	if kNode.freqHeadnode.head == kNode {
		kNode.freqHeadnode.head = kNode.freqHeadnode.head.down
	}
	// if it is taile node then update the new tail
	if kNode.freqHeadnode.tail == kNode {
		kNode.freqHeadnode.tail = kNode.freqHeadnode.tail.up
	}

	// after removeal if freq node is empty remove it
	if kNode.freqHeadnode.tail == nil && kNode.freqHeadnode.head == nil {
		freqList.removeFreqNode(kNode.freqHeadnode)
	}

	kNode.up = nil
	kNode.down = nil
	kNode.freqHeadnode = newFreqNode

	// insert node at new location
	if newFreqNode.head == nil {
		newFreqNode.head = kNode
		newFreqNode.tail = kNode
	} else {
		newFreqNode.tail.down = kNode
		newFreqNode.tail = newFreqNode.tail.down
	}
	return kNode
}

// function to insert key node in given freq node
//  1 --> 3 --> nil                         2 --> 3 --> nil
//  |     |         ==> insert x at 2 ==>   |     |
//  x     y                                 x     y
func (freql *freqLinkedList) insertKeyNodeAt(kNode *keyNode, freq int16) *keyNode {
	// get new freq node of get id exists
	currFreqNode := freql.getOrCreateFreqNode(kNode.freqHeadnode, freq)
	// append key node at tail
	kNode = freql.reallocateKeyNode(kNode, currFreqNode)
	return kNode
}

// function to rebalance list based on frequency pattern of given key
//  1 --> 3 --> nil                            2 --> 3 --> nil
//  |     |         ==> rebalance x to 2 ==>   |     |
//  x     y                                    x     y
func (freql *freqLinkedList) ReBalanceList(key string) *freqNode {
	// get the key node and corresponding freq node
	currKeyNode := keyMap[key]
	headFreqNode := currKeyNode.freqHeadnode
	// insert key node at immidate next value and reallrange it list
	currKeyNode = freql.insertKeyNodeAt(currKeyNode, headFreqNode.val+1)
	return currKeyNode.freqHeadnode
}

// function to remove key by name
//  1 --> 3 --> nil                         3 --> nil
//  |     |         ==> remove x ==>        |
//  x     y                                 y
func (freql *freqLinkedList) RemoveByKey(key string) bool {
	if _, ok := keyMap[key]; ok {
		currKeyNode := keyMap[key]
		newFreqNode := freql.getOrCreateFreqNode(nil, -1)
		freql.reallocateKeyNode(currKeyNode, newFreqNode)
		freql.removeFreqNode(newFreqNode)
		delete(keyMap, key)
		return true
	}
	return false
}

// function to remove list frequently used key
//  1 --> 3 --> nil                         3 --> nil
//  |     |         ==> remove ==>          |
//  x     y                                 y

func (freql *freqLinkedList) RemoveLRU() *keyNode {
	// remove first node form head
	if freql.head != nil {
		currFreqNode := freql.head
		nodeToRemove := currFreqNode.head

		delete(keyMap, nodeToRemove.key)

		currFreqNode.head = currFreqNode.head.down

		if currFreqNode.head == nil {
			freql.removeFreqNode(currFreqNode)
		} else {
			currFreqNode.head.up = nil
		}
		return nodeToRemove
	}
	return nil
}

// function to debug the list operations and printing the tree
func (freql *freqLinkedList) printList() {
	currNode := freql.head
	fmt.Printf("============LIST DATA===============\n")
	for currNode != nil {
		fmt.Printf("freq %d: ", currNode.val)

		curFreqNode := currNode.head
		for curFreqNode != nil {
			fmt.Printf("%s -> ", curFreqNode.key)
			curFreqNode = curFreqNode.down
		}
		fmt.Printf("nil\n")
		currNode = currNode.next
	}
	fmt.Printf("==============END==================\n\n\n")
}
