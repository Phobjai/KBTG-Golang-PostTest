package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Phobjai/assessment-tax/admin"
	"github.com/Phobjai/assessment-tax/initdb"
	"github.com/Phobjai/assessment-tax/tax"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"
)

func main() {
	initdb.InitDB()

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	e.POST("/tax/calculations", tax.CalculateTax)
	e.POST("/tax/calculations/upload-csv", tax.CalculateTaxFromCSV)

	//admin route
	adminGroup := e.Group("/admin/deductions")
	adminGroup.Use(middleware.BasicAuth(validateAdmin))
	adminGroup.POST("/personal", admin.UpdatePersonalDeduction)
	adminGroup.POST("/k-receipt", admin.UpdateKReceipt)

	port := os.Getenv("PORT")

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	//

	fmt.Println("shutting down the server")
}

func validateAdmin(username, password string, c echo.Context) (bool, error) {
	return username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD"), nil
}
