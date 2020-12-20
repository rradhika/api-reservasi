package models

import (
	. "github.com/rradhika/api-reservasi/datastore"
)

//ReservationData for fixed data
type ReservationData struct {
	ID          int64  `db:"id"`
	EmployeeID  int64  `db:"employee_id"`
	RoomID      int64  `db:"room_id"`
	StartDate   string `db:"start_date"`
	EndDate     string `db:"end_date"`
	CreatedDate string `db:"created_date"`
}

func (ReservationData) TableName() string {
	return "reservation_room"
}

//Create Data
func (p *ReservationData) CreateData(data *ReservationData) (*ReservationData, error) {
	var err error
	if err != nil {
		return nil, err
	}
	err = Db.Table("reservation_room").Create(&data).Error

	if err != nil {
		return nil, err
	}
	return data, err
}

//DeleteData Delete reservation data
func (p *ReservationData) DeleteData(EmployeeID int64) (emps ReservationData) {
	Db.
		Raw("DELETE FROM reservation_room WHERE employee_id = ?", EmployeeID).
		Scan(&emps)

	return
}
