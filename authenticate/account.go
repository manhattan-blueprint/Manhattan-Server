package main

import (
	"database/sql"
)

type Account struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (acc *Account) GetID(db *sql.DB) error {
	stmt := "SELECT user_id FROM account WHERE username=? AND password=?"
	// Prepared statements implemented by sql package
	return db.QueryRow(stmt, acc.Username, acc.Password).Scan(&acc.UserID)
}
