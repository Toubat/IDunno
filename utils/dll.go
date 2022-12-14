package utils

type DataKey string

type DataValue []byte

var DEFAULT_KEY DataKey = ""
var DEFAULT_VALUE DataValue = make(DataValue, 0)

type Node struct {
	key  DataKey
	val  DataValue
	freq int
	prev *Node
	next *Node
}

type DoublyLinkedList struct {
	size int
	head *Node
	tail *Node
}

func NewDLL() *DoublyLinkedList {
	dll := DoublyLinkedList{
		size: 0,
		head: &Node{key: DEFAULT_KEY, val: DEFAULT_VALUE, freq: 1},
		tail: &Node{key: DEFAULT_KEY, val: DEFAULT_VALUE, freq: 1},
	}
	dll.head.next = dll.tail
	dll.tail.prev = dll.head

	return &dll
}

func (dll *DoublyLinkedList) Len() int {
	return dll.size
}

func (dll *DoublyLinkedList) Push(node *Node) {
	dll.size++
	prev, curr := dll.tail.prev, dll.tail
	prev.next = node
	curr.prev = node
	node.prev = prev
	node.next = curr
}

func (dll *DoublyLinkedList) Pop() *Node {
	return dll.Remove(dll.head.next)
}

func (dll *DoublyLinkedList) Remove(node *Node) *Node {
	if dll.size == 0 {
		return nil
	}

	dll.size--
	node.prev.next = node.next
	node.next.prev = node.prev
	node.prev, node.next = nil, nil
	return node
}
