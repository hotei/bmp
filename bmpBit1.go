// bmpBit1.go

package bmp

import (
	"bufio"
	"image"
	"image/color"
	"io"
	"log"
)

// unpack a byte into bits using LSB ordering  leftmost bit is bit[0]
func unPack8(b byte) [8]byte {
	var x [8]byte
	if (b & 0x80) != 0 {
		x[0] = 1
	}
	if (b & 0x40) != 0 {
		x[1] = 1
	}
	if (b & 0x20) != 0 {
		x[2] = 1
	}
	if (b & 0x10) != 0 {
		x[3] = 1
	}
	if (b & 0x08) != 0 {
		x[4] = 1
	}
	if (b & 0x04) != 0 {
		x[5] = 1
	}
	if (b & 0x02) != 0 {
		x[6] = 1
	}
	if (b & 0x01) != 0 {
		x[7] = 1
	}
	return x
}

// decodePaletted1 reads a 1 bit-per-pixel BMP image from r.
func decodePaletted1(r io.Reader, c image.Config, b *BMP_T) (image.Image, error) {
	verbose.Printf("Entry to decodePaletted1\n")
	maxBits := len(b.aBitMapBits)

	paletted := image.NewPaletted(image.Rect(0, 0, c.Width, c.Height), c.ColorModel.(color.Palette))
	br := bufio.NewReader(r)
	var bytesRead int
	verbose.Printf("Height = %d   Width = %d  H*W = %d\n", c.Height, c.Width, c.Height*c.Width)
	verbose.Printf("maxBits = %d\n", maxBits<<3)
	verbose.Printf("paletted.Stride(%d)\n", paletted.Stride)
	lastPix := c.Height * c.Width
	dataSize := b.Infoheader.SizeImage << 3
	verbose.Printf("lastPix(%d) dataSize(%d)\n", lastPix, dataSize)
	// N.B. BMP images are stored bottom-up rather than top-down, left to right
	verbose.Printf("b.infoheader.biSizeImage/c.Height(%d)\n", b.Infoheader.SizeImage/uint32(c.Height))
	verbose.Printf("width mod 8 = %d\n", c.Width%8)
	rowWidth := c.Width / 8
	rowXtra := c.Width - (rowWidth * 8)
	verbose.Printf("rowWidth(%d) rowXtra(%d)\n", rowWidth, rowXtra)

	for y := c.Height - 1; y >= 0; y-- {
		var pix8 byte
		var err error
		var start, finish int
		for x := 0; x < rowWidth*8; x += 8 {
			if bytesRead >= maxBits {
				break
			}
			pix8, err = br.ReadByte()
			if err != nil {
				log.Printf("Read failed\n")
				return nil, err
			}
			bytesRead++
			b := unPack8(pix8)
			start = x + (y * c.Width)
			finish = start + 8
			if finish > lastPix {
				verbose.Printf("LastPix\n")
				finish = lastPix
			}
			if start > lastPix {
				start = lastPix
			}
			verbose.Printf("start(%d) finish(%d) byte[%d]  %v\n", start, finish, bytesRead, b)
			copy(paletted.Pix[start:finish], b[:])
		}
		// last byte of scanline may not have all bits used so piece it out
		if rowXtra != 0 {
			pix8, err = br.ReadByte()
			if err != nil {
				log.Printf("Read failed\n")
				return nil, err
			}
			bytesRead++
			b := unPack8(pix8)
			start += 8
			finish = start + rowXtra
			if finish > lastPix {
				verbose.Printf("LastPix\n")
				finish = lastPix
			}
			if start > lastPix {
				start = lastPix
			}
			verbose.Printf("+start(%d) finish(%d) byte[%d]  %v\n", start, finish, bytesRead, b)
			copy(paletted.Pix[start:finish], b[:rowXtra])
		}
		// scanlines are padded if necessary to multiple of uint32 (DWORD)
		for {
			if (bytesRead % 4) == 0 {
				break
			}
			pix8, err = br.ReadByte()
			if err != nil {
				log.Printf("Read failed\n")
				return nil, err
			}
			bytesRead++
			verbose.Printf("byte[%d]\n", bytesRead)
		}
	}
	return paletted, nil
}
