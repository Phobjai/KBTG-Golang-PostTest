package tax

import (
	"reflect"
	"testing"
)

func TestCalculateProgressiveTax(t *testing.T) {
	tests := []struct {
		name           string
		income         float64
		expectedTax    float64
		expectedLevels []TaxLevel
	}{
		{
			name:        "No Tax Income",
			income:      100000,
			expectedTax: 0,
			expectedLevels: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 0},
				{"500,001-1,000,000", 0},
				{"1,000,001-2,000,000", 0},
				{"2,000,001 ขึ้นไป", 0},
			},
		},
		{
			name:        "Middle Bracket Tax",
			income:      440000.0,
			expectedTax: 29000,
			expectedLevels: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 29000.0},
				{"500,001-1,000,000", 0},
				{"1,000,001-2,000,000", 0},
				{"2,000,001 ขึ้นไป", 0},
			},
		}, {
			name:        "Top of Third Bracket",
			income:      2000000,
			expectedTax: 310000,
			expectedLevels: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 35000},
				{"500,001-1,000,000", 75000},
				{"1,000,001-2,000,000", 200000},
				{"2,000,001 ขึ้นไป", 0},
			},
		},
		{
			name:        "Into the Highest Bracket",
			income:      3000000,
			expectedTax: 660000,
			expectedLevels: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 35000},
				{"500,001-1,000,000", 75000},
				{"1,000,001-2,000,000", 200000},
				{"2,000,001 ขึ้นไป", 350000},
			},
		},

		// Add more tests for each bracket...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTax, gotLevels := calculateProgressiveTax(tt.income)
			if gotTax != tt.expectedTax {
				t.Errorf("calculateProgressiveTax(%f) got tax %f, want tax %f", tt.income, gotTax, tt.expectedTax)
			}
			if !reflect.DeepEqual(gotLevels, tt.expectedLevels) {
				t.Errorf("calculateProgressiveTax(%f) got levels %v, want levels %v", tt.income, gotLevels, tt.expectedLevels)
			}
		})
	}
}

func TestCalculateAllowances(t *testing.T) {
	kReceiptMax := 50000.0 // Mocked maximum value for "k-receipt"

	tests := []struct {
		name        string
		allowances  []Allowance
		kReceiptMax float64
		expected    float64
	}{
		{"Donation Below Cap", []Allowance{{AllowanceType: "donation", Amount: 50000}}, kReceiptMax, 50000},
		{"Donation At Cap", []Allowance{{AllowanceType: "donation", Amount: 100000}}, kReceiptMax, 100000},
		{"Donation Above Cap", []Allowance{{AllowanceType: "donation", Amount: 200000}}, kReceiptMax, 100000},
		{"K-Receipt Below Cap", []Allowance{{AllowanceType: "k-receipt", Amount: 30000}}, kReceiptMax, 30000},
		{"K-Receipt At Cap", []Allowance{{AllowanceType: "k-receipt", Amount: 50000}}, kReceiptMax, 50000},
		{"K-Receipt Above Cap", []Allowance{{AllowanceType: "k-receipt", Amount: 70000}}, kReceiptMax, 50000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateAllowances(tt.allowances, tt.kReceiptMax)
			if err != nil {
				t.Errorf("calculateAllowances() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("calculateAllowances() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateReq(t *testing.T) {
	tests := []struct {
		name     string
		req      TaxRequest
		expected bool
		errMsg   string
	}{
		{
			name:     "Valid Request",
			req:      TaxRequest{TotalIncome: 300000, WHT: 20000, Allowances: []Allowance{{AllowanceType: "donation", Amount: 0}}},
			expected: true,
			errMsg:   "",
		},
		{
			name:     "Negative WHT",
			req:      TaxRequest{TotalIncome: 300000, WHT: -1000, Allowances: []Allowance{{AllowanceType: "donation", Amount: 0}}},
			expected: false,
			errMsg:   "'WHT' cannot be negative",
		},
		{
			name:     "Invalid totalincome",
			req:      TaxRequest{TotalIncome: -1, WHT: -1000, Allowances: []Allowance{{AllowanceType: "donation", Amount: 0}}},
			expected: false,
			errMsg:   "'totalIncome' must be specified and greater than zero",
		},
		{
			name:     "WHT greater than totalincome",
			req:      TaxRequest{TotalIncome: 400000, WHT: 2000000, Allowances: []Allowance{{AllowanceType: "donation", Amount: 0}}},
			expected: false,
			errMsg:   "'WHT' cannot be more than total income",
		},
		{
			name:     "allowanceType is not donation",
			req:      TaxRequest{TotalIncome: 400000, WHT: 0, Allowances: []Allowance{{AllowanceType: "xxx", Amount: 0}}},
			expected: false,
			errMsg:   "allowanceType must be 'donation'",
		},
		{
			name:     "allowanceType amount is negative",
			req:      TaxRequest{TotalIncome: 400000, WHT: 0, Allowances: []Allowance{{AllowanceType: "donation", Amount: -1}}},
			expected: false,
			errMsg:   "Allowance amounts cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, msg := validateReq(&tt.req)
			if valid != tt.expected {
				t.Errorf("validateReq(%+v) got valid %v, want %v", tt.req, valid, tt.expected)
			}
			if msg != tt.errMsg {
				t.Errorf("validateReq(%+v) got error message '%s', want '%s'", tt.req, msg, tt.errMsg)
			}
		})
	}
}
