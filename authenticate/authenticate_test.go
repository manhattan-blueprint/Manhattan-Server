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

/* Check invalid JSON is not accepted for registration */
func TestRegisterInvalidJSON(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"John":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check valid blank JSON is not accepted for registration */
func TestRegisterBlankJSON(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"username":"","password":""}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check valid user is accepted for registration and 64 character tokens are
** returned */
func TestRegisterUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check tokens returned are of length 64 and account type is player
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
	if m["account_type"] != "player" {
		t.Errorf("Expected account type player. Actual was %s", m["account_type"])
	}
}

/* Check existing username is not accepted for registration */
func TestRegisterExistingUsername(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate/register",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}
	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check invalid JSON is not accepted for login */
func TestLoginInvalidJSON(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Login with invalid JSON
	payload = []byte(`{"John":"Smith"}`)
	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check valid user is accepted for login and 64 character tokens are returned
** */
func TestLoginUser(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Login with correct credentials
	payload = []byte(`{"username":"John","password":"Smith"}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check tokens returned are of length 64 and account type is player
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
	if m["account_type"] != "player" {
		t.Errorf("Expected account type player. Actual was %s", m["account_type"])
	}
}

/* Check login combination of existing username but incorrect password is
** rejected */
func TestLoginIncorrectPassword(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Login with incorrect password
	payload = []byte(`{"username":"John","password":"Stretch"}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}

/* Check login combination of incorrect username but existing password is
** rejected */
func TestLoginIncorrectUsername(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Login with incorrect username but existing password
	payload = []byte(`{"username":"Stretch","password":"Smith"}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}

/* Check invalid JSON is not accepted for refresh */
func TestRefreshInvalidJSON(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"John":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check valid blank JSON is not accepted for refresh */
func TestRefreshBlankToken(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	payload := []byte(`{"refresh":""}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check refresh token of incorrect length is not accepted for refresh */
func TestRefreshIncorrectTokenLength(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// Token with 63 characters
	payload := []byte(`{"refresh":"abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1"}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Token with 65 characters
	payload = []byte(`{"refresh":"abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz123"}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check valid token is accepted for refresh and 64 character tokens are
** returned */
func TestRefreshToken(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	decoder := json.NewDecoder(res.Body)
	var tok Token
	err = decoder.Decode(&tok)

	// Refresh with correct token
	payload = []byte(fmt.Sprintf("{\"refresh\":\"%s\"}", tok.Refresh))
	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check tokens returned are of length 64 and account type is player
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
	if m["account_type"] != "player" {
		t.Errorf("Expected account type player. Actual was %s", m["account_type"])
	}
}

/* Check incorrect token is not accepted for refresh */
func TestRefreshIncorrectToken(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	decoder := json.NewDecoder(res.Body)
	var tok Token
	err = decoder.Decode(&tok)

	// Create incorrect token
	var incorrectTok string
	if tok.Refresh != "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz12" {
		incorrectTok = "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz12"
	} else {
		fmt.Printf("What are the chances!\n")
		incorrectTok = "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz13"
	}

	payload = []byte(fmt.Sprintf("{\"refresh\":\"%s\"}", incorrectTok))
	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}

/* Check token entry is removed after being refreshed */
func TestRefreshRemoveToken(t *testing.T) {
	clearTokenTable(t)
	clearAccountTable(t)

	// First register a user
	payload := []byte(`{"username":"John","password":"Smith"}`)

	req, err := http.NewRequest(http.MethodPost,
		"/api/v1/authenticate/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	decoder := json.NewDecoder(res.Body)
	var tok Token
	err = decoder.Decode(&tok)

	// Refresh with correct token
	payload = []byte(fmt.Sprintf("{\"refresh\":\"%s\"}", tok.Refresh))
	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Reuse the previous token
	req, err = http.NewRequest(http.MethodPost, "/api/v1/authenticate/refresh",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request")
	}
	res = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}
