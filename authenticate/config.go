package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Port int `json:"port"`
}

func GetConfiguration(fileName string) Configuration {
	var config Configuration
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		fmt.Println("error: ", err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error: ", err)
	}
	return config
}
