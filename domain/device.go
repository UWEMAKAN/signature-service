package domain

type SignatureDevice struct {
	ID               string `json:"id"`
	Algorithm        string `json:"algorithm"`
	Label            string `json:"label"`
	SignatureCounter int    `json:"signatureCounter"`
	LastSignature    string `json:"lastSignature"`
	PublicKey        string `json:"-"`
	PrivateKey       string `json:"-"`
}

type SignatureDeviceRequest struct {
	ID        string  `json:"id"`
	Algorithm string  `json:"algorithm"`
	Label     *string `json:"label"`
}

type SignTransactionResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

type SignTransactionRequest struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}
