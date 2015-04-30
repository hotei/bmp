// bmpRead.go

package bmp

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"log"
)

func ReadBMP(r io.Reader) (bfp *BMP_T, err error) {
	var bf BMP_T
	var bmpFileHdr BITMAPFILEHEADER_T
	var bmpInfoHdr BITMAPINFOHEADER_T

	bfp = &bf

	allBmpBytes, err := ioutil.ReadAll(r)
	if err != nil {
		log.Printf("ReadAll failed\n")
		return nil, err
	}
	lenBmp := len(allBmpBytes)
	if lenBmp < 54 {
		return nil, ErrShort
	}
	bmpBytes := make([]byte, 54)
	copy(bmpBytes, allBmpBytes[0:54])

	if bytes.Compare(bmpBytes[0:2], MagicBytes) != 0 {
		fmt.Printf("bmp: File ID (magic) bytes must be BM\n")
		return nil, ErrBadMagic
	}

	// MUST check header size first, don't continue if format not supported
	bmpInfoHdr.HdrSize = Uint32FromLSBytes(bmpBytes[14:18])
	switch bmpInfoHdr.HdrSize {
	case 12:
		fmt.Printf("bmp version OS/2 v1\n")
		return nil, ErrOS21NotSupported
	case 40:
		// by far most common version only one supporte at present
		fmt.Printf("bmp version 3\n")
	case 64:
		fmt.Printf("bmp version OS/2 v2\n")
		return nil, ErrOS22NotSupported
	case 108:
		fmt.Printf("bmp version 4\n")
		return nil, ErrV4NotSupported
	case 124:
		fmt.Printf("bmp version 5\n")
		return nil, ErrV5NotSupported
	default:
		fmt.Printf("bmp: can't recognize header size %d\n", bmpInfoHdr.HdrSize)
		return nil, ErrBadHeader
	}
	// since we have good header we proceed to pick it apart, checking
	// for sanity/consistency as we go
	copy(bmpFileHdr.bfMagic[0:2], bmpBytes[0:2])
	bmpFileHdr.bfSize = Uint32FromLSBytes(bmpBytes[2:6])
	// check expected file length against what we actually read in
	if bmpFileHdr.bfSize != uint32(len(allBmpBytes)) {
		fmt.Printf("header says file contains %d bytes, actually read %d\n",
			bmpFileHdr.bfSize, len(allBmpBytes))
		return nil, ErrShort
	}
	if bytes.Compare(bmpBytes[6:10], []byte{0, 0, 0, 0}) != 0 {
		log.Printf("bmp: nonzero bytes in reserved area\n")
		return nil, ErrBadHeader
	}
	// for whatever reason using mapOffset didn't work out
	// I don't know what the offset is from.
	// Apparently not from the start of the file.  In order to find the
	// bitmap we captured the size of the colortable and added it to the end of
	// the infoheader.  That worked.
	mapOffset := Uint32FromLSBytes(bmpBytes[10:14])
	verbose.Printf("mapOffset = %d\n", mapOffset)
	bmpFileHdr.bfOffsetBits = mapOffset

	verbose.Printf("Magic(%x)\n", bmpFileHdr.bfMagic)
	verbose.Printf("FileSize(%d)\n", bmpFileHdr.bfSize)
	verbose.Printf("OffsetBits(%d)\n", bmpFileHdr.bfOffsetBits)

	bmpInfoHdr.Width = Int32FromLSBytes(bmpBytes[18:22])
	verbose.Printf("width(%d)\n", int(bmpInfoHdr.Width))
	// I think neg width is always an err
	if bmpInfoHdr.Width <= 0 {
		fmt.Printf("bmp: width <= 0 ; found %d\n", bmpInfoHdr.Width)
		return nil, ErrGeneric
	}
	bmpInfoHdr.Height = Int32FromLSBytes(bmpBytes[22:26])
	verbose.Printf("height(%d)\n", int(bmpInfoHdr.Height))
	// OS2 can have inverted map so neg height is not an error
	if bmpInfoHdr.Height < 0 {
		verbose.Printf("top->down pixel order found (normal is bottom->up)\n")
	}
	// planes is always 1
	bmpInfoHdr.biPlanes = Uint16FromLSBytes(bmpBytes[26:28])
	if bmpInfoHdr.biPlanes != 1 {
		log.Printf("bmp: Bad number of planes, must be 1 but found %d\n", bmpInfoHdr.biPlanes)
		return nil, ErrBadHeader
	}
	bmpInfoHdr.Depth = Uint16FromLSBytes(bmpBytes[28:30])

	switch bmpInfoHdr.Depth {
	case 1:
		// working
	case 2:
		fmt.Printf("2 bit per pixel not supported (Windows CE only)\n")
		return nil, Err02NotSupported
	case 4:
		// working
	case 8:
		// working
	case 16:
		fmt.Printf("16 bit per pixel not supported\n")
		return nil, Err16NotSupported
	case 24:
		// working
	case 32:
		fmt.Printf("32 bit per pixel not supported\n")
		return nil, Err32NotSupported
	default:
		fmt.Printf("bmp: bad number of bits per pixel, must be 1/2/4/8/16/24/32 but got %d\n", bmpInfoHdr.Depth)
		return nil, ErrBadHeader
	}

	bmpInfoHdr.Compression = Uint32FromLSBytes(bmpBytes[30:34])

	switch bmpInfoHdr.Compression {
	case BI_RGB:
		// uncompressed - working
	case BI_RLE8:
		// RLE-8 - working
	case BI_RLE4:
		// RLE-8 - testing now
	case BI_BITFIELDS:
		fmt.Printf("bmp: BI_BITFIELDS is not handled\n")
		return nil, ErrGeneric
	default:
		fmt.Printf("bmp: compression value is not recognized - found (%d)\n", bmpInfoHdr.Compression)
		return nil, ErrGeneric
	}

	bmpInfoHdr.SizeImage = Uint32FromLSBytes(bmpBytes[34:38])
	bmpInfoHdr.biXPelsPerMeter = Int32FromLSBytes(bmpBytes[38:42])
	bmpInfoHdr.biYPelsPerMeter = Int32FromLSBytes(bmpBytes[42:46])
	bmpInfoHdr.biClrUsed = Uint32FromLSBytes(bmpBytes[46:50])
	bmpInfoHdr.biClrImportant = Uint32FromLSBytes(bmpBytes[50:54])

	numQuads := uint32((mapOffset - (bmpInfoHdr.HdrSize + 14)) >> 2) // /= 4
	verbose.Printf("numQuads(%d)\n", numQuads)
	bmpBytes = make([]byte, numQuads*4)
	verbose.Printf("copy %d to %d\n", bmpInfoHdr.HdrSize+14, bmpInfoHdr.HdrSize+14+numQuads*4)
	copy(bmpBytes, allBmpBytes[bmpInfoHdr.HdrSize+14:bmpInfoHdr.HdrSize+14+numQuads*4])
	verbose.Printf("read %d bytes of color table\n", len(bmpBytes))
	bf.aColors = make(color.Palette, numQuads)
	switch bmpInfoHdr.Depth {
	case 1, 2, 4, 8:
		for i := range bf.aColors {
			if uint32(i) >= numQuads {
				break
			}
			// BMP images are stored in BGR order rather than RGB order.
			// Every 4th byte is padding  (bmp source was padded with zero)
			bf.aColors[i] = color.RGBA{bmpBytes[4*i+2], bmpBytes[4*i+1], bmpBytes[4*i+0], 0xFF}
		}
	case 16, 24, 32: // color table is empty
	}
	mapSize := bmpInfoHdr.Width * bmpInfoHdr.Height
	switch bmpInfoHdr.Depth {
	case 1:
		mapSize >>= 3
	case 4:
		mapSize >>= 1
	case 8:
		mapSize = mapSize
	case 16:
		// not implemented
		return nil, ErrCantHappen
	case 24:
		mapSize *= 3
	case 32:
		// not implemented
		return nil, ErrCantHappen
	}
	if bmpInfoHdr.SizeImage != 0 {
		bf.aBitMapBits = make([]byte, bmpInfoHdr.SizeImage)
		copy(bf.aBitMapBits, allBmpBytes[bmpInfoHdr.HdrSize+14+numQuads*4:])
		n := len(bf.aBitMapBits)
		if uint32(n) != bmpInfoHdr.SizeImage {
			log.Printf("bmp: bad copy - expected %d bytes got %d\n", bmpInfoHdr.SizeImage, n)
			return nil, ErrShort
		}
		verbose.Printf("copied %d bytes into bf.aBitMapBits\n", n)
	} else {
		bf.aBitMapBits = make([]byte, mapSize)
		copy(bf.aBitMapBits, allBmpBytes[bmpInfoHdr.HdrSize+14+numQuads*4:])
		n := len(bf.aBitMapBits)
		if int32(n) != mapSize {
			log.Printf("bmp: bad copy - expected %d bytes got %d\n", bmpInfoHdr.SizeImage, n)
			return nil, ErrShort
		}
		verbose.Printf("copied %d bytes into bf.aBitMapBits\n", n)
	}
	verbose.Printf("len(Bits) = %d\n", len(bf.aBitMapBits))
	// copy our loose header elements into the struct we're returning
	// copy has to occur after they're fully built out
	bf.fileheader = bmpFileHdr
	bf.Infoheader = bmpInfoHdr

	if verbose {
		bf.Dump()
	}
	verbose.Printf("Exited readBMP() normally\n")
	return &bf, err
}

