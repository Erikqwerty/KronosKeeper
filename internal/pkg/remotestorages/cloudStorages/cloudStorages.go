package cloudStorages

import (
	"fmt"
	"path/filepath"
)

type File struct {
	Id       string
	Name     string
	Size     int64
	Parents  []string
	MimeType string
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
	fmt.Println(f.Parents)
	var path string
	fmt.Println("предки", f.Parents)
	for _, item := range f.Parents {
		path = filepath.Join(path, item)
	}
	return filepath.Join(path, f.Name)
}
