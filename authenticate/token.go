package main

import (
	"database/sql"
)

type Token struct {
	PairID        uint32 `json:"paid_id"`
	UserID        uint32 `json:"user_id"`
	Access        string `json:"access"`
	Refresh       string `json:"refresh"`
	AccessExpire  int64  `json:"access_expire"`
	RefreshExpire int64  `json:"refresh_expire"`
}

func (tok *Token) CreateToken(db *sql.DB) error {
	stmt := "INSERT INTO token VALUES (?, ?, ?, ?, ?, ?)"
	// Prepared statements implemented by sql package
	_, err := db.Exec(stmt, tok.PairID, tok.UserID, tok.Access, tok.Refresh,
		tok.AccessExpire, tok.RefreshExpire)
	if err != nil {
		return err
	}
	return nil
}

func (tok *Token) GetTokens(db *sql.DB) error {
	stmt := "SELECT access, refresh FROM token WHERE user_id=?"
	return db.QueryRow(stmt, tok.UserID).Scan(&tok.Access, &tok.Refresh)
}
