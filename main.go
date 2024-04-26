package main

import (
	"net/http"

	"github.com/Phobjai/assessment-tax/tax"
	"github.com/labstack/echo/v4"

	_ "github.com/lib/pq"
)

func main() {
	tax.InitDB()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
