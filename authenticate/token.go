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
	return err
}

func (tok *Token) RemoveToken(db *sql.DB) error {
	stmt := "DELETE FROM token WHERE refresh=?"
	_, err := db.Exec(stmt, tok.Refresh)
	return err
}

func (tok *Token) GetTokens(db *sql.DB) error {
	stmt := "SELECT access, refresh FROM token WHERE user_id=?"
	return db.QueryRow(stmt, tok.UserID).Scan(&tok.Access, &tok.Refresh)
}

func (tok *Token) GetID(db *sql.DB) error {
	stmt := "SELECT user_id FROM token WHERE refresh=?"
	return db.QueryRow(stmt, tok.Refresh).Scan(&tok.UserID)
}
