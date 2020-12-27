package ring

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/vx416/dcard-work/internal/app"
)

// New new a hashring
func New(cfg *Config) *HashRing {
	return &HashRing{
		Config: cfg,
		nodes:  newNodes(),
	}

}

// NewDefault new default hashing
func NewDefault(replicaCount int64) *HashRing {
	config := DefaultConfig()
	if replicaCount != 0 {
		config.numberOfReplicas = replicaCount
	}

	return &HashRing{
		Config: config,
		nodes:  newNodes(),
	}
}

func newNodes() *nodes {
	return &nodes{
		ids:      make([]string, 0, 10),
		keys:     NewSortedKeys(),
		idToKeys: make(map[string][]HashKey),
		keyToID:  make(map[uint32]string),
	}
}

// nodes 在 環上的節點，儲存 node 原始資料跟 hash key 的對應關係
type nodes struct {
	ids      []string
	keys     SortedKeys
	idToKeys map[string][]HashKey
	keyToID  map[uint32]string
}

func (ns *nodes) add(key HashKey, nodeID string) error {
	if _, exists := ns.keyToID[key.Val()]; exists {
		return fmt.Errorf("key (%d) collision", key.Val())
	}

	ns.keys.Insert(key)
	if _, ok := ns.idToKeys[nodeID]; !ok {
		ns.idToKeys[nodeID] = make([]HashKey, 0, 5)
		ns.ids = append(ns.ids, nodeID)
	}

	ns.idToKeys[nodeID] = append(ns.idToKeys[nodeID], key)
	ns.keyToID[key.Val()] = nodeID
	return nil
}

func (ns *nodes) findNode(key HashKey) (string, error) {
	hashKey := ns.keys.Find(key)
	if hashKey == nil {
		return "", fmt.Errorf("key(%d) not exists", key.Val())
	}

	nodeID, exits := ns.keyToID[hashKey.Val()]
	if !exits {
		return "", fmt.Errorf("key(%d) not exists", key.Val())
	}
	return nodeID, nil
}

func (ns *nodes) removeNode(nodeID string) error {
	if _, exists := ns.idToKeys[nodeID]; !exists {
		return fmt.Errorf("node(%s) not exists", nodeID)
	}

	keys := ns.idToKeys[nodeID]
	ns.keys.Del(keys...)
	delete(ns.idToKeys, nodeID)
	for _, key := range keys {
		delete(ns.keyToID, key.Val())
	}
	nodeIndex := 0
	for i, id := range ns.ids {
		if id == nodeID {
			nodeIndex = i
		}
	}
	oldIDs := ns.ids
	ns.ids = make([]string, len(oldIDs)-1)
	copy(ns.ids, oldIDs[:nodeIndex])
	copy(ns.ids[nodeIndex:], oldIDs[nodeIndex+1:])
	return nil
}

var _ app.Ringer = (*HashRing)(nil)

// HashRing hashing ring
type HashRing struct {
	sync.RWMutex
	*Config
	nodes *nodes
}

func (ring *HashRing) String() string {
	ring.RLock()
	defer ring.RUnlock()

	var str strings.Builder
	firstNode := ""
	for i, key := range ring.nodes.keys {
		nodeID := ring.nodes.keyToID[key.Val()]
		nodeStr := nodeID + "(" + strconv.Itoa(int(key.Val())) + ")"
		if i == 0 {
			firstNode = nodeStr
		}
		str.WriteString(nodeStr + " -> ")
	}
	str.WriteString(firstNode)

	return str.String()
}

// AddNodes 加入多個 node 進 ring
func (ring *HashRing) AddNodes(nodes ...string) error {
	ring.Lock()
	defer ring.Unlock()

	for _, nodeID := range nodes {
		err := ring.addNode(nodeID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetNode 找出 base string 屬於那個 node
func (ring *HashRing) GetNode(base string) (string, error) {
	ring.RLock()
	defer ring.RUnlock()

	key := ring.hashFunc(base)
	return ring.nodes.findNode(key)
}

// Nodes 列出目前有的 nodes
func (ring *HashRing) Nodes() []string {
	ring.RLock()
	defer ring.RUnlock()

	return ring.nodes.ids
}

// ResetNodes reset 環中的所有 nodes
func (ring *HashRing) ResetNodes(nodes ...string) error {
	ring.Lock()
	defer ring.Unlock()

	ring.nodes = newNodes()
	for _, nodeID := range nodes {
		err := ring.addNode(nodeID)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveNodes 移除環中的多個 nodes
func (ring *HashRing) RemoveNodes(nodes ...string) error {
	ring.Lock()
	defer ring.Unlock()

	for _, node := range nodes {
		err := ring.nodes.removeNode(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ring *HashRing) addNode(node string) error {
	keys := make([]HashKey, ring.numberOfReplicas)

	for i := range keys {
		nodeID := ring.nodeIDFunc(node, i)
		key := ring.hashFunc(nodeID)
		err := ring.nodes.add(key, node)
		if err != nil {
			return err
		}
	}
	return nil
}
