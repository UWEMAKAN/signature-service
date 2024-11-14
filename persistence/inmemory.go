package persistence

import (
	"encoding/base64"
	"sync"

	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)

type InMemorySignatureDeviceRepository struct {
    devices map[string]*domain.SignatureDevice
    mu      sync.RWMutex
}

func NewInMemorySignatureDeviceRepository() *InMemorySignatureDeviceRepository {
    return &InMemorySignatureDeviceRepository{
        devices: make(map[string]*domain.SignatureDevice),
    }
}

func (repo *InMemorySignatureDeviceRepository) CreateDevice(id, algorithm, publicKey, privateKey, label string) (*domain.SignatureDevice, error) {
    repo.mu.Lock()
    defer repo.mu.Unlock()

    if _, exists := repo.devices[id]; exists {
        return nil, utils.ErrDeviceAlreadyExists
    }

    device := &domain.SignatureDevice{
        ID:              id,
        Algorithm:       algorithm,
        PublicKey:       publicKey,
        PrivateKey:      privateKey,
        Label:           label,
        SignatureCounter: 0,
        LastSignature:   base64.StdEncoding.EncodeToString([]byte(id)),
    }
    repo.devices[id] = device
    return device, nil
}

func (repo *InMemorySignatureDeviceRepository) GetDevice(id string) (*domain.SignatureDevice, error) {
    repo.mu.RLock()
    defer repo.mu.RUnlock()

    device, exists := repo.devices[id]
    if !exists {
        return nil, utils.ErrDeviceNotFound
    }
    return device, nil
}

func (repo *InMemorySignatureDeviceRepository) ListDevices() ([]*domain.SignatureDevice, error) {
    repo.mu.RLock()
    defer repo.mu.RUnlock()

    devices := make([]*domain.SignatureDevice, 0, len(repo.devices))
    for _, device := range repo.devices {
        devices = append(devices, device)
    }
    return devices, nil
}

func (repo *InMemorySignatureDeviceRepository) SignAndIncrementCounter(deviceId, newSignature string) error {
    repo.mu.Lock()
    defer repo.mu.Unlock()

    device, exists := repo.devices[deviceId]
    if !exists {
        return utils.ErrDeviceNotFound
    }

    device.SignatureCounter++
    device.LastSignature = newSignature
    return nil
}
