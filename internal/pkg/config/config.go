// Пакет config содержит структуры для конфигурации приложения KronosKeeper.
package config

import (
	"github.com/BurntSushi/toml"
)

// Telegram представляет настройки уведомлений для Telegram.
type Telegram struct {
	Token  string `toml:"token"`   // API ключ Telegram
	ChatID string `toml:"chat_id"` // id чата
}

// RemoteStorages содержит настройки удаленных хранилищ данных.
type RemoteStorages struct {
	GCloud struct {
		CredentialsJSON string `toml:"credentials_json"` // Путь к JSON-файлу с учетными данными для Google Cloud
	} `toml:"gCloud"`
	GDrive struct {
		ApiKeyJson string `toml:"apiKeyJson"` // Путь к ключу GDrive для атентификации по OAuth2.0
		TokenFile  string `toml:"tokenFile"`  // Путь до токен файл где будет сохранен токен после атентификации
	} `toml:"gDrive"`
	Samba struct {
		SambaPath string `toml:"samba"`    // Путь к сетевому ресурсу Samba
		Username  string `toml:"username"` // Имя пользователя для доступа к ресурсу
		Password  string `toml:"password"` // Пароль для доступа к ресурсу
	} `toml:"samba"`
	NFS struct {
		Path         string `toml:"path"`          // Путь к NFS-шаре
		MountOptions string `toml:"mount_options"` // Опции монтирования
	} `toml:"nfs"`
}

// BackupUnitConfig содержит настройки юнитов/задач бекапов.
type BackupUnit struct {
	Name            string   `toml:"name"`            // Имя юнита/задачи бэкапа
	Retention       int      `toml:"retention"`       // Время хранения бэкапов (в днях)
	CrontabTask     string   `toml:"crontabTask"`     // Расписание Cron
	InputPaths      []string `toml:"inputPaths"`      // Пути для бэкапа
	OutputPath      string   `toml:"output"`          // Путь для сохранения бэкапов
	CompressFormat  string   `toml:"compressFormat"`  // Формат сжатия бэкапов
	CompressExclude []string `toml:"compressExclude"` // Исключения из сжатия
	MaxDiskUsage    string   `toml:"maxDiskUsage"`    // Максимальное использование диска
	UploadTo        []string `toml:"uploadTo"`        // Список удаленных хранилищ
	RemotePath      string   `toml:"remotePath"`      // Папка на удаленном хранилище для сохранения бэкапов
}

// Config представляет конфигурацию программы KronosKeeper.
type Config struct {
	LogPath        string          `toml:"log_path"`  // Путь к файлу журнала
	LogLevel       string          `toml:"log_level"` // Уровень журналирования
	Telegram       *Telegram       `toml:"telegram"`  // Настройки уведомлений
	RemoteStorages *RemoteStorages `toml:"storage"`   // Настройки удаленных хранилищ данных
	BackupUnits    []BackupUnit    `toml:"unit"`      // Настройки юнитов/задач бекапов
}

// NewConfig создает новый экземпляр конфигурации KronosKeeper.
func NewConfig(configPath string) (*Config, error) {
	conf := &Config{}
	_, err := toml.DecodeFile(configPath, conf) // загружаем конфигурацию из файла toml в config
	if err != nil {
		return nil, err
	}

	return conf, nil
}
