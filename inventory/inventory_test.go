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
	fmt.Println("hello inventory test")

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

	err = checkInventoryTableExistsEmpty()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}

// Check the inventory table exists and is empty
func checkInventoryTableExistsEmpty() error {

	// Check inventory table
	var itemCount Count
	itemStmt := "SELECT COUNT(*) FROM inventory"
	itemCount.Value = 1
	err := testA.DB.QueryRow(itemStmt).Scan(&itemCount.Value)
	if err != nil {
		return err
	} else if itemCount.Value != 0 {
		return errors.New("Inventory table is not empty")
	}
	return nil
}

func clearInventoryTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM inventory")
	if err != nil {
		t.Errorf("Failed to clear inventory table")
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
	clearInventoryTable(t)

	req, err := http.NewRequest("GET", "/api/v1/inventory", nil)
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
	clearInventoryTable(t)

	req, err := http.NewRequest("GET", "/api/v1/inventory", nil)
	req.Header.Set("Authorization",
		"Bearer ydzvGQg2EcjTTHLSVHb7JTpkSRDdd0hQu2n5YPEM4CTfnqQIrqnufSIIOWchPNSA")
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)
}

/* Check empty list is returned if user inventory is empty */
func TestGetEmptyInventory(t *testing.T) {
	clearInventoryTable(t)

	req, err := http.NewRequest("GET", "/api/v1/inventory", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check an empty items list is returned
	decoder := json.NewDecoder(res.Body)
	var inv Inventory
	err = decoder.Decode(&inv)
	if err != nil {
		t.Errorf("Failed to decode inventory response")
	}
	if len(inv.Items) != 0 {
		t.Errorf("Expected empty list. Actual length was %d", len(inv.Items))
	}
}

/* Check correct item list is added and returned */
func TestAddGetInventory(t *testing.T) {
	clearInventoryTable(t)

	// Add items to inventory
	payload := []byte(`{"items":[{"item_id":1,"quantity":4},{"item_id":9,"quantity":5}]}`)

	req, err := http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Get items from inventory
	req, err = http.NewRequest("GET", "/api/v1/inventory", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check the correct item ID, quantity pairs are returned
	decoder := json.NewDecoder(res.Body)
	var inv Inventory
	err = decoder.Decode(&inv)
	if err != nil {
		t.Errorf("Failed to decode inventory response")
	}
	if len(inv.Items) != 2 {
		t.Errorf("Expected two items. Actual number was %d", len(inv.Items))
	}
	if inv.Items[0].ItemID != 1 || inv.Items[1].ItemID != 9 {
		t.Errorf("Expected item IDs 1 and 9. Actual IDs were %d and %d",
			inv.Items[0].ItemID, inv.Items[1].ItemID)
	}
	if inv.Items[0].Quantity != 4 || inv.Items[1].Quantity != 5 {
		t.Errorf("Expected item quantities of 1 and 9. Actual quantities were %d and %d",
			inv.Items[0].Quantity, inv.Items[1].Quantity)
	}
}

/* Check empty item lists are not accepted for adding */
func TestAddEmptyInventory(t *testing.T) {
	clearInventoryTable(t)

	payload := []byte(`{"items":[]}`)

	req, err := http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check item lists with invalid item IDs are not accepted for adding */
func TestAddInvalidItemID(t *testing.T) {
	clearInventoryTable(t)

	// Item IDs must be greater than 0 and less than or equal 16
	payload := []byte(`{"items":[{"item_id":0,"quantity":4}]}`)

	req, err := http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	payload = []byte(`{"items":[{"item_id":17,"quantity":4}]}`)

	req, err = http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check item lists with quantities less than or equal zero are not accepted
** for adding */
func TestAddInvalidQuantity(t *testing.T) {
	clearInventoryTable(t)

	payload := []byte(`{"items":[{"item_id":1,"quantity":0}]}`)

	req, err := http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	payload = []byte(`{"items":[{"item_id":1,"quantity":-1}]}`)

	req, err = http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestDeleteInventory(t *testing.T) {
	clearInventoryTable(t)

	// Add items to inventory
	payload := []byte(`{"items":[{"item_id":1,"quantity":4},{"item_id":9,"quantity":5}]}`)

	req, err := http.NewRequest("POST", "/api/v1/inventory",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Delete items from inventory
	req, err = http.NewRequest("DELETE", "/api/v1/inventory", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Get the empty inventory
	req, err = http.NewRequest("GET", "/api/v1/inventory", nil)
	req.Header.Set("Authorization", ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check an empty items list is returned
	decoder := json.NewDecoder(res.Body)
	var inv Inventory
	err = decoder.Decode(&inv)
	if err != nil {
		t.Errorf("Failed to decode inventory response")
	}
	if len(inv.Items) != 0 {
		t.Errorf("Expected empty list. Actual length was %d", len(inv.Items))
	}
}
