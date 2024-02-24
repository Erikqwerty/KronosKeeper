package manager

import (
	"fmt"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/gCloud"
	gdrive "github.com/Erikqwerty/KronosKeeper/pkg/gDrive"
	"github.com/sirupsen/logrus"
)

type Kkmanager struct {
	conf *config.Config
	RemoteStorage
	Logger *logrus.Logger
}

type RemoteStorage struct {
	GCloud *gCloud.GCloud
	GDrive *gdrive.GDrive
}

func New(conf *config.Config) (*Kkmanager, error) {
	var err error
	kkm := &Kkmanager{
		conf:   conf,
		Logger: logrus.New(),
	}
	kkm.GCloud = gCloud.New(conf.RemoteStorages.GCloud.CredentialsJSON)
	kkm.GDrive, err = gdrive.New(conf.RemoteStorages.GDrive.ApiKeyJson, conf.RemoteStorages.GDrive.TokenFile)
	if err != nil {
		kkm.Logger.Errorf("Ошибка при инициализации kkmanager")
		return kkm, err
	}

	return kkm, err
}

func (kkm *Kkmanager) ListDir(remote remotestorages.Remoter, unitName string) {
	for _, unit := range kkm.conf.BackupUnits {
		if unit.Name == unitName {
			files, err := remote.ListDirItems(unit.RemotePath)
			if err != nil {
				kkm.Logger.Errorf("ошибка получения списка файлов в удаленном хранилеще %v", err)
			}
			for i, file := range files {
				fmt.Printf("%v. %v, %v, %v", i+1, file.Title, file.CreatedDate, file.FileSize)
			}
		}
	}
}
