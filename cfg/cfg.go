package cfg

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type config struct {
	Timeout    time.Duration
	Port       string
	DbUser     string
	DbPassword string
	DbName     string
	FileName   string
}

func GetConfig() *config {
	var config config
	file, err := os.Open("cfg/config.json")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	json.NewDecoder(file).Decode(&config)
	return &config
}
