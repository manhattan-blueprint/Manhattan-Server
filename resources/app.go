package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

type ResourcesResponse struct {
	Spawns []SpawnResponse `json:"spawns"`
}

type SpawnResponse struct {
	ItemID   uint32           `json:"item_id"`
	Location LocationResponse `json:"location"`
}

type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

const BEARER_PREFIX string = "Bearer "

// Radius to return resources from, in kilometres
const RESOURCE_RADIUS int = 1

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
		a.getResources).Methods("GET")
	a.Router.HandleFunc(fmt.Sprintf("%s/resources", prefix),
		a.addResources).Methods("POST")
	a.Router.HandleFunc(fmt.Sprintf("%s/resources", prefix),
		a.removeResources).Methods("DELETE")
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
func getIDFromToken(db *sql.DB, w http.ResponseWriter,
	r *http.Request) (uint32, error) {
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

/* Validate auth token and return resources within radius */
func (a *App) getResources(w http.ResponseWriter, r *http.Request) {
	_, err := getIDFromToken(a.DB, w, r)
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

	// Check latitude and longitude are valid
	if lat < -90.0 || lat > 90.0 {
		respondWithError(w, http.StatusBadRequest,
			"Invalid latitude, must be between -90 and 90")
		return
	}
	if long < -180.0 || lat > 180.0 {
		respondWithError(w, http.StatusBadRequest,
			"Invalid longitude, must be between -180 and 180")
		return
	}

	/* Get resources within radius, using the Haversine formula, where 6371 is
	** the radius of Earth in kilometres */
	stmt := "SELECT item_id, gcs_lat, gcs_long FROM (SELECT item_id, " +
		"gcs_lat, gcs_long, (6371 * ACOS(COS(RADIANS(?)) " +
		"* COS(RADIANS(gcs_lat)) * COS(RADIANS(gcs_long) - RADIANS(?)) " +
		"+ SIN(RADIANS(?)) * SIN(RADIANS(gcs_lat)))) AS distance FROM " +
		"resources " + fmt.Sprintf("HAVING distance < %d ", RESOURCE_RADIUS) +
		"ORDER BY distance ASC) AS with_distance"

	rows, err := a.DB.Query(stmt, lat, long, lat)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert result rows into resources response structure
	var resRes ResourcesResponse
	// Handle no resources case
	resRes.Spawns = make([]SpawnResponse, 0)
	defer rows.Close()
	for rows.Next() {
		var spawnRes SpawnResponse
		var locRes LocationResponse
		err = rows.Scan(&spawnRes.ItemID, &locRes.Latitude, &locRes.Longitude)
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
	id, err := getIDFromToken(a.DB, w, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	fmt.Printf("ID: %d\n", id)

	respondWithError(w, http.StatusNotImplemented, "To be implemented")
}

/* Validate auth token, check user is developer and remove resource(s) */
func (a *App) removeResources(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, w, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	fmt.Printf("ID: %d\n", id)

	respondWithError(w, http.StatusNotImplemented, "To be implemented")
}
