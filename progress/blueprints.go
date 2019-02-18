package main

import (
	"database/sql"
	"strings"
)

type Progress struct {
	Blueprints []Blueprint `json:"blueprints"`
}

type Blueprint struct {
	UserID uint32 `json:"user_id"`
	ItemID uint32 `json:"item_id"`
}

func (pro *Progress) AddProgress(db *sql.DB) error {
	stmt := "INSERT IGNORE INTO progress VALUES"
	values := []interface{}{}
	for i := 0; i < len(pro.Blueprints); i++ {
		stmt += " (?, ?),"
		values = append(values, pro.Blueprints[i].UserID, pro.Blueprints[i].ItemID)
	}
	// Remove trailing comma
	stmt = strings.TrimSuffix(stmt, ",")

	_, err := db.Exec(stmt, values...)

	return err
}
