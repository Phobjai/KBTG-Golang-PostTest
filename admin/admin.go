package admin

type DeductionUpdateRequest struct {
	Amount float64 `json:"amount"`
}

type DeductionUpdateResponse struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

type Err struct {
	Message string `json:"message"`
}

type KReceiptUpdateRequest struct {
	Amount float64 `json:"amount"`
}

type KReceiptUpdateResponse struct {
	KReceipt float64 `json:"kReceipt"`
}
