// Package gdrive предоставляет функции для взаимодействия с Google Drive API,
// включая загрузку файлов, просмотр содержимого папок и скачивание файлов.
package gDrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/remotestorages/cloudStorages"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

// GDrive представляет клиент Google Drive.
type GDrive struct {
	config      *oauth2.Config // Конфигурация OAuth2 для аутентификации.
	client      *http.Client   // HTTP-клиент для взаимодействия с API.
	service     *drive.Service // Сервис Google Drive.
	oAuthServer *http.Server   // HTTP-сервер для аутентификации.
	currentUser *drive.User    // Текущий пользователь Google Drive.

	tokenFile string // Файл для сохранения токена доступа.
}

// New создает новый экземпляр GDrive с заданным файлом учетных данных, и токеном клиента. Если токена нету он будет создан по переданному пути после атентификации.
func New(credentialsFile string, tokenFile string) (*GDrive, error) {
	// Чтение данных учетной записи из файла.
	credentials, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл с учетными данными: %v", err)
	}

	// Инициализация конфигурации OAuth2 из JSON-файла учетных данных.
	config, err := google.ConfigFromJSON(credentials, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("ошибка конфигурации OAuth 2.0: %v", err)
	}

	server := &http.Server{
		Addr: ":80",
	}

	return &GDrive{
		config:      config,
		oAuthServer: server,
		tokenFile:   tokenFile,
	}, nil
}

// NewClient запускает процесс аутентификации и создает нового клиента для взаимодействия с Google Drive API.
func (gd *GDrive) NewClient() error {

	// Попытка загрузить сохраненный токен.
	token, err := gd.loadToken()
	if err == nil && token.Valid() {
		// Если токен загружен и действителен, используйте его для создания клиента.
		gd.client = gd.config.Client(context.Background(), token)
		gd.service, err = drive.NewService(context.Background(), option.WithHTTPClient(gd.client))
		if err != nil {
			return fmt.Errorf("ошибка создания клиента Google Drive API: %v", err)
		}

		// Получение информации о текущем пользователе.
		user, err := gd.service.About.Get().Fields("user").Do()
		if err != nil {
			return fmt.Errorf("ошибка получения информации о пользователе: %v", err)
		}
		gd.currentUser = user.User

		return nil
	}

	errChan := make(chan error)
	// Запуск сервера аутентификации OAuth.
	go func() {
		err = gd.startOAuthServer()
		if err != nil {
			errChan <- fmt.Errorf("ошибка запуска сервера аутентификации: %v", err)
			return
		}
		errChan <- nil
	}()
	err = <-errChan
	if err != nil {
		return err
	}
	// Создание клиента Google Drive API.
	gd.service, err = drive.NewService(context.Background(), option.WithHTTPClient(gd.client))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента Google Drive API: %v", err)
	}

	// Получение информации о текущем пользователе.
	user, err := gd.service.About.Get().Fields("user").Do()
	if err != nil {
		return fmt.Errorf("ошибка получения информации о пользователе: %v", err)
	}
	gd.currentUser = user.User

	return nil
}

// startOAuthServer запускает HTTP-сервер для обработки процесса аутентификации OAuth.
func (gd *GDrive) startOAuthServer() error {
	// Вывод URL для аутентификации.
	gd.printAuthURL()

	// Обработчик для обработки обратного вызова аутентификации.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")

		if state != "state-token" {
			http.Error(w, "Неверный параметр состояния", http.StatusBadRequest)
			return
		}

		// Обмен кода авторизации на токен доступа.
		tok, err := gd.config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Не удалось обменять код на токен", http.StatusInternalServerError)
			return
		}
		gd.saveToken(tok)

		// Создание HTTP-клиента с учетом полученного токена.
		gd.client = gd.config.Client(context.Background(), tok)

		// Остановка HTTP-сервера после успешной аутентификации.
		if gd.client != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := gd.oAuthServer.Shutdown(ctx); err != nil {
					fmt.Printf("Ошибка остановки сервера: %v\n", err)
				}
			}()
		}
	})

	// Запуск HTTP-сервера аутентификации.
	if err := gd.oAuthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ошибка запуска сервера аутентификации Google Drive: %v", err)
	}

	return nil
}

// printAuthURL выводит URL для аутентификации.
func (gd *GDrive) printAuthURL() {
	url := gd.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Пожалуйста, перейдите по следующему URL и разрешите доступ:")
	fmt.Println(url)
}

// UploadFile загружает файл в Google Drive.
func (gd *GDrive) UploadFile(localPath string, remotePath string) error {
	// Открытие файла для чтения.
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	// Получение или создание папки для загрузки файла.
	folder, err := gd.getOrCreateFolder(remotePath)
	if err != nil {
		return fmt.Errorf("ошибка получения или создания папки: %v", err)
	}

	// Создание объекта файла для загрузки.
	f := &drive.File{
		Title:   filepath.Base(localPath),
		Parents: []*drive.ParentReference{{Id: folder.Id}},
	}

	// Загрузка файла в Google Drive.
	_, err = gd.service.Files.Insert(f).Media(file).Do()
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла: %v", err)
	}

	return nil
}

