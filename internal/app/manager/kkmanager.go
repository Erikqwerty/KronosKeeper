package manager

import (
	"fmt"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/cloudStorages"
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
	if conf.RemoteStorages != nil {
		if conf.RemoteStorages.GCloud.CredentialsJSON != "" {
			kkm.GCloud = gCloud.New(conf.RemoteStorages.GCloud.CredentialsJSON)
			if err := kkm.GCloud.NewClient(); err != nil {
				return kkm, err
			}
		}

		if conf.RemoteStorages.GDrive.ApiKeyJson != "" {
			kkm.GDrive, err = gDrive.New(conf.RemoteStorages.GDrive.ApiKeyJson, conf.RemoteStorages.GDrive.TokenFile)
			if err != nil {
				return kkm, err
			}
			if err := kkm.GDrive.NewClient(); err != nil {
				return kkm, err
			}
		}
	}

	return kkm, err
}

func (kkm *Kkmanager) ListDir(remote cloudStorages.Lister, path string) error {
	// получаем список файлов по пути path
	files, err := remote.ListDirItems(path)
	if err != nil {
		return fmt.Errorf("ошибка получения списка директорий для: %v, Error: %v", path, err)
	}
	// Выводим название директории
	fmt.Println(path + "/")
	for _, file := range files {
		if file.IsDir() {
			if err := kkm.ListDir(remote, file.PathGenerate()); err != nil {
				return fmt.Errorf("ошибка рекурсивного вывода структуры директории %v, Error: %v", file.PathGenerate(), err)
			}
		} else {
			fmt.Println("|--", file.Name)
		}
	}
	return nil
}

// func (kkm *Kkmanager) ListDir(remote cloudstorages.Lister, unitName string) {
// 	for _, unit := range kkm.Conf.BackupUnits {
// 		if unit.Name == unitName {
// 			files, err := remote.ListDirItems(unit.RemotePath)
// 			if err != nil {
// 				kkm.Logger.Error(err)
// 			}
// 			fmt.Printf("Список файлов в удаленной директории %s:\n", unit.RemotePath)
// 			for i, file := range files {
// 				if file.IsDir() {
// 					kkm.listDirRecursive(remote, file)
// 				}
// 				fmt.Printf("%v. %v - %v , id файла на диске: %v , тип файла: %v \n", i+1, file.Name, file.Size, file.Id, file.MimeType)
// 			}
// 		} else {
// 			kkm.Logger.Errorf("Такого юнита: '%v' не существует", unitName)
// 		}
// 	}
// }

// func (kkm *Kkmanager) listDirRecursive(remote cloudstorages.Lister, file cloudstorages.File) {
// 	path := file.PathGenerate()
// 	files, err := remote.ListDirItems(path)
// 	if err != nil {
// 		kkm.Logger.Error(err)
// 	}
// 	for i, f := range files {
// 		if file.IsDir() {
// 			kkm.listDirRecursive(remote, f)
// 		}
// 		fmt.Printf("%v. %v - %v , id файла на диске: %v , тип файла: %v \n", i+1, f.Name, f.Size, f.Id, f.MimeType)
// 	}
// }
