package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Erikqwerty/KronosKeeper/internal/app/manager"
	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
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
		log.Fatal(err)
	}
	kkmanager, err := manager.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = kkmanager.ListDir(kkmanager.RemoteStorage.GDrive, "testdirbackup")
	if err != nil {
		fmt.Println(err)
	}
}
