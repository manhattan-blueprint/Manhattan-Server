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
			/* TODO: This is really an internal server error, implement a custom
			** error type to return the correct HTTP code to respond to the user
			 */
			return id.Value, err
		}
	}
	return id.Value, nil
}

/* Validate auth token and return user inventory */
func (a *App) getInventory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, w, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
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
