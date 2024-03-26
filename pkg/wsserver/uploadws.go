package wsserver

import (
	"encoding/json"
	"io"
	"log"
	"path"
	"strconv"

	"golang.org/x/net/websocket"
)

type fileInfo struct { //номер файла в слайсе и path
	N    int
	Path string
}

func (s *wsServer) HandleUploadws(ws *websocket.Conn) {
	websocketTag := "uploadws"
	s.conns.Store(ws, websocketTag)
	buf := make([]byte, 1024)
	log.Printf(`New incoming WS "%s" connection from client: %s\n`, websocketTag, ws.Request().RemoteAddr)
	printSyncMapStringString(s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf(`WS "%s" connection closed: %s\n`, websocketTag, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				printSyncMapStringString(s.conns)
				break
			}
			log.Printf(`WS "%s" read error: %s`, websocketTag, err)
			continue
		}
		msg := buf[:n]
		log.Printf(`Recived from %s "%s": %s`, ws.Request().RemoteAddr, websocketTag, msg)

		npost, err := strconv.Atoi(string(msg))
		if err != nil {
			log.Printf("ошибка чтения номера файла из сообщения: %s\n", err)
			continue
		}
		//отрицательные числа обрабатываем как запрос самого нового файла
		//запросы индекса за границей слайса игнорируем
		if npost >= len(s.files) {
			continue
		} else if npost < 0 {
			npost = len(s.files) - 1
		}
		comandJSON, _ := json.Marshal(fileInfo{
			N:    npost,
			Path: path.Join(`/upload`, s.files[npost].Name()),
			//Name не возвращает путь к файлу, вложенные в папки файлы работать не будут
		})
		ws.Write(comandJSON)
	}
}
