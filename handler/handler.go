package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Welcome() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome!")
	}
}

func GetEmployees(db *gorm.DB) echo.HandlerFunc {

	var employee Employee

	dsn := "root:@tcp(localhost:3306)/spera"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Find(&employee).Error; err != nil {
		// error handling here
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, employee)
}
