// Пакет реализует создание резервной копии и отправку их в удаленное хранилище
package service

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/compress"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorage"
)

// Backup представляет собой структуру, объединяющую компрессию и удаленное хранилище
type Backup struct {
	*compress.Compress
	*remotestorage.Remotestorage
}

// BackupReport содержит отчет о создании резервной копии
type BackupReport struct {
	Local       *compress.CompressReport
	Remote      *remotestorage.PushReport
	CurrentTime string
}

// NewBackup создает новый объект сервиса резервного копирования
func NewBackup() *Backup {
	return &Backup{}
}

// CreateBackup создает резервную копию согласно конфигурации unit и загружает ее в удаленное хранилище, если remote не равно nil
func (b *Backup) CreateBackup(unit config.BackupUnitConfig, remote *config.StorageConfig) (*BackupReport, error) {
	c := &compress.Compress{
		ArchiveName: unit.Name,
		InputPaths:  unit.InputPaths,
		OutputPath:  unit.OutputPath,
		ExludeFile:  unit.CompressExclude,
	}
	var backupReport *BackupReport
	backupReport.CurrentTime = time.Now().Format("2006-01-02 15:04:05")

	cReport, err := c.Start(unit.CompressFormat)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создание архива: текст - %v", err)
	}
	backupReport.Local = cReport

	if remote != nil {
		b.Remotestorage, err = remotestorage.New(&remotestorage.PushConfig{
			PushTO:    unit.RemoteStorages,                                        // Передаем в какие удаленные хранилеща делать push
			Upload:    filepath.Join(cReport.ArchivePath, cReport.ArchiveName),    // указываем путь к архиву и имени архива
			RemoteDir: filepath.Join(unit.RemoteDir, unit.Name, cReport.YearMoth), // Передаем путь на удаленном хранилище какой должен быть
		}, remote)
		if err != nil {
			return backupReport, err
		}
		backupReport.Remote = b.PushBackup()
	}

	return backupReport, nil
}
