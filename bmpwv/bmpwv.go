// bmpwv.go (c) 2013 David Rook

// working but I don't like the current solution  
// 1) leaves files after creating bmp->png - when is it safe to remove - and how?
// 2) Is it thread-safe?


package main

import (
	"github.com/hotei/bmp"
	// go 1.X only below here
	"bytes"
	"bufio"
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
	hostIPstr = "127.0.0.1" 
	portNum = 8282
	virtualURL = "/virbmp/"
	vURL = "/virbmp/"
	serverRoot = "/www/bmp/"	
)

var (
	portNumString = fmt.Sprintf(":%d", portNum)
	listenOn = hostIPstr + portNumString
	g_fileNames []string
)

var myDir = []byte(`
<h3>Cant find index.html</h3>
<a href="bit8.bmp">Test with bit8.bmp</a><br>
<a href="whirlpool.jpg">show whirlpool</a><br>
<a href="two.html">test html file</a><br>
<img src="whirlpool.jpg"><br>
<img src="two.png"><br>
<!-- comment -->
`)


func checkName(pathname string, info os.FileInfo, err error) error {
	fmt.Printf("checking %s\n",pathname)
	if info == nil {
		fmt.Printf("WARNING --->  no stat info: %s\n", pathname)
		os.Exit(1)
	}
	if info.IsDir() {
		// return filepath.SkipDir
	} else { // regular file
		fmt.Printf("found %s %s\n",pathname,filepath.Ext(pathname))
		if filepath.Ext(pathname) == ".bmp" {
			fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, filepath.Base(pathname))
		}
	}
	return nil
}

// given fname.bmp return "<a href="fname.bmp">View fname.bmp</a><br>"
func makeLine(s string) []byte {
	//return []byte(fmt.Sprintf("<a href=\"%s\">View %s</a><br>\n",s,s))
	return []byte(fmt.Sprintf("<a href= \"%s\"><img src=\"%s\" height=200 width=200>%s Original</a><br>\n",s,s,s))
}

func init() {
	pathName := serverRoot
	g_fileNames = make([]string,0,20)
	myDir = []byte{}
	myDir = []byte(`<a href="256colorOS2v1.bmp">View 256colorOS2v1.bmp</a><br>
<a href="bit1bw-rnr.bmp">View bit1bw-rnr.bmp</a><br>
<a href="bit1color2.bmp">View bit1color2.bmp</a><br>
<a href="bit24-test.bmp">View bit24-test.bmp</a><br>
<a href="bit24-teststrip.bmp">View bit24-teststrip.bmp</a><br>
<a href="bit24uncomp-marbles.bmp">View bit24uncomp-marbles.bmp</a><br>
<a href="bit24uncomp-rnr.bmp">View bit24uncomp-rnr.bmp</a><br>
<a href="bit4-test.bmp">View bit4-test.bmp</a><br>
<a href="bit4comp-test.bmp">View bit4comp-test.bmp</a><br>
<a href="bit8-gray-rnr.bmp">View bit8-gray-rnr.bmp</a><br>
<a href="bit8-test.bmp">View bit8-test.bmp</a><br>
<a href="bit8.bmp">View bit8.bmp</a><br>
<a href="bit8comp-rnr.bmp">View bit8comp-rnr.bmp</a><br>
<a href="bit8comp-test.bmp">View bit8comp-test.bmp</a><br>
<a href="notBMP.bmp">View notBMP.bmp</a><br>`)

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
	fmt.Printf("g_fileNames = %v\n",g_fileNames)
	for _,val := range g_fileNames {
		fmt.Printf("%v\n",val)
		line := makeLine(val)
		myDir = append(myDir,line...)
	}
	t := []byte("<img src=\"bit8.bmp\">bit8.bmp<br>\n")
	myDir = append(myDir, t...)
}


// working and pretty
func vbmp(w http.ResponseWriter, r *http.Request) {
	imageName := serverRoot + r.URL.Path[len(vURL):]
	fmt.Printf("vbmp imageName = %s\n",imageName)
	if imageName == serverRoot {
		w.Write(myDir)
		return
	}
	bf,err := os.Open(imageName)
	if err != nil {
		fmt.Printf("bmpHandler cant open bmp %s\n",imageName)
		return
	}
	img,err := bmp.Decode(bf)
	b := make([]byte,0,10000)
	wo := bytes.NewBuffer(b)
	png.Encode(wo,img)
	w.Write(wo.Bytes())
}

// ---------------------------------------------------------------------------     M A I N

func main() {
//	http.HandleFunc(virtualURL, html)
	http.HandleFunc(vURL, vbmp)
	http.Handle(serverRoot,http.StripPrefix(serverRoot, http.FileServer(http.Dir(serverRoot))))
	log.Printf("Web server is ready at %s\n", listenOn)
	err := http.ListenAndServe(listenOn, nil)
	if err != nil {
		log.Printf("bmpwv: error running webserver %v", err)
	}
}

// working but not used
func dead_html(w http.ResponseWriter, r *http.Request) {
	var output []byte
	var err error
	fmt.Fprintf(w, "<!-- %s %v -->", r.Method, r.URL) // debug request input
	if len(r.URL.Path) == len(virtualURL) {
		// browse directory via index.html, don't allow 'raw' directory
		ndxBytes, err := ioutil.ReadFile(serverRoot + "index.html")
		if err != nil {
			w.Write(myDir)
			return
		}
		w.Write(ndxBytes)
		return
	}
	// if tail of URL.Path == ".bmp" then 
	urlOffset := len(virtualURL)
	fileName := r.URL.Path[urlOffset:]
	ext := filepath.Ext(fileName)
	fileName = serverRoot+fileName
	if ext == ".jpg" {
		var t = []byte("<img src=" + fileName + ">")
		w.Write(t)
		return		
	}		
	if ext == ".bmp" {
		dead_bmpHandler(w,r,fileName)
		return
	} else {
		output, err = ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Printf("error %v\n", err)
			return
		}
	}
	w.Write(output)
	//fmt.Fprintf(w, "<-- %q %q-->",fileName, ext)
}

// working but ugly
func dead_bmpHandler(w http.ResponseWriter, r *http.Request, imageName string) {
	fmt.Printf("bmpHandler imageName = %s\n",imageName)
	bf,err := os.Open(imageName)
	if err != nil {
		fmt.Printf("bmpHandler cant open bmp %s\n",imageName)
		return
	}
	img,err := bmp.Decode(bf)
	f,err := os.Create("/www/bmp/png.png")
	if err != nil {
		fmt.Printf("%v \n",err)
		return
	}
	wo := bufio.NewWriter(f)
	png.Encode(wo,img)	
	wo.Flush()
	var t = []byte("<img src=" + imageName + ">")
	w.Write(t)
}		

