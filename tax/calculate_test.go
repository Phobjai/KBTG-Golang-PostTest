package tax

import (
	"testing"
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
