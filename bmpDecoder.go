// bmpreader.go (c) 2013 David Rook
// started 2013-04-26 working well same day

// Read bmp file, decode and return image.Image
//
//  Features
//  ========
//    Handles 1 bit (Black and White) or BiColor
//    Handles 4 bit compressed (RLE-4) or uncompressed
//       with full or partial (<16 entry) colorMaps
//    Handles 8 bit compressed (RLE-8) or uncompressed
//       with full or partial (<256 entry) colorMaps
//    Handles 24 bit color uncompressed
//    These features covered over 99% of the files on my system YMMV
//
// Real BUGS  -##0{   None Known - but beware of limitations
//
//  Limitations
//  -----------
//    Doesn't handle delta compression
//    Doesn't handle 2 bit files (Windows CE only)
//    Doesn't handle 16 bit files
//    Doesn't handle 32 bit files
//    Doesn't handle OS2 v1 or OS2 v2 format bmps
//    Doesn't handle V4 or V5 format files
//
// (c) 2013 David Rook - License is BSD style - see LICENSE.md
//  Also see README.md for more info
//
package bmp

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
)

// Enumeration for possible compression values
const (
	BI_RGB            = 0
	BI_RLE8           = 1
	BI_RLE4           = 2
	BI_BITFIELDS      = 3
	BI_JPEG           = 4
	BI_PNG            = 5
	BI_ALPHABITFIELDS = 6
)

var (
	MagicBytes          = []byte{'B', 'M'}
	ErrGeneric          = errors.New("bmp: ?")
	Err02NotSupported   = errors.New("bmp:  2 bit format not supported")
	Err16NotSupported   = errors.New("bmp: 16 bit format not supported")
	Err32NotSupported   = errors.New("bmp: 32 bit format not supported")
	ErrBadHeader        = errors.New("bmp: File header not correct")
	ErrBadMagic         = errors.New("bmp: File Id bytes not BM")
	ErrCantHappen       = errors.New("bmp: Cant happen")
	ErrEmptyBitmap      = errors.New("bmp: Empty bitmap - no data")
	ErrNoDelta          = errors.New("bmp: Delta not supported yet")
	ErrOS21NotSupported = errors.New("bmp: OS2 v1 format not supported")
	ErrOS22NotSupported = errors.New("bmp: OS2 v2 format not supported")
	ErrShort            = errors.New("bmp: File too short")
	ErrV4NotSupported   = errors.New("bmp: V4 format not supported")
	ErrV5NotSupported   = errors.New("bmp: V5 format not supported")
)

type VerboseType bool

var verbose VerboseType

func init() {
	verbose = false
}

func (v VerboseType) Printf(s string, a ...interface{}) {
	if v {
		fmt.Printf(s, a...)
	}
}

type BITMAPFILEHEADER_T struct {
	bfMagic      [2]byte // must be "BM"   [0:2]
	bfSize       uint32  // size of file in bytes [2:6]
	bfReserved1  uint16  // must be zero	[6:8]
	bfReserved2  uint16  // must be zero	[8:10]
	bfOffsetBits uint32  // offset from what? (begOfFile?) to actual bitmap data [10:14]
}

type BITMAPINFO_T struct {
	bmiHeader BITMAPINFOHEADER_T
	bmiColors []RGBQUAD_T
}

/*
typedef struct tagBITMAPINFOHEADER {    // bmih
    DWORD   biSize;			// [14:18]
    LONG    biWidth;		// [18:22]
    LONG    biHeight;		// [22:26]
    WORD    biPlanes;		// [26:28]
    WORD    biBitCount;		// [28:30]
    DWORD   biCompression;	// [30:34]
    DWORD   biSizeImage;	// [34:38]
    LONG    biXPelsPerMeter;// [38:42]
    LONG    biYPelsPerMeter;// [42:46]
    DWORD   biClrUsed;		// [46:50]
    DWORD   biClrImportant; // [50:54]
} BITMAPINFOHEADER;
*/

// DIB device-independent bitmap
type BITMAPINFOHEADER_T struct {
	HdrSize         uint32 // bytes used by the InfoHeader struct - NOT always same as sizeof(BITMAPINFOHEADER_T)
	Width           int32  // width of bitmap in pixels
	Height          int32  // height of bitmap in pixels
	biPlanes        uint16 // number of planes - must be 1
	Depth           uint16 // bits per pixel - valid values are 1/2/4/8/16/24/32 but not all are implemented
	Compression     uint32 // BI_RGB - not compressed | BI_RLE4 | BI_RLE8
	SizeImage       uint32 // zero is ok if not compressed - otherwise ? height x width x depth
	biXPelsPerMeter int32  // vertical rez
	biYPelsPerMeter int32  // horizontal rez
	biClrUsed       uint32 // number of color indexes in color table actually used by bitmap - or see note
	biClrImportant  uint32 // number of 'important' color indexes in color table - if zero then all impt
}

