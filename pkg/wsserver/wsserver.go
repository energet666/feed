package wsserver

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"golang.org/x/net/websocket"
)

type wsServer struct {
	conns             *sync.Map     //[вебсокет]тип вебсокета
	contentPath       string        //директория с контентом и комментариями
	contentPathUpload string        //директория с контентом, дочерняя для "contentPath"
	contentPathMsg    string        //директория с комментариями, дочерняя для "contentPath"
	files             []fs.DirEntry //отсортированный по дате слайс файлов из директории "contentPathUpload". [0] - самый старый
}

func NewWsServer(contentPath string) (*wsServer, error) {
	var conns sync.Map

	up := filepath.Join(contentPath, "upload")
	ms := filepath.Join(contentPath, "msg")
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
		contentPath:       contentPath,
		contentPathUpload: up,
		contentPathMsg:    ms,
		files:             files,
	}, nil
}

func (s *wsServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("%-15s %21s %-10s %s", "HandleRoot", r.RemoteAddr, r.Method, r.URL.Path)
	w.Header().Set("Cache-Control", "no-store")
	http.FileServer(http.Dir("www")).ServeHTTP(w, r)
}

func (s *wsServer) HandleUploadDir(w http.ResponseWriter, r *http.Request) {
	log.Printf("%-15s %21s %-10s %s", "HandleUploadDir", r.RemoteAddr, r.Method, r.URL.Path)
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

func createDirIfNotExist(dirName string) error {
	_, err := os.Stat(dirName)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(dirName, 0644)
			if err != nil {
				return fmt.Errorf("не удалось создать директорию %s: %s", dirName, err)
			}
			log.Printf("Директория %s создана", dirName)
			return nil
		}
		return fmt.Errorf("что-то не так с директорией %s: %s", dirName, err)
	}
	return nil
}
