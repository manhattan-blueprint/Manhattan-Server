package main

import (
	crand "crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

type AccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenRequest struct {
	Refresh string `json:"refresh"`
}

type TokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type Count struct {
	Value int
}

// Expiration in years, months, days
var accessExpire = [3]int{0, 1, 0}
var refreshExpire = [3]int{1, 0, 0}

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
	a.Router.HandleFunc(fmt.Sprintf("%s/authenticate/register", prefix),
		a.registerUser).Methods("POST")
	a.Router.HandleFunc(fmt.Sprintf("%s/authenticate", prefix),
		a.validateLogin).Methods("POST")
	a.Router.HandleFunc(fmt.Sprintf("%s/authenticate/refresh", prefix),
		a.refreshTokens).Methods("POST")
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

/* Generate a unique given target id in a given table */
func generateID(db *sql.DB, table string, targetID string) (uint32, error) {
	seed := mrand.NewSource(time.Now().UnixNano())
	random := mrand.New(seed)
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

/* Generate a random token string of given length */
func generateToken(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b)[:n], err
}

/* Generate a random byte array */
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := crand.Read(b)
	return b, err
}

/* Create a new user and return auth tokens */
func (a *App) registerUser(w http.ResponseWriter, r *http.Request) {
	// Decode json body into account request
	decoder := json.NewDecoder(r.Body)
	var accReq AccountRequest
	err := decoder.Decode(&accReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			"Invalid username or password")
		return
	}
	// Check username and password fields are not blank
	if len(accReq.Username) == 0 || len(accReq.Password) == 0 {
		respondWithError(w, http.StatusBadRequest,
			"Invalid username or password")
		return
	}

	// Convert account request struct into database account struct
	acc := Account{
		Username: accReq.Username,
		Password: []byte(accReq.Password),
	}

	// Check no accounts with the same username exist
	usernameStmt := "SELECT COUNT(*) FROM account WHERE username=?"
	var usernameCount Count
	err = a.DB.QueryRow(usernameStmt, acc.Username).Scan(&usernameCount.Value)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if usernameCount.Value == 0 {
		// Create a unique user_id
		acc.UserID, err = generateID(a.DB, "account", "user_id")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Hash the password
		acc.Password, err = bcrypt.GenerateFromPassword(acc.Password,
			bcrypt.DefaultCost)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Create the account entry
		err = acc.CreateAccount(a.DB)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		respondWithError(w, http.StatusBadRequest, "Username already exists")
		return
	}

	tok := Token{UserID: acc.UserID}
	// Create a unique pair_id
	tok.PairID, err = generateID(a.DB, "token", "pair_id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create access token
	tok.Access, err = generateToken(64)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Change expiry
	tok.AccessExpire = time.Now().AddDate(accessExpire[0], accessExpire[1],
		accessExpire[2]).UnixNano()

	// Create refresh token
	tok.Refresh, err = generateToken(64)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Change expiry
	tok.RefreshExpire = time.Now().AddDate(refreshExpire[0], refreshExpire[1],
		refreshExpire[2]).UnixNano()

	// Create the token entry
	err = tok.CreateToken(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return token pair
	tokRes := TokenResponse{
		Access:  tok.Access,
		Refresh: tok.Refresh,
	}
	respondWithJSON(w, http.StatusOK, tokRes)
}

/* Validate a user login and return auth tokens */
func (a *App) validateLogin(w http.ResponseWriter, r *http.Request) {
	// Decode json body into account request
	decoder := json.NewDecoder(r.Body)
	var accReq AccountRequest
	err := decoder.Decode(&accReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			"Invalid username or password")
		return
	}
	if len(accReq.Username) == 0 || len(accReq.Password) == 0 {
		respondWithError(w, http.StatusBadRequest,
			"Invalid username or password")
		return
	}

	// Convert account request struct into database account struct
	acc := Account{Username: accReq.Username}

	// Get hashed password
	err = acc.GetPassword(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized,
				"The credentials provided do not match any user")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// Check password
	err = bcrypt.CompareHashAndPassword(acc.Password, []byte(accReq.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized,
			"The credentials provided do not match any user")
		return
	}

	// Get user_id
	err = acc.GetID(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusInternalServerError,
				"User not found after successful validation")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	tok := Token{UserID: acc.UserID}
	// Create a unique pair_id
	tok.PairID, err = generateID(a.DB, "token", "pair_id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create access token
	tok.Access, err = generateToken(64)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Change expiry
	tok.AccessExpire = time.Now().AddDate(accessExpire[0], accessExpire[1],
		accessExpire[2]).UnixNano()

	// Create refresh token
	tok.Refresh, err = generateToken(64)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Change expiry
	tok.RefreshExpire = time.Now().AddDate(refreshExpire[0], refreshExpire[1],
		refreshExpire[2]).UnixNano()

	// Create the token entry
	err = tok.CreateToken(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return token pair
	tokRes := TokenResponse{
		Access:  tok.Access,
		Refresh: tok.Refresh,
	}
	respondWithJSON(w, http.StatusOK, tokRes)
}

/* Validate a refresh token and return auth tokens */
func (a *App) refreshTokens(w http.ResponseWriter, r *http.Request) {
	// Decode json body into token request
	decoder := json.NewDecoder(r.Body)
	var tokReq TokenRequest
	err := decoder.Decode(&tokReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			"Invalid refresh token")
		return
	}
	if len(tokReq.Refresh) != 64 {
		respondWithError(w, http.StatusBadRequest,
			"Invalid refresh token")
		return
	}

	// Convert token request struct into database token struct
	tok := Token{Refresh: tokReq.Refresh}

	// Check refresh token exists
	refreshStmt := "SELECT COUNT(*) FROM token WHERE refresh=?"
	var refreshCount Count
	err = a.DB.QueryRow(refreshStmt, tok.Refresh).Scan(&refreshCount.Value)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if refreshCount.Value != 1 {
		// Also handles the near impossible case of two refresh tokens matching
		respondWithError(w, http.StatusUnauthorized,
			"The refresh token provided does not match any user")
		return
	}

	// Get user_id
	err = tok.GetID(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusInternalServerError,
				"User not found after successful validation")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Remove token
	err = tok.RemoveToken(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create a new token pair
	newTok := Token{UserID: tok.UserID}
	// Create a unique pair_id
	newTok.PairID, err = generateID(a.DB, "token", "pair_id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create access token
	newTok.Access, err = generateToken(64)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Change expiry
	newTok.AccessExpire = time.Now().AddDate(accessExpire[0], accessExpire[1],
		accessExpire[2]).UnixNano()

	// Create refresh token
	newTok.Refresh, err = generateToken(64)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Change expiry
	newTok.RefreshExpire = time.Now().AddDate(refreshExpire[0], refreshExpire[1],
		refreshExpire[2]).UnixNano()

	// Create the token entry
	err = newTok.CreateToken(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return token pair
	tokRes := TokenResponse{
		Access:  newTok.Access,
		Refresh: newTok.Refresh,
	}
	respondWithJSON(w, http.StatusOK, tokRes)
}
