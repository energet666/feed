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
	"strconv"

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

	http.Handle("/", http.HandlerFunc(myFIleServerHandler))
	http.Handle("/upload", websocket.Handler(s.HandleWsUpload))
	http.Handle("/ws", websocket.Handler(s.HandleWs))

	go http.ListenAndServe(":"+port, nil)

	log.Println("FileServer listen localhost:" + port + " ...")
	log.Println("WebSocketServer listen localhost:" + port + "/ws...")
	log.Println("WebSocketServer listen localhost:" + port + "/upload...")

	startBrowser("http://localhost:" + port)

	var read []byte
	for {
		fmt.Scan(&read)
		fmt.Println("read: ", string(read))
		s.Broadcast(read, "ws")
	}
}

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
