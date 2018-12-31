package main

import (
	"database/sql"
	"encoding/json"
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

const BEARER_PREFIX string = "Bearer "

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
	a.Router.HandleFunc(fmt.Sprintf("%s/inventory", prefix),
		a.getInventory).Methods("GET")
	a.Router.HandleFunc(fmt.Sprintf("%s/inventory", prefix),
		a.addInventory).Methods("POST")
	a.Router.HandleFunc(fmt.Sprintf("%s/inventory", prefix),
		a.removeInventory).Methods("DELETE")
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

func checkTokenAndGetID(db *sql.DB, w http.ResponseWriter,
	r *http.Request) (uint32, bool) {
	var id ID

	// Get raw Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest,
			"Authorization header required")
		return id.Value, false
	}

	// Check request is sending a bearer token
	if !strings.HasPrefix(authHeader, BEARER_PREFIX) {
		respondWithError(w, http.StatusBadRequest,
			"Bearer token required")
		return id.Value, false
	}

	// Get token string
	tokString := authHeader[len(BEARER_PREFIX):]

	// Check token and get user_id
	stmt := "SELECT user_id FROM token WHERE access=?"
	err := db.QueryRow(stmt, tokString).Scan(&id.Value)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized,
				"The access token provided does not match any user")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return id.Value, false
	}
	return id.Value, true
}

/* Validate auth token and return user inventory */
func (a *App) getInventory(w http.ResponseWriter, r *http.Request) {
	id, valid := checkTokenAndGetID(a.DB, w, r)
	if !valid {
		return
	}
	fmt.Printf("ID: %d\n", id)

	respondWithError(w, http.StatusNotImplemented, "To be implemented")
}

/* Validate auth token and add item to user inventory */
func (a *App) addInventory(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "To be implemented")
}

/* Validate auth token and remove item from user inventory */
func (a *App) removeInventory(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "To be implemented")
}
