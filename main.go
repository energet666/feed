package main

import (
	"feed/pkg/wsserver"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

func inUpload(ws *websocket.Conn) {
	files, _ := os.ReadDir("./www/upload")
	sort.Slice(files, func(i, j int) bool {
		finfoi, _ := files[i].Info()
		finfoj, _ := files[j].Info()
		return finfoi.ModTime().Before(finfoj.ModTime())
	})
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		ext = strings.ToLower(ext)
		if ext != "._msg" {
			ws.Write([]byte(`./upload/` + f.Name()))
		}
	}
}

func main() {
	port := "12345"
	if len(os.Args) > 1 {
		if len(os.Args[1]) > 5 {
			fmt.Println("Port must be a number in range 0..99999!")
			return
		}
		_, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("Port must be a number in range 0..99999!")
			return
		}
		port = os.Args[1]
	}

	s := wsserver.NewWsServer(inUpload)

	http.Handle("/uploadxml/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// the FormFile function takes in the POST input id file
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}

		defer file.Close()

		out, err := os.Create("./www/upload/" + header.Filename)
		if err != nil {
			fmt.Printf("Unable to create the file for writing. Check your write access privilege")
			return
		}

		defer out.Close()

		// write the content from POST to the file
		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s %d Bytes saved\n", header.Filename, header.Size)
		s.Broadcast([]byte(`./upload/`+header.Filename), "uploadws")
	}))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RemoteAddr, r.Method, r.RequestURI)
		http.FileServer(http.Dir("www")).ServeHTTP(w, r)
	}))

	http.Handle("/uploadws", websocket.Handler(s.HandleWsUpload))
	http.Handle("/ws", websocket.Handler(s.HandleWs))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		http.ListenAndServe(":"+port, nil)
		wg.Done()
	}()

	log.Println("FileServer listen localhost:" + port + " ...")
	log.Println("WebSocketServer listen localhost:" + port + "/ws...")
	log.Println("WebSocketServer listen localhost:" + port + "/uploadws...")

	startBrowser("http://localhost:" + port)

	// var read []byte
	// for {
	// 	fmt.Scan(&read)
	// 	fmt.Println("read: ", string(read))
	// 	s.Broadcast(read, "ws")
	// }
	wg.Wait()
}

func startBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
