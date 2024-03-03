package main

import (
	"encoding/binary"
	"math/big"

	"github.com/spaolacci/murmur3"
)

func hash(input string) string {
	h32 := murmur3.New32()
	h32.Write([]byte(input))
	bt := make([]byte, 4)
	binary.LittleEndian.PutUint32(bt, h32.Sum32())
	var i big.Int
	i.SetBytes(bt)
	return i.Text(62)
}
