// bmpic.go

package main

import (
	// mine
	"github.com/hotei/bmp"
	// Alien
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	// go version 1.X std lib only below
	// _ "code.google.com/p/go.image/bmp"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	testDir = "../testdata/"
)

var (
	testImages = []string{
		"notBMP.bmp",              // fails as req
		"256colorOS2v1.bmp",       // recognized as OS2v1 format, not decoded
		"bit1bw-rnr.bmp",          // 1 bit per pixel, uncompressed, 2 entry color table - working with my code
		"bit1color2.bmp",          // working with my code
		"bit4-test.bmp",           // working with my code
		"bit4comp-test.bmp",       // work in progress
		"bit8-gray-rnr.bmp",       // working with my code
		"bit8comp-test.bmp",       // RLE8 working with my code
		"bit8comp-rnr.bmp",        // RLE8
		"bit8-test.bmp",           // working with my code - original failed
		"bit24-test.bmp",          // airmoz
		"bit24uncomp-marbles.bmp", // working large 24 bit uncompressed
		"bit24uncomp-rnr.bmp",     // 24 bit per pixel, uncompressed, working with original
		"bit24-teststrip.bmp",     // working 24 bit uncompressed
		"whirlpool.jpg",           // fails as required with bad magic if called with bmp.Decode()
		//		"bit16-test.bmp", not implemented yet
		//		"bit32-test.bmp", not implemented yet
	}
	g_picNumFlag    int
	g_fnameFlag     string
	g_depthMap      map[int]int // count of how many files have what depth
	g_typeMap       map[int]int // count of file types based on header size
	ErrNotSupported = errors.New("bmp: not Version 3 - format not supported")
)

func init() {
	flag.IntVar(&g_picNumFlag, "pic", -1, "which pic to show or 0 to show all")
	flag.StringVar(&g_fnameFlag, "f", "", "which file")
	g_depthMap = make(map[int]int, 10)
	g_typeMap = make(map[int]int, 10)
}

func showOne(testNum int) {
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	fname := testImages[testNum]
	fmt.Printf("Working on %s\n", fname)
	file, err := os.Open(testDir + fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Decode the image.
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	ximg := xgraphics.NewConvert(X, img)
	ximg.XShowExtra(fname, true)

	xevent.Main(X)
	time.Sleep(4 * time.Second)
	xevent.Quit(X)
}

func showAll() {

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(testImages); i++ {
		fname := testImages[i]
		fmt.Printf("%d Working on %s\n", i, fname)
		file, err := os.Open(testDir + fname)
		if err != nil {
			continue
		}
		defer file.Close()
		var img image.Image
		// Decode the image.
		// using true forces bmp decoder, otherwise whatever is registered for ext is used
		// result slightly different if non-bmps fed to it
		if true {
			img, err = bmp.Decode(file)
		} else {
			img, _, err = image.Decode(file)
		}
		if err != nil {
			continue
		}
		ximg := xgraphics.NewConvert(X, img)
		ximg.XShowExtra(fname, true)
		time.Sleep(1 * time.Second)
	}
	xevent.Main(X)
	time.Sleep(4 * time.Second)
	xevent.Quit(X)
}

func showNamed(fname string) {
	fmt.Printf("Working on %s\n", fname)
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	// Decode the image.
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	ximg := xgraphics.NewConvert(X, img)
	ximg.XShowExtra(fname, true)
	xevent.Main(X)
	time.Sleep(4 * time.Second)
	xevent.Quit(X)
}

func readNamed(fname string) error {
	fmt.Printf("Working on %s\n", fname)
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := bmp.ReadBMP(file)
	if err != nil {
		fmt.Printf("!ERR -> problem with %s\n", fname)
		return err
	}
	b.Dump()

	g_depthMap[int(b.Infoheader.Depth)] += 1
	g_typeMap[int(b.Infoheader.HdrSize)] += 1
	if b.Infoheader.HdrSize != 40 {
		fmt.Printf("Header not BMP type 3 - unusual, may be valid\n")
		return ErrNotSupported
	}
	return nil
}

func GetAllArgs() []string {
	rv := make([]string, 0, 1000)
	f := os.Stdin // f is * osFile
	rdr := bufio.NewReader(f)
	alldone := false
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				alldone = true
			} else {
				log.Panicf("MDR: GetAllArgs read error")
			}
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			rv = append(rv, line)
		}
		if alldone {
			break
		}
	}
	if flag.Parsed() {
		args := flag.Args()
		for _, arg := range args {
			rv = append(rv, arg)
		}
	} else {
		fmt.Printf("Warning --> GetAllArgs: flags not parsed yet\n")
	}
	return rv
}

func main() {
	flag.Parse()
	if g_fnameFlag != "" {
		showNamed(g_fnameFlag)
		return
	}
	if g_picNumFlag > 0 {
		showOne(g_picNumFlag)
		return
	}
	if g_picNumFlag == 0 {
		showAll()
		return
	}
	var myArgs []string
	if flag.NArg() <= 0 {
		myArgs = GetAllArgs()
	} else {
		myArgs = flag.Args()
	}
	i := 0
	var err error
	for _, fname := range myArgs {
		fmt.Printf("\nFile #%d\n", i)
		err = readNamed(fname)
		if err != nil {
			fmt.Printf("!Err->%v on file %s\n", err, fname)
		}
		i++
	}
	fmt.Printf("\n\n\nprocessed %d files\n", len(myArgs))
	goodRead := 0
	for ndx, val := range g_depthMap {
		fmt.Printf("Depth[%2d] matches %3d files\n", ndx, val)
		goodRead += val
	}
	if goodRead != len(myArgs) {
		fmt.Printf("%d file(s) had errors\n", len(myArgs)-goodRead)
	}

	goodRead = 0
	for ndx, val := range g_typeMap {
		fmt.Printf("Type[%2d] matches %3d files\n", ndx, val)
		goodRead += val
	}
	if goodRead != len(myArgs) {
		fmt.Printf("%d file(s) had errors\n", len(myArgs)-goodRead)
	}

	fmt.Printf("<done>\n")
}
