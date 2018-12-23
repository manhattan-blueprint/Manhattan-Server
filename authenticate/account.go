package main

import (
	"database/sql"
)

type Account struct {
	UserID   uint32 `json:"user_id"`
	Username string `json:"username"`
	Password []byte `json:"password"`
}

func (acc *Account) CreateAccount(db *sql.DB) error {
	stmt := "INSERT INTO account VALUES (?, ?, ?)"
	// Prepared statements implemented by sql package
	_, err := db.Exec(stmt, acc.UserID, acc.Username, acc.Password)
	if err != nil {
		return err
	}
	return nil
}

func (acc *Account) GetPassword(db *sql.DB) error {
	stmt := "SELECT password FROM account WHERE username=?"
	return db.QueryRow(stmt, acc.Username).Scan(&acc.Password)
}

func (acc *Account) GetID(db *sql.DB) error {
	stmt := "SELECT user_id FROM account WHERE username=?"
	return db.QueryRow(stmt, acc.Username).Scan(&acc.UserID)
}