// DownloadFile скачивает файл с Google Drive.
func (gd *GDrive) DownloadFile(fileID string, localPath string) error {
	// Создание файла для записи.
	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer outFile.Close()

	// Загрузка содержимого файла с Google Drive.
	resp, err := gd.service.Files.Get(fileID).Download()
	if err != nil {
		return fmt.Errorf("ошибка скачивания файла: %v", err)
	}
	defer resp.Body.Close()

	// Копирование содержимого файла в локальный файл.
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка записи файла: %v", err)
	}

	fmt.Printf("Файл успешно скачен и сохранен по пути: %s\n", localPath)
	return nil
}

// ListDirItems выводит список файлов в указанной папке.
func (gd *GDrive) ListDirItems(remotePath string) ([]cloudStorages.File, error) {
	// Если путь пустой, устанавливаем его как корневая папка.
	if remotePath == "" {
		remotePath = "root"
	}

	// Получение ID папки по ее пути.
	folder, err := gd.getFolderByPath(remotePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения папки: %v", err)
	}

	// Формирование запроса к API для получения списка файлов в указанной папке.
	query := fmt.Sprintf("'%s' in parents", folder.Id)
	files, err := gd.service.Files.List().Q(query).Do()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка файлов: %v", err)
	}

	// Вывод списка файлов.
	var items []cloudStorages.File
	for _, file := range files.Items {
		// Преобразование списка родительских папок в путь.
		var parents []string
		for _, parent := range file.Parents {
			// Получаем информацию о каждой родительской папке и добавляем ее имя в список.
			parentInfo, err := gd.service.Files.Get(parent.Id).Fields("title").Do()
			if err != nil {
				return nil, fmt.Errorf("ошибка при получении информации о родительской папке: %v", err)
			}
			parents = append(parents, parentInfo.Title)
		}
		// Создание объекта файла с путем к родительским папкам.
		item := &cloudStorages.File{
			Id:       file.Id,
			Name:     file.Title,
			Size:     file.FileSize,
			Parents:  parents,
			MimeType: file.MimeType,
		}
		items = append(items, *item)
	}
	return items, err
}

// getOrCreateFolder получает или создает папку по указанному пути.
func (gd *GDrive) getOrCreateFolder(remotePath string) (*drive.File, error) {
	// Разбиение пути на компоненты.
	folders := strings.Split(remotePath, "/")

	// Поиск каждой папки в пути.
	parent := "root"
	for _, folder := range folders {
		if folder == "" {
			continue
		}

		// Поиск папки среди дочерних элементов текущей папки.
		query := fmt.Sprintf("title='%s' and trashed=false and mimeType='application/vnd.google-apps.folder' and '%s' in parents", folder, parent)
		folderList, err := gd.service.Files.List().Q(query).Do()
		if err != nil {
			return nil, fmt.Errorf("ошибка при поиске папки %s: %v", folder, err)
		}

		// Создание новой папки, если она не найдена.
		if len(folderList.Items) == 0 {
			newFolder := &drive.File{
				Title:    folder,
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []*drive.ParentReference{{Id: parent}},
			}
			createdFolder, err := gd.service.Files.Insert(newFolder).Do()
			if err != nil {
				return nil, fmt.Errorf("ошибка при создании папки %s: %v", folder, err)
			}
			parent = createdFolder.Id
		} else {
			// Использование существующей папки.
			parent = folderList.Items[0].Id
		}
	}

	// Возвращение объекта последней папки в пути.
	return &drive.File{Id: parent}, nil
}

// getFolderByPath получает папку по указанному пути.
func (gd *GDrive) getFolderByPath(remotePath string) (*drive.File, error) {
	// Разбиение пути на компоненты.
	folders := strings.Split(remotePath, "/")

	// Начало с корневой папки.
	parent := "root"
	for _, folder := range folders {
		if folder == "" {
			continue
		}

		// Поиск папки среди дочерних элементов текущей папки.
		query := fmt.Sprintf("title='%s' and trashed=false and mimeType='application/vnd.google-apps.folder' and '%s' in parents", folder, parent)
		folderList, err := gd.service.Files.List().Q(query).Do()
		if err != nil {
			return nil, fmt.Errorf("ошибка при поиске папки %s: %v", folder, err)
		}

		// Возврат ошибки, если папка не найдена.
		if len(folderList.Items) == 0 {
			return nil, fmt.Errorf("папка %s не найдена", folder)
		}

		// Использование ID найденной папки в следующем поиске.
		parent = folderList.Items[0].Id
	}

	// Возвращение объекта последней папки в пути.
	return &drive.File{Id: parent}, nil
}

func (gd *GDrive) saveToken(token *oauth2.Token) error {
	tokenData, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации токена: %v", err)
	}

	if err := os.WriteFile(gd.tokenFile, tokenData, 0644); err != nil {
		return fmt.Errorf("ошибка при сохранении токена в файл: %v", err)
	}

	return nil
}

// loadToken загружает сохраненный токен доступа из файла.
func (gd *GDrive) loadToken() (*oauth2.Token, error) {
	tokenData, err := os.ReadFile(gd.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении токена из файла: %v", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(tokenData, &token); err != nil {
		return nil, fmt.Errorf("ошибка при десериализации токена: %v", err)
	}

	return &token, nil
}
