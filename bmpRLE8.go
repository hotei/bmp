// bmprle8.go

package bmp

import (
	"bufio"
	"image"
	"image/color"
	"io"
	"log"
)

func unwindRLE8(r io.Reader, b *BMP_T) ([]byte, error) {
	maxReadBytes := len(b.aBitMapBits)
	verbose.Printf("Entry to unwindRLE8\n")
	verbose.Printf("Height = %d   Width = %d  H*W = %d\n",
		b.Infoheader.Height, b.Infoheader.Width, b.Infoheader.Height*b.Infoheader.Width)
	verbose.Printf("maxReadBytes = %d\n", maxReadBytes)
	rowWidth := b.Infoheader.Width
	if (rowWidth % 4) != 0 {
		rowWidth++
	}
	pixMap := make([]byte, 0, b.Infoheader.Height*rowWidth)
	verbose.Printf("PixMap will be [%d] when full\n", b.Infoheader.Height*rowWidth)
	br := bufio.NewReader(r)
	bytesRead := 0
	lineCt := 0
	verbose.Printf("len(pixMap)=%d cap(pixMap)=%d\n", len(pixMap), cap(pixMap))
	for {
		if len(pixMap) == cap(pixMap) {
			break
		}
		if bytesRead >= maxReadBytes {
			break
		}
		numPix, err := br.ReadByte()
		if err != nil {
			verbose.Printf("bytesRead(%d)\n", bytesRead)
			log.Printf("bmp: bad read in RLE8\n")
			return nil, err
		} else {
			bytesRead++
		}
		verbose.Printf("byte at hexoffset(%x), numPix(%x)\n", bytesRead, numPix)
		pixVal, err := br.ReadByte()
		if err != nil {
			verbose.Printf("bytesRead(%d)\n", bytesRead)
			log.Printf("bmp: bad read in RLE8\n")
			return nil, err
		} else {
			bytesRead++
		}
		verbose.Printf("byte at hexoffset(%x), pixVal(%x)\n", bytesRead, pixVal)
		if numPix > 0 { //  encoded mode
			//verbose.Printf("copying %d encoded pixels\n", numPix)
			for x := 0; x < int(numPix); x++ {
				pixMap = append(pixMap, pixVal)
			}
			continue
		} else { // absolute mode
			if inRangeByte(0, pixVal, 2) { // check for escaped mode
				switch pixVal {
				case 0: // end of line must be on DWORD boundary
					for {
						if (len(pixMap) % 4) == 0 {
							break
						}
						//verbose.Printf("Padding line with 0\n")
						pixMap = append(pixMap, 0)
					}
					verbose.Printf("end of line signal found\n")
					lineCt++
					verbose.Printf("len(pixMap)=%d  cap(pixMap)=%d   bytesRead(%d) lineCt(%d)\n", len(pixMap), cap(pixMap), bytesRead, lineCt)
					continue
				case 1: // end of bitmap
					verbose.Printf("end of pixmap source signal found\n")
					lineCt++
					verbose.Printf("len(pixMap)=%d  cap(pixMap)=%d   bytesRead(%d) lineCt(%d)\n", len(pixMap), cap(pixMap), bytesRead, lineCt)
					goto xit
				case 2: // Delta
					log.Printf("Delta value found but no handler available for it\n")
					return nil, ErrNoDelta
					deltax, err := br.ReadByte()
					deltay, err := br.ReadByte()
					// LINT req - leave here in case we build out delta code later
					deltax = deltax
					deltay = deltay
					err = err
					bytesRead += 2
					// need some magic here to advance over part of image (why would it be used?)
					// BUG(mdr): TODO - delta encoding not handled in unwindRLE8
				}
				log.Printf("can't happen\n")
				return nil, ErrCantHappen
			}
			numPix = pixVal
			verbose.Printf("copying %d absolute bytes\n", numPix)
			for i := 0; i < int(numPix); i++ {
				pixVal, err := br.ReadByte()
				if err != nil {
					verbose.Printf("bytesRead(%d)\n", bytesRead)
					log.Printf("bmp: bad read in RLE8\n")
					return nil, err
				} else {
					bytesRead++
				}
				pixMap = append(pixMap, pixVal)
			}
			if (bytesRead % 2) != 0 { // absolute run must be aligned to word boundary
				_, err := br.ReadByte()
				if err != nil {
					verbose.Printf("bytesRead(%d)\n", bytesRead)
					log.Printf("bmp: bad read in RLE8\n")
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
	verbose.Printf("Exit unwindRLE8() \n")
	return pixMap, nil
}

// decodePaletted8 reads an 8 bit-per-pixel BMP image from r.
func decodePaletted8(r io.Reader, c image.Config, b *BMP_T) (image.Image, error) {
	rowWidth := b.Infoheader.Width
	if (rowWidth % 4) != 0 {
		rowWidth++
	}
	maxBits := len(b.aBitMapBits)
	maxReadBytes := b.Infoheader.Height * rowWidth
	paletted := image.NewPaletted(image.Rect(0, 0, c.Width, c.Height), c.ColorModel.(color.Palette))
	verbose.Printf("Entry to decodePaletted8\n")
	verbose.Printf("Height = %d   Width = %d  H*W = %d\n", c.Height, c.Width, c.Height*c.Width)
	verbose.Printf("maxReadBytes = %d\n", maxReadBytes)
	verbose.Printf("paletted.Stride(%d)\n", paletted.Stride)
	//	lastPix := c.Height* c.Width
	// BMP images are stored bottom-up rather than top-down, left to right
	bytesRead := int32(0)
	tmp := make([]byte, 1)
	for y := c.Height - 1; y >= 0; y-- {
		if bytesRead >= int32(maxBits) {
			break
		}
		p := paletted.Pix[y*paletted.Stride : y*paletted.Stride+c.Width]
		n, err := r.Read(p)
		if err != nil {
			log.Printf("bmp: bad read in Pal8\n")
			return nil, err
		}
		if n != c.Width { // ok to ignore short read since coder could shortcut it
			verbose.Printf("short read n(%d) c.Width(%d)\n", n, c.Width)
		}
		bytesRead += int32(c.Width)
		verbose.Printf("bytesRead(%d)\n", bytesRead)
		if bytesRead >= int32(maxBits) {
			break
		}
		// Each row is 4-byte aligned
		for {
			if bytesRead >= int32(maxBits) {
				break
			}
			if (bytesRead % 4) == 0 {
				break
			}
			_, err := r.Read(tmp)
			if err != nil {
				log.Printf("bmp: bad read in Pal8\n")
				return nil, err
			}
			bytesRead++
			verbose.Printf("byte[%d] maxReadBytes(%d) maxBits(%d)\n", bytesRead, maxReadBytes, maxBits)
		}
	}
	verbose.Printf("decodePaletted8().bytesRead(%d)\n", bytesRead)
	return paletted, nil
}
