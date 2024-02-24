package manager

import (
	"fmt"
	"path/filepath"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	cloudstorages "github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/cloudStorages"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/cloudStorages/gCloud"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/cloudStorages/gDrive"

	"github.com/sirupsen/logrus"
)

type Kkmanager struct {
	Conf *config.Config
	RemoteStorage
	Logger *logrus.Logger
}

type RemoteStorage struct {
	GCloud *gCloud.GCloud
	GDrive *gDrive.GDrive
}

func New(conf *config.Config) (*Kkmanager, error) {
	var err error
	kkm := &Kkmanager{
		Conf:   conf,
		Logger: logrus.New(),
	}
	kkm.GCloud = gCloud.New(conf.RemoteStorages.GCloud.CredentialsJSON)
	kkm.GDrive, err = gDrive.New(conf.RemoteStorages.GDrive.ApiKeyJson, conf.RemoteStorages.GDrive.TokenFile)
	if err != nil {
		kkm.Logger.Errorf("Ошибка при инициализации kkmanager")
		return kkm, err
	}

	return kkm, err
}

func (kkm *Kkmanager) ListDir(remote cloudstorages.Lister, unitName string) {
	for _, unit := range kkm.Conf.BackupUnits {
		if unit.Name == unitName {
			files, err := remote.ListDirItems(unit.RemotePath)
			if err != nil {
				kkm.Logger.Error(err)
			}
			fmt.Printf("Список файлов в удаленной директории %s:\n", unit.RemotePath)
			for i, file := range files {
				var path string
				for _, parent := range file.Parents {
					path = filepath.Join(path, parent)
				}
				path = filepath.Join(path, file.Name)

				fmt.Printf("%v. %v - %v , id файла на диске: %v", i+1, path, file.Size, file.Id)
			}
		} else {
			kkm.Logger.Errorf("Такого юнита: '%v' не существует", unitName)
		}
	}
}
