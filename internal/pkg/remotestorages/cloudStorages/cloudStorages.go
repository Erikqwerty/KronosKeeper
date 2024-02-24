package cloudStorages

type File struct {
	Id      string
	Name    string
	Size    int64
	Parents []string
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