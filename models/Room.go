package models

import (
	. "github.com/rradhika/api-reservasi/datastore"
)

type Room struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Office string `db:"office"`
	Status string `db:"status"`
}

func (Room) TableName() string {
	return "meeting_room"
}

// MeetingRoom
func (p *Room) MeetingRoom(office string) (emps []Room) {

	Db.
		Raw("SELECT id, name, office, status FROM meeting_room WHERE status = ? AND office = ?", 1, office).
		Scan(&emps)

	return
}
