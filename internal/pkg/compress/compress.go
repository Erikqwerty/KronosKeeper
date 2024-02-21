// Пакет compress реализует функционал для создания архивов форматов ZIP, ...
package compress

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Compress представляет параметры для создания архива.
type Compress struct {
	ArchiveName string   // Имя для создаваемого архива
	InputPaths  []string // Элементы, которые необходимо включить в архив
	ExludeFile  []string // Исключения из архивации "*.log", "array.zip"
	OutputPath  string   // Путь сохранения архива
}

// Содержит отчет о результатах сжатия
type CompressReport struct {
	YearMoth    string // Год месяц создания - соответствует имени содержащей архив папке
	ArchiveName string // Имя архива - формат имени 23-10:34-unit.zip
	ArchivePath string // Полный путь до архива
}

// New создает новый экземпляр Compress.
func New() *Compress {
	return &Compress{}
}

// Start запускает процесс создания архива с указанным форматом.
func (c *Compress) Start(format string) (*CompressReport, error) {
	if c.ArchiveName == "" || len(c.InputPaths) == 0 || c.OutputPath == "" {
		return nil, fmt.Errorf("недостаточно параметров для запуска архивации")
	}
	// Создаем обьект времени для указания даты бекапов
	currentTime := time.Now()

	// Проверяем есть ли в папке назначения папка с именим юнита бекапа, если нет то создаем
	if err := c.ensureDirInOutputPath(c.ArchiveName); err != nil {
		return nil, err
	}
	c.OutputPath = filepath.Join(c.OutputPath, c.ArchiveName) // обнавляем путь до папки где нужно создать папку год месяц

	// Проверяем есть ли в папке с именим юнита папка с годом и месяцем, если нет то создаем
	dateYearMoth := currentTime.Format("2006-01")
	if err := c.ensureDirInOutputPath(dateYearMoth); err != nil {
		return nil, err
	}
	c.OutputPath = filepath.Join(c.OutputPath, currentTime.Format("2006-01")) // обнавляем путь до папки куда нужно положить бекап

	// Добавляем к имени архива день месяца и время
	dateDayTime := currentTime.Format("02-15:04")
	c.ArchiveName = dateDayTime + "-" + c.ArchiveName

	switch format {
	case "zip":
		if err := c.Zip(); err != nil {
			return nil, err
		}
		c.ArchiveName += ".zip"
	default:
		return nil, fmt.Errorf("формат сжатия ' %v ' не поддерживается", format)
	}

	return &CompressReport{
		YearMoth:    dateYearMoth,  // дата создания архива день год месяц
		ArchiveName: c.ArchiveName, // имя созданного архива
		ArchivePath: c.OutputPath,  // путь до архива
	}, nil
}

// ensureDirInOutputPath проверяет существует ли директория в с.OutputPath и если нет то создает ее
func (c *Compress) ensureDirInOutputPath(dir string) error {
	Foleder := filepath.Join(c.OutputPath, dir)
	if _, err := os.Stat(Foleder); os.IsNotExist(err) {
		// Папка не существует, создаем ее
		err := os.Mkdir(Foleder, 0755) // 0755 - права доступа к папке (rwxr-xr-x)
		if err != nil {
			// Если произошла ошибка при создании папки, возвращаем эту ошибку
			return err
		}
		return nil
	}
	return nil
}

// zip выполняет сжатие в формате ZIP по параметрам, указанным в структуре Compress.
func (c *Compress) Zip() error {
	if c.ArchiveName == "" || len(c.InputPaths) == 0 || c.OutputPath == "" {
		return fmt.Errorf("недостаточно параметров для запуска архивации")
	}
	// Формируем путь и имя архива
	archive := fmt.Sprintf(c.ArchiveName + ".zip")
	archivePath := filepath.Join(c.OutputPath, archive)

	// Создаем файл для записи архива
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создаем новый ZIP-архив
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Проходимся по директориям, которые нужно добавить в архив
	for _, dirToBackup := range c.InputPaths {
		err := filepath.Walk(dirToBackup, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Добавляем файлы в архив
			return c.addToZip(path, info, zipWriter, dirToBackup)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// addToZip добавляет файлы и директории в ZIP-архив.
func (c *Compress) addToZip(path string, info fs.FileInfo, zw *zip.Writer, dirToBackup string) error {
	// Проверяем, исключен ли файл из архивации
	exclude, err := c.isExcluded(info.Name())
	if err != nil {
		return err
	}
	if exclude {
		return nil
	}

	// Исключаем директорию, если она совпадает с директорией, в которой создаем архив
	if info.Name() == filepath.Base(c.OutputPath) {
		return nil
	}

	// Получаем метаданные файла для архива
	fileHeader, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("ошибка при получении метаданных файла: %v", err)
	}

	// Получаем относительный путь к файлу
	fileHeader.Name, err = filepath.Rel(dirToBackup, path)
	if err != nil {
		return fmt.Errorf("ошибка при получении относительного пути файла: %v", err)
	}

	// Убираем регистр из пути
	fileHeader.Name = strings.ToLower(fileHeader.Name)

	// Если это директория, добавляем "/" к имени файла
	if info.IsDir() {
		fileHeader.Name += "/"
	}

	// Создаем запись в ZIP-архиве
	ArchiveWriter, err := zw.CreateHeader(fileHeader)
	if err != nil {
		return fmt.Errorf("ошибка при создании записи в архиве: %v", err)
	}

	// Если это не директория, копируем содержимое файла в архив
	if !info.IsDir() {
		return c.writeToArchive(ArchiveWriter, path)
	}

	return nil
}

// isExcluded проверяет, исключен ли файл из архивации по его имени.
func (c *Compress) isExcluded(name string) (bool, error) {
	// Проверяем каждый паттерн исключения
	for _, pattern := range c.ExludeFile {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false, fmt.Errorf("ошибка при обработки исключения %v", err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// writeToArchive заполняет архив содержимым файла, указанного в path.
func (c *Compress) writeToArchive(ArchiveWriter io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("ошибка при получении информации о файле: %v", err)
	}

	// Пропускаем пустые файлы
	if fileInfo.Size() == 0 {
		return nil
	}

	// Копируем содержимое файла в архив
	_, err = io.Copy(ArchiveWriter, file)
	if err != nil {
		return fmt.Errorf("ошибка при копировании файла в архив: %v", err)
	}

	return nil
}
