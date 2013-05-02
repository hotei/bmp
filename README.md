<center>
bmp package
===========
</center>

LICENSE
-------

This implementation of the bmp package is (c) 2013 David Rook, released under a BSD style license found in MDR_license.md

bmp_test.go and bmpic.go are (c) 2013 David Rook also with MDR_license.md

Test bmp files are believed to be public domain.  If not please advise and I will replace them with a
link to the source.

Portions were derived or copied from work originally by the Go language project, released under
a license found in GoAuthors_license.md.  Specifically, the API was designed by the go authors,
and the entirety of the bmpRGBA.go, and reader_test.go files were written by the go authors.  A small portion of
bmpRLE8.go is excerpted from the original decodePaletted8().



NOTES
-----

Email: <hotei1352@gmail.com>

My goal was to expand the support for BMP.  The original source code was the package at
code.google.com/p/go.image/bmp Unfortunately it only worked on uncompressed 24 bit files and
some but not all uncompressed 8 bit files.  First I fixed the problems with the 8-bit colormaps.
Then I expanded by adding 1 bit, 4 bit compressed and uncompressed and 8 bit compressed.  In the process
speed and DRAM size were not considerations.  Some of the for { append } loops could be optimized but the initial
focus was on making it work correctly and for this simple seems better.  So far the results are encouraging.

The test program _bmpic.go_ displays the image using xgbutil.  I'm not sure how that will work on a Mac or MS-Windows.

The code was written from a semi-clean-room perspective.  By that I mean that I tossed pretty much the whole orignal
package and started from scratch programing  directly from the spec. 
The one exception to that was decodeRGBA() which worked correctly to begin with.
A side effect is that a lot of debugging via verbose.Printf() is scattered throughout the code.  For now I just set
verbose = false.

```
grep -v verbose < prog.go >prog2.go should remove the bulk of it if anyone is so inclined.
```

As I read the spec and decoded the headers I realized there is a bunch of stuff in there that never gets used.
I probably would have changed some of the names - and eventually did for those that needed to be exposed to public.
I think Depth is more accurate than biBitCount for pixel depth for instance.  Since I was unfamiliar with the bmp
spec I wasn't ready to be too bold.  Most of the names are unchanged or slightly shortened.

In keeping with the go philosophy I will eventually get rid of the cruft but its not hurting anything to keep it for now.
As a result I end up with a BMP_T struct that contains every byte of the BMP file.  Thus you carry
around a bit more memory while decoding since the original bitmap (compressed or not) will occupy more RAM than
the alternative which is an io.Reader.  In most cases this is ok since GC can occur once the image.image is returned
from Decode().

I chose to use another slightly memory-expensive method by unwinding the RLE data into a buffer which then has an 
io.Reader attached before passing it to the decoder.  It is possible to create the image as the RLE data is 
unwound but doing it the way I did allowed me to compare the unwound RLE data to the uncompressed version as a
way to validate the process.  It was simpler and helped at the time.

After scratching my head about how to do the DecodeConfig() I realized that this part of the API may indeed be
broken as Volker Dobler suggested.  The return values from DecodeConfig() don't have the info required
to do the full Decode().  About all they can do is provide suggestions for how big a window you'll need to display
the resulting image.  That's useful, but essentially it forced me to decode the file twice.  There's probably a
better way, but I don't see it yet.  I chose the simpler method since the danger of doing the same thing 
differently in Decode and DecodeConfig seemed more immediate than a slight performance hit.  In time that could
be revisited.

The colormaps provided in package image/color don't extend down to 4 / 2 / 1 bit maps so all the
paletted colormaps get delivered to the user the same way.  Since the bmp files are read only it's not a problem. 

The old test for the package was minimal and isn't easily expandable to other than 24 bit formats.  While it worked
for some files it didn't show them which was one of the main reasons I was doing this project.  So I added more tests.

The new testing _bmp_test.go_ checks to see that the image is readable without errors.  This test is available
regardless of OS. The intent was to feed it all the bmps I could find and see which would cause it to choke.  On my
system the new bmp package no longer chokes on implemented formats, but since I'm not on windows I only found about
9,000 of them.

Bit DEPTH

* 1 bit - 69 files
* 2 bit - 0 files
* 4 bit - 1228 files
* 8 bit - 7026 files
* 16 bit - 35 files
* 24 bit - 736 files
* 32 bit - 51 files
* broken - 63 files ( too short or bad magic )
* total - 9059 files with .bmp extension

To really know if the decoder worked on an unknown file requires visual inspection. _bmpic.go_ displays them but
as noted it assumes an X-11 environment.  (MS windows equivalent would be what?)

