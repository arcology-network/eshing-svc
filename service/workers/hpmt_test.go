package workers

import (
	"fmt"
	"testing"
)

func Test_array(t *testing.T) {

	type Node struct {
		Key   string
		Value []byte
	}
	nodes := []Node{
		Node{
			Key:   "1",
			Value: []byte{11, 12},
		},
		Node{
			Key:   "2",
			Value: []byte{13, 14},
		},
		Node{
			Key:   "1",
			Value: []byte{15, 16},
		},
		Node{
			Key:   "2",
			Value: []byte{17, 18},
		},
		Node{
			Key:   "1",
			Value: []byte{19, 20},
		},
	}

	accts := []string{}
	datas := map[string][][]byte{}

	for _, node := range nodes {
		if data, ok := datas[node.Key]; ok {
			data = append(data, node.Value)
			datas[node.Key] = data
		} else {
			accts = append(accts, node.Key)
			datas[node.Key] = [][]byte{node.Value}
		}
	}

	fmt.Printf("accts=%v\n", accts)
	for k, v := range datas {
		fmt.Printf("k=%v,v=%v\n", k, v)
	}

}
