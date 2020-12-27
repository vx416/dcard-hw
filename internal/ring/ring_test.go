package ring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRingWithData(replicaNum int64, testdata ...string) (*HashRing, error) {
	ring := NewDefault(replicaNum)
	err := ring.AddNodes(testdata...)
	return ring, err
}

func TestRing_AddRemoveNodes(t *testing.T) {
	type removeCase struct {
		node string
		err  bool
	}

	testcases := []struct {
		name        string
		nodes       []string
		replicaNum  int64
		removeCases []removeCase
	}{
		{
			name: "testcase_1", nodes: []string{"pig", "cat", "dog", "monkey", "tiger"},
			replicaNum: 3, removeCases: []removeCase{{"pig", false}, {"monkey", false}},
		},
		{
			name: "testcase_2", nodes: []string{"pig"},
			replicaNum: 3, removeCases: []removeCase{{"cat", true}},
		},
		{
			name: "testcase_3", nodes: []string{"pig"},
			replicaNum: 3, removeCases: []removeCase{{"pig", false}, {"pig", true}},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ring, err := newRingWithData(testcase.replicaNum, testcase.nodes...)
			assert.NoError(t, err)
			assert.Len(t, ring.Nodes(), len(testcase.nodes))
			assert.Len(t, ring.nodes.keys, len(testcase.nodes)*int(testcase.replicaNum))

			removed := 0
			for _, removeCase := range testcase.removeCases {
				err = ring.RemoveNodes(removeCase.node)
				if removeCase.err {
					assert.Error(t, err)
					return
				}
				removed++
				assert.NoError(t, err)
				assert.Len(t, ring.nodes.keys, (len(testcase.nodes)-(removed))*int(testcase.replicaNum))
				assert.NotContains(t, ring.Nodes(), removeCase.node)
				_, exits := ring.nodes.idToKeys[removeCase.node]
				assert.False(t, exits)
			}
		})
	}
}

func TestRing_FindNode(t *testing.T) {
	testcases := []struct {
		name     string
		replicas int64
		nodes    []string
		base     string
		err      bool
	}{
		{name: "testcase_1", replicas: 3, nodes: []string{"pig", "cat", "dog", "monkey", "tiger"}, base: "david"},
		{name: "testcase_2", replicas: 1, nodes: []string{"pig", "cat", "dog", "monkey", "tiger"}, base: "小明"},
		{name: "testcase_3", replicas: 5, nodes: []string{"pig"}, base: "小明"},
		{name: "testcase_4", replicas: 5, nodes: []string{}, base: "小明", err: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ring, err := newRingWithData(tc.replicas, tc.nodes...)
			assert.NoError(t, err)
			node, err := ring.GetNode(tc.base)
			if tc.err {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Contains(t, ring.Nodes(), node)
			node2, err := ring.GetNode(tc.base)
			assert.NoError(t, err)
			assert.Equal(t, node, node2)
		})
	}
}
