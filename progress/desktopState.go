package main

import (
	"database/sql"
)

type DesktopState struct {
	UserID    uint32 `json:"user_id"`
	GameState string `json:"game_state"`
}

func (deskState *DesktopState) AddState(db *sql.DB) error {
	stmt := "INSERT INTO desktop VALUES (?, ?) ON DUPLICATE KEY UPDATE state=VALUES(state)"
	_, err := db.Exec(stmt, deskState.UserID, deskState.GameState)
	return err
}
