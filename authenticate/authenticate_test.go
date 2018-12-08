package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
)

var testA App
var testConfig Configuration

func TestMain(m *testing.M) {
	fmt.Println("hello authenticate test")

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

	err = checkTablesExistEmpty()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}

// Check the account and token tables exist and are empty
func checkTablesExistEmpty() error {
	// Check account table
	var accountCount Count
	accountStmt := "SELECT COUNT(*) FROM account"
	accountCount.Value = 1
	err := testA.DB.QueryRow(accountStmt).Scan(&accountCount.Value)
	if err != nil {
		return err
	} else if accountCount.Value != 0 {
		return errors.New("Account table is not empty")
	}

	// Check token table
	var tokenCount Count
	tokenStmt := "SELECT COUNT(*) FROM token"
	tokenCount.Value = 1
	err = testA.DB.QueryRow(tokenStmt).Scan(&tokenCount.Value)
	if err != nil {
		return err
	} else if tokenCount.Value != 0 {
		return errors.New("Token table is not empty")
	}
	return nil
}