type RGBQUAD_T uint32

type BMP_T struct {
	fileheader  BITMAPFILEHEADER_T
	Infoheader  BITMAPINFOHEADER_T
	aColors     color.Palette //  []RGBQUAD_T
	aBitMapBits []byte
}

func init() {
	image.RegisterFormat("bmp", "BM????\x00\x00\x00\x00", Decode, DecodeConfig)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	verbose.Printf("init done\n")
}

func Decode(r io.Reader) (img image.Image, err error) {
	b, err := ReadBMP(r)
	if err != nil {
		return img, err
	}
	//c, err := decodeConfig(b)
	//verbose.Printf("c.Width(%d) c.Height(%d) c.ColorModel(%v)\n", c.Width, c.Height, c.ColorModel)
	//verbose.Printf("len(b.aBitMapBits) = %d\n", len(b.aBitMapBits))
	var c image.Config
	switch b.Infoheader.Depth {
	case 1, 2, 4, 8:
		c = image.Config{ColorModel: b.aColors, Width: int(b.Infoheader.Width), Height: int(b.Infoheader.Height)}
	case 16:
		fmt.Printf("16 bit per pixel not supported\n")
		return nil, Err16NotSupported
	case 24:
		verbose.Printf("24 colormodel=%v\n", color.RGBAModel)
		c = image.Config{ColorModel: color.RGBAModel, Width: int(b.Infoheader.Width), Height: int(b.Infoheader.Height)}
	case 32:
		fmt.Printf("32 bit per pixel not supported\n")
		return nil, Err32NotSupported
	default:
		log.Printf("bmp: can't happen \n")
		return nil, ErrCantHappen
	}

	if len(b.aBitMapBits) <= 0 {
		log.Printf("no bits in bitmap\n")
		return nil, ErrEmptyBitmap
	}
	nr := bytes.NewReader(b.aBitMapBits)
	switch b.Infoheader.Depth {
	case 1:
		img, err = decodePaletted1(nr, c, b)
	case 4:
		if b.Infoheader.Compression == BI_RLE4 {
			pixbufadr, err := unwindRLE4(nr, b)
			if err != nil {
				log.Printf("bmp: bad read from RLE4\n")
				return nil, err
			}
			nr = bytes.NewReader(pixbufadr)
		}
		img, err = decodePaletted4(nr, c, b)
	case 8:
		if b.Infoheader.Compression == BI_RLE8 {
			pixbufadr, err := unwindRLE8(nr, b)
			if err != nil {
				log.Printf("bmp: bad read from RLE8\n")
				return nil, err
			}
			nr = bytes.NewReader(pixbufadr)
		}
		img, err = decodePaletted8(nr, c, b)
	case 24:
		img, err = decodeRGBA(nr, c)
	default:
		log.Printf("bmp: can't happen Decode\n") // only 1/4/8/24 allowed by earlier logic
		return nil, ErrCantHappen
	}
	return img, err
}

func DecodeConfig(r io.Reader) (config image.Config, err error) {

	bf, err := ReadBMP(r) // not efficient but simple wins : read bitmap just to git header info

	switch bf.Infoheader.Depth {
	case 1, 2, 4, 8:
		return image.Config{ColorModel: bf.aColors, Width: int(bf.Infoheader.Width), Height: int(bf.Infoheader.Height)}, nil
	case 16:
		fmt.Printf("16 bit per pixel not supported\n")
		return config, Err16NotSupported
	case 24:
		verbose.Printf("24 colormodel=%v\n", color.RGBAModel)
		return image.Config{ColorModel: color.RGBAModel, Width: int(bf.Infoheader.Width), Height: int(bf.Infoheader.Height)}, nil
	case 32:
		fmt.Printf("32 bit per pixel not supported\n")
		return config, Err32NotSupported
	default:
		log.Printf("bmp: can't happen DecodeConfig\n") // only 1/4/8/24 allowed by earlier logic
		var noConfig image.Config                      // return an empty config struct with err
		return noConfig, ErrCantHappen
	}
	return config, err

}