func (b *BMP_T) Dump() {
	h := b.fileheader
	fmt.Printf("bfType(%x)\n", h.bfMagic)
	fmt.Printf("bfSize(%d)\n", h.bfSize)
	fmt.Printf("bfOffsetBits(%d)\n", h.bfOffsetBits)

	i := b.Infoheader
	fmt.Printf("HdrSize(%d)\n", i.HdrSize)
	fmt.Printf("Width(%d)\n", i.Width)
	fmt.Printf("Height(%d)\n", i.Height)
	fmt.Printf("H * W = %d\n", i.Width*i.Height)
	fmt.Printf("biPlanes(%d)\n", i.biPlanes)
	fmt.Printf("Depth(%d)\n", i.Depth)
	fmt.Printf("Compression(%d)\n", i.Compression)
	fmt.Printf("SizeImage(%d)\n", i.SizeImage)
	fmt.Printf("biXPelsPerMeter(%d)\n", i.biXPelsPerMeter)
	fmt.Printf("biYPelsPerMeter(%d)\n", i.biYPelsPerMeter)
	fmt.Printf("biClrUsed(%d)\n", i.biClrUsed)
	fmt.Printf("biClrImportant(%d)\n", i.biClrImportant)
	fmt.Printf("len(aColors) = %d\n", len(b.aColors))
	fmt.Printf("len(aBitMapBits) = %d\n", len(b.aBitMapBits))
	if false {
		for ndx, val := range b.aColors {
			fmt.Printf("color(%d) = %x\n", ndx, val)
		}
	}
	if true {
		const nBytes = 10
		var msg string
		if i.Compression != 0 {
			msg = "(compressed)"
		} else {
			msg = ""
		}
		fmt.Printf("First %d data %s : ", nBytes, msg)
		for ndx, val := range b.aBitMapBits {
			fmt.Printf("%02x ", val)
			if ndx > (nBytes - 1) {
				fmt.Printf("\n")
				break
			}
		}
	}
}
