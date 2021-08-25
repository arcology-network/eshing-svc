package lib

/*
#cgo LDFLAGS: -L./ -leshing
#include "./hpmt.external.h"
#include "./hpmt_external_test.h" // optional, for testing only
*/
import "C"

import (
	"errors"
	"unsafe"
)

const (
	HashType_160 = 20
	HashType_256 = 32
)

type Hpmt struct {
	hpmt    unsafe.Pointer
	batchId C.uint64_t
}

func (h *Hpmt) RootHash(src []byte) ([]byte, error) {
	totallen := len(src)
	if totallen%HashType_256 != 0 {
		return nil, errors.New("input data length err")
	}
	length := C.uint64_t(totallen / HashType_256)
	c_char := (*C.char)(unsafe.Pointer(&src[0]))

	ahash := make([]byte, HashType_256)
	a_char := (*C.char)(unsafe.Pointer(&ahash[0]))

	var err error
	h.batchId, err = C.BatchInsert(h.hpmt, c_char, length, a_char)

	return ahash, err
}

func (h *Hpmt) RootHashRaw(src []byte, lengths []int) ([]byte, error) {
	c_char := (*C.char)(unsafe.Pointer(&src[0]))

	cunts := len(lengths)
	c_count := C.uint64_t(cunts)

	clength := make([]C.uint64_t, cunts)

	for i := 0; i < len(lengths); i++ {
		clength[i] = C.uint64_t(lengths[i])
	}
	c_length := (*C.uint64_t)(unsafe.Pointer(&clength[0]))

	ahash := make([]byte, HashType_256)
	a_char := (*C.char)(unsafe.Pointer(&ahash[0]))
	var err error
	h.batchId, err = C.BatchInsertFromRaw(h.hpmt, c_char, c_length, c_count, a_char)
	return ahash, err

}

func (h *Hpmt) GetLog() string {
	buffersize := 4 * 1024
	c_size := C.uint64_t(buffersize)
	buffer := make([]byte, buffersize)
	log_char := (*C.char)(unsafe.Pointer(&buffer[0]))
	C.GetLogMsg(h.hpmt, log_char, c_size)

	var rstsize int
	for i, c := range buffer {
		if c == 0 {
			rstsize = i
			break
		}
	}
	return string(buffer[:rstsize])
}

func (h *Hpmt) Confirm() (bool, error) {
	isSucc, err := C.Confirm(h.hpmt, h.batchId)
	return bool(isSucc), err
}
func (h *Hpmt) TreeStop() {
	C.Stop(h.hpmt)
}
func HpmtStart(Cfgpath string) *Hpmt {
	h := Hpmt{}
	h.hpmt = C.Start(C.CString(Cfgpath))
	return &h
}
