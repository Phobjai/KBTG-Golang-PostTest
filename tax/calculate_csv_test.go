package tax

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Mock fetchDeductionFunc for testing
func mockFetchDeduction() (float64, error) {
	return 60000, nil // Assume deduction is always 10000 for testing
}

func TestCalculateTaxFromCSV(t *testing.T) {
	// Create an instance of Echo
	e := echo.New()

	// Create a multipart form with a CSV file
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "taxes.csv")
	if err != nil {
		t.Fatal("Failed to create form file", err)
	}

	// Write CSV data to the form
	_, err = fw.Write([]byte("totalIncome,wht,donation\n500000,0,0\n600000,40000,20000\n750000,50000,15000"))
	if err != nil {
		t.Fatal("Failed to write to form file", err)
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	// Mocking deduction function
	fetchDeductionFunc = mockFetchDeduction
	defer func() { fetchDeductionFunc = fetchDeduction }() // Restore original function after test

	// Test
	if assert.NoError(t, CalculateTaxFromCSV(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"taxes":[{"totalIncome":500000,"tax":29000},{"totalIncome":600000,"taxRefund":2000},{"totalIncome":750000,"tax":11250}]}`
		assert.JSONEq(t, expected, rec.Body.String(), "Response body should match expected tax calculations")
	}
}
