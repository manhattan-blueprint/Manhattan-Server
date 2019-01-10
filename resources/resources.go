package main

import (
	"fmt"
	"log"
)

var config Configuration

func main() {
	fmt.Println("hello resources")

	// Get the configuration
	config, err := GetConfiguration("conf.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialise and run
	a := App{}
	err = a.Initialise(config.DBUsername, config.DBPassword, config.DBHost,
		config.DBName, config.Developers)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(a.Run(config.Port))
}
