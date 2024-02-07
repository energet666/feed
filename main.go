package main

import (
	"feed/pkg/wsserver"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"

	"github.com/BurntSushi/toml"
	"golang.org/x/net/websocket"
)

type opts struct {
	Port        string
	ContentPath string
}

func main() {
	port := "12345"
	optsData, err := os.ReadFile("config.toml")
	if err != nil {
		fmt.Println(err)
		return
	}
	var inputOpts opts
	_, err = toml.Decode(string(optsData), &inputOpts)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(inputOpts.Port) > 5 {
		fmt.Println("Port must be a number in range 0..99999!")
		return
	}
	_, err = strconv.Atoi(inputOpts.Port)
	if err != nil {
		fmt.Println("Port must be a number in range 0..99999!")
		return
	}
	port = inputOpts.Port

	s := wsserver.NewWsServer(inputOpts.ContentPath)

	http.Handle("/", http.HandlerFunc(s.HandleRoot))
	http.Handle("/upload/", http.HandlerFunc(s.HandleUploadDir))
	http.Handle("/uploadxml/", http.HandlerFunc(s.HandleUploadxml))
	http.Handle("/uploadws", websocket.Handler(s.HandleUploadws))
	http.Handle("/ws", websocket.Handler(s.HandleWs))
	http.Handle("/eventws", websocket.Handler(s.HandleEventws))

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
