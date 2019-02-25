package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Port       int    `json:"port"`
	DBUsername string `json:"dbUsername"`
	DBPassword string `json:"dbPassword"`
	DBHost     string `json:"dbHost"`
	DBName     string `json:"dbName"`
}

func GetConfiguration(fileName string) (Configuration, error) {
	var config Configuration
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
