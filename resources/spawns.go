package main

import (
	"database/sql"
	"strings"
)

type Resources struct {
	Spawns []Spawn `json:"spawns"`
}

type Spawn struct {
	SpawnID        uint32  `json:"spawn_id"`
	ItemID         uint32  `json:"item_id"`
	GCSLat         float64 `json:"gcs_lat"`
	GCSLong        float64 `json:"gcs_long"`
	ResourceExpire int64   `json:"resource_expire"`
}

/* Add a variable number of resources */
func (res *Resources) AddResources(db *sql.DB) error {
	stmt := "INSERT INTO resources VALUES"
	values := []interface{}{}
	for i := 0; i < len(res.Spawns); i++ {
		stmt += " (?, ?, ?, ?, ?),"
		values = append(values, res.Spawns[i].SpawnID, res.Spawns[i].ItemID,
			res.Spawns[i].GCSLat, res.Spawns[i].GCSLong,
			res.Spawns[i].ResourceExpire)
	}
	// Remove the trailing comma
	stmt = strings.TrimSuffix(stmt, ",")

	_, err := db.Exec(stmt, values...)

	return err
}
