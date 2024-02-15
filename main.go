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

	s := wsserver.NewWsServer(param.ContentPath)

	http.Handle("/", http.HandlerFunc(s.HandleRoot))
	http.Handle("/upload/", http.HandlerFunc(s.HandleUploadDir))
	http.Handle("/uploadxml/", http.HandlerFunc(s.HandleUploadxml))

	http.Handle("/ws", websocket.Handler(s.HandleWs))
	http.Handle("/uploadws", websocket.Handler(s.HandleUploadws))
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

type parameters struct {
	Port        uint16
	ContentPath string
}

func readParametersFromToml(tomlFileName string, dataStruct *parameters) error {
	// optsData, err := os.ReadFile(tomlFileName)
	// if err != nil {
	// 	return fmt.Errorf("ошибка открытия файла конфигурации: %s", err)
	// }
	// _, err = toml.Decode(string(optsData), &dataStruct)
	_, err := toml.DecodeFile(tomlFileName, &dataStruct)
	if err != nil {
		if os.IsNotExist(err) {
			mokupParams := parameters{
				Port:        12345,
				ContentPath: "/some/path",
			}
			f, _ := os.Create(tomlFileName)
			toml.NewEncoder(f).Encode(mokupParams)
			return fmt.Errorf("не найден файл конфигурации %s. Файл создан, впишите в него необходимые значения параметров", tomlFileName)
		}
		return fmt.Errorf("ошибка декодирования файла конфигурации: %s", err)
	}
	// if tempDataStruct.Port > 0xFFFF {
	// 	return fmt.Errorf("Port must be a number in range 0..65535!")
	// }
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
