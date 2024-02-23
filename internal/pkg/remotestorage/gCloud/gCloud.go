package gCloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GCloud представляет объект для работы с Google Cloud.
type GCloud struct {
	CredentialsJSON string          // Путь к файлу Credentials.json, необходимому для аутентификации Google Cloud API.
	ctx             context.Context // Контекст для выполнения операций API.
	client          *drive.Service  // Клиент Google Cloud API.
}

// New создает новый экземпляр GCloud с указанием пути к файлу JSON с учетными данными.
func New(credentialsJSON string) *GCloud {
	return &GCloud{
		CredentialsJSON: credentialsJSON,
		ctx:             context.Background(),
	}
}

// NewClient создает новый клиент Google Cloud API с использованием учетных данных из файла JSON.
func (gc *GCloud) NewClient() error {
	client, err := drive.NewService(gc.ctx, option.WithCredentialsFile(gc.CredentialsJSON))
	if err != nil {
		return fmt.Errorf("не удалось создать клиент Google Cloud API: %v", err)
	}
	gc.client = client
	return nil
}

// UploadFile загружает файл на Google Cloud в указанную удаленную директорию.
func (gc *GCloud) UploadFile(localFilePath, remoteDir string) error {
	// Открываем локальный файл для чтения.
	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл для чтения: %v", err)
	}
	defer file.Close()

	// Создаем директории на Google Cloud, если они не существуют.
	if err := gc.ensureDirectoriesExist(remoteDir); err != nil {
		return err
	}

	// Получаем идентификатор директории на Google Cloud, куда будем загружать файл.
	remoteDirID, err := gc.getFolderIDByPath(remoteDir)
	if err != nil {
		return err
	}

	// Создаем метаданные для файла.
	fileMetadata := &drive.File{
		Name:    filepath.Base(localFilePath),
		Parents: []string{remoteDirID},
	}

	// Загружаем файл на Google Cloud.
	_, err = gc.client.Files.Create(fileMetadata).Media(file).Do()
	if err != nil {
		return fmt.Errorf("не удалось загрузить файл: %v", err)
	}

	return nil
}

// ensureDirectoriesExist создает необходимые директории на Google Cloud для заданного пути.
func (gc *GCloud) ensureDirectoriesExist(remoteDir string) error {
	// Разделяем путь на отдельные компоненты.
	folderNames := strings.Split(remoteDir, "/")

	// Начинаем с корневой папки Google Cloud.
	parentID := "root"

	// Перебираем каждый компонент пути.
	for _, folderName := range folderNames {
		// Проверяем, существует ли папка с текущим именем в текущей родительской папке.
		folderID, err := gc.getFolderIDByNameAndParentID(folderName, parentID)
		if err != nil {
			return fmt.Errorf("не удалось получить идентификатор папки %s: %v", folderName, err)
		}

		// Если папка не существует, создаем ее.
		if folderID == "" {
			newFolder, err := gc.createFolder(folderName, parentID)
			if err != nil {
				return fmt.Errorf("не удалось создать папку %s: %v", folderName, err)
			}
			parentID = newFolder.Id
		} else {
			// Если папка существует, используем ее в качестве родительской для следующей итерации.
			parentID = folderID
		}
	}

	return nil
}

// getFolderIDByNameAndParentID возвращает идентификатор папки по ее имени и родительскому идентификатору.
func (gc *GCloud) getFolderIDByNameAndParentID(folderName, parentID string) (string, error) {
	query := fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and name='%s' and '%s' in parents", folderName, parentID)
	fileList, err := gc.client.Files.List().Q(query).Fields("files(id)").Do()
	if err != nil {
		return "", err
	}
	if len(fileList.Files) > 0 {
		return fileList.Files[0].Id, nil
	}
	return "", nil
}

// createFolder создает новую папку на Google Cloud с заданным именем и родительским идентификатором.
func (gc *GCloud) createFolder(folderName, parentID string) (*drive.File, error) {
	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	return gc.client.Files.Create(folder).Do()
}

// getFolderIDByPath возвращает идентификатор папки на Google Cloud по ее пути.
func (gc *GCloud) getFolderIDByPath(folderPath string) (string, error) {
	// Разделяем путь на отдельные компоненты.
	folderNames := strings.Split(folderPath, "/")

	// Начинаем с корневой папки Google Cloud.
	parentID := "root"

	// Перебираем каждый компонент пути.
	for _, folderName := range folderNames {
		// Выполняем запрос к Google Cloud API для получения списка папок в текущей родительской папке.
		fileList, err := gc.client.Files.List().Q(fmt.Sprintf("'%s' in parents and mimeType='application/vnd.google-apps.folder' and name='%s'", parentID, folderName)).Do()
		if err != nil {
			return "", fmt.Errorf("не удалось получить список папок: %v", err)
		}

		// Проверяем, найдена ли папка с заданным именем в текущей родительской папке.
		if len(fileList.Files) == 0 {
			return "", fmt.Errorf("папка '%s' не найдена в родительской папке с идентификатором '%s'", folderName, parentID)
		}

		// Получаем идентификатор найденной папки.
		parentID = fileList.Files[0].Id
	}

	// Возвращаем идентификатор последней найденной папки.
	return parentID, nil
}

// ListFilesAndFolders возвращает список файлов и папок в указанной папке на Google Cloud.
func (gc *GCloud) ListFilesAndFolders(folderPath string) ([]*drive.File, error) {
	// Получаем идентификатор папки по ее пути.
	folderID, err := gc.getFolderIDByPath(folderPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить идентификатор папки по пути %s: %v", folderPath, err)
	}

	// Выполняем запрос к Google Cloud API для получения списка файлов и папок в указанной папке.
	fileList, err := gc.client.Files.List().Q(fmt.Sprintf("'%s' in parents", folderID)).Do()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список файлов и папок: %v", err)
	}

	// Возвращаем список файлов и папок.
	return fileList.Files, nil
}
