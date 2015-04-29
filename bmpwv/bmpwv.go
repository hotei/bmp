// bmpwv.go (c) 2013 David Rook

package main

import (
	// go 1.X only below here
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Alien imports
	"github.com/hotei/bmp"
	"github.com/russross/blackfriday" // Russ Ross 2012-11-22 version
)

const (
	hostIPstr  = "127.0.0.1"
	portNum    = 8282
	serverRoot = "/www/"

	bmpURL = "/bmp/"
	imgURL = "/image/"
	mdURL  = "/md/"
	// notused pngURL = "/png/"
)

var (
	portNumString = fmt.Sprintf(":%d", portNum)
	listenOnPort  = hostIPstr + portNumString
	g_fileNames   []string
	myBmpDir      = []byte{}
	myMdDir       = []byte{}
	myImgDir      = []byte{}
	myPngDir      = []byte{}
)

func checkImgName(pathname string, info os.FileInfo, err error) error {
	fmt.Printf("checking image %s\n", pathname)
	if info == nil {
		fmt.Printf("WARNING --->  no stat info: %s\n", pathname)
		os.Exit(1)
	}
	if info.IsDir() {
		// return filepath.SkipDir
	} else { // regular file
		//fmt.Printf("found %s %s\n", pathname, filepath.Ext(pathname))
		// really need to match against lowercase pathname
		ext := strings.ToLower(filepath.Ext(pathname))
		var images = []string{".bmp", ".jpeg", ".jpg", ".png", ".svg"}
		images = images
		if ext == ".bmp" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
			return nil
		}
		if ext == ".jpeg" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
			return nil
		}
		if ext == ".jpg" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
			return nil
		}
		if ext == ".png" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
			return nil
		}
		if ext == ".svg" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
			return nil
		}
		// BUG(mdr): need an error returned if no matches made?
	}
	return nil
}

func checkBmpName(pathname string, info os.FileInfo, err error) error {
	fmt.Printf("checking bmp %s\n", pathname)
	if info == nil {
		fmt.Printf("WARNING --->  no stat info: %s\n", pathname)
		os.Exit(1)
	}
	if info.IsDir() {
		// return filepath.SkipDir
	} else { // regular file
		//fmt.Printf("found %s %s\n", pathname, filepath.Ext(pathname))
		if filepath.Ext(pathname) == ".bmp" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
		}
	}
	return nil
}

func checkMdName(pathname string, info os.FileInfo, err error) error {
	fmt.Printf("checking md %s\n", pathname)
	if info == nil {
		fmt.Printf("WARNING --->  no stat info: %s\n", pathname)
		os.Exit(1)
	}
	if info.IsDir() {
		// return filepath.SkipDir
		// g_fileNames = append(g_fileNames, pathname)
		return nil
	} else { // regular file
		//fmt.Printf("found %s %s\n", pathname, filepath.Ext(pathname))
		ext := filepath.Ext(pathname)
		if ext == ".md" || ext == ".markdown" || ext == ".mdown" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
		}
	}
	return nil
}

// show thumbnail as clickable link to original image
func makeBmpLine(s string) []byte {
	//return []byte(fmt.Sprintf("<a href=\"%s\">View %s</a><br>\n",s,s))
	workDir := serverRoot + bmpURL[1:]
	s = s[len(workDir):]
	x := []byte(fmt.Sprintf("<img src=\"%s\" height=100 width=150> <a href= \"%s\"> %s</a><br>\n", bmpURL+s, bmpURL+s, s))
	fmt.Printf("%s\n", x)
	return x
}

func makeMdLine(s string) []byte {
	workDir := serverRoot + mdURL[1:]
	s = s[len(workDir):]
	return []byte(fmt.Sprintf("<a href=\"%s\">%s</a><br>", s, s))
}

func makeImgLine(s string) []byte {
	//return []byte(fmt.Sprintf("<a href=\"%s\">View %s</a><br>\n",s,s))
	workDir := serverRoot + imgURL[1:]
	s = s[len(workDir):]
	x := []byte(fmt.Sprintf("<img src=\"%s\" height=100 width=150> <a href= \"%s\"> %s</a><br>\n", imgURL+s, imgURL+s, s))
	fmt.Printf("%s\n", x)
	return x
}

