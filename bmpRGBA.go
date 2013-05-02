// bmpRGBA.go

package bmp

import (
	"fmt"
	"image"
	"io"
)

// decodeRGBA reads a 24 bit-per-pixel BMP image from r.
func decodeRGBA(r io.Reader, c image.Config) (image.Image, error) {
	verbose.Printf("Entry to decodeRGBA\n")
	verbose.Printf("c.Width(%d) c.Height(%d) c.ColorModel(%v)\n", c.Width, c.Height, c.ColorModel)
	rgba := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	// There are 3 bytes per pixel, and each row is 4-byte aligned.
	b := make([]byte, (3*c.Width+3)&^3)
	verbose.Printf("len(b) = %d\n", len(b))
	// BMP images are stored bottom-up rather than top-down.
	for y := c.Height - 1; y >= 0; y-- {
		n, err := r.Read(b)
		if err != nil {
			fmt.Printf("Read %d bytes\n", n)
			return nil, err
		}
		p := rgba.Pix[y*rgba.Stride : y*rgba.Stride+c.Width*4]
		for i, j := 0, 0; i < len(p); i, j = i+4, j+3 {
			// BMP images are stored in BGR order rather than RGB order.
			p[i+0] = b[j+2]
			p[i+1] = b[j+1]
			p[i+2] = b[j+0]
			p[i+3] = 0xFF
		}
	}
	return rgba, nil
}
