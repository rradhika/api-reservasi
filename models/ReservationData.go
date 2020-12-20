package models

import (
	. "github.com/rradhika/api-reservasi/datastore"
)

//ReservationData for fixed data
type ReservationData struct {
	ID          int64  `db:"id"`
	EmployeeID  int64  `db:"employee_id"`
	RoomID      int64  `db:"room_id"`
	DateStart   string `db:"start_date"`
	DateEnd     string `db:"end_date"`
	CreatedDate string `db:"created_date"`
}

func (ReservationData) TableName() string {
	return "reservation_data"
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
