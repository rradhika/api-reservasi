package model

type Employee struct {
	ID       int64  `db:"esid"`
	Nik      string `db:"nik"`
	Name     string `db:"name"`
	Telegram string `db:"telegram"`
}

func (Employee) TableName() string {
	return "employee"
}
