package wsserver

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/websocket"
)

type wsServer struct {
	conns          map[*websocket.Conn]string
	uploadCallback func(ws *websocket.Conn)
}

func NewWsServer(f func(ws *websocket.Conn)) *wsServer {
	return &wsServer{
		conns:          make(map[*websocket.Conn]string),
		uploadCallback: f,
	}
}

func (s *wsServer) HandleWsUpload(ws *websocket.Conn) {
	s.conns[ws] = "upload"
	buf := make([]byte, 1024)
	fmt.Println("new incoming \"upload\" connection from client:", ws.Request().RemoteAddr)
	fmt.Println("WS list: ", s.conns)
	var state int = 0
	var fo *os.File
	var name string
	var size int
	var nn int = 0

	s.uploadCallback(ws)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("connection closed: ", ws.Request().RemoteAddr)
				delete(s.conns, ws)
				fmt.Println("WS list: ", s.conns)
				break
			}
			fmt.Println("read error: ", err)
			continue
		}
		msg := buf[:n]
		switch state {
		case 0:
			name = string(msg)
			fmt.Println("File name: ", string(name))
			os.Create("./www/upload/" + name)
			nn = 0
			state = 1
		case 1:
			size, _ = strconv.Atoi(string(msg))
			fmt.Println("Size: ", size)
			state = 2
		case 2:
			fo, err = os.OpenFile("./www/upload/"+name, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("OpenFile err: ", err)
			}
			n, err := fo.Write(msg)
			nn += n
			if err != nil {
				fmt.Println("Write err: ", err)
			}
			fo.Close()
			if nn == size {
				fmt.Println("Uploading complite!")
				state = 0

				s.Broadcast(HtmlBlockFromFileName(name), "upload")
			}
		}
	}
}

func HtmlBlockFromFileName(name string) []byte {
	spl := strings.Split(name, ".")
	ext := spl[len(spl)-1]
	switch strings.ToLower(ext) {
	case "mp4":
		return ([]byte(`<video controls>
								<source  src="./upload/` + name + `" type="video/mp4">
								Your browser does not support the video tag.
							</video>`))
	case "mp3":
		return ([]byte(`<a href="./upload/` + name + `" download>` + name + `</a>
							<audio controls>
								<source loading="lazy" src="./upload/` + name + `" type="audio/mpeg">
								Your browser does not support the audio element.
							</audio>`))
	case "jpg", "png", "jpeg":
		return ([]byte(`<img loading="lazy" src="./upload/` + name + `">`))
	default:
		return ([]byte(`<a href="./upload/` + name + `">` + name + `</a>`))
	}
}

func (s *wsServer) HandleWs(ws *websocket.Conn) {
	s.conns[ws] = "ws"
	buf := make([]byte, 1024)
	fmt.Println("New incoming \"ws\" connection from client:", ws.Request().RemoteAddr)
	fmt.Println("WS list: ", s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed: ", ws.Request().RemoteAddr)
				delete(s.conns, ws)
				fmt.Println("WS list: ", s.conns)
				break
			}
			fmt.Println("Read error: ", err)
			continue
		}
		msg := buf[:n]
		s.Broadcast(msg, "ws")
	}
}

func (s *wsServer) Broadcast(b []byte, t string) {
	for ws := range s.conns {
		if s.conns[ws] == t {
			go func(wst *websocket.Conn) {
				wst.Write(b)
			}(ws)
		}
	}
}