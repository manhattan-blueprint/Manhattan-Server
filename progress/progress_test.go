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

	err = checkDesktopTableExistsEmpty()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}

/* Check the progress table exists and is empty */
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

/* Check the desktop table exists and is empty */
func checkDesktopTableExistsEmpty() error {

	// Check desktop table
	var stateCount Count
	stateStmt := "SELECT COUNT(*) FROM desktop"
	stateCount.Value = 1
	err := testA.DB.QueryRow(stateStmt).Scan(&stateCount.Value)
	if err != nil {
		return err
	} else if stateCount.Value != 0 {
		return errors.New("Desktop table is not empty")
	}
	return nil
}

func clearProgressTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM progress")
	if err != nil {
		t.Errorf("Failed to clear progress table")
	}
}

func clearDesktopTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM desktop")
	if err != nil {
		t.Errorf("Failed to clear desktop table")
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
	payload := []byte(`{"blueprints":[{"item_id":11},{"item_id":18}]}`)

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
	if pro.Blueprints[0].ItemID != 11 || pro.Blueprints[1].ItemID != 18 {
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

	/* Item IDs must be greater than 0 and less than or equal 32 and correspond to
	** a blueprint, i.e. item schema type of 2 or 4 */
	payload := []byte(`{"blueprints":[{"item_id":0}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	payload = []byte(`{"blueprints":[{"item_id":33}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Check type 1 is not accepted
	payload = []byte(`{"blueprints":[{"item_id":3}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Check type 2 is accepted
	payload = []byte(`{"blueprints":[{"item_id":11}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check type 3 is not accepted
	payload = []byte(`{"blueprints":[{"item_id":14}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Check type 4 is accepted
	payload = []byte(`{"blueprints":[{"item_id":23}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check type 5 is not accepted
	payload = []byte(`{"blueprints":[{"item_id":32}]}`)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/progress",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Check blueprint list with both correct and incorrect item IDs is rejected
	payload = []byte(`{"blueprints":[{"item_id":1}, {"item_id":31}]}`)

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

	/* TODO: This really should check the returned JSON matches the item schema file */
}

/* Check correct desktop state is added and returned */
func TestAddGetDesktopState(t *testing.T) {
	clearDesktopTable(t)

	payload := []byte(`{"mapState":{"grid":[{"Key":{"x":-2.0,"y":5.0},"Value":{"id":1,"input":[],"output":[]}}]},"heldItemState":{"indexOfHeldItem":0},"inventoryState":{"inventoryContents":[{"Key":1,"Value":[{"hexID":0,"quantity":1}]}],"inventorySize":24}}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/progress/desktop-state",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, err = http.NewRequest(http.MethodGet, "/api/v1/progress/desktop-state", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	if !bytes.Equal(res.Body.Bytes(), payload) {
		t.Errorf("JSON returned does not match the JSON added")
	}
}

/* Check empty JSON is returned if no desktop state exists */
func TestGetEmptyDesktopState(t *testing.T) {
	clearDesktopTable(t)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/progress/desktop-state", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	expectedPayload := []byte(`{}`)

	if !bytes.Equal(res.Body.Bytes(), expectedPayload) {
		t.Errorf("JSON returned is non-empty")
	}
}
