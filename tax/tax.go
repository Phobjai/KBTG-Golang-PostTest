package tax

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type Err struct {
	Message string `json:"message"`
}

type AdminConfig struct {
	Deduction float64 `json:"deduction"`
	KReceipt  float64 `json:"k_receipt"`
}

type TaxRequest struct {
	TotalIncome float64     `json:"totalIncome"`
	WHT         float64     `json:"wht"`
	Allowances  []Allowance `json:"allowances"`
}

type Allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type TaxResponse struct {
	Tax       float64    `json:"tax"`
	TaxLevels []TaxLevel `json:"taxLevel"`
	TaxRefund float64    `json:"taxRefund,omitempty"`
}

type TaxCalculation struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax,omitempty"`
	TaxRefund   float64 `json:"taxRefund,omitempty"`
}
