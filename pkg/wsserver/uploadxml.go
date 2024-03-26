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
	"strings"
)

func (s *wsServer) HandleUploadxml(w http.ResponseWriter, r *http.Request) {
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	filename := header.Filename
	//Избавляемся от запрещенных символов
	filename = strings.ReplaceAll(filename, "#", "_")
	filenamePath := filepath.Join(s.contentPathUpload, filename)
	n := 0
	for {
		_, err := os.Stat(filenamePath)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
		}
		filename = fmt.Sprintf("%d_%s", n, header.Filename)
		filenamePath = filepath.Join(s.contentPathUpload, filename)
		n += 1
	}

	out, err := os.Create(filenamePath)
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
	log.Printf("%s %d Bytes saved\n", filename, header.Size)
	comandJSON, _ := json.Marshal(
		fileInfo{
			N:    len(s.files) - 1,
			Path: `/upload/` + filename,
		})
	s.Broadcast(comandJSON, "uploadws")
}
