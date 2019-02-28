package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

type ID struct {
	Value uint32
}

type Count struct {
	Value int
}

type AccountType struct {
	Value string
}

type ResourcesResReq struct {
	Spawns []SpawnResReq `json:"spawns"`
}

type SpawnResReq struct {
	ItemID   uint32         `json:"item_id"`
	Location LocationResReq `json:"location"`
	Quantity uint32         `json:"quantity"`
}

type LocationResReq struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

const BEARER_PREFIX string = "Bearer "
const MAX_ITEM_ID uint32 = 16

// Radius to return resources from, in kilometres
const RESOURCE_RADIUS int = 1

// Expiration in years, months, days
var resourceExpire = [3]int{0, 1, 0}

/* Initialise database connection, mux router and routes */
func (a *App) Initialise(dbUser, dbPassword, dbHost, dbName string) error {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword,
		dbHost, dbName)
	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	a.Router = mux.NewRouter()
	a.initialiseRoutes()
	return nil
}

/* Run the server, listening on the given port */
func (a *App) Run(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(port)), a.Router)
}

/* Map routes to functions */
func (a *App) initialiseRoutes() {
	prefix := "/api/v1"
	a.Router.HandleFunc(fmt.Sprintf("%s/resources", prefix),
		a.getResources).Methods(http.MethodGet)
	a.Router.HandleFunc(fmt.Sprintf("%s/resources", prefix),
		a.addResources).Methods(http.MethodPost)
	a.Router.HandleFunc(fmt.Sprintf("%s/resources", prefix),
		a.removeResources).Methods(http.MethodDelete)
}

/* Respond with a error JSON */
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

/* Respond with a JSON */
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

/* Respond with an empty JSON */
func respondWithEmptyJSON(w http.ResponseWriter, code int) {
	response := []byte("{}")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

/* Validate auth token and get user ID */
func getIDFromToken(db *sql.DB, r *http.Request) (uint32, error) {
	var id ID

	// Get raw Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return id.Value, errors.New("Authorization header required")
	}

	// Check request is sending a bearer token
	if !strings.HasPrefix(authHeader, BEARER_PREFIX) {
		return id.Value, errors.New("Bearer token required")
	}

	// Get token string
	tokString := authHeader[len(BEARER_PREFIX):]

	// Check token and get user_id
	stmt := "SELECT user_id FROM token WHERE access=?"
	err := db.QueryRow(stmt, tokString).Scan(&id.Value)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return id.Value,
				errors.New("The access token provided does not match any user")
		default:
			/* TODO: This is really an internal server error; implement a custom
			** error type to return the correct HTTP code to respond to the user
			 */
			return id.Value, err
		}
	}
	return id.Value, nil
}

/* Validate username is a developer */
func checkDeveloper(db *sql.DB, id uint32) error {
	stmt := "SELECT account_type FROM account WHERE user_id=?"
	var accType AccountType
	err := db.QueryRow(stmt, id).Scan(&accType.Value)
	if err != nil {
		return errors.New("User not found")
	}

	if accType.Value != "developer" {
		return errors.New("User must be a developer")
	}
	return nil
}

/* Check sent spawn list is valid */
func checkValidResources(res ResourcesResReq) error {
	if len(res.Spawns) <= 0 {
		return errors.New("Empty spawn list")
	}
	for i := 0; i < len(res.Spawns); i++ {
		if res.Spawns[i].ItemID <= 0 || res.Spawns[i].ItemID > MAX_ITEM_ID {
			return errors.New("Invalid item ID in list")
		}
		err := checkValidLatLong(res.Spawns[i].Location.Latitude,
			res.Spawns[i].Location.Longitude)
		if err != nil {
			return err
		}
		if res.Spawns[i].Quantity <= 0 {
			return errors.New("Invalid resource quantity in list")
		}
	}
	return nil
}

/* Check floats are valid latitude and longitude */
func checkValidLatLong(lat, long float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("Invalid latitude, must be between -90 and 90")
	}
	if long < -180 || long > 180 {
		return errors.New("Invalid longitude, must be between -180 and 180")
	}
	return nil
}

