package models

import (
	"time"

	. "github.com/rradhika/api-reservasi/datastore"
)

type Employee struct {
	Esid           int64  `db:"esid"`
	Name           string `db:"name"`
	Telegram       string `db:"telegram"`
	TelegramChatID string `db:"telegram_chat_id"`
	StartDate      string `db:"start_date"`
	EndDate        string `db:"end_date"`
	RoomName       string `db:"room_name"`
}

func (Employee) TableName() string {
	return "employee"
}

// GetEmployees
func (p *Employee) GetEmployees() (emps []Employee) {
	Db.Where("deleted_at is null").Find(&emps)
	return
}

// GetEmployee
func (p *Employee) GetEmployee(telegram string) (emps Employee) {
	Db.
		Raw("SELECT * FROM employee WHERE telegram = ?", "@"+telegram).
		Scan(&emps)

	return
}

// CheckEmployee
func (p *Employee) CheckEmployee(telegram string) (ada bool) {
	var emps []Employee

	Db.
		Raw("SELECT telegram FROM employee WHERE telegram = ?", "@"+telegram).
		Scan(&emps)
	ada = len(emps) > 0
	return
}

// CheckPenggunaRuangan Employee and join
func (p *Employee) CheckPenggunaRuangan(room_id int64) (emps []Employee) {

	Db.
		Raw("SELECT e.esid, e.name, e.telegram, con.employee_id, con.room_id, con.start_date, con.end_date, con.created_date, mr.id as mid, mr.name as room_name FROM reservation_room con LEFT JOIN employee e ON(e.esid = con.employee_id) LEFT JOIN meeting_room mr ON(mr.id = con.room_id) WHERE (DATE(con.start_date) <= ? AND DATE(con.end_date) >= ?) AND con.room_id = ?", time.Now().Format(LayoutSQL), time.Now().Format(LayoutSQL), room_id).
		Scan(&emps)

	return
}

// CheckPenggunaRuangan Employee and join
func (p *Employee) CheckToday() (ada bool) {
	var emps []Employee
	// fmt.Println(time.Now().Format(LayoutSQL))
	Db.
		Raw("SELECT start_date, end_date, created_date FROM reservation_room WHERE DATE(start_date) <= ? AND DATE(end_date) >= ?", time.Now().Format(LayoutSQL), time.Now().Format(LayoutSQL)).
		Scan(&emps)
	ada = len(emps) > 0
	return
}