It bothered me a little that the original didn't have a seekable input source.
The original code assumed that the info-header would always be that of a version 3 bmp(40 bytes).  If
the file actually contains a different size header then the read will fail.  The
switch to a ReadSeeker was done by using ioutil.ReadAll() to read the whole file at once. Then we can just slice the
parts we want from the result. A possible downside of this is that for a while there will be two copies of the bitmap in RAM.

NOTE BENE - Panics: As the package is being debugged it helps to know where errors happen(line number) and why(stack trace).
Panic is good for that.  It's not so good to see a panic in a supposedly working imported package. In that case an error return is usually
the better choice since not all errors are fatal and the caller might decide to take corrective action and continue.
Panic will happen if []byte -> integer conversion are fed the wrong size slice.  Be careful. 
I personally prefer explicit error messages so there are a lot of them.

TODO
----

1. HI-pri
	1. Bug fixing - none currently known
	1. TBD

1. MED-pri
	1. add 32 bit format

1. LO-pri
	1. add 16 bit format
	1. add 2 bit format (?)
	1. cruft removal

BMP formats
-----------

* 1 bit - can be found in black and white from early fax scanners or any other two colors from ? source
	* uncompressed is supported ( read-only)
	* no compressed version of this
* 2 bit - not common format - was used only for Windows CE
	* uncompressed only <font color=red>not supported</font> [MSDN spec][3] ... <font color=red>need sample</font> low priority)
* 4 bit - fairly common in early days of Windows when 16 color graphics cards were the norm
	* uncompressed is supported
	* RLE-4 compression is supported 
* 8 bit - very common - sub-repo version did not work if colormap was not full.  That's fixed now.
	* uncompressed is supported (originally with full colormap now also with partial colormap)
	* RLE-8 compression is supported with full or partial colormap
* 16 bit - not common in my filesystem - no utility here other than completeness
	* <font color=red>not supported yet</font> - have sample, [good spec][2]
	* bitfields instead of compression
* 24 bit - common but lack of compression means large files, other formats may be more suitable for most things
	* uncompressed is supported - version from sub-repo worked ok
	* no compressed version
* 32 bit - not common in my filesystem - no utility here other than completeness
	* <font color=red>not supported yet</font>  - have sample, [good spec][2]
	* bitfields instead of compression

Resources
---------
* godoc for my [package][6] at go.pkgdoc.org

* The [original google "sub-repo" source] [4]

* The [original golang-nuts thread] [5]

* The original package references this out-of-date / partial [BMP specification][7].
	It provided much of the information I needed, but lacked the values for a few constants (compression related), and
	had nothing at all on bmp header versioning.
	
* There's a [wiki entry][8] that's useful but also incomplete

* http://www.fileformat.info/format/bmp/egff.htm was a bit more helpful, decent spec, good history info

* test bmps came from several sources
	* original testdata video-001.bmp was loaded into GIMP and resaved as 1 bit, 8 bit RLE, 8 bit gray
	* http://vaxa.wvnet.edu/vmswww/bmp.html (jason1@pobox.com / github.com/jsummers/bmpsuite) Jason Summers provided a C
 	program that creates test files in a variety of less common formats such as 4 bit compressed and uncompressed, 8 bit 
	compressed and 1 bit bi-color. Jason's remarks say the output of the program is Public Domain. Thanks Jason!
	* http://videos-cdn.mozilla.net/serv/air_mozilla/public/slides/Test Patterns provided the large images in 16,24 and 32 bit
 	depths
 	* marbles and the 24 bit test strip came from [samples.fileformat.info][1]


Since this was a nearly "clean-room" development I didn't look at other code.  That probably made debugging more painful than 
necessary but the whole thing took only a weekend.


[1]: http://samples.fileformat.info/format/bmp/sample/index.htm	"fileformat.info"
[2]: http://www.fileformat.info/format/bmp/egff.htm "good bmp spec sheet"
[3]: http://msdn.microsoft.com/en-us/library/ms959648.aspx "Win CE bmp spec"
[4]: http://code.google.com/p/go.image/bmp "sub-repository for bmp"
[5]: https://groups.google.com/forum/?fromgroups=#!topic/golang-nuts/7yniCyZW7iY "golang-nuts bmp thread"
[6]: http://go.pkgdoc.org/github.com/hotei/bmp
[7]: http://www.digicamsoft.com/bmp/bmp.html "a bmp spec"
[8]: http://en.wikipedia.org/wiki/BMP_file_format "wiki bmp spec"
<end> README.md Last Update 2013-05-2

