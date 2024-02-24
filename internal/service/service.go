// Пакет реализует создание резервной копии и отправку их в удаленное хранилище
package service

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/compress"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages"
)

// Backup представляет собой структуру, объединяющую компрессию и удаленное хранилище
type Backup struct {
	*compress.Compress
	*remotestorages.Remotestorages
}

// BackupReport содержит отчет о создании резервной копии
type BackupReport struct {
	Local       *compress.CompressReport      // отчет о локальное резервной копии
	Remote      *remotestorages.UploadReports // отчет о загрузки в удаленные хранилеща
	CurrentTime string
}

// NewBackup создает новый объект сервиса резервного копирования
func NewBackup() *Backup {
	return &Backup{}
}

// CreateBackup создает резервную копию согласно конфигурации unit и загружает ее в удаленное хранилище, если remote не равно nil
func (b *Backup) CreateBackup(unit config.BackupUnit, remote *config.RemoteStorages) (*BackupReport, error) {
	backupReport := &BackupReport{
		Local:       nil,
		Remote:      nil,
		CurrentTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	c := &compress.Compress{
		ArchiveName: unit.Name,
		InputPaths:  unit.InputPaths,
		OutputPath:  unit.OutputPath,
		ExludeFile:  unit.CompressExclude,
	}

	cReport, err := c.Start(unit.CompressFormat)
	if err != nil {
		return nil, fmt.Errorf("CreateBackup, ошибка при создание архива, текст: %v", err)
	}
	backupReport.Local = cReport

	if remote != nil {
		b.Remotestorages, err = remotestorages.New(&remotestorages.UploadConfig{
			UploadTO:   unit.UploadTo,                                               // Передаем в какие удаленные хранилеща делать push
			LocalPath:  filepath.Join(cReport.ArchivePath, cReport.ArchiveName),     // указываем путь к архиву и имени архива
			RemotePath: filepath.Join(unit.RemotePath, unit.Name, cReport.YearMoth), // Передаем путь на удаленном хранилище какой должен быть
		}, remote)
		if err != nil {
			return backupReport, err
		}
		backupReport.Remote = b.UploadBackups()
	}

	return backupReport, nil
}
