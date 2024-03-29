package main

import (
	"database/sql"
)

type Account struct {
	UserID      uint32 `json:"user_id"`
	Username    string `json:"username"`
	Password    []byte `json:"password"`
	AccountType string `json:"account_type"`
}

func (acc *Account) CreateAccount(db *sql.DB) error {
	stmt := "INSERT INTO account VALUES (?, ?, ?, ?)"
	// Prepared statements implemented by sql package
	_, err := db.Exec(stmt, acc.UserID, acc.Username, acc.Password,
		acc.AccountType)
	return err
}

func (acc *Account) GetPassword(db *sql.DB) error {
	stmt := "SELECT password FROM account WHERE username=?"
	return db.QueryRow(stmt, acc.Username).Scan(&acc.Password)
}

func (acc *Account) GetIDAndType(db *sql.DB) error {
	stmt := "SELECT user_id, account_type FROM account WHERE username=?"
	return db.QueryRow(stmt, acc.Username).Scan(&acc.UserID, &acc.AccountType)
}

func (acc *Account) GetType(db *sql.DB) error {
	stmt := "SELECT account_type FROM account WHERE user_id=?"
	return db.QueryRow(stmt, acc.UserID).Scan(&acc.AccountType)
}
