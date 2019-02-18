package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
)

type Count struct {
	Value int
}

const ACCESS_TOKEN string = "Bearer ydzvGQg2EcjTTHLSVHb7JTpkSRDdd0hQu2n5YPEM4CTfnqQIrqnufSIIOWchPNSZ"

var testA App
var testConfig Configuration

func TestMain(m *testing.M) {
	fmt.Println("hello progress test")

	// Get the configuration
	testConfig, err := GetConfiguration("conf.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialise the router and database connection
	testA = App{}
	err = testA.Initialise(testConfig.DBUsername, testConfig.DBPassword,
		testConfig.DBHost, fmt.Sprintf("%s_test", testConfig.DBName))
	if err != nil {
		log.Fatal(err)
	}

	err = checkProgressTableExistsEmpty()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}

// Check the inventory table exists and is empty
func checkProgressTableExistsEmpty() error {

	// Check progress table
	var blueprintCount Count
	blueprintStmt := "SELECT COUNT(*) FROM progress"
	blueprintCount.Value = 1
	err := testA.DB.QueryRow(blueprintStmt).Scan(&blueprintCount.Value)
	if err != nil {
		return err
	} else if blueprintCount.Value != 0 {
		return errors.New("Progress table is not empty")
	}
	return nil
}
