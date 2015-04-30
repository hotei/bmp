// bmprle4.go

package bmp

import (
	"bufio"
	//"bytes"
	//"fmt"
	"image"
	"image/color"
	"io"
	"log"
	//"os"
)

func unPack2(b byte) [2]byte {
	var x [2]byte
	x[0] = (b >> 4) & 0x0f
	x[1] = b & 0xf
	return x
}

func unwindRLE4(r io.Reader, b *BMP_T) ([]byte, error) {
	maxReadBytes := len(b.aBitMapBits)
	verbose.Printf("Entry to unwindRLE4\n")
	verbose.Printf("Height = %d   Width = %d  H*W = %d\n",
		b.Infoheader.Height, b.Infoheader.Width, b.Infoheader.Height*b.Infoheader.Width)
	verbose.Printf("maxBits = %d\n", maxReadBytes)
	rowWidth := b.Infoheader.Width
	if (rowWidth % 2) != 0 {
		rowWidth++
	}
	// use uncompressed size/2 as rough cap for pixMap
	pixMap := make([]byte, 0, b.Infoheader.Height*rowWidth/2)

	verbose.Printf("PixMap will be [%d] when full\n", b.Infoheader.Height*rowWidth/2)
	br := bufio.NewReader(r)
	//BOF := b.fileheader.bfOffsetBits % 4
	bytesRead := 0
	lineCt := 0
	verbose.Printf("len(pixMap)=%d cap(pixMap)=%d\n", len(pixMap), cap(pixMap))
	for {
		if len(pixMap) == cap(pixMap) {
			break
		}
		if bytesRead >= maxReadBytes { // don't read past source end
			break
		}
		numPix, err := br.ReadByte()
		if err != nil {
			log.Printf("bmp: bad read in RLE4\n")
			return nil, err
		} else {
			bytesRead++
		}
		verbose.Printf("byte at hexoffset(%x), numPix(%x)\n", bytesRead, numPix)
		pixVal, err := br.ReadByte()
		if err != nil {
			log.Printf("bmp: bad read in RLE4\n")
			return nil, err
		} else {
			bytesRead++
		}
		verbose.Printf("byte at hexoffset(%x), pixVal(%x)\n", bytesRead, pixVal)
		if numPix > 0 { //  encoded mode
			//verbose.Printf("copying %d encoded pixels\n", numPix)
			loopCt := numPix / 2
			loopXtra := numPix - (loopCt * 2)
			for x := 0; x < int(loopCt); x++ {
				pixMap = append(pixMap, pixVal)
			}
			if loopXtra != 0 {
				pixMap = append(pixMap, pixVal&0xf0)
			}
			continue
		} else { // absolute mode if numPix == 0, can be escaped if second byte 0..2
			if inRangeByte(0, pixVal, 2) { // check for escaped mode
				switch pixVal {
				case 0: // end of line must be on DWORD boundary counting from BOF, not BOpixmap
					for {
						if (len(pixMap) % 4) == 0 {
							break
						}
						//verbose.Printf("Padding line with 0\n")
						pixMap = append(pixMap, 0)
					}
					verbose.Printf("end of line signal found  ")
					lineCt++
					verbose.Printf("len(pixMap)=%d  cap(pixMap)=%d   bytesRead(%d) lineCt(%d)\n", len(pixMap), cap(pixMap), bytesRead, lineCt)
					continue
				case 1: // end of bitmap
					verbose.Printf("end of pixmap source signal found  ")
					lineCt++
					verbose.Printf("len(pixMap)=%d  cap(pixMap)=%d   bytesRead(%d) lineCt(%d)\n", len(pixMap), cap(pixMap), bytesRead, lineCt)
					goto xit
				case 2: // Delta
					// BUG(mdr): TODO - delta encoding not handled in unwindRLE4
					log.Printf("Delta value found but no delta handler available\n")
					return nil, ErrNoDelta
					deltax, err := br.ReadByte()
					deltay, err := br.ReadByte()
					// LINT req - leave here in case we build out the delta code later
					deltax = deltax
					deltay = deltay
					err = err
					bytesRead += 2
					// need some magic here to advance over part of image (why would it be used?)
				}
				log.Printf("can't happen\n")
				return nil, ErrCantHappen
			}
			numPix = pixVal
			verbose.Printf("copying %d absolute pixels\n", numPix)
			loopCt := numPix / 2
			loopXtra := numPix - (loopCt * 2)
			for x := 0; x < int(loopCt); x++ {
				pixVal, err := br.ReadByte()
				if err != nil {
					verbose.Printf("bytesRead(%d)\n", bytesRead)
					log.Printf("bmp: bad read in RLE4\n")
					return nil, err
				} else {
					bytesRead++
				}
				pixMap = append(pixMap, pixVal)
			}
			if loopXtra != 0 {
				pixMap = append(pixMap, pixVal&0xf0)
			}

			if (bytesRead % 2) != 0 { // absolute run must end at word boundary
				_, err := br.ReadByte()
				if err != nil {
					verbose.Printf("bytesRead(%d)\n", bytesRead)
					log.Printf("bmp: bad read in RLE4\n")
					return nil, err
				} else {
					bytesRead++
				}
			}
		}
		verbose.Printf("pixMap[%d]  bytesRead(%d)\n", len(pixMap), bytesRead)
	}

xit:
	verbose.Printf("len(pixMap)=%d  cap(pixMap)=%d   bytesRead(%d) lineCt(%d)\n", len(pixMap), cap(pixMap), bytesRead, lineCt)
	if len(pixMap) != cap(pixMap) {
		verbose.Printf("!Err-> mismatched len & cap - short by(%d)bytes is bad\n", cap(pixMap)-len(pixMap))
	}
	if bytesRead != len(b.aBitMapBits) {
		verbose.Printf("!Err-> mismatched len(source) & bytesRead is bad\n")
		verbose.Printf("bytesRead is %d but should be %d\n", bytesRead, len(b.aBitMapBits))
	}
	// BUG(mdr): OVERKILL? - we fill out pixmap with null bytes if end of source data before map is full
	for {
		if len(pixMap) >= cap(pixMap) {
			break
		}
		//verbose.Printf("padding map at exit\n")
		pixMap = append(pixMap, 0x0)
	}
	b.aBitMapBits = pixMap
	verbose.Printf("pixMap[%d]  bytesRead(%d)\n", len(pixMap), bytesRead)
	verbose.Printf("Exit unwindRLE4() \n")
	return pixMap, nil
}

