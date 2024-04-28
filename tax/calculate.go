package tax

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/Phobjai/assessment-tax/initdb"
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

	if deduction < 10000 {
		deduction = 10000
	}

	kReceiptMax, err := fetchKReceiptMax() // Assume fetchKReceiptMax is defined elsewhere
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "Failed to fetch k-receipt maximum"})
	}

	totalAllowances, err := calculateAllowances(req.Allowances, kReceiptMax)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "Failed to fetch allowance data"})
	}
	netIncome := req.TotalIncome - deduction - totalAllowances
	calculatedTax, taxLevels := calculateProgressiveTax(netIncome)

	fmt.Println("Calculated tax is ", calculatedTax)
	taxAfterWHT := calculatedTax - req.WHT
	fmt.Println("taxAfterWHT  is ", taxAfterWHT)

	// Prepare the response with tax level details
	response := TaxResponse{
		Tax:       max(0, taxAfterWHT), // Ensure tax is not negative in the response
		TaxLevels: taxLevels,
	}

	if taxAfterWHT < 0 { // If the tax after WHT is negative, it means a refund is due
		response.TaxRefund = -taxAfterWHT // Refund amount is the negative tax
	}

	return c.JSON(http.StatusOK, response)
}

func calculateAllowances(allowances []Allowance, kReceiptMax float64) (float64, error) {

	totalAllowances := 0.0
	for _, allowance := range allowances {
		if allowance.AllowanceType == "donation" && allowance.Amount > 100000 {
			totalAllowances += 100000
		} else if allowance.AllowanceType == "k-receipt" && allowance.Amount > kReceiptMax {
			totalAllowances += kReceiptMax
		} else {
			totalAllowances += allowance.Amount
		}
	}
	return totalAllowances, nil
}

func fetchDeduction() (float64, error) {
	var deduction float64
	err := initdb.DB.QueryRow("SELECT deduction FROM admin_config").Scan(&deduction)
	if err != nil {
		return 0, err
	}
	return deduction, err
}

func fetchKReceiptMax() (float64, error) {
	var kReceiptMax float64
	err := initdb.DB.QueryRow("SELECT kreceipt FROM admin_config").Scan(&kReceiptMax)
	if err != nil {
		return 0, err
	}
	return kReceiptMax, nil
}

func calculateProgressiveTax(income float64) (float64, []TaxLevel) {
	tax := 0.0
	var taxLevels []TaxLevel

	brackets := []struct {
		UpperBound float64
		TaxRate    float64
		Level      string // Ensure this matches the field name in your TaxLevel struct
	}{
		{150000, 0.0, "0-150,000"},
		{500000, 0.10, "150,001-500,000"},
		{1000000, 0.15, "500,001-1,000,000"},
		{2000000, 0.20, "1,000,001-2,000,000"},
		{math.MaxFloat64, 0.35, "2,000,001 ขึ้นไป"},
	}

	previousLimit := 0.0
	for _, bracket := range brackets {
		if income > previousLimit {
			taxableIncome := math.Min(income, bracket.UpperBound) - previousLimit
			taxAtThisLevel := taxableIncome * bracket.TaxRate
			tax += taxAtThisLevel
			taxLevels = append(taxLevels, TaxLevel{Level: bracket.Level, Tax: taxAtThisLevel})
		} else {
			taxLevels = append(taxLevels, TaxLevel{Level: bracket.Level, Tax: 0})
		}
		previousLimit = bracket.UpperBound
	}

	return tax, taxLevels
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
		if allowance.AllowanceType != "donation" && allowance.AllowanceType != "k-receipt" {
			return false, "allowanceType must be 'donation'"
		}
		if allowance.Amount < 0 {
			return false, "Allowance amounts cannot be negative"
		}
	}
	return true, ""
}
