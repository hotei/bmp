// util.go

package bmp

import (
	"fmt"
	"log"
	"os"
)

func inRangeByte(a, b, c byte) bool {
	if a > c { // swap ends if necessary
		a, c = c, a
	}
	if b < a {
		return false
	}
	if b > c {
		return false
	}
	return true
}

func FatalError(err error) {
	fmt.Printf("%v\n", err)
	os.Exit(1)
}

// reverse function is LSBytesFromUint32
// test is
func Uint32FromLSBytes(b []byte) uint32 {
	if len(b) != 4 {
		log.Panicf("bmp: Slice must be exactly 4 bytes\n")
	}
	var rc uint32
	rc = uint32(b[3])
	rc <<= 8
	rc |= uint32(b[2])
	rc <<= 8
	rc |= uint32(b[1])
	rc <<= 8
	rc |= uint32(b[0])
	return rc
}

// reverse function is LSBytesFromUint16
// test is
func Uint16FromLSBytes(b []byte) uint16 {
	if len(b) != 2 {
		log.Panicf("bmp: Slice must be exactly 2 bytes\n")
	}
	var rc uint16
	rc |= uint16(b[1])
	rc <<= 8
	rc |= uint16(b[0])
	return rc
}

func Int64FromLSBytes(b []byte) int64 {
	if len(b) != 8 {
		log.Panicf("bmp: Slice must be exactly 8 bytes\n")
	}
	// unsure which is faster but we have to use second version on < 64bit machines anyway
	// could handle with build flags but too much trouble for now
	//	return  int32(b[0]) | (int32(b[1]) << 8) | (int32(b[2])<<16) | (int32(b[3]) << 24) |
	//	(int32(b[4]) <<32) | (int32(b[5]) << 40) | (int32(b[6])<<48) | (int32(b[7]) << 56)
	rv := int64(0)
	for i := 7; i >= 0; i-- {
		rv += int64(b[i])
		if i == 0 {
			break
		}
		rv <<= 8
	}
	return rv
}

func Int32FromLSBytes(b []byte) int32 {
	if len(b) != 4 {
		log.Panicf("bmp: Slice must be exactly 4 bytes\n")
	}
	return int32(b[0]) | (int32(b[1]) << 8) | (int32(b[2]) << 16) | (int32(b[3]) << 24)
}

func Int16FromLSBytes(b []byte) int16 {
	if len(b) != 2 {
		log.Panicf("bmp: Slice must be exactly 2 bytes\n")
	}
	return int16(b[0]) | (int16(b[1]) << 8)
}
