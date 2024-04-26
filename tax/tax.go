package tax

type Err struct {
	Message string `json:"message"`
}

type AdminConfig struct {
	Deduction float64 `json:"deduction"`
	KReceipt  float64 `json:"k_receipt"`
}
