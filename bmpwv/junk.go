// +build ignore

// ContentType = ?
	http.ServeFile(w,r,"/www/bmp/png.png")
	
	imageName := r.URL.Path
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
	// ReadSeeker required for ServeContent, os.File or bytes.Buffer
	//t := time.Now()
	// http.ServeContent(w,r,imageName,t,f)
	
	time.Sleep(time.Second)
	var t = []byte("<img src=\"png.png\">")
	w.Write(t)
	
	
	
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
	fileName = serverRoot + fileName
	if ext == ".jpg" {
		var t = []byte("<img src=" + fileName + ">")
		w.Write(t)
		return
	}
	if ext == ".bmp" {
		dead_bmpHandler(w, r, fileName)
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
	fmt.Printf("bmpHandler imageName = %s\n", imageName)
	bf, err := os.Open(imageName)
	if err != nil {
		fmt.Printf("bmpHandler cant open bmp %s\n", imageName)
		return
	}
	img, err := bmp.Decode(bf)
	f, err := os.Create("/www/bmp/png.png")
	if err != nil {
		fmt.Printf("%v \n", err)
		return
	}
	wo := bufio.NewWriter(f)
	png.Encode(wo, img)
	wo.Flush()
	var t = []byte("<img src=" + imageName + ">")
	w.Write(t)
}