/* Generate a unique given target id in a given table */
func generateID(db *sql.DB, table, targetID string) (uint32, error) {
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s=?", table, targetID)
	var id uint32
	idCount := Count{Value: 1}
	for idCount.Value != 0 {
		id = random.Uint32()
		err := db.QueryRow(stmt, id).Scan(&idCount.Value)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

/* Validate auth token and return resources within radius */
func (a *App) getResources(w http.ResponseWriter, r *http.Request) {
	_, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Get latitude and longitude parameters
	latString := r.URL.Query().Get("lat")
	longString := r.URL.Query().Get("long")
	if len(latString) == 0 || len(longString) == 0 {
		respondWithError(w, http.StatusBadRequest,
			"Latitude and longitude parameters are required")
		return
	}
	lat, err := strconv.ParseFloat(latString, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			"Could not convert latitude to float")
		return
	}
	long, err := strconv.ParseFloat(longString, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			"Could not convert longitude to float")
		return
	}

	err = checkValidLatLong(lat, long)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	/* Get resources within radius, using the Haversine formula, where 6371 is
	** the radius of Earth in kilometres */
	stmt := "SELECT item_id, gcs_lat, gcs_long, quantity FROM (SELECT item_id, gcs_lat, gcs_long, quantity, (6371 * ACOS(COS(RADIANS(?)) * COS(RADIANS(gcs_lat)) * COS(RADIANS(gcs_long) - RADIANS(?)) + SIN(RADIANS(?)) * SIN(RADIANS(gcs_lat)))) AS distance FROM resources " +
		fmt.Sprintf("HAVING distance < %d ", RESOURCE_RADIUS) +
		"ORDER BY distance ASC) AS with_distance"

	rows, err := a.DB.Query(stmt, lat, long, lat)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert result rows into resources response structure
	var resRes ResourcesResReq
	// Handle no resources case
	resRes.Spawns = make([]SpawnResReq, 0)
	defer rows.Close()
	for rows.Next() {
		var spawnRes SpawnResReq
		var locRes LocationResReq
		err = rows.Scan(&spawnRes.ItemID, &locRes.Latitude, &locRes.Longitude,
			&spawnRes.Quantity)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		spawnRes.Location = locRes
		resRes.Spawns = append(resRes.Spawns, spawnRes)
	}
	// Handle any errors encountered during iteration
	err = rows.Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, resRes)
}

/* Validate auth token, check user is developer and add resource(s) */
func (a *App) addResources(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = checkDeveloper(a.DB, id)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Decode json body into resources request
	decoder := json.NewDecoder(r.Body)
	var resReq ResourcesResReq
	err = decoder.Decode(&resReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid spawn list")
		return
	}

	err = checkValidResources(resReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert resources request into database resources struct
	var res Resources
	for i := 0; i < len(resReq.Spawns); i++ {
		var spawn Spawn
		// Create a unique spawn_id
		spawn.SpawnID, err = generateID(a.DB, "resources", "spawn_id")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		spawn.ItemID = resReq.Spawns[i].ItemID
		spawn.GCSLat = resReq.Spawns[i].Location.Latitude
		spawn.GCSLong = resReq.Spawns[i].Location.Longitude
		spawn.Quantity = resReq.Spawns[i].Quantity

		// Change expiry
		spawn.ResourceExpire = time.Now().AddDate(resourceExpire[0],
			resourceExpire[1], resourceExpire[2]).UnixNano()

		res.Spawns = append(res.Spawns, spawn)
	}

	// Query database
	err = res.AddResources(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEmptyJSON(w, http.StatusOK)
}

/* Validate auth token, check user is developer and remove resource(s) */
func (a *App) removeResources(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = checkDeveloper(a.DB, id)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Decode json body into resources request
	decoder := json.NewDecoder(r.Body)
	var resReq ResourcesResReq
	err = decoder.Decode(&resReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid spawn list")
		return
	}

	err = checkValidResources(resReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := "DELETE FROM resources WHERE (item_id, gcs_lat, gcs_long, quantity) IN ("
	values := []interface{}{}
	for i := 0; i < len(resReq.Spawns); i++ {
		stmt += "(?, ?, ?, ?), "
		values = append(values, resReq.Spawns[i].ItemID,
			resReq.Spawns[i].Location.Latitude,
			resReq.Spawns[i].Location.Longitude, resReq.Spawns[i].Quantity)
	}
	// Remove the trailing space and comma, and add closing parenthesis
	stmt = strings.TrimSuffix(stmt, ", ")
	stmt += ")"

	_, err = a.DB.Exec(stmt, values...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEmptyJSON(w, http.StatusOK)
}
