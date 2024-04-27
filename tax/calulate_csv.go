package tax

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CalculateTaxFromCSV(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid file"})
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	reader := csv.NewReader(src)
	var results []TaxCalculation

	if _, err := reader.Read(); err != nil { // Skip header
		return err
	}

	deduction, err := fetchDeductionFunc()
	if err != nil {
		return handleDeductionError(err, c)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, Err{Message: "Error reading CSV"})
		}

		totalIncome, wht, donation, err := parseRecord(record)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		grossTax, _ := calculateProgressiveTax(totalIncome - deduction - donation)
		netTax := grossTax - wht

		if netTax < 0 {
			results = append(results, TaxCalculation{TotalIncome: totalIncome, TaxRefund: -netTax})
		} else {
			results = append(results, TaxCalculation{TotalIncome: totalIncome, Tax: netTax})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"taxes": results})
}

func handleDeductionError(err error, c echo.Context) error {
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, Err{Message: "Deduction data not found"})
	}
	return c.JSON(http.StatusInternalServerError, Err{Message: "Failed to fetch deduction data"})
}

func parseRecord(record []string) (totalIncome, wht, donation float64, err error) {
	totalIncome, err = strconv.ParseFloat(record[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid totalIncome format")
	}

	wht, err = strconv.ParseFloat(record[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid WHT format")
	}

	donation, err = strconv.ParseFloat(record[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid donation format")
	}

	return totalIncome, wht, donation, nil
}
