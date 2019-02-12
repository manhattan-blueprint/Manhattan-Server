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

const DEV_ACCESS_TOKEN string = "Bearer ydzvGQg2EcjTTHLSVHb7JTpkSRDdd0hQu2n5YPEM4CTfnqQIrqnufSIIOWchPNSZ"

const NORMAL_ACCESS_TOKEN string = "Bearer CwlBrHSOzAC2NMLDREmeLeSdAeGMWcczp7KH2Ks9hWJtsMAey82kdRlggoqG0Yjr"

var testA App
var testConfig Configuration

func TestMain(m *testing.M) {
	fmt.Println("hello resources test")

	// Get the configuration
	testConfig, err := GetConfiguration("conf.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialise the router and database connection
	testA = App{}
	err = testA.Initialise(testConfig.DBUsername, testConfig.DBPassword,
		testConfig.DBHost, fmt.Sprintf("%s_test", testConfig.DBName),
		testConfig.Developers)
	if err != nil {
		log.Fatal(err)
	}

	err = checkResourcesTableExistsEmpty()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}

// Check the resources table exists and is empty
func checkResourcesTableExistsEmpty() error {

	// Check resources table
	var spawnCount Count
	spawnStmt := "SELECT COUNT(*) FROM resources"
	spawnCount.Value = 1
	err := testA.DB.QueryRow(spawnStmt).Scan(&spawnCount.Value)
	if err != nil {
		return err
	} else if spawnCount.Value != 0 {
		return errors.New("Resources table is not empty")
	}
	return nil
}

func clearResourcesTable(t *testing.T) {
	_, err := testA.DB.Exec("DELETE FROM resources")
	if err != nil {
		t.Errorf("Failed to clear resources table")
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

/* Check omitting latitude and longitude parameters is not accepted */
func TestGetMissingParameters(t *testing.T) {
	clearResourcesTable(t)

	// No URL parameters
	req, err := http.NewRequest(http.MethodGet, "/api/v1/resources", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Irrelevant URL parameters
	req, err = http.NewRequest(http.MethodGet, "/api/v1/resources?animal=duck",
		nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Missing longitude parameter
	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=50.123456", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Missing latitude parameter
	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?long=-1.123456", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

}

/* Check invalid latitude and longitude parameters are not accepted */
func TestGetInvalidParameters(t *testing.T) {
	clearResourcesTable(t)

	// Latitude and longitude must be numeric
	req, err := http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=will&long=smith", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Latitude must be between -90 and 90 inclusive
	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=-91&long=0", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=91&long=0", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Longitude must be between -180 and 180 inclusive
	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=0&long=-181", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=0&long=181", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

}

/* Check empty list is returned if no resources are spawned */
func TestGetEmptyResources(t *testing.T) {
	clearResourcesTable(t)

	req, err := http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=51.4560&long=2.6030", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check an empty spawns list is returned
	decoder := json.NewDecoder(res.Body)
	var resources ResourcesResReq
	err = decoder.Decode(&resources)
	if err != nil {
		t.Errorf("Failed to decode resources response")
	}
	if len(resources.Spawns) != 0 {
		t.Errorf("Expected empty list. Actual length was %d",
			len(resources.Spawns))
	}
}

/* Check adding resource and getting resource within radius of coordinate */
func TestAddGetResourcesWithinRadius(t *testing.T) {
	clearResourcesTable(t)

	// Add resource
	payload := []byte(`{"spawns":[{"item_id":5,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":3}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=51.456825&long=-2.601893", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	/* Check one resource with the correct item ID, latitude and longitude is
	** returned, assumes radius is set to 1km */
	decoder := json.NewDecoder(res.Body)
	var resources ResourcesResReq
	err = decoder.Decode(&resources)
	if err != nil {
		t.Errorf("Failed to decode resources response")
	}
	if len(resources.Spawns) != 1 {
		t.Errorf("Expected 1 resource. Actual number was %d",
			len(resources.Spawns))
	}
	if resources.Spawns[0].ItemID != 5 {
		t.Errorf("Expected item ID 5. Actual ID was %d",
			resources.Spawns[0].ItemID)
	}
	if resources.Spawns[0].Location.Latitude != 51.456061 {
		t.Errorf("Expected latitude 51.456061. Actual was %10.8f",
			resources.Spawns[0].Location.Latitude)
	}
	if resources.Spawns[0].Location.Longitude != -2.603104 {
		t.Errorf("Expected longitude -2.603104. Actual was %11.8f",
			resources.Spawns[0].Location.Longitude)
	}
	if resources.Spawns[0].Quantity != 3 {
		t.Errorf("Expected quantity of 3. Actual quantity was %d",
			resources.Spawns[0].Quantity)
	}
}

/* Check adding resource and getting resource outside radius of coordinate,
** assumes radius is set to 1km */
func TestAddGetResourcesOutsideRadius(t *testing.T) {
	clearResourcesTable(t)

	// Add resource
	payload := []byte(`{"spawns":[{"item_id":5,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":3}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=51.464514&long=-2.609992", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	/* Check no resources are returned */
	decoder := json.NewDecoder(res.Body)
	var resources ResourcesResReq
	err = decoder.Decode(&resources)
	if err != nil {
		t.Errorf("Failed to decode resources response")
	}
	if len(resources.Spawns) != 0 {
		t.Errorf("Expected no resources. Actual number was %d",
			len(resources.Spawns))
	}
}

/* Check adding resources and getting resources */
func TestAddGetMultipleResources(t *testing.T) {
	clearResourcesTable(t)

	// Add resource
	payload := []byte(`{"spawns":[{"item_id":1,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":3},{"item_id":2,"location":{"latitude":51.454244,"longitude":-2.607209},"quantity":4},{"item_id":3,"location":{"latitude":51.472149,"longitude":-2.625381},"quantity":5},{"item_id":4,"location":{"latitude":51.449645,"longitude":-2.581146},"quantity":6}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=51.456112&long=-2.605962", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check two resources are returned and the item IDs are correct
	decoder := json.NewDecoder(res.Body)
	var resources ResourcesResReq
	err = decoder.Decode(&resources)
	if err != nil {
		t.Errorf("Failed to decode resources response")
	}
	if len(resources.Spawns) != 2 {
		t.Errorf("Expected 2 resources. Actual number was %d",
			len(resources.Spawns))
	}
	if !(resources.Spawns[0].ItemID == 1 || resources.Spawns[0].ItemID == 2) {
		t.Errorf("Incorrect resources returned")
	}
	if !(resources.Spawns[1].ItemID == 1 || resources.Spawns[1].ItemID == 2) {
		t.Errorf("Incorrect resources returned")
	}
}

/* Check developer status is correctly returned */
func TestGetDeveloperStatus(t *testing.T) {
	clearResourcesTable(t)

	// Check developer status of normal account
	req, err := http.NewRequest(http.MethodGet, "/api/v1/resources/dev",
		nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check returned developer status is false
	decoder := json.NewDecoder(res.Body)
	var devRes DeveloperResponse
	err = decoder.Decode(&devRes)
	if err != nil {
		t.Errorf("Failed to decode developer response")
	}
	if devRes.Developer != false {
		t.Errorf("Expected developer status to be false. Actual was %t",
			devRes.Developer)
	}

	// Check developer status of developer account
	req, err = http.NewRequest(http.MethodGet, "/api/v1/resources/dev",
		nil)
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check returned developer status is true
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&devRes)
	if err != nil {
		t.Errorf("Failed to decode developer response")
	}
	if devRes.Developer != true {
		t.Errorf("Expected developer status to be true. Actual was %t",
			devRes.Developer)
	}

}

/* Check only developer accounts can add and remove resources */
func TestAddRemoveOnlyDeveloper(t *testing.T) {
	clearResourcesTable(t)

	// Add resource with a normal account
	payload := []byte(`{"spawns":[{"item_id":5,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":3}]}`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)

	// Add resource with developer account
	req, err = http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Remove resource with normal account
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)

	// Remove resource with developer account
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)
}

/* Check empty spawn lists are not accepted for adding and removing */
func TestAddRemoveEmptyResources(t *testing.T) {
	clearResourcesTable(t)

	payload := []byte(`{"spawns":[]}`)

	// Add
	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Remove
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check spawn lists with invalid item IDs are not accepting for adding and
** removing */
func TestAddRemoveInvalidItemID(t *testing.T) {
	clearResourcesTable(t)

	payload := []byte(`{"spawns":[{"item_id":17,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":3}]}`)

	// Add
	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Remove
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

}

/* Check spawn lists with invalid latitudes or longitudes are not accepted for
** adding and removing */
func TestAddRemoveInvalidLatLong(t *testing.T) {
	clearResourcesTable(t)

	// Invalid latitude
	payload := []byte(`{"spawns":[{"item_id":5,"location":{"latitude":151.456061,"longitude":-2.603104},"quantity":3}]}`)

	// Add
	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Remove
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Invalid longitude
	payload = []byte(`{"spawns":[{"item_id":5,"location":{"latitude":51.456061,"longitude":-182.603104}}]}`)

	// Add
	req, err = http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Remove
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)
}

/* Check spawn lists with invalid quantities are not accepted for
** adding and removing */
func TestAddRemoveInvalidQuantity(t *testing.T) {
	clearResourcesTable(t)

	payload := []byte(`{"spawns":[{"item_id":17,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":0}]}`)

	// Add
	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	// Remove
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

}

/* Check removing resource removes resource */
func TestRemoveResource(t *testing.T) {
	clearResourcesTable(t)

	payload := []byte(`{"spawns":[{"item_id":5,"location":{"latitude":51.456061,"longitude":-2.603104},"quantity":3}]}`)

	// Add
	req, err := http.NewRequest(http.MethodPost, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res := executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Remove
	req, err = http.NewRequest(http.MethodDelete, "/api/v1/resources",
		bytes.NewBuffer(payload))
	req.Header.Set("Authorization", DEV_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Get
	req, err = http.NewRequest(http.MethodGet,
		"/api/v1/resources?lat=51.456061&long=-2.603104", nil)
	req.Header.Set("Authorization", NORMAL_ACCESS_TOKEN)
	if err != nil {
		t.Errorf("Failed to create request")
	}

	res = executeRequest(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// Check an empty spawns list is returned
	decoder := json.NewDecoder(res.Body)
	var resources ResourcesResReq
	err = decoder.Decode(&resources)
	if err != nil {
		t.Errorf("Failed to decode resources response")
	}
	if len(resources.Spawns) != 0 {
		t.Errorf("Expected empty list. Actual length was %d",
			len(resources.Spawns))
	}
}
