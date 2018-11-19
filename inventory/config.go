package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Port int `json:"port"`
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
