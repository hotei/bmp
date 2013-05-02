// bmp_test.go (c) 2013 David Rook

package bmp

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"testing"
)

const testDir = "./testdata/"

var (
	testImages = []string{
		"bit1bw-rnr.bmp",          // 1 bit per pixel, uncompressed, 2 entry color table - working with my code
		"bit1color2.bmp",          // working with my code
		"bit4-test.bmp",           // working with my code
		"bit4comp-test.bmp",       // RLE4 working with my code
		"bit8-gray-rnr.bmp",       // working with my code
		"bit8comp-test.bmp",       // RLE8 working with my code
		"bit8comp-rnr.bmp",        // RLE8 working with my code
		"bit8-test.bmp",           // working with my code - original failed to read header correctly
		"bit24uncomp-marbles.bmp", // working large 24 bit uncompressed
		"bit24uncomp-rnr.bmp",     // 24 bit per pixel, uncompressed, working with original
		"bit24-teststrip.bmp",     // working 24 bit uncompressed
		"whirlpool.jpg",           // fails as required with bad magic if called with bmp.Decode(), ok with image.Decode()
		"notBMP.bmp",              // fails as required
		//"bit24-test.bmp",	// air-moz
		//"bit16-test.bmp",
		//"bit32-test.bmp",
	}
)

func Test_0001(t *testing.T) {
	fmt.Printf("Test_0001\n")
	errCt := 0
	for i := 0; i < len(testImages); i++ {
		fname := testImages[i]
		fmt.Printf("%d Working on %s...\n", i, fname)

		bf, err := os.Open(testDir + fname)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		img, _, err := image.Decode(bf)
		img = img // LINT
		if err != nil {
			fmt.Printf("%s Failed\n", fname)
			t.Errorf("%v\n", err)
			errCt++
		} else {
			bf.Close()
			fmt.Printf("%s Passed\n", fname)
		}
	}
	if errCt == 0 {
		fmt.Printf("Test_0001 Pass\n")
	} else {
		fmt.Printf("Test_0001 Fail\n")
	}
}
