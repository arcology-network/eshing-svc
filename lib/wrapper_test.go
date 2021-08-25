package lib

import (
	"fmt"
	"testing"
)

func Test_roothash(t *testing.T) {
	cfgpath := "./config.json"
	h := HpmtStart(cfgpath)

	tx := []byte{}
	for j := 0; j < HashType_256*5; j++ {
		tx = append(tx, byte(j))
	}
	hash, err := h.RootHash(tx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("hash=%x,err=%v\n", hash, err)

	t.Log("Test_roothash success")

	h.TreeStop()
}

func Test_roothashRaw(t *testing.T) {
	cfgpath := "./config.json"
	h := HpmtStart(cfgpath)

	tx := []byte{}
	lengs := make([]int, 5)
	for j := 0; j < HashType_256*5; j++ {
		tx = append(tx, byte(j))

	}
	for i := 0; i < 5; i++ {
		lengs[i] = HashType_256
	}
	hash, err := h.RootHashRaw(tx, lengs)
	if err != nil {
		panic(err)
	}
	logs := h.GetLog()
	fmt.Printf("hash=%x,err=%v,logs=%v\n", hash, err, logs)

	t.Log("Test_roothash success")

	h.TreeStop()
}
