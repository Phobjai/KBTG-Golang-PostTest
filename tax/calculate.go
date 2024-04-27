package tax

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	fetchDeductionFunc = fetchDeduction
)

func CalculateTax(c echo.Context) error {
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields() // Ensure no unknown fields are sent

	var req TaxRequest
	if err := decoder.Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request: JSON contains unknown fields or incorrect format"})
	}

	valid, errMsg := validateReq(&req)
	if !valid {
		return c.JSON(http.StatusBadRequest, Err{Message: errMsg})
	}

	deduction, err := fetchDeductionFunc()
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, Err{Message: "Deduction data not found"})
		}
		return c.JSON(http.StatusInternalServerError, Err{Message: "Failed to fetch deduction data"})
	}

	totalDonations := calculateDonationAllowance(req.Allowances)

	netIncome := req.TotalIncome - deduction - totalDonations
	calculatedTax := calculateProgressiveTax(netIncome)
	taxAfterWHT := calculatedTax - req.WHT

	// Prepare the response
	response := TaxResponse{Tax: taxAfterWHT}
	if taxAfterWHT < 0 { // If the tax after WHT is negative, it means a refund is due
		response.Tax = 0
		response.TaxRefund = -taxAfterWHT // Refund amount is the negative tax
	}

	return c.JSON(http.StatusOK, response)
}

func calculateDonationAllowance(allowances []Allowance) float64 {
	totalDonations := 0.0
	for _, allowance := range allowances {
		if allowance.AllowanceType == "donation" {
			if allowance.Amount > 100000 {
				totalDonations += 100000
			} else {
				totalDonations += allowance.Amount
			}
		}
	}
	return totalDonations
}

func fetchDeduction() (float64, error) {
	var deduction float64
	err := db.QueryRow("SELECT deduction FROM admin_config").Scan(&deduction)
	return deduction, err
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

func validateReq(req *TaxRequest) (bool, string) {
	if req.TotalIncome <= 0 {
		return false, "'totalIncome' must be specified and greater than zero"
	}
	if req.WHT < 0 {
		return false, "'WHT' cannot be negative"
	}

	if req.WHT > req.TotalIncome {
		return false, "'WHT' cannot be more than total income"
	}
	for _, allowance := range req.Allowances {
		if allowance.AllowanceType != "donation" {
			return false, "allowanceType must be 'donation'"
		}
		if allowance.Amount < 0 {
			return false, "Allowance amounts cannot be negative"
		}
	}
	return true, ""
}
