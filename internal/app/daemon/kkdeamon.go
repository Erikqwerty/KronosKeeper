// Пакет daemon предоставляет функциональность для запуска и управления демоном KronosKeeper.
package daemon

import (
	"fmt"
	"os"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/TaskRunner"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/notifications/telegram"
	"github.com/Erikqwerty/KronosKeeper/internal/service"
	"github.com/sirupsen/logrus"
)

// KronosKeeperDeamon представляет собой демона KronosKeeper.
type KronosKeeperDeamon struct {
	config *config.Config
	Logger *logrus.Logger
	*TaskRunner.TaskRunner
	*service.Backup
	tg *telegram.TelegramBot
}

// New создает новый экземпляр KronosKeeperDeamon с заданной конфигурацией.
func New(conf *config.Config) *KronosKeeperDeamon {
	kkd := &KronosKeeperDeamon{
		config:     conf,
		Logger:     logrus.New(),
		Backup:     service.NewBackup(),
		TaskRunner: TaskRunner.NewTaskRunner(),
	}

	// Инициализация TelegramBot, если есть конфигурация
	if conf.Telegram != nil {
		var err error
		kkd.tg, err = telegram.NewTelegramBot(conf.Telegram)
		if err != nil {
			kkd.Logger.Warning("Ошибка при запуске телеграм бота", err)
		}
	}

	return kkd
}

// Start запускает демона KronosKeeperDeamon.
func (kkd *KronosKeeperDeamon) Start() error {
	if err := kkd.configureLogger(); err != nil {
		return err
	}

	kkd.Logger.Info("KronosKeeperDeamon Start!")

	// Запуск телеграм бота если в конфигурации имееться подобная настройка
	if kkd.config.Telegram != nil {
		go func() {
			if err := kkd.tg.Start(); err != nil {
				kkd.Logger.Warningf("ошибка отправки сообщения текст ошибки: %v", err)
			}
		}()
	}
	if len(kkd.config.BackupUnits) > 0 {
		if err := kkd.addTasksBackup(); err != nil {
			return err
		}
	} else {
		kkd.Logger.Info("Нету не одной задачи резервного копирования")
	}

	return nil
}

// Stop останавливает демона KronosKeeperDeamon.
func (kkd *KronosKeeperDeamon) Stop() error {
	kkd.TaskRunner.Stop()
	return nil
}

// addTasksBackup добавляет задачи резервного копирования в планировщик задач.
func (kkd *KronosKeeperDeamon) addTasksBackup() error {
	for _, unit := range kkd.config.BackupUnits {
		err := kkd.AddTask(unit.CrontabTask, func() {
			backupReport, err := kkd.CreateBackup(unit, kkd.config.Storage)
			if err != nil {
				kkd.writeLogAndNotifyError(fmt.Sprintf("Ошибка при запуске создания резервной копии для Unit %v: %v", unit.Name, err))
			}
			if err := kkd.handleBackupReport(backupReport); err != nil {
				kkd.writeLogAndNotifyError(fmt.Sprintf("Ошибка при обработке отчета о резервном копировании для Unit %v: %v", unit.Name, err))
			}
		})
		if err != nil {
			kkd.writeLogAndNotifyError(fmt.Sprintf("Не удалось запустить задачу резервного копирования по расписанию для Unit: %v", unit.Name))
			return err
		}
	}

	kkd.TaskRunner.Start() // Запускаем планировщик задач
	kkd.writeLogAndNotify("Запуск задач резервного копирования по расписанию запущен!")
	return nil
}

// configureLogger настраивает логгер на уровень логирования и место хранения логов.
func (kkd *KronosKeeperDeamon) configureLogger() error {
	level, err := logrus.ParseLevel(kkd.config.LogLevel)
	if err != nil {
		return fmt.Errorf("некорректный уровень логирования: %v", err)
	}

	logfile, err := os.OpenFile(kkd.config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		kkd.Logger.Info("Не удалось открыть файл лога. Логирование будет осуществляться только в стандартный вывод.")
	} else {
		kkd.Logger.SetOutput(logfile)
	}

	kkd.Logger.SetLevel(level)
	return nil
}

// handleBackupReport обрабатывает отчеты о резервном копировании и записывает результаты в лог, а также отправляет уведомления в телеграм.
func (kkd *KronosKeeperDeamon) handleBackupReport(backupReport *service.BackupReport) error {
	var ERRORS error // Общая ошибка, в которую будут записываться все ошибки

	// Обработка локальной резервной копии
	if backupReport.Local != nil {
		msg := fmt.Sprintf("%v - Успешно созданна резервная копия в %v на локальном диске", backupReport.CurrentTime, backupReport.Local.ArchiveName)
		kkd.writeLogAndNotify(msg)
	}

	// Обработка удаленных хранилищ
	if backupReport.Remote != nil {
		// Обработка загрузки на Google Cloud
		if backupReport.Remote.GCloud.Status {
			msg := fmt.Sprintf("%v - Успешная загрузка резервной копии %v на google Cloud disk", backupReport.CurrentTime, backupReport.Local.ArchiveName)
			kkd.writeLogAndNotify(msg)
		} else if backupReport.Remote.GCloud.Err != nil {
			err := backupReport.Remote.GCloud.Err
			kkd.writeLogAndNotifyError(err.Error())
			ERRORS = fmt.Errorf("%v: %v", ERRORS, err) // Добавляем ошибку в общий список ошибок
		}
		// Обработка загрузки на Google Drive
		if backupReport.Remote.GDrive.Status {
			msg := fmt.Sprintf("%v - Успешная загрузка резервной копии %v на google Drive disk", backupReport.CurrentTime, backupReport.Local.ArchiveName)
			kkd.writeLogAndNotify(msg)
		} else if backupReport.Remote.GDrive.Err != nil {
			err := backupReport.Remote.GCloud.Err
			kkd.writeLogAndNotifyError(err.Error())
			ERRORS = fmt.Errorf("%v: %v", ERRORS, err) // Добавляем ошибку в общий список ошибок
		}
		// Обработка других удаленных хранилищ
	}

	return ERRORS
}

// writeLogAndNotify записывает информацию в лог и отправляет уведомления в телеграм.
func (kkd *KronosKeeperDeamon) writeLogAndNotify(msg string) {
	kkd.Logger.Infof(msg)
	kkd.notifyTelegram(msg)
}

// writeLogAndNotifyError записывает ошибку в лог и отправляет уведомления в телеграм.
func (kkd *KronosKeeperDeamon) writeLogAndNotifyError(msg string) {
	kkd.Logger.Errorf(msg)
	kkd.notifyTelegram(msg)
}

// notifyTelegram отправляет уведомление в телеграм.
func (kkd *KronosKeeperDeamon) notifyTelegram(msg string) {
	if kkd.config.Telegram != nil { // Проверка наличия настроек Telegram
		// Проверяем, что TelegramBot был инициализирован
		if kkd.tg != nil && kkd.tg.BotAPI != nil {
			if err := kkd.tg.SendMessage(msg); err != nil {
				kkd.Logger.Warning(err)
			}
		} else {
			kkd.Logger.Warning("TelegramBot не был инициализирован, уведомление не может быть отправлено")
		}
	} else {
		kkd.Logger.Warning("Отсутствуют настройки Telegram, уведомление не может быть отправлено")
	}
}
