package main

import (
	"feed/pkg/wsserver"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"

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
		ws.Write(wsserver.HtmlBlockFromFileName(f.Name()))
	}
}

func main() {
	s := wsserver.NewWsServer(inUpload)

	// http.Handle("/", myHandler{http.Dir("www")})
	http.Handle("/", http.HandlerFunc(myFIleServerHandler))
	http.Handle("/upload", websocket.Handler(s.HandleWsUpload))
	http.Handle("/ws", websocket.Handler(s.HandleWs))

	go http.ListenAndServe(":12345", nil)

	log.Println("FileServer listen localhost:12345 ...")
	log.Println("WebSocketServer listen localhost:12345/ws...")
	log.Println("WebSocketServer listen localhost:12345/upload...")

	startBrowser("http://localhost:12345")

	var read []byte
	for {
		fmt.Scan(&read)
		s.Broadcast(read, "ws")
	}
}

// type myHandler struct {
// 	root http.FileSystem
// }

// func (f myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println(r.RemoteAddr, r.Method, r.RequestURI)
// 	http.FileServer(f.root).ServeHTTP(w, r)
// }

func myFIleServerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, r.Method, r.RequestURI)
	http.FileServer(http.Dir("www")).ServeHTTP(w, r)
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
