package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)

func TestValidateDevice(t *testing.T) {
	requires := require.New(t)
	id, _ := uuid.NewRandom()
	requires.True(validateUUID(id.String()))
	requires.False(validateUUID(utils.RandomString(16)))
}

func TestValidateData(t *testing.T) {
	requires := require.New(t)
	data := "0_Data_Signature"
	requires.True(validateData(data))
	requires.False(validateData(""))
	requires.False(validateData("0_Data"))
	requires.False(validateData(" _ _ "))
}

func TestValidateAlgorithm(t *testing.T) {
	requires := require.New(t)
	requires.True(validateAlgorithm(utils.Algorithms[0]))
	requires.True(validateAlgorithm(utils.Algorithms[1]))
	requires.False(validateAlgorithm("UNKNOWN"))
}

func TestValidateLabel(t *testing.T) {
	requires := require.New(t)
	requires.True(validateLabel(utils.RandomString(6)))
	requires.False(validateLabel(""))
}

func TestValidateSignatureDeviceRequest(t *testing.T) {
	requires := require.New(t)
	deviceId, _ := uuid.NewRandom()
	label := utils.RandomString(6)
	empty := ""
	request := &domain.SignatureDeviceRequest{
		ID:        deviceId.String(),
		Algorithm: utils.Algorithms[0],
		Label:     &label,
	}

	testCases := []struct {
		name          string
		request       *domain.SignatureDeviceRequest
		checkResponse func([]string)
	}{
		{
			name:    "validateSignatureDeviceRequest_OK",
			request: request,
			checkResponse: func(s []string) {
				requires.Empty(s)
			},
		},
		{
			name: "validateSignatureDeviceRequest_Failed",
			request: &domain.SignatureDeviceRequest{
				Label: &empty,
			},
			checkResponse: func(s []string) {
				requires.NotEmpty(s)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.checkResponse(validateSignatureDeviceRequest(tc.request))
		})
	}
}

func TestValidateTransactionSignatureRequest(t *testing.T) {
	requires := require.New(t)
	deviceId, _ := uuid.NewRandom()
	data := "0_DATA_Signature"
	testCases := []struct {
		name          string
		request       *domain.SignTransactionRequest
		checkResponse func([]string)
	}{
		{
			name: "validateTransactionSignatureRequest_OK",
			request: &domain.SignTransactionRequest{
				ID:   deviceId.String(),
				Data: data,
			},
			checkResponse: func(s []string) {
				requires.Empty(s)
			},
		},
		{
			name:    "validateTransactionSignatureRequest_OK",
			request: &domain.SignTransactionRequest{},
			checkResponse: func(s []string) {
				requires.NotEmpty(s)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.checkResponse(validateTransactionSignatureRequest(tc.request))
		})
	}
}
