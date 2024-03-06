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
	"text/tabwriter"
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
	var conns sync.Map
	var connsUpload sync.Map
	up := filepath.Join(contentPath, "/upload")
	ms := filepath.Join(contentPath, "/msg")

	err := createDirIfNotExist(up)
	if err != nil {
		return nil, err
	}
	err = createDirIfNotExist(ms)
	if err != nil {
		return nil, err
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
	s.connsUpload.Store(ws, int(len(s.files)-1)) //самый свежий файл(последний в слайсе)
	buf := make([]byte, 1024)
	log.Println(`New incoming WS "uploadws" connection from client:`, ws.Request().RemoteAddr)
	printSyncMapStringString(s.conns)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println(`WS "uploadws" connection closed: `, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				s.connsUpload.Delete(ws)
				printSyncMapStringString(s.conns)
				break
			}
			log.Println(`WS "uploadws" read error: `, err)
			continue
		}
		msg := buf[:n]
		log.Println("Recived from uploadws: " + string(msg))
		switch string(msg) {
		case "old":
			if n, _ := s.connsUpload.Load(ws); n.(int) != 0 {
				s.connsUpload.Store(ws, n.(int)-1)
				nfile, _ := s.connsUpload.Load(ws)
				comandJSON, _ := json.Marshal(comand{
					Cmd: `append`,
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

type message struct {
	Id  string //имя файла, к которому оставлен коментарий
	Txt string //текст коментария
}

func (s *wsServer) HandleWs(ws *websocket.Conn) {
	s.conns.Store(ws, "ws")
	buf := make([]byte, 1024)
	log.Println(`New incoming WS "ws" connection from client:`, ws.Request().RemoteAddr)
	// fmt.Println("WS list: ", s.conns)
	printSyncMapStringString(s.conns)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println(`WS "ws" connection closed: `, ws.Request().RemoteAddr)
				s.conns.Delete(ws)
				// fmt.Println("WS list: ", s.conns)
				printSyncMapStringString(s.conns)
				break
			}
			log.Println(`WS "ws" read error: `, err)
			continue
		}
		msg := buf[:n]

		var msgStruct message
		json.Unmarshal(msg, &msgStruct)
		fo, err := os.OpenFile(
			filepath.Join(s.contentPathMsg, strings.TrimPrefix(msgStruct.Id, "/upload/")+"._msg"),
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0644)
		if err != nil {
			log.Println(err)
		}
		fo.Write([]byte(msgStruct.Txt + "\n"))
		fo.Close()
		s.Broadcast(msg, "ws")
	}
}

// type MouseXY struct {
// 	X int
// 	Y int
// }

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
		// var mxy MouseXY
		// json.Unmarshal(msg, &mxy)
		// fmt.Println(string(msg))
		// fmt.Println(mxy.X, mxy.Y)
	}
}

func (s *wsServer) HandleUploadxml(w http.ResponseWriter, r *http.Request) {
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	out, err := os.Create(filepath.Join(s.contentPathUpload, header.Filename))
	if err != nil {
		log.Printf("Unable to create the file for writing. Check your write access privilege")
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
	comandJSON, _ := json.Marshal(
		comand{
			Cmd: `prepend`,
			Arg: `/upload/` + header.Filename,
		})
	s.Broadcast(comandJSON, "uploadws")
}

func (s *wsServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleRoot:", r.RemoteAddr, r.Method, r.URL.Path)
	w.Header().Set("Cache-Control", "no-store")
	http.FileServer(http.Dir("www")).ServeHTTP(w, r)
}

func (s *wsServer) HandleUploadDir(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleUploadDir:", r.RemoteAddr, r.Method, r.URL.Path)
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

func printSyncMapStringString(m *sync.Map) {
	w := tabwriter.NewWriter(os.Stdout, 5, 0, 2, ' ', 0)
	m.Range(func(key, value any) bool {
		fmt.Fprintf(w, "%v\t%v\n", key.(*websocket.Conn).Request().RemoteAddr, value.(string))
		return true
	})
	w.Flush()
}

func createDirIfNotExist(fname string) error {
	_, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Директория %s не найдена", fname)
			err := os.Mkdir(fname, 0644)
			if err != nil {
				return fmt.Errorf("не удалось создать директорию %s: %s", fname, err)
			}
			log.Printf("Директория %s создана", fname)
		} else {
			return fmt.Errorf("что-то не так с директорией %s: %s", fname, err)
		}
	}
	return nil
}
