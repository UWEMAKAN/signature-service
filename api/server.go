package api

import (
	"encoding/json"
	"net/http"

	"github.com/uwemakan/signing-service/crypto"
	"github.com/uwemakan/signing-service/persistence"
	"github.com/uwemakan/signing-service/services"
	"github.com/uwemakan/signing-service/utils"
)

// Response is the generic API response container.
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	listenAddress          string
	signatureDeviceService services.SignatureService
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		signatureDeviceService: services.NewSignatureService(
			services.SignatureServiceParams{
				Repo:           persistence.NewInMemorySignatureDeviceRepository(),
				KeyPairFactory: crypto.NewKeyPairFactory(),
				SignerFactory:  crypto.NewSignerFactory(),
			},
		),
	}
}

// Run registers all HandlerFuncs for the existing HTTP routes and starts the Server.
func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.Handle("/api/v0/health", http.HandlerFunc(s.Health))
	mux.Handle("/api/v0/signature-devices", http.HandlerFunc(s.Handler))
	mux.Handle("/api/v0/signature-devices/", http.HandlerFunc(s.GetSignatureDevice))
	mux.Handle("/api/v0/signature-devices/sign", http.HandlerFunc(s.SignTransaction))

	return http.ListenAndServe(s.listenAddress, mux)
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// HandleError matches errors to their corresponding http status codes
func HandleError(w http.ResponseWriter, err error) {
	switch err {
	case utils.ErrInvalidSignatureCounter,
		utils.ErrInvalidLastSignature,
		utils.ErrInvalidData,
		utils.ErrDeviceAlreadyExists,
		utils.ErrUnsupportedAlgorithm,
		utils.ErrInvalidDeviceId:
		WriteErrorResponse(w, http.StatusBadRequest, []string{err.Error()})
	case utils.ErrDeviceNotFound:
		WriteErrorResponse(w, http.StatusNotFound, []string{err.Error()})
	default:
		WriteErrorResponse(w, http.StatusInternalServerError, []string{
			http.StatusText(http.StatusInternalServerError),
		})
	}
}
