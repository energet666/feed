package wsserver

import (
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/net/websocket"
)

func (s *wsServer) HandleEventws(ws *websocket.Conn) {
	s.conns.Store(ws, "eventws")
	buf := make([]byte, 1024)
	log.Println(`New incoming WS "eventws" connection from client:`, ws.Request().RemoteAddr)
	printSyncMapStringString(s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println(`WS "eventws" connection closed: `, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				printSyncMapStringString(s.conns)
				break
			}
			log.Println(`WS "eventws" read error: `, err)
			continue
		}
		msg := buf[:n]
		log.Printf("WS message from %s \"eventws\": %s\n", ws.Request().RemoteAddr, msg)
		fmt.Fprintf(ws, "%s %s", time.Now().Format(time.TimeOnly), msg) //echo
	}
}
