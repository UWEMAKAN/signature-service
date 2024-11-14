package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)

func TestHandler(t *testing.T) {
	requires := require.New(t)
	server := NewServer(":0")
	recorder := httptest.NewRecorder()

	url := "/api/v0/signature-devices"

	request, err := http.NewRequest(http.MethodPut, url, nil)
	requires.NoError(err)

	server.Handler(recorder, request)
}

func TestCreateSignatureDevice(t *testing.T) {
	requires := require.New(t)
	id, _ := uuid.NewRandom()
	testCases := []struct {
		name          string
		request       any
		setup         func(*Server)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "CreateSignatureDevice_OK",
			request: &domain.SignatureDeviceRequest{
				ID:        id.String(),
				Algorithm: utils.Algorithms[0],
			},
			setup: func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusCreated, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response Response
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				b, err := json.Marshal(response.Data)
				requires.NoError(err)
				var device domain.SignatureDevice
				err = json.Unmarshal(b, &device)
				requires.NoError(err)
				requires.Equal(id.String(), device.ID)
				requires.Equal(utils.Algorithms[0], device.Algorithm)
				requires.Equal(base64.StdEncoding.EncodeToString([]byte(id.String())), device.LastSignature)
				requires.Equal(0, device.SignatureCounter)
			},
		},
		{
			name:    "CreateSignatureDevice_UNPROCESSABLE_ENTITY",
			request: id.String(),
			setup:   func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusUnprocessableEntity, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.NotEmpty(response.Errors)
				requires.Equal(response.Errors[0], http.StatusText(http.StatusUnprocessableEntity))
			},
		},
		{
			name:    "CreateSignatureDevice_BAD_REQUEST",
			request: &domain.SignatureDeviceRequest{},
			setup:   func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusBadRequest, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.Len(response.Errors, 2)
				requires.Contains(response.Errors, fmt.Sprintf("invalid device id: %s is not a valid UUID", ""))
				requires.Contains(response.Errors, fmt.Sprintf("algorithm must be one of %s", utils.Algorithms))
			},
		},
		{
			name: "CreateSignatureDevice_ALREADY_EXIST",
			request: &domain.SignatureDeviceRequest{
				ID:        id.String(),
				Algorithm: utils.Algorithms[0],
			},
			setup: func(s *Server) {
				b, err := json.Marshal(domain.SignatureDeviceRequest{
					ID:        id.String(),
					Algorithm: utils.Algorithms[0],
				})
				requires.NoError(err)

				url := "/api/v0/signature-devices"
				body := bytes.NewReader(b)

				request, err := http.NewRequest(http.MethodPost, url, body)
				requires.NoError(err)
				recorder := httptest.NewRecorder()
				s.Handler(recorder, request)
			},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusBadRequest, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.NotEmpty(response.Errors)
				requires.Equal(response.Errors[0], utils.ErrDeviceAlreadyExists.Error())
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(":0")
			tc.setup(server)
			recorder := httptest.NewRecorder()

			b, err := json.Marshal(tc.request)
			requires.NoError(err)

			url := "/api/v0/signature-devices"
			body := bytes.NewReader(b)

			request, err := http.NewRequest(http.MethodPost, url, body)
			requires.NoError(err)

			server.Handler(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListSignatureDevices(t *testing.T) {
	requires := require.New(t)
	id, _ := uuid.NewRandom()
	testCases := []struct {
		name          string
		setup         func(*Server)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name:  "ListSignatureDevices_EMPTY",
			setup: func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusOK, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response Response
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				b, err := json.Marshal(response.Data)
				requires.NoError(err)
				var devices []domain.SignatureDevice
				err = json.Unmarshal(b, &devices)
				requires.NoError(err)
				requires.Empty(devices)
			},
		},
		{
			name: "ListSignatureDevice_OK",
			setup: func(s *Server) {
				b, err := json.Marshal(domain.SignatureDeviceRequest{
					ID:        id.String(),
					Algorithm: utils.Algorithms[0],
				})
				requires.NoError(err)

				url := "/api/v0/signature-devices"
				body := bytes.NewReader(b)

				request, err := http.NewRequest(http.MethodPost, url, body)
				requires.NoError(err)
				recorder := httptest.NewRecorder()
				s.Handler(recorder, request)
			},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusOK, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response Response
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				b, err := json.Marshal(response.Data)
				requires.NoError(err)
				var devices []domain.SignatureDevice
				err = json.Unmarshal(b, &devices)
				requires.NoError(err)
				requires.Len(devices, 1)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(":0")
			tc.setup(server)
			recorder := httptest.NewRecorder()

			url := "/api/v0/signature-devices"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			requires.NoError(err)

			server.Handler(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetSignatureDevice(t *testing.T) {
	requires := require.New(t)
	id, _ := uuid.NewRandom()
	testCases := []struct {
		name          string
		id            string
		setup         func(*Server)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "GetSignatureDevice_OK",
			id:   id.String(),
			setup: func(s *Server) {
				b, err := json.Marshal(domain.SignatureDeviceRequest{
					ID:        id.String(),
					Algorithm: utils.Algorithms[0],
				})
				requires.NoError(err)

				url := "/api/v0/signature-devices"
				body := bytes.NewReader(b)

				request, err := http.NewRequest(http.MethodPost, url, body)
				requires.NoError(err)
				recorder := httptest.NewRecorder()
				s.Handler(recorder, request)
			},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusOK, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response Response
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				b, err := json.Marshal(response.Data)
				requires.NoError(err)
				var device domain.SignatureDevice
				err = json.Unmarshal(b, &device)
				requires.NoError(err)
				requires.Equal(id.String(), device.ID)
				requires.Equal(utils.Algorithms[0], device.Algorithm)
				requires.Equal(base64.StdEncoding.EncodeToString([]byte(id.String())), device.LastSignature)
				requires.Equal(0, device.SignatureCounter)
			},
		},
		{
			name:  "GetSignatureDevice_NOT_FOUND",
			id:    id.String(),
			setup: func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusNotFound, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.Len(response.Errors, 1)
				requires.Equal(response.Errors[0], utils.ErrDeviceNotFound.Error())
			},
		},
		{
			name:  "GetSignatureDevice_BAD_REQUEST",
			id:    utils.RandomString(12),
			setup: func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusBadRequest, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.Len(response.Errors, 1)
				requires.Equal(response.Errors[0], "device ID must be a valid UUID")
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(":0")
			tc.setup(server)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v0/signature-devices/%s", tc.id)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			requires.NoError(err)

			server.GetSignatureDevice(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestSignTransaction(t *testing.T) {
	requires := require.New(t)
	id, _ := uuid.NewRandom()
	id2, _ := uuid.NewRandom()
	data := fmt.Sprintf("0_TESTDATA_%s", base64.StdEncoding.EncodeToString([]byte(id.String())))
	testCases := []struct {
		name          string
		request       any
		setup         func(*Server)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "SignTransaction_OK",
			request: &domain.SignTransactionRequest{
				ID:   id.String(),
				Data: data,
			},
			setup: func(s *Server) {
				b, err := json.Marshal(domain.SignatureDeviceRequest{
					ID:        id.String(),
					Algorithm: utils.Algorithms[0],
				})
				requires.NoError(err)

				url := "/api/v0/signature-devices"
				body := bytes.NewReader(b)

				request, err := http.NewRequest(http.MethodPost, url, body)
				requires.NoError(err)
				recorder := httptest.NewRecorder()
				s.Handler(recorder, request)
			},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusOK, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response Response
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				b, err := json.Marshal(response.Data)
				requires.NoError(err)
				var signatureData domain.SignTransactionResponse
				err = json.Unmarshal(b, &signatureData)
				requires.NoError(err)
				requires.NotZero(signatureData.Signature)
				requires.Equal(data, signatureData.SignedData)
			},
		},
		{
			name:    "SignTransaction_UNPROCESSABLE_ENTITY",
			request: id.String(),
			setup:   func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusUnprocessableEntity, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.NotEmpty(response.Errors)
				requires.Equal(response.Errors[0], http.StatusText(http.StatusUnprocessableEntity))
			},
		},
		{
			name:    "SignTransaction_BAD_REQUEST",
			request: &domain.SignTransactionRequest{},
			setup:   func(s *Server) {},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusBadRequest, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.Len(response.Errors, 2)
				requires.Contains(response.Errors, fmt.Sprintf("invalid device id: %s is not a valid UUID", ""))
				requires.Contains(response.Errors, "invalid data: data must be in the format signatureCounter_data_lastSignature")
			},
		},
		{
			name: "SignTransaction_NOT_FOUND",
			request: &domain.SignTransactionRequest{
				ID:   id2.String(),
				Data: data,
			},
			setup: func(s *Server) {
				b, err := json.Marshal(domain.SignatureDeviceRequest{
					ID:        id.String(),
					Algorithm: utils.Algorithms[0],
				})
				requires.NoError(err)

				url := "/api/v0/signature-devices"
				body := bytes.NewReader(b)

				request, err := http.NewRequest(http.MethodPost, url, body)
				requires.NoError(err)
				recorder := httptest.NewRecorder()
				s.Handler(recorder, request)
			},
			checkResponse: func(rr *httptest.ResponseRecorder) {
				requires.Equal(http.StatusNotFound, rr.Code)

				body, err := io.ReadAll(rr.Body)
				requires.NoError(err)
				requires.NotNil(body)
				var response ErrorResponse
				err = json.Unmarshal(body, &response)
				requires.NoError(err)
				requires.NotEmpty(response.Errors)
				requires.Equal(response.Errors[0], utils.ErrDeviceNotFound.Error())
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(":0")
			tc.setup(server)
			recorder := httptest.NewRecorder()

			b, err := json.Marshal(tc.request)
			requires.NoError(err)

			url := "/api/v0/signature-devices/sign"
			body := bytes.NewReader(b)

			request, err := http.NewRequest(http.MethodPost, url, body)
			requires.NoError(err)

			server.SignTransaction(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
