// Пакет remotestorages реализует взаимодействие с различными удаленными хранилищами.
package remotestorages

import (
	"fmt"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/gCloud"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/nfs"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/samba"
	gdrive "github.com/Erikqwerty/KronosKeeper/pkg/gDrive"
	"google.golang.org/api/drive/v2"
)

// Remotestorages представляет собой структуру для работы с удаленными хранилищами.
type Remotestorages struct {
	UploadConfig
	gCloud *gCloud.GCloud
	gDrive *gdrive.GDrive
	smb    *samba.Samba
	nfs    *nfs.NFS
}

type Remoter interface {
	UploadFile(string, string) error
	ListDirItems(string) ([]*drive.File, error)
}

// UploadConfig содержит настройки для отправки резервных копий в удаленное хранилище.
type UploadConfig struct {
	UploadTO   []string // куда заливать резервные копии, например: []string{"gCloud", "nfs"}
	LocalPath  string   // Путь к загружаемой директории
	RemotePath string   // Путь к папке, в которую загружаем
}

// UploadReports представляет отчет о выполнении операции отправки резервных копий в удаленное хранилище.
type UploadReports struct {
	GCloud struct {
		Status bool  // выполнелся ли push на удаленное хранилище
		Err    error // есть ли ошибка в отправке резервных копий
	}
	NFS struct {
		Status bool  // выполнелся ли push на удаленное хранилище
		Err    error // есть ли ошибка в отправке резервных копий
	}
	Samba struct {
		Status bool  // выполнелся ли push на удаленное хранилище
		Err    error // есть ли ошибка в отправке резервных копий
	}
	GDrive struct {
		Status bool  // выполнелся ли push на удаленное хранилище
		Err    error // есть ли ошибка в отправке резервных копий
	}
}

// NewReports создает новый объект UploadReports.
func NewReports() *UploadReports {
	return &UploadReports{
		GCloud: struct {
			Status bool
			Err    error
		}{false, nil},
		NFS: struct {
			Status bool
			Err    error
		}{false, nil},
		Samba: struct {
			Status bool
			Err    error
		}{false, nil},
		GDrive: struct {
			Status bool
			Err    error
		}{false, nil},
	}
}

// New создает новый объект Remotestorages на основе конфигурации.
func New(UploadConfig *UploadConfig, remote *config.RemoteStorages) (*Remotestorages, error) {
	remotestorages := &Remotestorages{
		UploadConfig: *UploadConfig,
		gCloud:       gCloud.New(remote.GCloud.CredentialsJSON),
	}

	if remote.GDrive.ApiKeyJson != "" {
		gDrive, err := gdrive.New(remote.GDrive.ApiKeyJson, remote.GDrive.TokenFile)
		if err != nil {
			return remotestorages, err
		}
		remotestorages.gDrive = gDrive
	}

	return remotestorages, nil
}

// UploadBackups выполняет операцию отправки резервных копий в удаленные хранилища.
func (r *Remotestorages) UploadBackups() *UploadReports {
	UploadReports := NewReports()

	for _, TO := range r.UploadTO {
		switch TO {
		case "gCloud":
			if err := r.gCloud.NewClient(); err != nil {
				UploadReports.GCloud.Err = fmt.Errorf("UploadBackups: %v", err)
				continue
			}
			if err := r.gCloud.UploadFile(r.LocalPath, r.RemotePath); err != nil {
				UploadReports.GCloud.Err = fmt.Errorf("UploadBackups: %v", err)
				continue
			}
			UploadReports.GCloud.Status = true

		case "gDrive":
			if err := r.gDrive.NewClient(); err != nil {
				UploadReports.GDrive.Err = fmt.Errorf("UploadBackups: %v", err)
			}
			if err := r.gDrive.UploadFile(r.LocalPath, r.RemotePath); err != nil {
				UploadReports.GDrive.Err = fmt.Errorf("UploadBackups: %v", err)
			}
			UploadReports.GDrive.Status = true

		}
	}

	return UploadReports
}
