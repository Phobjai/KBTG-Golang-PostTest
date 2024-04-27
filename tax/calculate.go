package tax

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CalculateTax(c echo.Context) error {
	var req TaxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request data"})
	}

	// Fetch deduction from the database
	var deduction float64
	err := db.QueryRow("SELECT deduction FROM admin_config").Scan(&deduction)

	fmt.Println("deduction on db is : ", deduction)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Deduction data not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch deduction data", "details": err.Error()})
	}

	// Calculate net income
	netIncome := req.TotalIncome - deduction

	// Calculate tax based on progressive tax rates
	tax := calculateProgressiveTax(netIncome)

	response := TaxResponse{
		Tax: tax,
	}

	return c.JSON(http.StatusOK, response)
}

func calculateProgressiveTax(income float64) float64 {
	tax := 0.0
	if income > 2000000 {
		tax += (income - 2000000) * 0.35
		income = 2000000
	}
	if income > 1000000 {
		tax += (income - 1000000) * 0.20
		income = 1000000
	}
	if income > 500000 {
		tax += (income - 500000) * 0.15
		income = 500000
	}
	if income > 150000 {
		tax += (income - 150000) * 0.10
	}

	return tax
}
