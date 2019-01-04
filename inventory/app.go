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

type InventoryResponse struct {
	Items []ItemResponse `json:"items"`
}

type ItemResponse struct {
	ItemID   uint32 `json:"item_id"`
	Quantity uint32 `json:"quantity"`
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

/* Check sent item list is valid */
func checkValidInventory(inv Inventory) error {
	if len(inv.Items) <= 0 {
		return errors.New("Empty item list")
	}
	for i := 0; i < len(inv.Items); i++ {
		if inv.Items[i].ItemID <= 0 || inv.Items[i].ItemID > 16 {
			return errors.New("Invalid item ID in list")
		}
		if inv.Items[i].Quantity <= 0 {
			return errors.New("Invalid item quantity in list")
		}
	}
	return nil
}

/* Return user inventory */
func (a *App) getInventory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, w, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	stmt := "SELECT item_id, quantity FROM inventory WHERE user_id=?"
	rows, err := a.DB.Query(stmt, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert result rows into inventory response structure
	var invRes InventoryResponse
	// Handle empty inventory case
	invRes.Items = make([]ItemResponse, 0)
	defer rows.Close()
	for rows.Next() {
		var itemRes ItemResponse
		err = rows.Scan(&itemRes.ItemID, &itemRes.Quantity)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		invRes.Items = append(invRes.Items, itemRes)
	}
	// Handle any errors encountered during iteration
	err = rows.Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, invRes)
}

/* Add item(s) to user inventory */
func (a *App) addInventory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, w, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Decode json body into inventory struct
	decoder := json.NewDecoder(r.Body)
	var inv Inventory
	err = decoder.Decode(&inv)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item list")
		return
	}

	err = checkValidInventory(inv)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add user ID to item(s)
	for i := 0; i < len(inv.Items); i++ {
		inv.Items[i].UserID = id
	}

	// Query database
	err = inv.AddInventory(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithEmptyJSON(w, http.StatusOK)
}

/* Remove all items from user inventory */
func (a *App) removeInventory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromToken(a.DB, w, r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	stmt := "DELETE FROM inventory WHERE user_id=?"
	_, err = a.DB.Exec(stmt, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEmptyJSON(w, http.StatusOK)
}
