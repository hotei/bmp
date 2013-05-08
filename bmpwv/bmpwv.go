// bmpwv.go (c) 2013 David Rook

package main

import (
	"github.com/hotei/bmp"
	// go 1.X only below here
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

//"time"
)

const (
	hostIPstr  = "127.0.0.1"
	portNum    = 8282
	virtualURL = "/virbmp/"
	vURL       = "/virbmp/"
	serverRoot = "/www/bmp/"
)

var (
	portNumString = fmt.Sprintf(":%d", portNum)
	listenOn      = hostIPstr + portNumString
	g_fileNames   []string
)

var myDir = []byte{}

func checkName(pathname string, info os.FileInfo, err error) error {
	//fmt.Printf("checking %s\n", pathname)
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
			g_fileNames = append(g_fileNames, filepath.Base(pathname))
		}
	}
	return nil
}

// show thumbnail as clickable link to original image
func makeLine(s string) []byte {
	//return []byte(fmt.Sprintf("<a href=\"%s\">View %s</a><br>\n",s,s))
	return []byte(fmt.Sprintf("<a href= \"%s\"><img src=\"%s\" height=200 width=300 align=center> %s Original Size</a><br>\n", s, s, s))
}

func init() {
	pathName := serverRoot
	g_fileNames = make([]string, 0, 20)
	myDir = []byte(`<html><!-- comment --><head><title>Test BMP package</title></head><body>click on image to see in original size<br>`) // {}
	stats, err := os.Stat(pathName)
	if err != nil {
		fmt.Printf("Can't get fileinfo for %s\n", pathName)
		os.Exit(1)
	}
	if stats.IsDir() {
		filepath.Walk(pathName, checkName)
	} else {
		fmt.Printf("this argument must be a directory (but %s isn't)\n", pathName)
		os.Exit(-1)
	}
	//fmt.Printf("g_fileNames = %v\n", g_fileNames)
	for _, val := range g_fileNames {
		//fmt.Printf("%v\n", val)
		line := makeLine(val)
		myDir = append(myDir, line...)
	}
	t := []byte(`</body></html>`)
	myDir = append(myDir, t...)
}

// working and pretty
func vbmp(w http.ResponseWriter, r *http.Request) {
	imageName := serverRoot + r.URL.Path[len(vURL):]
	fmt.Printf("vbmp: imageName = %s\n", imageName)
	if imageName == serverRoot {
		w.Write(myDir)
		return
	}
	bf, err := os.Open(imageName)
	if err != nil {
		fmt.Printf("vbmp: cant open bmp %s\n", imageName)
		return
	}
	img, err := bmp.Decode(bf)
	if err != nil {
		fmt.Printf("vbmp: bmp decode failed for %s\n", imageName)
		w.Write([]byte(fmt.Sprintf("Decode failed for %s\n",imageName)))
		return	
	}
	b := make([]byte, 0, 10000)
	wo := bytes.NewBuffer(b)
	err = png.Encode(wo, img)
	if err != nil {
		fmt.Printf("vbmp: png encode failed for %s\n", imageName)
		return	
	}
	w.Write(wo.Bytes())
}

// ---------------------------------------------------------------------------     M A I N

func main() {
	//	http.HandleFunc(virtualURL, html)
	http.HandleFunc(vURL, vbmp)
	http.Handle(serverRoot, http.StripPrefix(serverRoot, http.FileServer(http.Dir(serverRoot))))
	log.Printf("bmpwv is ready to serve at %s\n", listenOn)
	err := http.ListenAndServe(listenOn, nil)
	if err != nil {
		log.Printf("bmpwv: error running webserver %v", err)
	}
}


