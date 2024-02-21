// Пакет remotestorage реализует взаимодействие с различными удаленными хранилищами.
package remotestorage

import (
	"fmt"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorage/gCloud"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorage/nfs"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorage/samba"
)

// Remotestorage представляет собой структуру для работы с удаленными хранилищами.
type Remotestorage struct {
	PushConfig
	gCloud *gCloud.GCloud
	smb    *samba.Samba
	nfs    *nfs.NFS
}

// PushConfig содержит настройки для отправки резервных копий в удаленное хранилище.
type PushConfig struct {
	PushTO    []string // куда заливать резервные копии, например: []string{"gCloud", "nfs"}
	Upload    string   // Путь к загружаемой директории
	RemoteDir string   // Путь к папке, в которую загружаем
}

// PushReport представляет отчет о выполнении операции отправки резервных копий в удаленное хранилище.
type PushReport struct {
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
}

// New создает новый объект Remotestorage на основе конфигурации.
func New(pushConfig *PushConfig, remote *config.StorageConfig) *Remotestorage {
	return &Remotestorage{
		PushConfig: *pushConfig,
		gCloud:     gCloud.New(remote.GCloud.CredentialsJSON),
	}
}

// NewReport создает новый объект PushReport.
func NewReport() *PushReport {
	return &PushReport{
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
	}
}

// PushBackup выполняет операцию отправки резервных копий в удаленные хранилища.
func (r *Remotestorage) PushBackup() *PushReport {
	pushReport := NewReport()

	for _, TO := range r.PushTO {
		switch TO {
		case "gCloud":
			if err := r.gCloud.NewClient(); err != nil {
				pushReport.GCloud.Err = fmt.Errorf("ошибка создания клиента: %v", err)
				continue
			}
			if err := r.gCloud.UploadFile(r.Upload, r.RemoteDir); err != nil {
				pushReport.GCloud.Err = fmt.Errorf("ошибка загрузки файла: %v", err)
				continue
			}
			pushReport.GCloud.Status = true
		case "nfs":
			if err := r.nfs.Push(); err != nil {
				pushReport.NFS.Err = err
				continue
			}
			pushReport.NFS.Status = true
		case "samba":
			if err := r.smb.Push(); err != nil {
				pushReport.Samba.Err = err
				continue
			}
			pushReport.Samba.Status = true
		}
	}
	return pushReport
}
