package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
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
	conf := config.NewConfig()
	_, err := toml.DecodeFile(configPath, conf) // загружаем конфигурацию из файла toml в config
	if err != nil {
		log.Fatal(err)
	}

}
