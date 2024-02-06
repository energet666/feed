package wsserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/net/websocket"
)

type wsServer struct {
	conns map[*websocket.Conn]string
}

func NewWsServer() *wsServer {
	return &wsServer{
		conns: make(map[*websocket.Conn]string),
	}
}

func (s *wsServer) HandleUploadws(ws *websocket.Conn) {
	s.conns[ws] = "uploadws"
	buf := make([]byte, 1024)
	fmt.Println(`new incoming "uploadws" connection from client:`, ws.Request().RemoteAddr)
	fmt.Println("WS list: ", s.conns)

	files, _ := os.ReadDir("./www/upload")
	sort.Slice(files, func(i, j int) bool {
		finfoi, _ := files[i].Info()
		finfoj, _ := files[j].Info()
		return finfoi.ModTime().Before(finfoj.ModTime())
	})
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		ext = strings.ToLower(ext)
		if ext != "._msg" {
			ws.Write([]byte(`./upload/` + f.Name()))
		}
	}

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("connection closed: ", ws.Request().RemoteAddr)
				delete(s.conns, ws)
				fmt.Println("WS list: ", s.conns)
				break
			}
			fmt.Println("uploadws read error: ", err)
			continue
		}
		msg := buf[:n]
		fmt.Println("Recived from uploadws: " + string(msg))
	}
}

type Message struct {
	Id  string //имя файла, к которому оставлен коментарий
	Txt string //текст коментария
}

func (s *wsServer) HandleWs(ws *websocket.Conn) {
	s.conns[ws] = "ws"
	buf := make([]byte, 1024)
	fmt.Println(`New incoming "ws" connection from client:`, ws.Request().RemoteAddr)
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

		var msgs Message
		json.Unmarshal(msg, &msgs)
		fo, _ := os.OpenFile("./www/"+msgs.Id+"._msg", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		fo.Write([]byte(msgs.Txt + "\n"))
		fo.Close()
		s.Broadcast(msg, "ws")
	}
}

type MouseXY struct {
	X int
	Y int
}

func (s *wsServer) HandleEventws(ws *websocket.Conn) {
	s.conns[ws] = "eventws"
	buf := make([]byte, 1024)
	fmt.Println(`New incoming "eventws" connection from client:`, ws.Request().RemoteAddr)
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

		var mxy MouseXY
		json.Unmarshal(msg, &mxy)
		fmt.Println(string(msg))
		fmt.Println(mxy.X, mxy.Y)

		ws.Write([]byte(msg))
	}
}

func (s *wsServer) HandleUploadxml(w http.ResponseWriter, r *http.Request) {
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	out, err := os.Create("./www/upload/" + header.Filename)
	if err != nil {
		fmt.Printf("Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s %d Bytes saved\n", header.Filename, header.Size)
	s.Broadcast([]byte(`./upload/`+header.Filename), "uploadws")
}

func (s *wsServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, r.Method, r.RequestURI)
	w.Header().Set("Cache-Control", "no-cache")
	http.FileServer(http.Dir("www")).ServeHTTP(w, r)
}

func (s *wsServer) Broadcast(b []byte, t string) {
	for ws, tt := range s.conns {
		if tt == t {
			// go func(wst *websocket.Conn) {
			// 	wst.Write(b)
			// }(ws)
			ws := ws //без этого не захватывается переменная в замыкании
			go func() {
				ws.Write(b)
			}()
		}
	}
}
