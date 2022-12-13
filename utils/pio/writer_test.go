package pio

import (
	"fmt"
	"testing"
)

func TestPutU32BE(t *testing.T) {
	var u32 uint32 = uint32(1094861636)
	var out []byte = []byte("ZZZZ")
	PutU32BE(out, u32)
	fmt.Printf("u64 %d converto []byte is %s", u32, out)
}

func TestPutU64BE(t *testing.T) {
	var u64 uint64 = uint64(4702394921090429748)
	var out []byte = []byte("12345678")
	PutU64BE(out, u64)
	fmt.Printf("u64 %d converto []byte is %s", u64, out)
}
