package services

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uwemakan/signing-service/crypto"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/persistence"
	"github.com/uwemakan/signing-service/utils"
)

func TestGetSignatureDevice(t *testing.T) {
	requires := require.New(t)

	service := NewSignatureService(SignatureServiceParams{
		Repo:           persistence.NewInMemorySignatureDeviceRepository(),
		KeyPairFactory: crypto.NewKeyPairFactory(),
		SignerFactory:  crypto.NewSignerFactory(),
	})

	deviceId := utils.RandomString(16)
	device, err := service.GetSignatureDevice(deviceId)
	requires.Error(err)
	requires.ErrorIs(err, utils.ErrDeviceNotFound)
	requires.Nil(device)

	algorithm := utils.Algorithms[0]
	label := utils.RandomString(8)
	device, err = service.CreateSignatureDevice(&domain.SignatureDeviceRequest{
		ID:        deviceId,
		Algorithm: algorithm,
		Label:     &label,
	})
	requires.NoError(err)
	requires.NotNil(device)
	d, err := service.GetSignatureDevice(deviceId)
	requires.NoError(err)
	requires.Equal(device, d)
}

func TestListSignatureDevices(t *testing.T) {
	requires := require.New(t)

	service := NewSignatureService(SignatureServiceParams{
		Repo:           persistence.NewInMemorySignatureDeviceRepository(),
		KeyPairFactory: crypto.NewKeyPairFactory(),
		SignerFactory:  crypto.NewSignerFactory(),
	})

	devices, err := service.ListSignatureDevices()
	requires.NoError(err)
	requires.Len(devices, 0)

	deviceId := utils.RandomString(16)
	algorithm := utils.Algorithms[0]
	label := utils.RandomString(8)
	device, err := service.CreateSignatureDevice(&domain.SignatureDeviceRequest{
		ID:        deviceId,
		Algorithm: algorithm,
		Label:     &label,
	})
	requires.NoError(err)
	requires.NotNil(device)
	devices, err = service.ListSignatureDevices()
	requires.NoError(err)
	requires.Len(devices, 1)
	requires.Equal(device, devices[0])
}

func TestCreateSignatureDevice(t *testing.T) {
	requires := require.New(t)

	deviceId := utils.RandomString(16)
	algorithm := utils.Algorithms[0]
	label := utils.RandomString(8)
	testCases := []struct {
		name          string
		request       *domain.SignatureDeviceRequest
		setup         func(SignatureService)
		checkResponse func(*domain.SignatureDevice, error)
	}{
		{
			name: "CreateSignatureDevice_OK",
			request: &domain.SignatureDeviceRequest{
				ID:        deviceId,
				Algorithm: algorithm,
				Label:     &label,
			},
			setup: func(ss SignatureService) {},
			checkResponse: func(sd *domain.SignatureDevice, err error) {
				requires.NoError(err)
				requires.NotNil(sd)
				requires.Equal(deviceId, sd.ID)
				requires.Equal(algorithm, sd.Algorithm)
				requires.Equal(label, sd.Label)
				requires.Equal(0, sd.SignatureCounter)
				requires.Equal(base64.StdEncoding.EncodeToString([]byte(deviceId)), sd.LastSignature)
			},
		},
		{
			name: "CreateSignatureDevice_Device_Exist",
			request: &domain.SignatureDeviceRequest{
				ID:        deviceId,
				Algorithm: algorithm,
				Label:     &label,
			},
			setup: func(ss SignatureService) {
				ss.CreateSignatureDevice(&domain.SignatureDeviceRequest{
					ID:        deviceId,
					Algorithm: algorithm,
					Label:     &label,
				})
			},
			checkResponse: func(sd *domain.SignatureDevice, err error) {
				requires.Error(err)
				requires.ErrorIs(err, utils.ErrDeviceAlreadyExists)
				requires.Nil(sd)
			},
		},
		{
			name: "CreateSignatureDevice_Unknown_Algorithm",
			request: &domain.SignatureDeviceRequest{
				ID:        deviceId,
				Algorithm: "UNKNOWN",
				Label:     &label,
			},
			setup: func(ss SignatureService) {},
			checkResponse: func(sd *domain.SignatureDevice, err error) {
				requires.Error(err)
				requires.ErrorIs(err, utils.ErrUnsupportedAlgorithm)
				requires.Nil(sd)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewSignatureService(SignatureServiceParams{
				Repo:           persistence.NewInMemorySignatureDeviceRepository(),
				KeyPairFactory: crypto.NewKeyPairFactory(),
				SignerFactory:  crypto.NewSignerFactory(),
			})
			tc.setup(service)
			d, err := service.CreateSignatureDevice(tc.request)
			tc.checkResponse(d, err)
		})
	}
}

