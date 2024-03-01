package cloudStorages

import (
	"path/filepath"
)

type File struct {
	Id       string   // ID файла на диске
	Name     string   // Имя файла
	Size     int64    // Размер файла
	Parents  []string // Папка вышестоящего уровня, а надо бы путь к файлу.
	MimeType string   // Тип файла
}

type Cloud interface {
	Lister
	Uploader
	Downloader
}

type Lister interface {
	ListDirItems(string) ([]File, error)
}

type Uploader interface {
	UploadFile(localPath string, remotePath string) error
}

type Downloader interface {
	DownloadFile(fileID string, localPath string) error
}

func (f *File) IsDir() bool {
	return f.MimeType == "application/vnd.google-apps.folder"
}

func (f *File) PathGenerate() string {
	var path string
	for _, item := range f.Parents {
		path = filepath.Join(path, item)
	}
	return filepath.Join(path, f.Name)
}
