package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.ListSignatureDevices(w, r)
	case http.MethodPost:
		s.CreateSignatureDevice(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) CreateSignatureDevice(response http.ResponseWriter, request *http.Request) {
	b := request.Body
	var deviceRequest domain.SignatureDeviceRequest
	err := json.NewDecoder(b).Decode(&deviceRequest)
	if err != nil {
		WriteErrorResponse(response, http.StatusUnprocessableEntity, []string{
			http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}
	errs := validateSignatureDeviceRequest(&deviceRequest)
	if len(errs) > 0 {
		WriteErrorResponse(response, http.StatusBadRequest, errs)
		return
	}
	device, err := s.signatureDeviceService.CreateSignatureDevice(&deviceRequest)
	if err != nil {
		HandleError(response, err)
		return
	}
	WriteAPIResponse(response, http.StatusCreated, device)
}

func (s *Server) GetSignatureDevice(response http.ResponseWriter, request *http.Request) {
	id := strings.TrimPrefix(request.URL.Path, "/api/v0/signature-devices/")

	if id == "" || !validateUUID(id) {
		HandleError(response, utils.ErrInvalidDeviceId)
		return
	}
	device, err := s.signatureDeviceService.GetSignatureDevice(id)
	if err != nil {
		HandleError(response, err)
		return
	}

	WriteAPIResponse(response, http.StatusOK, device)
}

func (s *Server) ListSignatureDevices(response http.ResponseWriter, request *http.Request) {
	devices, err := s.signatureDeviceService.ListSignatureDevices()
	if err != nil {
		HandleError(response, err)
		return
	}

	WriteAPIResponse(response, http.StatusOK, devices)
}

func (s *Server) SignTransaction(response http.ResponseWriter, request *http.Request) {
	var signatureRequest domain.SignTransactionRequest
	err := json.NewDecoder(request.Body).Decode(&signatureRequest)
	if err != nil {
		WriteErrorResponse(response, http.StatusUnprocessableEntity, []string{
			http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}
	errs := validateTransactionSignatureRequest(&signatureRequest)
	if len(errs) > 0 {
		WriteErrorResponse(response, http.StatusBadRequest, errs)
		return
	}
	signatureData, err := s.signatureDeviceService.SignTransaction(signatureRequest.ID, signatureRequest.Data)
	if err != nil {
		HandleError(response, err)
		return
	}

	WriteAPIResponse(response, http.StatusOK, signatureData)
}
