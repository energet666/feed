package wsserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func (s *wsServer) HandleUploadxml(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("ошибка получения файла из формы: %s\n", err)
		return
	}
	defer file.Close()

	filename := header.Filename
	filenamePath := filepath.Join(s.contentPathUpload, filename)

	//Переименовываем файл до тех пор пока имя не станет уникальным в папке с контентом
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
		log.Printf("unable to create the file for writing: %s\n", err)
		return
	}
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		log.Printf("ошибка копирования файла из формы в файловую систему: %s\n", err)
		return
	}
	outInfo, _ := out.Stat()
	s.files = append(s.files, fs.FileInfoToDirEntry(outInfo))
	log.Printf("%s - %d Bytes saved\n", filename, header.Size)
	comandJSON, _ := json.Marshal(
		fileInfo{
			N:    len(s.files) - 1,
			Path: path.Join(`/upload`, filename),
		})
	s.Broadcast(comandJSON, "uploadws")
}
