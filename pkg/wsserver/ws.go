package wsserver

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/websocket"
)

type message struct {
	Id  string //имя файла, к которому оставлен коментарий
	Txt string //текст коментария
}

func (s *wsServer) HandleWs(ws *websocket.Conn) {
	websocketTag := "ws"
	s.conns.Store(ws, websocketTag)
	buf := make([]byte, 1024)
	log.Printf("New incoming WS %q connection from client: %s", websocketTag, ws.Request().RemoteAddr)
	printSyncMapStringString(s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("WS %q connection closed: %s", websocketTag, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				printSyncMapStringString(s.conns)
				break
			}
			log.Printf("WS %q read error: %s", websocketTag, err)
			continue
		}
		msg := buf[:n]

		var msgStruct message
		err = json.Unmarshal(msg, &msgStruct)
		if err != nil {
			log.Printf("ошибка декодирования JSON сообщения: %s", err)
			continue
		}

		filepathRel, err := filepath.Rel("/upload", msgStruct.Id)
		if err != nil {
			log.Printf("ошибка получения относительного пути: %s", err)
			continue
		}
		fo, err := os.OpenFile(
			filepath.Join(
				s.contentPathMsg,
				filepathRel+"._msg",
			),
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0644,
		)
		if err != nil {
			log.Printf("ошибка открытия/создания файла с комментариями: %s", err)
			continue
		}
		defer fo.Close()
		fo.WriteString(msgStruct.Txt + "\n")
		s.Broadcast(msg, websocketTag)
	}
}
