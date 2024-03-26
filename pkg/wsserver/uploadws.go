package wsserver

import (
	"encoding/json"
	"io"
	"log"
	"strconv"

	"golang.org/x/net/websocket"
)

type fileInfo struct { //номер файла в слайсе и path
	N    int
	Path string
}

func (s *wsServer) HandleUploadws(ws *websocket.Conn) {
	s.conns.Store(ws, "uploadws")
	buf := make([]byte, 1024)
	log.Println(`New incoming WS "uploadws" connection from client:`, ws.Request().RemoteAddr)
	printSyncMapStringString(s.conns)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println(`WS "uploadws" connection closed: `, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				printSyncMapStringString(s.conns)
				break
			}
			log.Println(`WS "uploadws" read error: `, err)
			continue
		}
		msg := buf[:n]
		log.Printf(`Recived from %s "uploadws": %s`, ws.Request().RemoteAddr, msg)

		npost, err := strconv.Atoi(string(msg))
		if err != nil {
			log.Printf("ошибка чтения номера файла из сообщения: %s\n", err)
			continue
		}
		//отрицательные числа обрабатываем как запрос самого нового файла
		//запросы индекса за границей слайса игнорируем
		//если слайс пустой то тоже игнорируем
		if npost >= len(s.files) || len(s.files) == 0 {
			continue
		} else if npost < 0 {
			npost = len(s.files) - 1
		}
		comandJSON, _ := json.Marshal(fileInfo{
			N:    npost,
			Path: `/upload/` + s.files[npost].Name(),
		})
		ws.Write(comandJSON)
	}
}