func init() {
	fmt.Printf("Starting init()\n")

	pathName := serverRoot + bmpURL[1:]
	g_fileNames = make([]string, 0, 20)
	myBmpDir = []byte(`<html><!-- comment --><head><title>Test BMP package</title></head><body>click on image to see in original size<br>`) // {}
	stats, err := os.Stat(pathName)
	if err != nil {
		fmt.Printf("Can't get fileinfo for %s\n", pathName)
		os.Exit(1)
	}
	if stats.IsDir() {
		filepath.Walk(pathName, checkBmpName)
	} else {
		fmt.Printf("this argument must be a directory (but %s isn't)\n", pathName)
		os.Exit(-1)
	}
	//fmt.Printf("g_fileNames = %v\n", g_fileNames)
	for _, val := range g_fileNames {
		fmt.Printf("%v\n", val)
		line := makeBmpLine(val)
		myBmpDir = append(myBmpDir, line...)
	}
	t := []byte(`</body></html>`)
	myBmpDir = append(myBmpDir, t...)

	// =====================
	pathName = serverRoot + mdURL[1:]
	g_fileNames = make([]string, 0, 20)
	myMdDir = []byte(`<html><!-- comment --><head><title>Test MD package</title></head><body>click to read<br>`) // {}
	stats, err = os.Stat(pathName)
	if err != nil {
		fmt.Printf("Can't get fileinfo for %s\n", pathName)
		os.Exit(1)
	}
	if stats.IsDir() {
		filepath.Walk(pathName, checkMdName)
	} else {
		fmt.Printf("this argument must be a directory (but %s isn't)\n", pathName)
		os.Exit(-1)
	}
	//fmt.Printf("g_fileNames = %v\n", g_fileNames)
	for _, val := range g_fileNames {
		//fmt.Printf("%v\n", val)
		line := makeMdLine(val)
		myMdDir = append(myMdDir, line...)
	}
	t = []byte(`</body></html>`)
	myMdDir = append(myMdDir, t...)

	// =====================
	pathName = serverRoot + imgURL[1:]
	g_fileNames = make([]string, 0, 20)
	myImgDir = []byte(`<html><!-- comment --><head><title>Test IMAGES package</title></head><body>click on image to see in original size<br>`) // {}
	stats, err = os.Stat(pathName)
	if err != nil {
		fmt.Printf("Can't get fileinfo for %s\n", pathName)
		os.Exit(1)
	}
	if stats.IsDir() {
		filepath.Walk(pathName, checkImgName)
	} else {
		fmt.Printf("this argument must be a directory (but %s isn't)\n", pathName)
		os.Exit(-1)
	}
	//fmt.Printf("g_fileNames = %v\n", g_fileNames)
	for _, val := range g_fileNames {
		fmt.Printf("%v\n", val)
		line := makeImgLine(val)
		myImgDir = append(myImgDir, line...)
	}
	t = []byte(`</body></html>`)
	myImgDir = append(myImgDir, t...)

}

func mdHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("A: r.URL.Path %q\n", r.URL.Path)
	if r.URL.Path == mdURL {
		w.Write(myMdDir)
		return
	}
	var output []byte
	var err error
	workDir := serverRoot + mdURL[1:]
	fileName := workDir + r.URL.Path[len(mdURL):]
	fmt.Printf("mdHandler: fname = %s\n", fileName)
	ext := filepath.Ext(fileName)
	if ext == ".md" || ext == ".markdown" || ext == ".mdown" {
		output = htmlFromMd(fileName)
	} else {
		output, err = ioutil.ReadFile(fileName)
		if err != nil {
			errStr := fmt.Sprintf("mdHandler: %v\n", err)
			fmt.Printf("%s\n", errStr)
			w.Write([]byte(fmt.Sprintf("404 - Not Found\n")))
			return
		}
	}
	w.Write(output)
}

func htmlFromMd(fname string) []byte {
	var output []byte
	input, err := ioutil.ReadFile(fname)
	if err != nil {
		tmp := fmt.Sprintf("Problem reading input, can't open %s", fname)
		output = []byte(tmp)
	} else {
		output = blackfriday.MarkdownCommon(input) // MarkdownBasic(input) also possible
	}
	if false { // debug use only
		os.Stdout.Write(input)
		os.Stdout.Write(output)
	}
	return output
}

// tested ok with bmp,png,jpeg,jpg
func imageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("imageHandler: r.URL.Path %q\n", r.URL.Path)
	if r.URL.Path == imgURL {
		fmt.Printf("image dir sent %d lines\n", len(myImgDir))
		w.Write(myImgDir)
		return
	}
	workDir := serverRoot + imgURL[1:]
	fmt.Printf("workDir(%s)\n", workDir)
	fmt.Printf("r.URL.Path(%s)\n", r.URL.Path)
	imageName := workDir + r.URL.Path[len(imgURL):]
	fmt.Printf("imageHandler: imageName = %s\n", imageName)
	ext := strings.ToLower(filepath.Ext(imageName))
	//fmt.Printf("ext = %s\n",ext)
	if ext == ".bmp" {
		bmpWriteOut(imageName, w)
		return
	}
	if ext == ".png" {
		pngWriteOut(imageName, w)
		return
	}
	if ext == ".jpeg" {
		jpegWriteOut(imageName, w)
		return
	}
	if ext == ".jpg" {
		jpegWriteOut(imageName, w)
		return
	}
	if ext == ".svg" {
		rawWriteOut(imageName, w)
		return
	}
}

