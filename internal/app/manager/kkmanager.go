package manager

import (
	"fmt"
	"path/filepath"
	"regexp"

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

func (kkm *Kkmanager) listDir(remote cloudStorages.Lister, path string) error {

	// получаем список файлов по пути path
	files, err := remote.ListDirItems(path)
	if err != nil {
		return fmt.Errorf("ошибка получения списка директорий для: %v, Error: %v", path, err)
	}
	if isValidDate(filepath.Base(path)) {
		fmt.Println(path) //выводим путь до папки с бекапами
	}

	return func() error {
		for id, file := range files {
			// обнуляем путь до бекапа юнита если мы попали в папку формата unitname/2024-02 чтобы избежать формирования ошибочных путей дальше
			if isValidDate(filepath.Base(path)) {
				path = filepath.Dir(path)
			}

			if file.IsDir() {
				path = filepath.Join(path, file.Name)
				if err := kkm.listDir(remote, path); err != nil {
					return fmt.Errorf("ошибка рекурсивного вывода структуры директории %v, Error: %v", file.PathGenerate(), err)
				}
			} else {
				fmt.Printf("|__ %v. %v |  %v | Размер: %v \n", id+1, file.Id, file.Name, file.SizeSuffix())
			}
		}
		return nil
	}()
}

func isValidDate(date string) bool {
	// Определяем регулярное выражение для проверки формата YYYY-MM
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}$`)
	return dateRegex.MatchString(date)
}

func (kkm *Kkmanager) ListBackupsUnit(unitName string) error {
	for _, unit := range kkm.Conf.BackupUnits {
		if unit.Name == unitName {
			for _, remote := range unit.UploadTo {
				fmt.Print("_________________________________________________________________________________")
				switch remote {
				case "gDrive":
					fmt.Printf("\n				Список бекапов на %v:		\n", remote)
					fmt.Println("‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾")
					if err := kkm.listDir(kkm.GDrive, unit.RemotePath); err != nil {
						return err
					}
					continue
				case "gCloud":
					fmt.Printf("\n				Список бекапов на %v:		\n", remote)
					fmt.Println("‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾")
					if err := kkm.listDir(kkm.GCloud, unit.RemotePath); err != nil {
						return err
					}
					continue
				default:
					kkm.Logger.Infof("У данного юнита: %v нету параметров удаленного копирования", unitName)
				}
			}
		} else {
			kkm.Logger.Warnf("Юнита с таким именим: %v не существует", unitName)
		}
	}
	return nil
}
