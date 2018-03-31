package main

import "fmt"
import "flag"
import "encoding/json"
import "os"

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
	file, _ := os.Open(*confPath)
	decoder := json.NewDecoder(file)
	var conf = Configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("error:", err)
	}
	lorebot := NewLorebot(&conf)
	lorebot.Start()
}
