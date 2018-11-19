package main

import (
	"database/sql"
	"fmt"
)

type account struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (acc *account) getID(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT user_id FROM account WHERE username='%s' AND password='%s'",
		acc.Username, acc.Password)
	return db.QueryRow(statement).Scan(&acc.UserID)
}
