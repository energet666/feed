package wsserver

import (
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/net/websocket"
)

func (s *wsServer) HandleEventws(ws *websocket.Conn) {
	websocketTag := "eventws"
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
		log.Printf(`WS message from %s "%s": %s\n`, ws.Request().RemoteAddr, websocketTag, msg)
		fmt.Fprintf(ws, "%s %s", time.Now().Format(time.TimeOnly), msg) //echo
	}
}
