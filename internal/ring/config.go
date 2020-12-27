package ring

import (
	"fmt"

	"github.com/spaolacci/murmur3"
)

// NodeIDGenerator 組合 nodeID 以及 replicasNo helper function
type NodeIDGenerator func(nodeID string, replicasNo int) string

// Config ring configuration
type Config struct {
	numberOfReplicas int64
	hashFunc         HashFunc
	nodeIDFunc       NodeIDGenerator
}

// DefaultConfig 預設 config
func DefaultConfig() *Config {
	return &Config{
		numberOfReplicas: 3,
		hashFunc:         MurMurHashing,
		nodeIDFunc:       defaultNodeIDGen,
	}
}

func defaultNodeIDGen(nodeID string, replicNo int) string {
	return fmt.Sprintf("%s-%d", nodeID, replicNo)
}

// HashFunc hashing function
type HashFunc func(key string) HashKey

// MurMurHashing murmur hasinh
func MurMurHashing(key string) HashKey {
	h := murmur3.New32()
	h.Write([]byte(key))
	return hashKey(h.Sum32())
}
