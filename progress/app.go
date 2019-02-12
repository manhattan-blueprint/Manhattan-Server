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

type ProgressResponse struct {
	Blueprints []BlueprintResponse `json:"blueprints"`
}

type BlueprintResponse struct {
	ItemID uint32 `json:"item_id"`
}

const BEARER_PREFIX string = "Bearer "
const MAX_ITEM_ID uint32 = 16

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
	a.Router.HandleFunc(fmt.Sprintf("%s/progress", prefix),
		a.getProgress).Methods(http.MethodGet)
	a.Router.HandleFunc(fmt.Sprintf("%s/progress", prefix),
		a.addProgress).Methods(http.MethodPost)
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

/* Check sent blueprint list is valid */
func checkValidProgress(pro Progress) error {
	if len(pro.Blueprints) <= 0 {
		return errors.New("Empty blueprint list")
	}
	for i := 0; i < len(pro.Blueprints); i++ {
		/* Does not check if item ID is actually a blueprint, only if valid item ID.
		** Ideally this would check against the item schema, to see if the item
		** ID(s) sent have the correct type
		 */
		if pro.Blueprints[i].ItemID <= 0 || pro.Blueprints[i].ItemID > MAX_ITEM_ID {
			return errors.New("Invalid item ID in list")
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
