package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

type config struct {
	Apps map[string]appsInfo
}

type appsInfo struct {
	Name            string
	Namespace       string
	TillerNamespace string
	Enabled         bool
	Chart           string
	Version         string
	ValuesFile      string
}

func main() {
	var conf config
	if _, err := toml.DecodeFile("example.toml", &conf); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Version: ", conf.Apps["card-zetryer"].Version)
	app := conf.Apps["card-zetryer"]
	app.Version = "1.4.3"
	conf.Apps["card-zetryer"] = app
	fmt.Println("Version: ", conf.Apps["card-zetryer"].Version)

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(conf); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Result: ", buf.String())

	file, err := os.OpenFile(
		"writeFile.toml",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Write(buf.Bytes())
}
