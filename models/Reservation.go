package models

import (
	"fmt"

	. "github.com/rradhika/api-reservasi/datastore"
)

//Reservation for temporary
type Reservation struct {
	ID        int64  `db:"id"`
	ChatID    int64  `db:"chat_id"`
	RoomID    int64  `db:"room_id"`
	DateStart string `db:"date_start"`
	DateEnd   string `db:"date_end"`
}

func (Reservation) TableName() string {
	return "reservation_temp"
}

// CheckChatID Check Chat ID apakah sudah ada atau tidak?
func (p *Reservation) CheckChatID(ChatID int64) (ada bool) {
	var emps []Reservation
	// fmt.Println(time.Now().Format(LayoutSQL))
	Db.
		Raw("SELECT chat_id FROM reservation_temp WHERE chat_id = ?", ChatID).
		Scan(&emps)
	ada = len(emps) > 0
	return
}

// GetTemp
func (p *Reservation) GetTemp(ChatID int64) (emps Reservation) {

	Db.
		Raw("SELECT chat_id,room_id,date_start,date_end FROM reservation_temp WHERE chat_id = ?", ChatID).
		Scan(&emps)

	return
}

//Create Temporary data
func (p *Reservation) Create(data *Reservation) (*Reservation, error) {
	var err error
	if err != nil {
		return nil, err
	}
	err = Db.Table("reservation_temp").Create(&data).Error

	if err != nil {
		return nil, err
	}
	return data, err
}

//UpdateRoom Temporary data
func (p *Reservation) UpdateRoom(data *Reservation) (*Reservation, error) {
	var err error
	if err != nil {
		return nil, err
	}
	// fmt.Println(data.RoomID)
	err = Db.Table("reservation_temp").Where("chat_id = ?", data.ChatID).Update("room_id", data.RoomID).Error
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	return data, err
}

//UpdateStartDate Temporary data
func (p *Reservation) UpdateStartDate(data *Reservation) (*Reservation, error) {
	var err error
	if err != nil {
		return nil, err
	}
	// fmt.Println(data.RoomID)
	err = Db.Table("reservation_temp").Where("chat_id = ?", data.ChatID).Update("date_start", data.DateStart).Error
	if err != nil {
		return nil, err
	}
	return data, err
}

//UpdateEndDate Temporary data
func (p *Reservation) UpdateEndDate(data *Reservation) (*Reservation, error) {
	var err error
	if err != nil {
		return nil, err
	}
	// fmt.Println(data.RoomID)
	err = Db.Table("reservation_temp").Where("chat_id = ?", data.ChatID).Update("date_end", data.DateEnd).Error
	if err != nil {
		return nil, err
	}
	return data, err
}

//DeleteTemp Temporary data
func (p *Reservation) DeleteTemp(ChatID int64) (emps Reservation) {
	Db.
		Raw("DELETE FROM reservation_temp WHERE chat_id = ?", ChatID).
		Scan(&emps)

	return
}
