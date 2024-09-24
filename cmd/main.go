package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/shynn2/cmd-gram/internal/api"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "c:/Users/daji2/go/src/cmd-gram/pkg/client/postgresql/api.toml", "path to config file")
}

func main() {
	flag.Parse()
	config := api.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	api := api.New(mux.NewRouter(), config)

	if err := api.Start(); err != nil {
		log.Fatal(err)
	}
}
