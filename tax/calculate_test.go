package tax

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCalculateProgressiveTax(t *testing.T) {
	tests := []struct {
		name     string
		income   float64
		expected float64
	}{
		{"Zero Income", 0, 0},
		{"Below First Threshold", 100000, 0},
		{"At First Threshold", 150000, 0},
		{"Above First Threshold", 200000, 5000},     // 200000 - 150000 = 50000, 50000 * 10% = 5000
		{"Middle Bracket", 600000, 50000},           // First 350000 taxed at 10%, next 100000 at 15%
		{"At Second Threshold", 1000000, 110000},    // 85000 + 75000 = 90000
		{"Above Second Threshold", 1500000, 210000}, // 90000 + (500000 * 20%) = 190000
		{"Top Bracket", 2500000, 485000},            // 190000 + (500000 * 35%) = 467500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateProgressiveTax(tt.income); got != tt.expected {
				t.Errorf("calculateProgressiveTax(%f) = %f; want %f", tt.income, got, tt.expected)
			}
		})
	}
}

func TestCalculateTax(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name               string
		requestBody        string
		mockDeduction      float64
		mockDeductionErr   error
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "Valid Request",
			requestBody:        `{"totalIncome": 500000, "wht": 100, "allowances": [{"allowanceType": "donation", "amount": 50}]}`,
			mockDeduction:      60000,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"tax":29000}`,
		},
		{
			name:               "Database Error",
			requestBody:        `{"totalIncome": 500000, "wht": 100, "allowances": [{"allowanceType": "donation", "amount": 50}]}`,
			mockDeductionErr:   sql.ErrConnDone,
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   `{"message":"Failed to fetch deduction data"}`,
		},
		{
			name:               "Deduction Not Found",
			requestBody:        `{"totalIncome": 500000, "wht": 100, "allowances": [{"allowanceType": "donation", "amount": 50}]}`,
			mockDeductionErr:   sql.ErrNoRows,
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   `{"message":"Deduction data not found"}`,
		},
		{
			name:               "Invalid Request - Negative WHT",
			requestBody:        `{"totalIncome": 500000, "wht": -100, "allowances": [{"allowanceType": "donation", "amount": 50}]}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"'WHT' cannot be negative"}`,
		},
		{
			name:               "Invalid Request - Missing totalIncome",
			requestBody:        `{"wht": 100, "allowances": [{"allowanceType": "donation", "amount": 50}]}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"'totalIncome' must be specified and greater than zero"}`,
		},
		{
			name:               "Invalid Request - Malformed JSON",
			requestBody:        `{"totalIncome": 500000, "wht": 100, "allowances": "not an array"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"Invalid request: JSON contains unknown fields or incorrect format"}`,
		}, {
			name:               "Invalid Allowance - Negative Amount",
			requestBody:        `{"totalIncome": 500000, "wht": 100, "allowances": [{"allowanceType": "donation", "amount": -50}]}`,
			mockDeduction:      60000,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"Allowance amounts cannot be negative"}`,
		},
		{
			name:               "Invalid Allowance - Empty Allowance Type",
			requestBody:        `{"totalIncome": 500000, "wht": 100, "allowances": [{"allowanceType": "", "amount": 50}]}`,
			mockDeduction:      60000,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"Allowance type cannot be empty"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchDeductionFunc = func() (float64, error) {
				return tt.mockDeduction, tt.mockDeductionErr
			}
			defer func() { fetchDeductionFunc = fetchDeduction }()

			req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := CalculateTax(c); err != nil {
				t.Fatal("CalculateTax failed:", err)
			}

			if rec.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, rec.Code)
			}

			if strings.TrimSpace(rec.Body.String()) != tt.expectedResponse {
				t.Errorf("Expected response %s, got %s", tt.expectedResponse, strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}
