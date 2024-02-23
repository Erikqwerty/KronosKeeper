package main

import (
	"flag"
	"fmt"
	"log"

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
	fmt.Println(conf)

}
