package main

import (
	"database/sql"
)

type Token struct {
	UserID        int    `json:"user_id"`
	Access        string `json:"access"`
	Refresh       string `json:"refresh"`
	AccessExpire  int    `json:"access_expire"`
	RefreshExpire int    `json:"refresh_expire"`
}

func (tok *Token) GetTokens(db *sql.DB) error {
	stmt := "SELECT access, refresh FROM token WHERE user_id=?"
	// Prepared statements implemented by sql package
	return db.QueryRow(stmt, tok.UserID).Scan(&tok.Access, &tok.Refresh)
}
