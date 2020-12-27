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

// CheckWaktuEmployee mem-validasi employee tidak boleh memesan 2 tempat di waktu yang sama
func (p *ReservationData) CheckWaktuEmployee(data *ReservationData) (ada bool) {
	var emps []ReservationData
	Db.
		Raw("SELECT * FROM reservation_room WHERE (start_date <= ? OR end_date >= ?) AND employee_id = ? ", data.StartDate, data.EndDate, data.EmployeeID).
		Scan(&emps)
	ada = len(emps) == 0
	return
}

// CheckWaktuRuangan mem-validasi ruangan tidak boleh dipesan diwaktu yg sama
func (p *ReservationData) CheckWaktuRuangan(data *ReservationData) (ada bool) {
	var emps []ReservationData
	Db.
		Raw("SELECT * FROM reservation_room WHERE (start_date <= ? OR end_date >= ?) AND room_id = ? ", data.StartDate, data.EndDate, data.RoomID).
		Scan(&emps)
	ada = len(emps) == 0
	return
}

//CreateData untuk create
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
