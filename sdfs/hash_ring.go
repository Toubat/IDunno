package sdfs

import (
	"fmt"
	"mp4/api"
	"mp4/utils"

	"sort"
	"strings"
)

type HashNode struct {
	HashKey int
	Process *api.Process
}

type HashRing []HashNode

const HASH_SIZE = 1024

func NewHashRing() *HashRing {
	return &HashRing{}
}

func (hr *HashRing) Refresh(processes []*api.Process) {
	*hr = make(HashRing, 0)
	keys := make(map[int]bool, 0)

	for _, p := range processes {
		// binary probing whenever hash collision, expected O(logn)
		i := 0
		for i < HASH_SIZE {
			if _, found := keys[(utils.Hash(p.Address())|i)%HASH_SIZE]; found {
				i++
			} else {
				break
			}
		}

		hashKey := (utils.Hash(p.Address()) | i) % HASH_SIZE
		keys[hashKey] = true

		*hr = append(*hr, HashNode{
			HashKey: hashKey,
			Process: p,
		})
	}
	sort.Sort(hr)
}

func (hr *HashRing) FindReplicas(key string, numReplicas int) []*api.Process {
	hashKey := utils.Hash(key) % HASH_SIZE
	replicas := make([]*api.Process, 0)

	// find insert position
	start := 0
	for i, node := range *hr {
		if node.HashKey >= hashKey {
			start = i
			break
		}
	}

	i := 0
	for i < min(len(*hr), numReplicas) { // get numReplicas replicas
		replicas = append(replicas, (*hr)[(start+i)%len(*hr)].Process)
		i++
	}

	return replicas
}

func (hr *HashRing) FindSuccessors(process *api.Process, numSuccessors int) []*api.Process {
	successors := make([]*api.Process, 0)

	index := hr.FindProcessIndex(process)
	if index == -1 {
		return successors
	}

	for i := 0; i < numSuccessors; i++ {
		successor := (*hr)[(index+i+1)%len(*hr)].Process
		if api.IsSameProcess(successor, process) {
			break
		}

		successors = append(successors, successor)
	}

	return successors
}

func (hr *HashRing) FindPredecessor(process *api.Process) *api.Process {
	index := hr.FindProcessIndex(process)
	if index == -1 || hr.Len() == 1 {
		return nil
	}

	return (*hr)[(index-1+hr.Len())%hr.Len()].Process
}

func (hr *HashRing) FindProcessIndex(process *api.Process) int {
	hashKey := utils.Hash(process.Address()) % HASH_SIZE

	for i, node := range *hr {
		if node.HashKey == hashKey {
			return i
		}
	}

	return -1
}

func (hr *HashRing) GetRouteProcess(key string) *api.Process {
	if hr.Len() == 0 {
		return nil
	}

	hashKey := utils.Hash(key) % HASH_SIZE

	for _, node := range *hr {
		if node.HashKey >= hashKey {
			return node.Process
		}
	}

	return (*hr)[0].Process
}

func (hr *HashRing) DebugValue() string {
	// <process_hash, port> ---> <hash_key, port> ---> ...
	values := make([]string, 0)
	for _, node := range *hr {
		values = append(values, fmt.Sprintf("<%d,%v>", node.HashKey, node.Process.Port))
	}

	// join by --->
	return strings.Join(values, " ---> ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (hr *HashRing) Len() int {
	return len(*hr)
}

func (hr *HashRing) Less(i, j int) bool {
	return (*hr)[i].HashKey < (*hr)[j].HashKey
}

func (hr *HashRing) Swap(i, j int) {
	(*hr)[i], (*hr)[j] = (*hr)[j], (*hr)[i]
}
