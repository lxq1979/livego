package pio

import (
	"fmt"
	"testing"
)

func TestU16BE(t *testing.T) {
	str := "ABC"
	character := []byte(str)
	fmt.Printf("string %s convert to U16BE is %d\n", str, U16BE(character))
}

func TestU32BE(t *testing.T) {
	str := "ABCD"
	character := []byte(str)
	fmt.Printf("string %s convert to U32BE is %d\n", str, U32BE(character))
}

func TestU64BE(t *testing.T) {
	str := "ABCD1234"
	character := []byte(str)
	fmt.Printf("string %s convert to U64BE is %d\n", str, U64BE(character))
}