// just hand the raw file over to the browser
func rawWriteOut(fileName string, w http.ResponseWriter) {
	rawBuf, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("rawWriteOut: cant open file %s\n", fileName)
		return
	}
	w.Write(rawBuf)
}

func jpegWriteOut(imageName string, w http.ResponseWriter) {
	bf, err := os.Open(imageName)
	if err != nil {
		fmt.Printf("jpegWriteOut: cant open image %s\n", imageName)
		return
	}
	img, err := jpeg.Decode(bf)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("jpegWriteOut: Decode failed for %s of error:%v\n", imageName, err)))
		return
	}
	b := make([]byte, 0, 1000*1000) // b will expand as needed
	wo := bytes.NewBuffer(b)
	err = png.Encode(wo, img)
	if err != nil {
		fmt.Printf("imageHandler: png encode failed for %s\n", imageName)
		return
	}
	w.Write(wo.Bytes())
}

// testing
func pngWriteOut(imageName string, w http.ResponseWriter) {
	fmt.Printf("pngWriteOut: imageName = %s\n", imageName)
	bf, err := os.Open(imageName)
	if err != nil {
		fmt.Printf("pngWriteOut: cant open image %s\n", imageName)
		return
	}
	img, err := png.Decode(bf)
	if err != nil {
		fmt.Printf("pngWriteOut: image decode failed for %s png\n", imageName)
		w.Write([]byte(fmt.Sprintf("image Decode failed for %s png error:%v\n", imageName, err)))
		return
	}
	b := make([]byte, 0, 1000*1000) // b will expand as needed
	wo := bytes.NewBuffer(b)
	err = png.Encode(wo, img)
	if err != nil {
		fmt.Printf("pngWriteOut: png encode failed for %s\n", imageName)
		return
	}
	w.Write(wo.Bytes())
}

func bmpWriteOut(imageName string, w http.ResponseWriter) {
	fmt.Printf("bmpWriteOut: imageName = %s\n", imageName)
	bf, err := os.Open(imageName)
	if err != nil {
		fmt.Printf("bmpWriteOut: cant open bmp %s\n", imageName)
		return
	}
	img, err := bmp.Decode(bf)
	if err != nil {
		fmt.Printf("bmpWriteOut: bmp decode failed for %s\n", imageName)
		w.Write([]byte(fmt.Sprintf("Decode failed for %s\n", imageName)))
		return
	}
	b := make([]byte, 0, 10000)
	wo := bytes.NewBuffer(b)
	err = png.Encode(wo, img)
	if err != nil {
		fmt.Printf("bmpWriteOut: png encode failed for %s\n", imageName)
		return
	}
	w.Write(wo.Bytes())
}

func bmpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("bmpHandler: r.URL.Path %q\n", r.URL.Path)
	if r.URL.Path == bmpURL {
		w.Write(myBmpDir)
		return
	}
	fileName := serverRoot + r.URL.Path[1:]
	fmt.Printf("bmpHandler: fname = %s\n", fileName)
	ext := strings.ToLower(filepath.Ext(fileName))
	//fmt.Printf("ext = %s\n",ext)
	if ext == ".bmp" {
		bmpWriteOut(fileName, w)
		return
	}
}

// ---------------------------------------------------------------------------     M A I N

func main() {
	// Handle(serverRoot, is like a dir missing an index "ftp-style"
	// http.Handle(serverRoot, http.StripPrefix(serverRoot, http.FileServer(http.Dir(serverRoot))))
	http.HandleFunc(mdURL, mdHandler)
	http.HandleFunc(imgURL, imageHandler)
	http.HandleFunc(bmpURL, bmpHandler)

	log.Printf("load markdown urls with %s/www/md\n", listenOnPort)
	log.Printf("load image urls with %s/www/image\n", listenOnPort)
	log.Printf("load bmp urls with %s/www/bmp\n", listenOnPort)
	log.Printf("bmpwv is ready to serve at %s\n", listenOnPort)
	err := http.ListenAndServe(listenOnPort, nil)
	if err != nil {
		log.Printf("bmpwv: error running webserver %v", err)
	}
}
