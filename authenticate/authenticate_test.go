package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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

func clearAccountTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM account")
	if err != nil {
		t.Errorf("Failed to clear account table")
	}
}

func clearTokenTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM token")
	if err != nil {
		t.Errorf("Failed to clear token table")
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	testA.Router.ServeHTTP(rec, req)
	return rec
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code: %d. Actual: %d\n", expected, actual)
	}
}

// Check invalid JSONs are not accepted for registration
func TestRegisterInvalidUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"will":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

// Check valid JSONs are accepted for registration and tokens returned
func TestRegisterValidUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"username":"will","password":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check tokens returned are of length 64
	var m map[string]string
	json.Unmarshal(res.Body.Bytes(), &m)
	if len(m["access"]) != 64 {
		t.Errorf("Expected access token of length 64. Actual length was %d",
			len(m["access"]))
	}
	if len(m["refresh"]) != 64 {
		t.Errorf("Expected refresh token of length 64. Actual length was %d",
			len(m["refresh"]))
	}
}

// Check duplicate usernames are not accepted for registration
func TestRegisterDuplicateUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"username":"will","password":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

// Check invalid JSONs are not accepted for account validation
func TestValidateInvalidUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"will","password":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Validate with invalid JSON
	payload = []byte(`{"will":"smith"}`)
	req, err = http.NewRequest("POST", "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

// Check valid and correct JSONs are accepted and token returned
func TestValidateValidUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"will","password":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Validate with valid JSON
	payload = []byte(`{"username":"will","password":"smith"}`)

	req, err = http.NewRequest("POST", "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check tokens returned are of length 64
	var m map[string]string
	json.Unmarshal(res.Body.Bytes(), &m)
	if len(m["access"]) != 64 {
		t.Errorf("Expected access token of length 64. Actual length was %d",
			len(m["access"]))
	}
	if len(m["refresh"]) != 64 {
		t.Errorf("Expected refresh token of length 64. Actual length was %d",
			len(m["refresh"]))
	}
}

// Check combination of existing username but incorrect password is rejected
func TestValidateIncorrectPassword(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"will","password":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Validate with incorrect password
	payload = []byte(`{"username":"will","password":"stretch"}`)

	req, err = http.NewRequest("POST", "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

// Check combination of incorrect username but existing password is rejected
func TestValidateIncorrectUsername(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"will","password":"smith"}`)

	req, err := http.NewRequest("POST", "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Validate with incorrect password
	payload = []byte(`{"username":"stretch","password":"smith"}`)

	req, err = http.NewRequest("POST", "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}
