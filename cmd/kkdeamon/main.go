package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Erikqwerty/KronosKeeper/internal/app/daemon"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/sirupsen/logrus"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/kronoskeeper.toml", "Path to configure file")
}

func main() {
	flag.Parse()

	conf, err := config.NewConfig(configPath)
	if err != nil {
		logrus.Fatal(err)
	}

	kkd := daemon.New(conf)

	// Устанавливаем канал для обработки сигналов завершения работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем демона
	if err := kkd.Start(); err != nil {
		logrus.Fatal(err)
	}

	// Ожидаем сигнала завершения работы
	<-stop

	kkd.Logger.Info("Принят сигнал завершения работы. Остановка демона...")

	// Останавливаем демона
	if err := kkd.Stop(); err != nil {
		kkd.Logger.Errorf("Ошибка при остановке демона: %v\n", err)
	}

	kkd.Logger.Info("Демон успешно остановлен.")
}
