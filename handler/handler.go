package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rradhika/api-reservasi/models"
	"gorm.io/gorm"
)

func Welcome() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome!")
	}
}

func GetEmployees(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var em = []models.Employee{}

		if err := db.Find(&em).Error; err != nil {
			// error handling here
			log.Fatal(err)
		}

		return c.JSON(http.StatusOK, em)
	}
}
