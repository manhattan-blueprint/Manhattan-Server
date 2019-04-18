package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

type AccountType struct {
	Value string
}

type ProgressResponse struct {
	Blueprints []BlueprintResponse `json:"blueprints"`
}

type BlueprintResponse struct {
	ItemID uint32 `json:"item_id"`
}

type LeaderboardResponse struct {
	LeaderboardElements []LeaderboardElementResponse `json:"leaderboard"`
}

type LeaderboardElementResponse struct {
	Username string `json:"username"`
	ItemID   uint32 `json:"item_id"`
}

type ItemSchema struct {
	Items []SchemaItem `json:"items"`
}

type SchemaItem struct {
	ItemID    uint32                  `json:"item_id"`
	Name      string                  `json:"name"`
	Type      uint32                  `json:"type"`
	Blueprint []SchemaBlueprintRecipe `json:"blueprint"`
	MachineID uint32                  `json:"machine_id"`
	Recipe    []SchemaBlueprintRecipe `json:"recipe"`
	Fuel      []SchemaFuel            `json:"fuel"`
}

type SchemaBlueprintRecipe struct {
	ItemID   uint32 `json:"item_id"`
	Quantity uint32 `json:"quantity"`
}

type SchemaFuel struct {
	ItemID uint32 `json:"item_id"`
}

const BEARER_PREFIX string = "Bearer "
const MAX_ITEM_ID uint32 = 32
const ITEM_SCHEMA string = "serve/item-schema-v2.json"

const (
	TYPE_PRIMARY_RESOURCE      = 1
	TYPE_BLUEPRINT_PLACEABLE   = 2
	TYPE_MACHINERY_UNPLACEABLE = 3
	TYPE_BLUEPRINT_UNPLACEABLE = 4
	TYPE_INTANGIBLE            = 5
)

var itemSchema ItemSchema
var itemTypeMap map[uint32]uint32

/* Initialise database connection, mux router, routes and item schema */
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

	err = GetItemSchema(ITEM_SCHEMA)
	if err != nil {
		return err
	}

	return nil
}

/* Run the server, listening on the given port */
func (a *App) Run(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(port)), a.Router)
}

func GetItemSchema(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&itemSchema)
	if err != nil {
		return err
	}

	// Build type map
	itemTypeMap = make(map[uint32]uint32)
	for i := 0; i < len(itemSchema.Items); i++ {
		itemTypeMap[itemSchema.Items[i].ItemID] = itemSchema.Items[i].Type
	}

	return nil
}

/* Map routes to functions */
func (a *App) initialiseRoutes() {
	prefix := "/api/v1"
	a.Router.HandleFunc(fmt.Sprintf("%s/progress", prefix),
		a.getProgress).Methods(http.MethodGet)
	a.Router.HandleFunc(fmt.Sprintf("%s/progress", prefix),
		a.addProgress).Methods(http.MethodPost)
	a.Router.HandleFunc(fmt.Sprintf("%s/progress/leaderboard", prefix),
		a.getLeaderboard).Methods(http.MethodGet)
	a.Router.HandleFunc(fmt.Sprintf("%s/progress/desktop-state", prefix),
		a.addDesktopState).Methods(http.MethodPost)
	a.Router.HandleFunc(fmt.Sprintf("%s/progress/desktop-state", prefix),
		a.getDesktopState).Methods(http.MethodGet)
	// Serve item schema
	a.Router.HandleFunc(fmt.Sprintf("%s/item-schema", prefix),
		a.getItemSchema).Methods(http.MethodGet)
}

/* Respond with a error JSON */
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

/* Respond with a JSON */
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
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

/* Respond with a raw byte array */
func respondWithRaw(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body)
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

/* Validate user_id is a developer */
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

/* Check sent blueprint list is valid */
func checkValidProgress(pro Progress) error {
	if len(pro.Blueprints) <= 0 {
		return errors.New("Empty blueprint list")
	}
	// Check IDs exist and are blueprints
	for i := 0; i < len(pro.Blueprints); i++ {
		value, ok := itemTypeMap[pro.Blueprints[i].ItemID]
		if ok {
			if !(value == TYPE_BLUEPRINT_PLACEABLE ||
				value == TYPE_BLUEPRINT_UNPLACEABLE) {
				return errors.New("Invalid ID in blueprint list")
			}
		} else {
			return errors.New("Out of range ID in blueprint list")
		}
	}
	return nil
}

/* Return user progress */
func (a *App) getProgress(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	stmt := "SELECT item_id FROM progress WHERE user_id=?"
	rows, err := a.DB.Query(stmt, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	// Convert results rows into progress response structure
	var proRes ProgressResponse
	// Handle empty progress case
	proRes.Blueprints = make([]BlueprintResponse, 0)
	for rows.Next() {
		var bluRes BlueprintResponse
		err = rows.Scan(&bluRes.ItemID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		proRes.Blueprints = append(proRes.Blueprints, bluRes)
	}
	// Handle any errors encountered during iteration
	err = rows.Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, proRes)
}

/* Add user progress */
func (a *App) addProgress(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Decode json body into progress struct
	decoder := json.NewDecoder(r.Body)
	var pro Progress
	err = decoder.Decode(&pro)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid blueprint list")
		return
	}

	err = checkValidProgress(pro)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add user ID to blueprint(s)
	for i := 0; i < len(pro.Blueprints); i++ {
		pro.Blueprints[i].UserID = id
	}

	// Query database
	err = pro.AddProgress(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEmptyJSON(w, http.StatusOK)
}

/* Return all player progress */
func (a *App) getLeaderboard(w http.ResponseWriter, r *http.Request) {
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

	stmt := "SELECT account.username, progress.item_id FROM progress INNER JOIN account ON progress.user_id = account.user_id"

	rows, err := a.DB.Query(stmt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var leadRes LeaderboardResponse
	leadRes.LeaderboardElements = make([]LeaderboardElementResponse, 0)
	for rows.Next() {
		var leadEleRes LeaderboardElementResponse
		err = rows.Scan(&leadEleRes.Username, &leadEleRes.ItemID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		leadRes.LeaderboardElements = append(leadRes.LeaderboardElements, leadEleRes)
	}
	// Handle any errors encountered during iteration
	err = rows.Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, leadRes)
}

/* Return item schema */
func (a *App) getItemSchema(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, itemSchema)
}

/* Add player desktop state as a JSON */
func (a *App) addDesktopState(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Read body into byte array
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	var deskState DesktopState
	deskState.UserID = id
	deskState.GameState = string(body)

	// Query database
	err = deskState.AddState(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEmptyJSON(w, http.StatusOK)
}

/* Get player desktop state as a JSON */
func (a *App) getDesktopState(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	stmt := "SELECT state FROM desktop WHERE user_id=?"
	var body []byte
	body = make([]byte, 0)
	err = a.DB.QueryRow(stmt, id).Scan(&body)
	if err != nil {
		respondWithEmptyJSON(w, http.StatusOK)
	}

	respondWithRaw(w, http.StatusOK, body)
}
