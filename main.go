package main

import (
	"feed/pkg/wsserver"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/net/websocket"
)

func main() {
	var param parameters
	err := readParametersFromToml("config.toml", &param)
	if err != nil {
		log.Println("ошибка чтения параметров:", err)
		return
	}
	fmt.Printf("%+v\n", param)
	port := fmt.Sprint(param.Port)

	s, err := wsserver.NewWsServer(param.ContentPath)
	if err != nil {
		log.Println("ошибка создания сервера:", err)
		return
	}

	myMux := http.NewServeMux()

	myMux.Handle("/", http.HandlerFunc(s.HandleRoot))
	myMux.Handle("/upload/{path...}", http.HandlerFunc(s.HandleUploadDir))
	myMux.Handle("/uploadxml/", http.HandlerFunc(s.HandleUploadxml))

	myMux.Handle("/ws", websocket.Handler(s.HandleWs))
	myMux.Handle("/uploadws", websocket.Handler(s.HandleUploadws))
	myMux.Handle("/eventws", websocket.Handler(s.HandleEventws))

	httpServ := &http.Server{
		Addr:           ":" + port,
		Handler:        myMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		if err := httpServ.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	log.Println("FileServer listen localhost:" + port + " ...")
	log.Println("WebSocketServer listen localhost:" + port + "/ws...")
	log.Println("WebSocketServer listen localhost:" + port + "/uploadws...")

	startBrowser("http://localhost:" + port)

	wg.Wait()
}

type parameters struct {
	Port        uint16
	ContentPath string
}

func readParametersFromToml(tomlFileName string, dataStruct *parameters) error {
	_, err := toml.DecodeFile(tomlFileName, &dataStruct)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("не найден файл конфигурации %s\n", tomlFileName)
			mokupParams := parameters{
				Port:        12345,
				ContentPath: "/some/path",
			}
			f, err := os.Create(tomlFileName)
			if err != nil {
				return fmt.Errorf("не удалось создать файл конфигурации %s: %s", tomlFileName, err)
			}
			toml.NewEncoder(f).Encode(mokupParams)
			return fmt.Errorf("файл %s создан, впишите в него необходимые значения параметров и снова запустите пприложение", tomlFileName)
		}
		return fmt.Errorf("ошибка декодирования файла конфигурации: %s", err)
	}
	dataStruct.ContentPath = strings.ReplaceAll(dataStruct.ContentPath, `\`, `/`)
	dataStruct.ContentPath = path.Clean(dataStruct.ContentPath)
	_, err = os.Stat(dataStruct.ContentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("директория %s не существует", dataStruct.ContentPath)
		}
		return fmt.Errorf("ошибка получения информации о директории %s: %s", dataStruct.ContentPath, err)
	}
	return nil
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
