package wsserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type wsServer struct {
	conns             *sync.Map
	connsUpload       *sync.Map
	contentPath       string
	contentPathUpload string
	contentPathMsg    string
	files             []fs.DirEntry
}

func NewWsServer(contentPath string) (*wsServer, error) {
	up := filepath.Join(contentPath, "/upload")
	ms := filepath.Join(contentPath, "/msg")
	var conns sync.Map
	var connsUpload sync.Map
	_, err := os.Stat(up)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Директория %s не найдена", up)
			err := os.Mkdir(up, 0644)
			if err != nil {
				return nil, fmt.Errorf("не удалось создать директорию %s: %s", up, err)
			}
			log.Printf("Директория %s создана", up)
		} else {
			return nil, fmt.Errorf("что-то не так с директорией %s: %s", up, err)
		}
	}
	_, err = os.Stat(ms)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Директория %s не найдена", ms)
			err := os.Mkdir(ms, 0644)
			if err != nil {
				return nil, fmt.Errorf("не удалось создать директорию %s: %s", ms, err)
			}
			log.Printf("Директория %s создана", ms)
		} else {
			return nil, fmt.Errorf("что-то не так с директорией %s: %s", ms, err)
		}
	}
	files, _ := os.ReadDir(up)
	sort.Slice(files, func(i, j int) bool {
		finfoi, _ := files[i].Info()
		finfoj, _ := files[j].Info()
		return finfoi.ModTime().Before(finfoj.ModTime())
	})
	return &wsServer{
		conns:             &conns,
		connsUpload:       &connsUpload,
		contentPath:       contentPath,
		contentPathUpload: up,
		contentPathMsg:    ms,
		files:             files,
	}, nil
}

type comand struct {
	Cmd string
	Arg string
}

func (s *wsServer) HandleUploadws(ws *websocket.Conn) {
	s.conns.Store(ws, "uploadws")
	s.connsUpload.Store(ws, int(len(s.files)-1)) //самый свежий файл
	buf := make([]byte, 1024)
	fmt.Println(`new incoming "uploadws" connection from client:`, ws.Request().RemoteAddr)
	fmt.Println("WS list: ", s.conns)

	// ws.Write([]byte(`/upload/` + s.files[s.connsUpload[ws]].Name()))
	fnum, _ := s.connsUpload.Load(ws)
	comandJSON, _ := json.Marshal(comand{
		Cmd: "append",
		Arg: `/upload/` + s.files[fnum.(int)].Name(),
	})
	ws.Write(comandJSON)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("connection closed: ", ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				s.connsUpload.Delete(ws)
				fmt.Println("WS list: ", s.conns)
				break
			}
			fmt.Println("uploadws read error: ", err)
			continue
		}
		msg := buf[:n]
		fmt.Println("Recived from uploadws: " + string(msg))
		switch string(msg) {
		case "old":
			if n, _ := s.connsUpload.Load(ws); n.(int) != 0 {
				s.connsUpload.Store(ws, n.(int)-1)
				// ws.Write([]byte(`/upload/` + s.files[s.connsUpload[ws]].Name()))
				nfile, _ := s.connsUpload.Load(ws)
				comandJSON, _ := json.Marshal(comand{
					Cmd: "append",
					Arg: `/upload/` + s.files[nfile.(int)].Name(),
				})
				ws.Write(comandJSON)
			}
			// case "new":
			// 	if s.connsUpload[ws] != uint64(len(s.files)-1) {
			// 		s.connsUpload[ws] += 1
			// 		ws.Write([]byte(`/upload/` + s.files[s.connsUpload[ws]].Name()))
			// 	}
		}
	}
}

type Message struct {
	Id  string //имя файла, к которому оставлен коментарий
	Txt string //текст коментария
}

func (s *wsServer) HandleWs(ws *websocket.Conn) {
	s.conns.Store(ws, "ws")
	buf := make([]byte, 1024)
	fmt.Println(`New incoming "ws" connection from client:`, ws.Request().RemoteAddr)
	fmt.Println("WS list: ", s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed: ", ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				fmt.Println("WS list: ", s.conns)
				break
			}
			fmt.Println("Read error: ", err)
			continue
		}
		msg := buf[:n]

		var msgs Message
		json.Unmarshal(msg, &msgs)
		fo, err := os.OpenFile(
			filepath.Join(s.contentPathMsg, strings.TrimPrefix(msgs.Id, "/upload/")+"._msg"),
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0644)
		if err != nil {
			log.Println(err)
		}
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
	s.conns.Store(ws, "eventws")
	buf := make([]byte, 1024)
	fmt.Println(`New incoming "eventws" connection from client:`, ws.Request().RemoteAddr)
	fmt.Println("WS list: ", s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed: ", ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				fmt.Println("WS list: ", s.conns)
				break
			}
			fmt.Println("Read error: ", err)
			continue
		}
		msg := buf[:n]
		log.Println(string(msg))
		// var mxy MouseXY
		// json.Unmarshal(msg, &mxy)
		// fmt.Println(string(msg))
		// fmt.Println(mxy.X, mxy.Y)
		ws.Write([]byte(time.Now().Format(time.TimeOnly) + " Я ещё жив!"))
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

	out, err := os.Create(filepath.Join(s.contentPathUpload, header.Filename))
	if err != nil {
		fmt.Printf("Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println(err)
	}
	outInfo, _ := out.Stat()
	s.files = append(s.files, fs.FileInfoToDirEntry(outInfo))
	fmt.Printf("%s %d Bytes saved\n", header.Filename, header.Size)
	// s.Broadcast([]byte(`/upload/`+header.Filename), "uploadws")
	comandJSON, _ := json.Marshal(comand{
		Cmd: "prepend",
		Arg: `/upload/` + header.Filename,
	})
	s.Broadcast(comandJSON, "uploadws")
}

func (s *wsServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, r.Method, r.RequestURI)
	w.Header().Set("Cache-Control", "no-store")
	http.FileServer(http.Dir("www")).ServeHTTP(w, r)
}

func (s *wsServer) HandleUploadDir(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.Method, r.URL.Path)
	r.URL.Path = r.PathValue("path")
	if strings.ToLower(filepath.Ext(r.URL.Path)) == "._msg" {
		w.Header().Set("Cache-Control", "no-store")
		http.FileServer(http.Dir(s.contentPathMsg)).ServeHTTP(w, r)
	} else {
		http.FileServer(http.Dir(s.contentPathUpload)).ServeHTTP(w, r)
	}
}

func (s *wsServer) Broadcast(b []byte, t string) {
	s.conns.Range(func(ws, tt any) bool {
		if tt.(string) == t {
			go func() {
				ws.(*websocket.Conn).Write(b)
			}()
		}
		return true
	})
}