// decodePaletted reads a 4 bit-per-pixel BMP image from r.
func decodePaletted4(r io.Reader, c image.Config, b *BMP_T) (image.Image, error) {
	maxBits := len(b.aBitMapBits)
	verbose.Printf("Entry to decodePaletted4\n")
	paletted := image.NewPaletted(image.Rect(0, 0, c.Width, c.Height), c.ColorModel.(color.Palette))
	br := bufio.NewReader(r)
	var bytesRead int
	verbose.Printf("Height = %d   Width = %d  H*W = %d\n", c.Height, c.Width, c.Height*c.Width)
	verbose.Printf("maxBits = %d\n", maxBits)
	verbose.Printf("paletted.Stride(%d)\n", paletted.Stride)
	lastPix := c.Height * c.Width
	rowWidth := c.Width / 2
	rowXtra := c.Width - (rowWidth * 2)
	// N.B. BMP images are stored bottom-up rather than top-down, left to right
	for y := c.Height - 1; y >= 0; y-- {
		var pix2 byte
		var err error
		var start, finish int
		for x := 0; x < rowWidth*2; x += 2 {
			if bytesRead >= maxBits {
				break
			}
			pix2, err := br.ReadByte()
			if err != nil {
				log.Printf("bmp: bad read in Pal4\n")
				return nil, err
			}
			bytesRead++
			b := unPack2(pix2)
			start := x + (y * c.Width)
			finish := start + 2
			if finish > lastPix {
				finish = lastPix
			}
			if start > lastPix {
				start = lastPix
			}
			//verbose.Printf("start(%d) finish(%d) byte[%d]  %v\n", start, finish, bytesRead, b)
			copy(paletted.Pix[start:finish], b[:])
		}
		// last byte of scanline may not have all bits used so piece it out
		if rowXtra != 0 {
			//verbose.Printf("adding last pixel to line\n")
			pix2, err = br.ReadByte()
			if err != nil {
				log.Printf("bmp: bad read in Pal4\n")
				return nil, err
			}
			bytesRead++
			b := unPack2(pix2)
			start += 2
			finish = start + rowXtra
			if finish > lastPix {
				verbose.Printf("LastPix\n")
				finish = lastPix
			}
			if start > lastPix {
				start = lastPix
			}
			//verbose.Printf("+start(%d) finish(%d) byte[%d]  %v\n", start, finish, bytesRead, b)
			copy(paletted.Pix[start:finish], b[:rowXtra])
		}
		// scanlines are padded if necessary to multiple of uint32 (DWORD)
		for {
			if (bytesRead % 4) == 0 {
				break
			}
			pix2, err = br.ReadByte()
			if err != nil {
				log.Printf("bmp: bad read in Pal4\n")
				return nil, err
			}
			bytesRead++
			verbose.Printf("byte[%d]\n", bytesRead)
		}

	}
	return paletted, nil
}
