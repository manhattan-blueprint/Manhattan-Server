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
	"strings"
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

func clearProgressTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM progress")
	if err != nil {
		t.Errorf("Failed to clear progress table")
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

/* Check invalid header is not accepted */
func TestInvalidHeader(t *testing.T) {
	clearProgressTable(t)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	// Header key field must be set to 'Authorization'
	req.Header.Set("Auth", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)

	// Header value field must have the 'Bearer ' prefix
	req.Header.Set("Authorization", strings.TrimPrefix(ACCESS_TOKEN, "Bearer "))

	res = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}

/* Check incorrect access token is not accepted */
func TestIncorrectToken(t *testing.T) {
	clearProgressTable(t)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	req.Header.Set("Authorization",
		"Bearer ydzvGQg2EcjTTHLSVHb7JTpkSRDdd0hQu2n5YPEM4CTfnqQIrqnufSIIOWchPNSA")
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}

/* Check correct access token is accepted */
func TestCorrectToken(t *testing.T) {
	clearProgressTable(t)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)
}

/* Check empty list is returned if user progress is empty */
func TestGetEmptyProgress(t *testing.T) {
	clearProgressTable(t)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check an empty items list is returned
	decoder := json.NewDecoder(res.Body)
	var pro Progress
	err = decoder.Decode(&pro)
	if err != nil {
		t.Errorf("Failed to decode progress response")
	}
	if len(pro.Blueprints) != 0 {
		t.Errorf("Expected empty list. Actual length was %d", len(pro.Blueprints))
	}
}

/* Check correct blueprint list is added and returned */
func TestAddGetProgress(t *testing.T) {
	clearProgressTable(t)

	// Add blueprints to progress
	payload := []byte(`{"blueprints":[{"item_id":1},{"item_id":9}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Get blueprints from progress
	req, err = http.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check the correct item IDs are returned
	decoder := json.NewDecoder(res.Body)
	var pro Progress
	err = decoder.Decode(&pro)
	if err != nil {
		t.Errorf("Failed to decode progress response")
	}
	if len(pro.Blueprints) != 2 {
		t.Errorf("Expected two blueprints. Actual number was %d",
			len(pro.Blueprints))
	}
	if pro.Blueprints[0].ItemID != 1 || pro.Blueprints[1].ItemID != 9 {
		t.Errorf("Expected item IDs 1 and 9. Actual IDs were %d and %d",
			pro.Blueprints[0].ItemID, pro.Blueprints[1].ItemID)
	}
}

/* Check empty blueprint lists are not accepted for adding */
func TestAddEmptyProgress(t *testing.T) {
	clearProgressTable(t)

	payload := []byte(`{"blueprints":[]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check blueprint lists with invalid item IDs are not accepted for adding */
func TestAddInvalidItemID(t *testing.T) {
	clearProgressTable(t)

	// Item IDs must be greater than 0 and less than or equal 16
	payload := []byte(`{"blueprints":[{"item_id":0}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	payload = []byte(`{"blueprints":[{"item_id":17}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check item schema is returned */
func TestGetItemSchema(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/v1/item-schema", nil)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	/* This really should check the returned JSON matches the item schema file */
}
