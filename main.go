package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type Configuration struct {
	Token      string
	PGHost     string
	PGPort     int
	PGUser     string
	PGPassword string
	PGDbname   string
	BotID      string
}

func main() {
	fmt.Println("Starting lorebot")
	confPath := flag.String("conf", "conf.json", "Path to json configuration file")
	flag.Parse()

	file, err := os.Open(*confPath)
	if err != nil {
		log.Fatalf("failed to open config: %v", err)
	}

	var conf Configuration
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	lorebot := NewLorebot(&conf)
	lorebot.Start()
}
