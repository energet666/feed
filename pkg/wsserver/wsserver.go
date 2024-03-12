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
	conns             *sync.Map     //[вебсокет]тип вебсокета
	connsUpload       *sync.Map     //[вебсокет типа "uploadws"]индекс последнего отданного файла из слайса "files"
	contentPath       string        //директория с контентом и комментариями
	contentPathUpload string        //директория с контентом, дочерняя для "contentPath"
	contentPathMsg    string        //директория с комментариями, дочерняя для "contentPath"
	files             []fs.DirEntry //отсортированный по дате слайс файлов из директории "contentPathUpload". [0] - самый старый
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
			n, ok := s.connsUpload.Load(ws)
			nt := len(s.files)
			if ok {
				nt = n.(int)
			}
			if nt < 1 {
				continue
			}
			comandJSON, _ := json.Marshal(comand{
				Cmd: `append`,
				Arg: `/upload/` + s.files[nt-1].Name(),
			})
			ws.Write(comandJSON)
			s.connsUpload.Store(ws, nt-1)
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
		log.Printf("Unable to create the file for writing: %s", err)
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
	log.Printf("%s %d Bytes saved\n", header.Filename, header.Size)
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

type BufferedWriterForTabwriter []byte

func (m *BufferedWriterForTabwriter) Write(p []byte) (int, error) {
	*m = append(*m, p...)
	return len(p), nil
}
func (m *BufferedWriterForTabwriter) Print() {
	fmt.Printf("%s\n", *m)
	*m = []byte{}
}

func printSyncMapStringString(m *sync.Map) {
	var writer BufferedWriterForTabwriter
	w := tabwriter.NewWriter(&writer, 5, 0, 2, ' ', 0)
	m.Range(func(key, value any) bool {
		fmt.Fprintf(w, "%v\t%v\n", key.(*websocket.Conn).Request().RemoteAddr, value.(string))
		return true
	})
	w.Flush()
	writer.Print()
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