func TestSignTransaction(t *testing.T) {
	requires := require.New(t)

	deviceId := utils.RandomString(16)
	algorithm := utils.Algorithms[0]
	label := utils.RandomString(8)
	data := fmt.Sprintf("0_TestData_%s", base64.StdEncoding.EncodeToString([]byte(deviceId)))
	testCases := []struct {
		name          string
		deviceId      string
		data          string
		setup         func(SignatureService)
		checkResponse func(*domain.SignTransactionResponse, error)
	}{
		{
			name:     "SignTransaction_OK",
			deviceId: deviceId,
			data:     data,
			setup: func(ss SignatureService) {
				ss.CreateSignatureDevice(&domain.SignatureDeviceRequest{
					ID:        deviceId,
					Algorithm: algorithm,
					Label:     &label,
				})
			},
			checkResponse: func(sr *domain.SignTransactionResponse, err error) {
				requires.NoError(err)
				requires.NotNil(sr)
				requires.Equal(data, sr.SignedData)
			},
		},
		{
			name:     "SignTransaction_Device_Not_Found",
			deviceId: deviceId,
			data:     data,
			setup:    func(ss SignatureService) {},
			checkResponse: func(sr *domain.SignTransactionResponse, err error) {
				requires.Error(err)
				requires.ErrorIs(err, utils.ErrDeviceNotFound)
				requires.Nil(sr)
			},
		},
		{
			name:     "SignTransaction_SignatureCount_Error",
			deviceId: deviceId,
			data:     fmt.Sprintf("1_TestData_%s", base64.StdEncoding.EncodeToString([]byte(deviceId))),
			setup: func(ss SignatureService) {
				ss.CreateSignatureDevice(&domain.SignatureDeviceRequest{
					ID:        deviceId,
					Algorithm: algorithm,
					Label:     &label,
				})
			},
			checkResponse: func(sr *domain.SignTransactionResponse, err error) {
				requires.Error(err)
				requires.ErrorIs(err, utils.ErrInvalidSignatureCounter)
				requires.Nil(sr)
			},
		},
		{
			name:     "SignTransaction_SignatureCount_Error",
			deviceId: deviceId,
			data:     fmt.Sprintf("0_TestData_%s", utils.RandomString(16)),
			setup: func(ss SignatureService) {
				ss.CreateSignatureDevice(&domain.SignatureDeviceRequest{
					ID:        deviceId,
					Algorithm: algorithm,
					Label:     &label,
				})
			},
			checkResponse: func(sr *domain.SignTransactionResponse, err error) {
				requires.Error(err)
				requires.ErrorIs(err, utils.ErrInvalidLastSignature)
				requires.Nil(sr)
			},
		},
		{
			name:     "SignTransaction_UNKNOWN_Algorithm",
			deviceId: deviceId,
			data:     fmt.Sprintf("0_TestData_%s", utils.RandomString(16)),
			setup: func(ss SignatureService) {
				ss.CreateSignatureDevice(&domain.SignatureDeviceRequest{
					ID:        deviceId,
					Algorithm: algorithm,
					Label:     &label,
				})
			},
			checkResponse: func(sr *domain.SignTransactionResponse, err error) {
				requires.Error(err)
				requires.ErrorIs(err, utils.ErrInvalidLastSignature)
				requires.Nil(sr)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewSignatureService(SignatureServiceParams{
				Repo:           persistence.NewInMemorySignatureDeviceRepository(),
				KeyPairFactory: crypto.NewKeyPairFactory(),
				SignerFactory:  crypto.NewSignerFactory(),
			})
			tc.setup(service)
			sr, err := service.SignTransaction(tc.deviceId, tc.data)
			tc.checkResponse(sr, err)
		})
	}
}
