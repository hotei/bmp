// docs

package bmp

var x = `

BI_RLE8

When the biCompression member is set to BI_RLE8, the bitmap is compressed
using a run-length encoding format for an 8-bit bitmap. This format may be
compressed in either of two modes: encoded and absolute. Both modes can occur
anywhere throughout a single bitmap.

Encoded mode consists of two bytes: the first byte specifies the number of
consecutive pixels to be drawn using the color index contained in the second
byte. In addition, the first byte of the pair can be set to zero to indicate
an escape that denotes an end of line, end of bitmap, or a delta. The
interpretation of the escape depends on the value of the second byte of the
pair. The following list shows the meaning of the second byte:

Value   Meaning

0       End of line. 
1       End of bitmap. 
2       Delta. The two bytes following the escape contain unsigned values
indicating the horizontal and vertical offset of the next pixel from the
current position.

Absolute mode is signaled by the first byte set to zero and the second byte
set to a value between 0x03 and 0xFF. In absolute mode, the second byte
represents the number of bytes that follow, each of which contains the color
index of a single pixel. When the second byte is set to 2 or less, the escape
has the same meaning as in encoded mode. In absolute mode, each run must be
aligned on a word boundary.  The following example shows the hexadecimal
values of an 8-bit compressed bitmap:

BI_RLE4

When the biCompression member is set to BI_RLE4, the bitmap is compressed
using a run-length encoding (RLE) format for a 4-bit bitmap, which also uses
encoded and absolute modes. In encoded mode, the first byte of the pair
contains the number of pixels to be drawn using the color indexes in the
second byte. The second byte contains two color indexes, one in its
high-order nibble (that is, its low-order four bits) and one in its low-order
nibble. The first of the pixels is drawn using the color specified by the
high-order nibble, the second is drawn using the color in the low-order
nibble, the third is drawn with the color in the high-order nibble, and so
on, until all the pixels specified by the first byte have been drawn.  In
absolute mode, the first byte contains zero, the second byte contains the
number of color indexes that follow, and subsequent bytes contain color
indexes in their high- and low-order nibbles, one color index for each pixel.
In absolute mode, each run must be aligned on a word boundary. The
end-of-line, end-of-bitmap, and delta escapes also apply to BI_RLE4.



`
