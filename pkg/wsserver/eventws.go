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
		log.Printf("%-15s %21s %-10q %s", "Recived from", ws.Request().RemoteAddr, websocketTag, msg)
		fmt.Fprintf(ws, "%s %s", time.Now().Format(time.TimeOnly), msg) //echo
	}
}
