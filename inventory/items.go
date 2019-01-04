package main

import (
	"database/sql"
	"strings"
)

type Inventory struct {
	Items []Item `json:"items"`
}

type Item struct {
	UserID   uint32 `json:"user_id"`
	ItemID   uint32 `json:"item_id"`
	Quantity uint32 `json:"quantity"`
}

/* Add a variable number of items to user inventory, checking if an entry
** exists for the given user ID, item ID pair. If one does exist, the
** quantity is added to the existing entry, otherwise a new entry is added
 */
func (inv *Inventory) AddInventory(db *sql.DB) error {
	stmt := "INSERT INTO inventory (user_id, item_id, quantity) VALUES"
	values := []interface{}{}
	for i := 0; i < len(inv.Items); i++ {
		stmt += " (?, ?, ?),"
		values = append(values, inv.Items[i].UserID, inv.Items[i].ItemID,
			inv.Items[i].Quantity)
	}
	// Remove the trailing comma
	stmt = strings.TrimSuffix(stmt, ",")
	stmt += " ON DUPLICATE KEY UPDATE quantity = quantity + VALUES (quantity)"

	_, err := db.Exec(stmt, values...)

	return err
}
