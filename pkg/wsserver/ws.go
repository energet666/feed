package wsserver

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/websocket"
)

type message struct {
	Id  string //имя файла, к которому оставлен коментарий
	Txt string //текст коментария
}

func (s *wsServer) HandleWs(ws *websocket.Conn) {
	s.conns.Store(ws, "ws")
	buf := make([]byte, 1024)
	log.Println(`New incoming WS "ws" connection from client:`, ws.Request().RemoteAddr)
	printSyncMapStringString(s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println(`WS "ws" connection closed: `, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				printSyncMapStringString(s.conns)
				break
			}
			log.Println(`WS "ws" read error: `, err)
			continue
		}
		msg := buf[:n]

		var msgStruct message
		err = json.Unmarshal(msg, &msgStruct)
		if err != nil {
			log.Println(err)
		}
		fo, err := os.OpenFile(
			filepath.Join(s.contentPathMsg, strings.TrimPrefix(msgStruct.Id, "/upload/")+"._msg"),
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0644)
		if err != nil {
			log.Println(err)
		}
		fo.WriteString(msgStruct.Txt + "\n")
		fo.Close()
		s.Broadcast(msg, "ws")
	}
}
