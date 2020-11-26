package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

//M is exported
type M map[string]interface{}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		data := M{"Message": "Hello", "Counter": 2}
		return c.JSON(http.StatusOK, data)
	})

	e.GET("/test", func(c echo.Context) error {
		name := c.QueryParam("name")
		data := fmt.Sprintf("Hello %s", name)

		return c.String(http.StatusOK, data)
	})
	e.Logger.Fatal(e.Start(":3001"))
}
