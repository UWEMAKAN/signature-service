package persistence

import (
	"encoding/base64"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)

func createDevice(t *testing.T) (*domain.SignatureDevice, *InMemorySignatureDeviceRepository) {
	requires := require.New(t)
	repo := NewInMemorySignatureDeviceRepository()
	deviceId := utils.RandomString(16)
	label := utils.RandomString(6)
	publicKey := utils.RandomString(16)
	privateKey := utils.RandomString(16)
	algorithm := utils.Algorithms[0]

	device, err := repo.CreateDevice(deviceId, algorithm, publicKey, privateKey, label)
	requires.NoError(err)
	requires.NotNil(device)
	return device, repo
}

func TestCreateDevice(t *testing.T) {
	requires := require.New(t)
	device, repo := createDevice(t)
	newDevice, err := repo.CreateDevice(device.ID, device.Algorithm, device.PublicKey, device.PrivateKey, device.Label)
	requires.Error(err)
	requires.ErrorIs(err, utils.ErrDeviceAlreadyExists)
	requires.Nil(newDevice)
}

func TestGetDevice(t *testing.T) {
	requires := require.New(t)
	device, repo := createDevice(t)

	d1, err := repo.GetDevice(device.ID)
	requires.NoError(err)
	requires.NotNil(d1)

	d2, err := repo.GetDevice(utils.RandomString(6))
	requires.Error(err)
	requires.Nil(d2)
}

func TestListDevices(t *testing.T) {
	requires := require.New(t)
	_, repo := createDevice(t)

	devices, err := repo.ListDevices()
	requires.NoError(err)
	requires.Len(devices, 1)
}

func TestSignAndIncrementCounter(t *testing.T) {
	requires := require.New(t)
	device, repo := createDevice(t)
	newSignature := utils.RandomString(24)

	requires.Equal(0, device.SignatureCounter)
	requires.Equal(base64.StdEncoding.EncodeToString([]byte(device.ID)), device.LastSignature)

	err := repo.SignAndIncrementCounter(device.ID, newSignature)
	requires.NoError(err)
	device, err = repo.GetDevice(device.ID)
	requires.NoError(err)
	requires.NotNil(device)
	requires.Equal(1, device.SignatureCounter)
	requires.Equal(newSignature, device.LastSignature)

	err = repo.SignAndIncrementCounter(utils.RandomString(16), newSignature)
	requires.Error(err)
	requires.ErrorIs(err, utils.ErrDeviceNotFound)
}

func TestLoadSignAndIncrementCounter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestLoadSignAndIncrementCounter in short mode.")
	}
	repo := NewInMemorySignatureDeviceRepository()
	var wg sync.WaitGroup
	nd := 10
	ms := 1000
	for range nd {
		wg.Add(1)
		go func(t *testing.T, m int) {
			defer wg.Done()
			requires := require.New(t)
			deviceId := utils.RandomString(16)
			label := utils.RandomString(6)
			publicKey := utils.RandomString(16)
			privateKey := utils.RandomString(16)
			algorithm := utils.Algorithms[0]

			device, err := repo.CreateDevice(deviceId, algorithm, publicKey, privateKey, label)
			requires.NoError(err)
			requires.NotNil(device)
			for range m {
				wg.Add(1)
				go func(tt *testing.T) {
					defer wg.Done()
					requires := require.New(t)
					err := repo.SignAndIncrementCounter(device.ID, utils.RandomString(10))
					requires.NoError(err)
				}(t)
			}
		}(t, ms)
	}
	wg.Wait()
	requires := require.New(t)
	devices, err := repo.ListDevices()
	requires.NoError(err)
	requires.NotNil(devices)
	requires.Len(devices, nd)
	for _, device := range devices {
		requires.Equal(ms, device.SignatureCounter)
		requires.NotEqual(base64.StdEncoding.EncodeToString([]byte(device.ID)), device.ID)
	}
}
