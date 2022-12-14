package utils

type LFUCache struct {
	capacity  int
	size      int
	minFreq   int
	nodeMap   map[DataKey]*Node
	freqLists map[int]*DoublyLinkedList
}

const KiloByte = 1024
const MegaByte = 1024 * KiloByte
const GigaByte = 1024 * MegaByte

func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity:  capacity,
		size:      0,
		minFreq:   0,
		nodeMap:   make(map[DataKey]*Node),
		freqLists: make(map[int]*DoublyLinkedList),
	}
}

func (cache *LFUCache) SetDefault(freq int) {
	if _, ok := cache.freqLists[freq]; !ok {
		cache.freqLists[freq] = NewDLL()
	}
}

func (cache *LFUCache) Contains(key DataKey) bool {
	_, ok := cache.nodeMap[key]
	return ok
}

func (cache *LFUCache) Get(key DataKey) (DataValue, bool) {
	if cache.minFreq == 0 || !cache.Contains(key) {
		return DEFAULT_VALUE, false
	}

	node := cache.nodeMap[key]
	cache.freqLists[node.freq].Remove(node)
	cache.SetDefault(node.freq + 1)
	cache.freqLists[node.freq+1].Push(node)

	if node.freq == cache.minFreq && cache.freqLists[node.freq].Len() == 0 {
		cache.minFreq++
	}

	node.freq++
	return node.val, true
}

func (cache *LFUCache) Put(key DataKey, value DataValue) {
	if cache.capacity == 0 {
		return
	}

	if cache.Contains(key) {
		cache.Get(key)
		cache.size -= len(cache.nodeMap[key].val)
		cache.size += len(value)
		cache.nodeMap[key].val = value
	} else {
		node := &Node{key: key, val: value, freq: 1}
		cache.size += len(value)
		cache.nodeMap[key] = node
		cache.SetDefault(1)
		cache.freqLists[1].Push(node)
		cache.minFreq = 1
	}

	cache.RecycleLRU()
}

func (cache *LFUCache) RecycleLRU() {
	for cache.size > cache.capacity {
		popped := cache.freqLists[cache.minFreq].Pop()
		delete(cache.nodeMap, popped.key)
		cache.size -= len(popped.val)
	}
}
